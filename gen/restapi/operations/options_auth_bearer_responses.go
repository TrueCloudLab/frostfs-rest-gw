// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// OptionsAuthBearerOKCode is the HTTP code returned for type OptionsAuthBearerOK
const OptionsAuthBearerOKCode int = 200

/*OptionsAuthBearerOK CORS

swagger:response optionsAuthBearerOK
*/
type OptionsAuthBearerOK struct {
	/*

	 */
	AccessControlAllowHeaders string `json:"Access-Control-Allow-Headers"`
	/*

	 */
	AccessControlAllowOrigin string `json:"Access-Control-Allow-Origin"`
}

// NewOptionsAuthBearerOK creates OptionsAuthBearerOK with default headers values
func NewOptionsAuthBearerOK() *OptionsAuthBearerOK {

	return &OptionsAuthBearerOK{}
}

// WithAccessControlAllowHeaders adds the accessControlAllowHeaders to the options auth bearer o k response
func (o *OptionsAuthBearerOK) WithAccessControlAllowHeaders(accessControlAllowHeaders string) *OptionsAuthBearerOK {
	o.AccessControlAllowHeaders = accessControlAllowHeaders
	return o
}

// SetAccessControlAllowHeaders sets the accessControlAllowHeaders to the options auth bearer o k response
func (o *OptionsAuthBearerOK) SetAccessControlAllowHeaders(accessControlAllowHeaders string) {
	o.AccessControlAllowHeaders = accessControlAllowHeaders
}

// WithAccessControlAllowOrigin adds the accessControlAllowOrigin to the options auth bearer o k response
func (o *OptionsAuthBearerOK) WithAccessControlAllowOrigin(accessControlAllowOrigin string) *OptionsAuthBearerOK {
	o.AccessControlAllowOrigin = accessControlAllowOrigin
	return o
}

// SetAccessControlAllowOrigin sets the accessControlAllowOrigin to the options auth bearer o k response
func (o *OptionsAuthBearerOK) SetAccessControlAllowOrigin(accessControlAllowOrigin string) {
	o.AccessControlAllowOrigin = accessControlAllowOrigin
}

// WriteResponse to the client
func (o *OptionsAuthBearerOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	// response header Access-Control-Allow-Headers

	accessControlAllowHeaders := o.AccessControlAllowHeaders
	if accessControlAllowHeaders != "" {
		rw.Header().Set("Access-Control-Allow-Headers", accessControlAllowHeaders)
	}

	// response header Access-Control-Allow-Origin

	accessControlAllowOrigin := o.AccessControlAllowOrigin
	if accessControlAllowOrigin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", accessControlAllowOrigin)
	}

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}