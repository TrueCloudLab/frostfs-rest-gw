package handlers

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/TrueCloudLab/frostfs-api-go/v2/acl"
	"github.com/TrueCloudLab/frostfs-api-go/v2/refs"
	sessionv2 "github.com/TrueCloudLab/frostfs-api-go/v2/session"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/models"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi/operations"
	"github.com/TrueCloudLab/frostfs-rest-gw/internal/util"
	neofsecdsa "github.com/TrueCloudLab/frostfs-sdk-go/crypto/ecdsa"
	"github.com/TrueCloudLab/frostfs-sdk-go/pool"
	"github.com/TrueCloudLab/frostfs-sdk-go/user"
	"github.com/go-openapi/runtime/middleware"
	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

const defaultTokenExpDuration = 100 // in epoch

type headersParams struct {
	XBearerLifetime    uint64
	XBearerOwnerID     string
	XBearerForAllUsers bool
}

type objectTokenParams struct {
	headersParams
	Records []*models.Record
	Name    string
}

type containerTokenParams struct {
	headersParams
	Rule *models.Rule
	Name string
}

func newHeaderParams(params operations.AuthParams) headersParams {
	prm := headersParams{
		XBearerOwnerID:     params.XBearerOwnerID,
		XBearerForAllUsers: *params.XBearerForAllUsers,
	}

	if params.XBearerLifetime != nil && *params.XBearerLifetime > 0 {
		prm.XBearerLifetime = uint64(*params.XBearerLifetime)
	}

	return prm
}

func newObjectParams(common headersParams, token *models.Bearer) objectTokenParams {
	return objectTokenParams{
		headersParams: common,
		Records:       token.Object,
		Name:          token.Name,
	}
}

func newContainerParams(common headersParams, token *models.Bearer) containerTokenParams {
	return containerTokenParams{
		headersParams: common,
		Rule:          token.Container,
		Name:          token.Name,
	}
}

// PostAuth handler that forms bearer token to sign.
func (a *API) PostAuth(params operations.AuthParams) middleware.Responder {
	ctx := params.HTTPRequest.Context()
	commonPrm := newHeaderParams(params)

	tokenNames := make(map[string]struct{})
	response := make([]*models.TokenResponse, len(params.Tokens))
	for i, token := range params.Tokens {
		if _, ok := tokenNames[token.Name]; ok {
			err := fmt.Errorf("duplicated token name '%s'", token.Name)
			return operations.NewAuthBadRequest().WithPayload(util.NewErrorResponse(err))
		}
		tokenNames[token.Name] = struct{}{}

		isObject, err := IsObjectToken(token)
		if err != nil {
			return operations.NewAuthBadRequest().WithPayload(util.NewErrorResponse(err))
		}

		if isObject {
			prm := newObjectParams(commonPrm, token)
			response[i], err = prepareObjectToken(ctx, prm, a.pool, *a.owner)
		} else {
			prm := newContainerParams(commonPrm, token)
			response[i], err = prepareContainerTokens(ctx, prm, a.pool, a.key.PublicKey())
		}
		if err != nil {
			return operations.NewAuthBadRequest().WithPayload(util.NewErrorResponse(err))
		}
	}

	return operations.NewAuthOK().
		WithPayload(response).
		WithAccessControlAllowOrigin("*")
}

// FormBinaryBearer handler that forms binary bearer token using headers with body and signature.
func (a *API) FormBinaryBearer(params operations.FormBinaryBearerParams, principal *models.Principal) middleware.Responder {
	btoken, err := getBearerToken(principal, params.XBearerSignature, params.XBearerSignatureKey, *params.WalletConnect, false)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid bearer token", err)
		return operations.NewFormBinaryBearerBadRequest().WithPayload(resp)
	}

	resp := &models.BinaryBearer{
		Token: util.NewString(base64.StdEncoding.EncodeToString(btoken.Marshal())),
	}

	return operations.NewFormBinaryBearerOK().WithPayload(resp)
}

func prepareObjectToken(ctx context.Context, params objectTokenParams, pool *pool.Pool, owner user.ID) (*models.TokenResponse, error) {
	btoken, err := util.ToNativeObjectToken(params.Records)
	if err != nil {
		return nil, fmt.Errorf("couldn't transform token to native: %w", err)
	}

	if !params.XBearerForAllUsers {
		btoken.ForUser(owner)
	}

	iat, exp, err := getTokenLifetime(ctx, pool, params.XBearerLifetime)
	if err != nil {
		return nil, fmt.Errorf("couldn't get lifetime: %w", err)
	}
	btoken.SetIat(iat)
	btoken.SetExp(exp)

	var v2token acl.BearerToken
	btoken.WriteToV2(&v2token)
	binaryBearer := v2token.GetBody().StableMarshal(nil)

	return &models.TokenResponse{
		Name:  params.Name,
		Type:  models.NewTokenType(models.TokenTypeObject),
		Token: util.NewString(base64.StdEncoding.EncodeToString(binaryBearer)),
	}, nil
}

func prepareContainerTokens(ctx context.Context, params containerTokenParams, pool *pool.Pool, key *keys.PublicKey) (*models.TokenResponse, error) {
	iat, exp, err := getTokenLifetime(ctx, pool, params.XBearerLifetime)
	if err != nil {
		return nil, fmt.Errorf("couldn't get lifetime: %w", err)
	}

	var ownerID user.ID
	if err = ownerID.DecodeString(params.XBearerOwnerID); err != nil {
		return nil, fmt.Errorf("invalid bearer owner: %w", err)
	}

	stoken, err := util.ToNativeContainerToken(params.Rule)
	if err != nil {
		return nil, fmt.Errorf("couldn't transform rule to native session token: %w", err)
	}

	stoken.SetID(uuid.New())
	stoken.SetIat(iat)
	stoken.SetExp(exp)

	authKey := neofsecdsa.PublicKey(*key)
	stoken.SetAuthKey(&authKey)

	var v2token sessionv2.Token
	stoken.WriteToV2(&v2token)

	var issuer refs.OwnerID
	ownerID.WriteToV2(&issuer)
	v2token.GetBody().SetOwnerID(&issuer)

	binaryToken := v2token.GetBody().StableMarshal(nil)

	return &models.TokenResponse{
		Name:  params.Name,
		Type:  models.NewTokenType(models.TokenTypeContainer),
		Token: util.NewString(base64.StdEncoding.EncodeToString(binaryToken)),
	}, nil
}

func getCurrentEpoch(ctx context.Context, p *pool.Pool) (uint64, error) {
	netInfo, err := p.NetworkInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("couldn't get netwokr info: %w", err)
	}

	return netInfo.CurrentEpoch(), nil
}

func getTokenLifetime(ctx context.Context, p *pool.Pool, expDuration uint64) (uint64, uint64, error) {
	currEpoch, err := getCurrentEpoch(ctx, p)
	if err != nil {
		return 0, 0, err
	}

	var lifetimeDuration uint64 = defaultTokenExpDuration
	if expDuration != 0 {
		lifetimeDuration = expDuration
	}

	return currEpoch, currEpoch + lifetimeDuration, nil
}
