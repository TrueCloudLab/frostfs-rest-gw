package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/errors"
	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-rest-gw/gen/models"
	"github.com/nspcc-dev/neofs-rest-gw/gen/restapi/operations"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"go.uber.org/zap"
)

// API is a REST v1 request handler.
type API struct {
	log              *zap.Logger
	pool             *pool.Pool
	key              *keys.PrivateKey
	defaultTimestamp bool
}

// PrmAPI groups parameters to init rest API.
type PrmAPI struct {
	Logger           *zap.Logger
	Pool             *pool.Pool
	Key              *keys.PrivateKey
	DefaultTimestamp bool
}

type BearerToken struct {
	Token     string
	Signature string
	Key       string
}

// ContextKey is used for context.Context value. The value requires a key that is not primitive type.
type ContextKey string

const (
	// BearerPrefix is the prefix for authorization token.
	BearerPrefix = "Bearer "

	// ContextKeyRequestID is the ContextKey for RequestID.
	ContextKeyRequestID ContextKey = "requestID"
)

// New creates a new API using specified logger, connection pool and other parameters.
func New(prm *PrmAPI) *API {
	return &API{
		log:              prm.Logger,
		pool:             prm.Pool,
		key:              prm.Key,
		defaultTimestamp: prm.DefaultTimestamp,
	}
}

func (a *API) Configure(api *operations.NeofsRestGwAPI) http.Handler {
	api.ServeError = errors.ServeError

	api.AuthHandler = operations.AuthHandlerFunc(a.PostAuth)

	api.PutObjectHandler = operations.PutObjectHandlerFunc(a.PutObjects)
	api.GetObjectInfoHandler = operations.GetObjectInfoHandlerFunc(a.GetObjectInfo)

	api.PutContainerHandler = operations.PutContainerHandlerFunc(a.PutContainers)
	api.GetContainerHandler = operations.GetContainerHandlerFunc(a.GetContainer)
	api.DeleteContainerHandler = operations.DeleteContainerHandlerFunc(a.DeleteContainer)
	api.PutContainerEACLHandler = operations.PutContainerEACLHandlerFunc(a.PutContainerEACL)
	api.GetContainerEACLHandler = operations.GetContainerEACLHandlerFunc(a.GetContainerEACL)

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

	api.ServerShutdown = func() {}

	return a.setupGlobalMiddleware(api.Serve(setupMiddlewares))
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

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
