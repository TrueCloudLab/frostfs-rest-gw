// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NewGetStorageGroupParams creates a new GetStorageGroupParams object
// with the default values initialized.
func NewGetStorageGroupParams() GetStorageGroupParams {

	var (
		// initialize parameters with default values

		fullBearerDefault = bool(false)

		walletConnectDefault = bool(false)
	)

	return GetStorageGroupParams{
		FullBearer: &fullBearerDefault,

		WalletConnect: &walletConnectDefault,
	}
}

// GetStorageGroupParams contains all the bound params for the get storage group operation
// typically these are obtained from a http.Request
//
// swagger:parameters getStorageGroup
type GetStorageGroupParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Base64 encoded signature for bearer token.
	  In: header
	*/
	XBearerSignature *string
	/*Hex encoded the public part of the key that signed the bearer token.
	  In: header
	*/
	XBearerSignatureKey *string
	/*Base58 encoded container id.
	  Required: true
	  In: path
	*/
	ContainerID string
	/*Provided bearer token is final or gate should assemble it using signature.
	  In: query
	  Default: false
	*/
	FullBearer *bool
	/*Base58 encoded storage group id.
	  Required: true
	  In: path
	*/
	StorageGroupID string
	/*Use wallet connect signature scheme or native FrostFS signature.
	  In: query
	  Default: false
	*/
	WalletConnect *bool
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetStorageGroupParams() beforehand.
func (o *GetStorageGroupParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	if err := o.bindXBearerSignature(r.Header[http.CanonicalHeaderKey("X-Bearer-Signature")], true, route.Formats); err != nil {
		res = append(res, err)
	}

	if err := o.bindXBearerSignatureKey(r.Header[http.CanonicalHeaderKey("X-Bearer-Signature-Key")], true, route.Formats); err != nil {
		res = append(res, err)
	}

	rContainerID, rhkContainerID, _ := route.Params.GetOK("containerId")
	if err := o.bindContainerID(rContainerID, rhkContainerID, route.Formats); err != nil {
		res = append(res, err)
	}

	qFullBearer, qhkFullBearer, _ := qs.GetOK("fullBearer")
	if err := o.bindFullBearer(qFullBearer, qhkFullBearer, route.Formats); err != nil {
		res = append(res, err)
	}

	rStorageGroupID, rhkStorageGroupID, _ := route.Params.GetOK("storageGroupId")
	if err := o.bindStorageGroupID(rStorageGroupID, rhkStorageGroupID, route.Formats); err != nil {
		res = append(res, err)
	}

	qWalletConnect, qhkWalletConnect, _ := qs.GetOK("walletConnect")
	if err := o.bindWalletConnect(qWalletConnect, qhkWalletConnect, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindXBearerSignature binds and validates parameter XBearerSignature from header.
func (o *GetStorageGroupParams) bindXBearerSignature(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.XBearerSignature = &raw

	return nil
}

// bindXBearerSignatureKey binds and validates parameter XBearerSignatureKey from header.
func (o *GetStorageGroupParams) bindXBearerSignatureKey(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.XBearerSignatureKey = &raw

	return nil
}

// bindContainerID binds and validates parameter ContainerID from path.
func (o *GetStorageGroupParams) bindContainerID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.ContainerID = raw

	return nil
}

// bindFullBearer binds and validates parameter FullBearer from query.
func (o *GetStorageGroupParams) bindFullBearer(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetStorageGroupParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("fullBearer", "query", "bool", raw)
	}
	o.FullBearer = &value

	return nil
}

// bindStorageGroupID binds and validates parameter StorageGroupID from path.
func (o *GetStorageGroupParams) bindStorageGroupID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.StorageGroupID = raw

	return nil
}

// bindWalletConnect binds and validates parameter WalletConnect from query.
func (o *GetStorageGroupParams) bindWalletConnect(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetStorageGroupParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("walletConnect", "query", "bool", raw)
	}
	o.WalletConnect = &value

	return nil
}
