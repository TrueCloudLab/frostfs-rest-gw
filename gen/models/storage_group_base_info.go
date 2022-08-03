// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// StorageGroupBaseInfo Storage group info for listing.
//
// swagger:model StorageGroupBaseInfo
type StorageGroupBaseInfo struct {

	// address
	// Required: true
	Address *Address `json:"address"`

	// expiration epoch
	// Required: true
	ExpirationEpoch *string `json:"expirationEpoch"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this storage group base info
func (m *StorageGroupBaseInfo) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAddress(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateExpirationEpoch(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *StorageGroupBaseInfo) validateAddress(formats strfmt.Registry) error {

	if err := validate.Required("address", "body", m.Address); err != nil {
		return err
	}

	if m.Address != nil {
		if err := m.Address.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("address")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("address")
			}
			return err
		}
	}

	return nil
}

func (m *StorageGroupBaseInfo) validateExpirationEpoch(formats strfmt.Registry) error {

	if err := validate.Required("expirationEpoch", "body", m.ExpirationEpoch); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this storage group base info based on the context it is used
func (m *StorageGroupBaseInfo) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateAddress(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *StorageGroupBaseInfo) contextValidateAddress(ctx context.Context, formats strfmt.Registry) error {

	if m.Address != nil {
		if err := m.Address.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("address")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("address")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *StorageGroupBaseInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *StorageGroupBaseInfo) UnmarshalBinary(b []byte) error {
	var res StorageGroupBaseInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}