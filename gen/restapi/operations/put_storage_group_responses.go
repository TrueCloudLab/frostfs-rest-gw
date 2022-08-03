// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/nspcc-dev/neofs-rest-gw/gen/models"
)

// PutStorageGroupOKCode is the HTTP code returned for type PutStorageGroupOK
const PutStorageGroupOKCode int = 200

/*PutStorageGroupOK Address of uploaded storage group.

swagger:response putStorageGroupOK
*/
type PutStorageGroupOK struct {

	/*
	  In: Body
	*/
	Payload *models.Address `json:"body,omitempty"`
}

// NewPutStorageGroupOK creates PutStorageGroupOK with default headers values
func NewPutStorageGroupOK() *PutStorageGroupOK {

	return &PutStorageGroupOK{}
}

// WithPayload adds the payload to the put storage group o k response
func (o *PutStorageGroupOK) WithPayload(payload *models.Address) *PutStorageGroupOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put storage group o k response
func (o *PutStorageGroupOK) SetPayload(payload *models.Address) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutStorageGroupOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutStorageGroupBadRequestCode is the HTTP code returned for type PutStorageGroupBadRequest
const PutStorageGroupBadRequestCode int = 400

/*PutStorageGroupBadRequest Bad request.

swagger:response putStorageGroupBadRequest
*/
type PutStorageGroupBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewPutStorageGroupBadRequest creates PutStorageGroupBadRequest with default headers values
func NewPutStorageGroupBadRequest() *PutStorageGroupBadRequest {

	return &PutStorageGroupBadRequest{}
}

// WithPayload adds the payload to the put storage group bad request response
func (o *PutStorageGroupBadRequest) WithPayload(payload *models.ErrorResponse) *PutStorageGroupBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put storage group bad request response
func (o *PutStorageGroupBadRequest) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutStorageGroupBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}