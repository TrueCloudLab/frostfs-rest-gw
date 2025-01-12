// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// OptionsContainersPutListOKCode is the HTTP code returned for type OptionsContainersPutListOK
const OptionsContainersPutListOKCode int = 200

/*OptionsContainersPutListOK CORS

swagger:response optionsContainersPutListOK
*/
type OptionsContainersPutListOK struct {
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

// NewOptionsContainersPutListOK creates OptionsContainersPutListOK with default headers values
func NewOptionsContainersPutListOK() *OptionsContainersPutListOK {

	return &OptionsContainersPutListOK{}
}

// WithAccessControlAllowHeaders adds the accessControlAllowHeaders to the options containers put list o k response
func (o *OptionsContainersPutListOK) WithAccessControlAllowHeaders(accessControlAllowHeaders string) *OptionsContainersPutListOK {
	o.AccessControlAllowHeaders = accessControlAllowHeaders
	return o
}

// SetAccessControlAllowHeaders sets the accessControlAllowHeaders to the options containers put list o k response
func (o *OptionsContainersPutListOK) SetAccessControlAllowHeaders(accessControlAllowHeaders string) {
	o.AccessControlAllowHeaders = accessControlAllowHeaders
}

// WithAccessControlAllowMethods adds the accessControlAllowMethods to the options containers put list o k response
func (o *OptionsContainersPutListOK) WithAccessControlAllowMethods(accessControlAllowMethods string) *OptionsContainersPutListOK {
	o.AccessControlAllowMethods = accessControlAllowMethods
	return o
}

// SetAccessControlAllowMethods sets the accessControlAllowMethods to the options containers put list o k response
func (o *OptionsContainersPutListOK) SetAccessControlAllowMethods(accessControlAllowMethods string) {
	o.AccessControlAllowMethods = accessControlAllowMethods
}

// WithAccessControlAllowOrigin adds the accessControlAllowOrigin to the options containers put list o k response
func (o *OptionsContainersPutListOK) WithAccessControlAllowOrigin(accessControlAllowOrigin string) *OptionsContainersPutListOK {
	o.AccessControlAllowOrigin = accessControlAllowOrigin
	return o
}

// SetAccessControlAllowOrigin sets the accessControlAllowOrigin to the options containers put list o k response
func (o *OptionsContainersPutListOK) SetAccessControlAllowOrigin(accessControlAllowOrigin string) {
	o.AccessControlAllowOrigin = accessControlAllowOrigin
}

// WriteResponse to the client
func (o *OptionsContainersPutListOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

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
