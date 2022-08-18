// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/nspcc-dev/neofs-rest-gw/gen/models"
)

// FormBinaryBearerHandlerFunc turns a function with the right signature into a form binary bearer handler
type FormBinaryBearerHandlerFunc func(FormBinaryBearerParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn FormBinaryBearerHandlerFunc) Handle(params FormBinaryBearerParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// FormBinaryBearerHandler interface for that can handle valid form binary bearer params
type FormBinaryBearerHandler interface {
	Handle(FormBinaryBearerParams, *models.Principal) middleware.Responder
}

// NewFormBinaryBearer creates a new http.Handler for the form binary bearer operation
func NewFormBinaryBearer(ctx *middleware.Context, handler FormBinaryBearerHandler) *FormBinaryBearer {
	return &FormBinaryBearer{Context: ctx, Handler: handler}
}

/* FormBinaryBearer swagger:route GET /auth/bearer formBinaryBearer

Form binary bearer token

*/
type FormBinaryBearer struct {
	Context *middleware.Context
	Handler FormBinaryBearerHandler
}

func (o *FormBinaryBearer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewFormBinaryBearerParams()
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