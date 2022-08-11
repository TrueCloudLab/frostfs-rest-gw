// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// OptionsObjectsGetDeleteOKCode is the HTTP code returned for type OptionsObjectsGetDeleteOK
const OptionsObjectsGetDeleteOKCode int = 200

/*OptionsObjectsGetDeleteOK CORS

swagger:response optionsObjectsGetDeleteOK
*/
type OptionsObjectsGetDeleteOK struct {
	/*

	 */
	AccessControlAllowHeaders string `json:"Access-Control-Allow-Headers"`
	/*

	 */
	AccessControlAllowMethods string `json:"Access-Control-Allow-Methods"`
	/*

	 */
	AccessControlAllowOrigin string `json:"Access-Control-Allow-Origin"`
}

// NewOptionsObjectsGetDeleteOK creates OptionsObjectsGetDeleteOK with default headers values
func NewOptionsObjectsGetDeleteOK() *OptionsObjectsGetDeleteOK {

	return &OptionsObjectsGetDeleteOK{}
}

// WithAccessControlAllowHeaders adds the accessControlAllowHeaders to the options objects get delete o k response
func (o *OptionsObjectsGetDeleteOK) WithAccessControlAllowHeaders(accessControlAllowHeaders string) *OptionsObjectsGetDeleteOK {
	o.AccessControlAllowHeaders = accessControlAllowHeaders
	return o
}

// SetAccessControlAllowHeaders sets the accessControlAllowHeaders to the options objects get delete o k response
func (o *OptionsObjectsGetDeleteOK) SetAccessControlAllowHeaders(accessControlAllowHeaders string) {
	o.AccessControlAllowHeaders = accessControlAllowHeaders
}

// WithAccessControlAllowMethods adds the accessControlAllowMethods to the options objects get delete o k response
func (o *OptionsObjectsGetDeleteOK) WithAccessControlAllowMethods(accessControlAllowMethods string) *OptionsObjectsGetDeleteOK {
	o.AccessControlAllowMethods = accessControlAllowMethods
	return o
}

// SetAccessControlAllowMethods sets the accessControlAllowMethods to the options objects get delete o k response
func (o *OptionsObjectsGetDeleteOK) SetAccessControlAllowMethods(accessControlAllowMethods string) {
	o.AccessControlAllowMethods = accessControlAllowMethods
}

// WithAccessControlAllowOrigin adds the accessControlAllowOrigin to the options objects get delete o k response
func (o *OptionsObjectsGetDeleteOK) WithAccessControlAllowOrigin(accessControlAllowOrigin string) *OptionsObjectsGetDeleteOK {
	o.AccessControlAllowOrigin = accessControlAllowOrigin
	return o
}

// SetAccessControlAllowOrigin sets the accessControlAllowOrigin to the options objects get delete o k response
func (o *OptionsObjectsGetDeleteOK) SetAccessControlAllowOrigin(accessControlAllowOrigin string) {
	o.AccessControlAllowOrigin = accessControlAllowOrigin
}

// WriteResponse to the client
func (o *OptionsObjectsGetDeleteOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	// response header Access-Control-Allow-Headers

	accessControlAllowHeaders := o.AccessControlAllowHeaders
	if accessControlAllowHeaders != "" {
		rw.Header().Set("Access-Control-Allow-Headers", accessControlAllowHeaders)
	}

	// response header Access-Control-Allow-Methods

	accessControlAllowMethods := o.AccessControlAllowMethods
	if accessControlAllowMethods != "" {
		rw.Header().Set("Access-Control-Allow-Methods", accessControlAllowMethods)
	}

	// response header Access-Control-Allow-Origin

	accessControlAllowOrigin := o.AccessControlAllowOrigin
	if accessControlAllowOrigin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", accessControlAllowOrigin)
	}

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}