// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
)

// NewOptionsContainersGetDeleteParams creates a new OptionsContainersGetDeleteParams object
//
// There are no default values defined in the spec.
func NewOptionsContainersGetDeleteParams() OptionsContainersGetDeleteParams {

	return OptionsContainersGetDeleteParams{}
}

// OptionsContainersGetDeleteParams contains all the bound params for the options containers get delete operation
// typically these are obtained from a http.Request
//
// swagger:parameters optionsContainersGetDelete
type OptionsContainersGetDeleteParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Base58 encoded container id.
	  Required: true
	  In: path
	*/
	ContainerID string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewOptionsContainersGetDeleteParams() beforehand.
func (o *OptionsContainersGetDeleteParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rContainerID, rhkContainerID, _ := route.Params.GetOK("containerId")
	if err := o.bindContainerID(rContainerID, rhkContainerID, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindContainerID binds and validates parameter ContainerID from path.
func (o *OptionsContainersGetDeleteParams) bindContainerID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.ContainerID = raw

	return nil
}
