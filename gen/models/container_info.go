// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// ContainerInfo Information about container.
// Example: {"attribute":[{"key":"Timestamp","value":"1648810072"},{"key":"Name","value":"container"}],"basicAcl":"0x1fbf9fff","containerId":"5HZTn5qkRnmgSz9gSrw22CEdPPk6nQhkwf2Mgzyvkikv","containerName":"container","ownerId":"NbUgTSFvPmsRxmGeWpuuGeJUoRoi6PErcM","placementPolicy":"REP 2","version":"2.11"}
//
// swagger:model ContainerInfo
type ContainerInfo struct {

	// attributes
	// Required: true
	Attributes []*Attribute `json:"attributes"`

	// basic Acl
	// Required: true
	BasicACL *string `json:"basicAcl"`

	// The friendly name for the basicAcl field.
	CannedACL string `json:"cannedAcl,omitempty"`

	// container Id
	// Required: true
	ContainerID *string `json:"containerId"`

	// container name
	// Required: true
	ContainerName *string `json:"containerName"`

	// owner Id
	// Required: true
	OwnerID *string `json:"ownerId"`

	// placement policy
	// Required: true
	PlacementPolicy *string `json:"placementPolicy"`

	// version
	// Required: true
	Version *string `json:"version"`
}

// Validate validates this container info
func (m *ContainerInfo) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAttributes(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateBasicACL(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateContainerID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateContainerName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOwnerID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePlacementPolicy(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVersion(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ContainerInfo) validateAttributes(formats strfmt.Registry) error {

	if err := validate.Required("attributes", "body", m.Attributes); err != nil {
		return err
	}

	for i := 0; i < len(m.Attributes); i++ {
		if swag.IsZero(m.Attributes[i]) { // not required
			continue
		}

		if m.Attributes[i] != nil {
			if err := m.Attributes[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("attributes" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("attributes" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *ContainerInfo) validateBasicACL(formats strfmt.Registry) error {

	if err := validate.Required("basicAcl", "body", m.BasicACL); err != nil {
		return err
	}

	return nil
}

func (m *ContainerInfo) validateContainerID(formats strfmt.Registry) error {

	if err := validate.Required("containerId", "body", m.ContainerID); err != nil {
		return err
	}

	return nil
}

func (m *ContainerInfo) validateContainerName(formats strfmt.Registry) error {

	if err := validate.Required("containerName", "body", m.ContainerName); err != nil {
		return err
	}

	return nil
}

func (m *ContainerInfo) validateOwnerID(formats strfmt.Registry) error {

	if err := validate.Required("ownerId", "body", m.OwnerID); err != nil {
		return err
	}

	return nil
}

func (m *ContainerInfo) validatePlacementPolicy(formats strfmt.Registry) error {

	if err := validate.Required("placementPolicy", "body", m.PlacementPolicy); err != nil {
		return err
	}

	return nil
}

func (m *ContainerInfo) validateVersion(formats strfmt.Registry) error {

	if err := validate.Required("version", "body", m.Version); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this container info based on the context it is used
func (m *ContainerInfo) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateAttributes(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ContainerInfo) contextValidateAttributes(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Attributes); i++ {

		if m.Attributes[i] != nil {
			if err := m.Attributes[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("attributes" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("attributes" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *ContainerInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ContainerInfo) UnmarshalBinary(b []byte) error {
	var res ContainerInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
