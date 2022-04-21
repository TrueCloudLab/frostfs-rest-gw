// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/nspcc-dev/neofs-rest-gw/gen/models"
)

// PutContainerEACLOKCode is the HTTP code returned for type PutContainerEACLOK
const PutContainerEACLOKCode int = 200

/*PutContainerEACLOK Successfule EACL upading

swagger:response putContainerEAclOK
*/
type PutContainerEACLOK struct {
}

// NewPutContainerEACLOK creates PutContainerEACLOK with default headers values
func NewPutContainerEACLOK() *PutContainerEACLOK {

	return &PutContainerEACLOK{}
}

// WriteResponse to the client
func (o *PutContainerEACLOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// PutContainerEACLBadRequestCode is the HTTP code returned for type PutContainerEACLBadRequest
const PutContainerEACLBadRequestCode int = 400

/*PutContainerEACLBadRequest Bad request

swagger:response putContainerEAclBadRequest
*/
type PutContainerEACLBadRequest struct {

	/*
	  In: Body
	*/
	Payload models.Error `json:"body,omitempty"`
}

// NewPutContainerEACLBadRequest creates PutContainerEACLBadRequest with default headers values
func NewPutContainerEACLBadRequest() *PutContainerEACLBadRequest {

	return &PutContainerEACLBadRequest{}
}

// WithPayload adds the payload to the put container e Acl bad request response
func (o *PutContainerEACLBadRequest) WithPayload(payload models.Error) *PutContainerEACLBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put container e Acl bad request response
func (o *PutContainerEACLBadRequest) SetPayload(payload models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutContainerEACLBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}