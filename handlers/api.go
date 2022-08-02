package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	sessionv2 "github.com/TrueCloudLab/frostfs-api-go/v2/session"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/models"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi/operations"
	"github.com/TrueCloudLab/frostfs-rest-gw/internal/util"
	"github.com/TrueCloudLab/frostfs-rest-gw/metrics"
	"github.com/TrueCloudLab/frostfs-sdk-go/pool"
	"github.com/TrueCloudLab/frostfs-sdk-go/user"
	"github.com/go-openapi/errors"
	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"go.uber.org/zap"
)

// API is a REST v1 request handler.
type API struct {
	log              *zap.Logger
	pool             *pool.Pool
	key              *keys.PrivateKey
	owner            *user.ID
	defaultTimestamp bool

	gateMetric             *metrics.GateMetrics
	prometheusService      *metrics.Service
	pprofService           *metrics.Service
	serviceShutdownTimeout time.Duration
}

// PrmAPI groups parameters to init rest API.
type PrmAPI struct {
	Logger           *zap.Logger
	Pool             *pool.Pool
	Key              *keys.PrivateKey
	DefaultTimestamp bool

	GateMetric             *metrics.GateMetrics
	PrometheusService      *metrics.Service
	PprofService           *metrics.Service
	ServiceShutdownTimeout time.Duration
}

type BearerToken struct {
	Token     string
	Signature string
	Key       string
}

type SessionToken struct {
	BearerToken
	Verb sessionv2.ContainerSessionVerb
}

// ContextKey is used for context.Context value. The value requires a key that is not primitive type.
type ContextKey string

const (
	// BearerPrefix is the prefix for authorization token.
	BearerPrefix = "Bearer "

	// ContextKeyRequestID is the ContextKey for RequestID.
	ContextKeyRequestID ContextKey = "requestID"

	docsPrefix = "/docs"

	accessControlAllowOriginHeader = "Access-Control-Allow-Origin"
)

// New creates a new API using specified logger, connection pool and other parameters.
func New(prm *PrmAPI) *API {
	var owner user.ID
	user.IDFromKey(&owner, prm.Key.PrivateKey.PublicKey)

	return &API{
		log:              prm.Logger,
		pool:             prm.Pool,
		key:              prm.Key,
		owner:            &owner,
		defaultTimestamp: prm.DefaultTimestamp,

		prometheusService:      prm.PrometheusService,
		pprofService:           prm.PprofService,
		gateMetric:             prm.GateMetric,
		serviceShutdownTimeout: prm.ServiceShutdownTimeout,
	}
}

func (a *API) Configure(api *operations.FrostfsRestGwAPI) http.Handler {
	api.ServeError = errors.ServeError

	api.UseSwaggerUI()

	api.OptionsAuthHandler = operations.OptionsAuthHandlerFunc(a.OptionsAuth)
	api.AuthHandler = operations.AuthHandlerFunc(a.PostAuth)

	api.OptionsAuthBearerHandler = operations.OptionsAuthBearerHandlerFunc(a.OptionsAuthBearer)
	api.FormBinaryBearerHandler = operations.FormBinaryBearerHandlerFunc(a.FormBinaryBearer)

	api.GetBalanceHandler = operations.GetBalanceHandlerFunc(a.Balance)

	api.OptionsObjectsPutHandler = operations.OptionsObjectsPutHandlerFunc(a.OptionsObjectsPut)
	api.PutObjectHandler = operations.PutObjectHandlerFunc(a.PutObjects)

	api.OptionsObjectsGetDeleteHandler = operations.OptionsObjectsGetDeleteHandlerFunc(a.OptionsObjectsGetDelete)
	api.GetObjectInfoHandler = operations.GetObjectInfoHandlerFunc(a.GetObjectInfo)
	api.DeleteObjectHandler = operations.DeleteObjectHandlerFunc(a.DeleteObject)

	api.OptionsObjectsSearchHandler = operations.OptionsObjectsSearchHandlerFunc(a.OptionsObjectSearch)
	api.SearchObjectsHandler = operations.SearchObjectsHandlerFunc(a.SearchObjects)

	api.OptionsContainersPutListHandler = operations.OptionsContainersPutListHandlerFunc(a.OptionsContainersPutList)
	api.PutContainerHandler = operations.PutContainerHandlerFunc(a.PutContainers)
	api.ListContainersHandler = operations.ListContainersHandlerFunc(a.ListContainer)

	api.OptionsContainersGetDeleteHandler = operations.OptionsContainersGetDeleteHandlerFunc(a.OptionsContainersGetDelete)
	api.GetContainerHandler = operations.GetContainerHandlerFunc(a.GetContainer)
	api.DeleteContainerHandler = operations.DeleteContainerHandlerFunc(a.DeleteContainer)

	api.OptionsContainersEACLHandler = operations.OptionsContainersEACLHandlerFunc(a.OptionsContainersEACL)
	api.PutContainerEACLHandler = operations.PutContainerEACLHandlerFunc(a.PutContainerEACL)
	api.GetContainerEACLHandler = operations.GetContainerEACLHandlerFunc(a.GetContainerEACL)
	api.ListContainersHandler = operations.ListContainersHandlerFunc(a.ListContainer)

	api.PutStorageGroupHandler = operations.PutStorageGroupHandlerFunc(a.PutStorageGroup)
	api.ListStorageGroupsHandler = operations.ListStorageGroupsHandlerFunc(a.ListStorageGroups)

	api.BearerAuthAuth = func(s string) (*models.Principal, error) {
		if !strings.HasPrefix(s, BearerPrefix) {
			return nil, fmt.Errorf("has not bearer token")
		}
		if s = strings.TrimPrefix(s, BearerPrefix); len(s) == 0 {
			return nil, fmt.Errorf("bearer token is empty")
		}

		return (*models.Principal)(&s), nil
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {
		shutDownCtx, cancel := context.WithTimeout(context.Background(), a.serviceShutdownTimeout)
		defer cancel()

		a.prometheusService.ShutDown(shutDownCtx)
		a.pprofService.ShutDown(shutDownCtx)
	}

	return a.setupGlobalMiddleware(a.docMiddleware(api.Serve(setupMiddlewares)))
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func (a *API) setupGlobalMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		a.log.Info("request", zap.String("remote", r.RemoteAddr),
			zap.String("method", r.Method), zap.String("url", r.URL.String()),
			zap.String("id", requestID))

		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)

		w.Header().Set(accessControlAllowOriginHeader, allOrigins)

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) docMiddleware(handler http.Handler) http.Handler {
	fh := http.StripPrefix(docsPrefix, http.FileServer(http.Dir("static/docs")))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, docsPrefix) {
			fh.ServeHTTP(w, r)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}

func (a *API) logAndGetErrorResponse(msg string, err error, fields ...zap.Field) *models.ErrorResponse {
	fields = append(fields, zap.Error(err))
	a.log.Error(msg, fields...)
	return util.NewErrorResponse(fmt.Errorf("%s: %w", msg, err))
}

func (a API) StartCallback() {
	if a.gateMetric == nil {
		return
	}

	a.gateMetric.SetHealth(1)
}

func (a API) RunServices() {
	go a.pprofService.Start()
	go a.prometheusService.Start()
}
