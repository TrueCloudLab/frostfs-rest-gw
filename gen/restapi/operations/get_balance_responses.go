// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/nspcc-dev/neofs-rest-gw/gen/models"
)

// GetBalanceOKCode is the HTTP code returned for type GetBalanceOK
const GetBalanceOKCode int = 200

/*GetBalanceOK Balance of address in NeoFS

swagger:response getBalanceOK
*/
type GetBalanceOK struct {

	/*
	  In: Body
	*/
	Payload *models.Balance `json:"body,omitempty"`
}

// NewGetBalanceOK creates GetBalanceOK with default headers values
func NewGetBalanceOK() *GetBalanceOK {

	return &GetBalanceOK{}
}

// WithPayload adds the payload to the get balance o k response
func (o *GetBalanceOK) WithPayload(payload *models.Balance) *GetBalanceOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get balance o k response
func (o *GetBalanceOK) SetPayload(payload *models.Balance) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetBalanceOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetBalanceBadRequestCode is the HTTP code returned for type GetBalanceBadRequest
const GetBalanceBadRequestCode int = 400

/*GetBalanceBadRequest Bad request

swagger:response getBalanceBadRequest
*/
type GetBalanceBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewGetBalanceBadRequest creates GetBalanceBadRequest with default headers values
func NewGetBalanceBadRequest() *GetBalanceBadRequest {

	return &GetBalanceBadRequest{}
}

// WithPayload adds the payload to the get balance bad request response
func (o *GetBalanceBadRequest) WithPayload(payload *models.ErrorResponse) *GetBalanceBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get balance bad request response
func (o *GetBalanceBadRequest) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetBalanceBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}