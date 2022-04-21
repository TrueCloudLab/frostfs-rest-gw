// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/nspcc-dev/neofs-rest-gw/gen/models"
)

// GetObjectInfoHandlerFunc turns a function with the right signature into a get object info handler
type GetObjectInfoHandlerFunc func(GetObjectInfoParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn GetObjectInfoHandlerFunc) Handle(params GetObjectInfoParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// GetObjectInfoHandler interface for that can handle valid get object info params
type GetObjectInfoHandler interface {
	Handle(GetObjectInfoParams, *models.Principal) middleware.Responder
}

// NewGetObjectInfo creates a new http.Handler for the get object info operation
func NewGetObjectInfo(ctx *middleware.Context, handler GetObjectInfoHandler) *GetObjectInfo {
	return &GetObjectInfo{Context: ctx, Handler: handler}
}

/* GetObjectInfo swagger:route GET /objects/{containerId}/{objectId} getObjectInfo

Get object info by address

*/
type GetObjectInfo struct {
	Context *middleware.Context
	Handler GetObjectInfoHandler
}

func (o *GetObjectInfo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetObjectInfoParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		*r = *aCtx
	}
	var principal *models.Principal
	if uprinc != nil {
		principal = uprinc.(*models.Principal) // this is really a models.Principal, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}