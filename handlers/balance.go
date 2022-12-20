package handlers

import (
	"strconv"

	"github.com/TrueCloudLab/frostfs-rest-gw/gen/models"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi/operations"
	"github.com/TrueCloudLab/frostfs-rest-gw/internal/util"
	"github.com/TrueCloudLab/frostfs-sdk-go/pool"
	"github.com/TrueCloudLab/frostfs-sdk-go/user"
	"github.com/go-openapi/runtime/middleware"
)

// Balance handler that get balance from FrostFS.
func (a *API) Balance(params operations.GetBalanceParams) middleware.Responder {
	var ownerID user.ID
	if err := ownerID.DecodeString(params.Address); err != nil {
		resp := a.logAndGetErrorResponse("parse address", err)
		return operations.NewGetBalanceBadRequest().WithPayload(resp)
	}

	var prm pool.PrmBalanceGet
	prm.SetAccount(ownerID)

	frostfsBalance, err := a.pool.Balance(params.HTTPRequest.Context(), prm)
	if err != nil {
		resp := a.logAndGetErrorResponse("get balance", err)
		return operations.NewGetBalanceBadRequest().WithPayload(resp)
	}

	var resp models.Balance
	resp.Address = util.NewString(params.Address)
	resp.Value = util.NewString(strconv.FormatInt(frostfsBalance.Value(), 10))
	resp.Precision = util.NewInteger(int64(frostfsBalance.Precision()))

	return operations.NewGetBalanceOK().
		WithPayload(&resp).
		WithAccessControlAllowOrigin("*")
}
