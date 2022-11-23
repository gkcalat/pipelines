// Code generated by go-swagger; DO NOT EDIT.

package run_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// APIError Error message.
// swagger:model apiError
type APIError struct {

	// Error details.
	ErrorDetails string `json:"error_details,omitempty"`

	// Error message.
	ErrorMessage string `json:"error_message,omitempty"`
}

// Validate validates this api error
func (m *APIError) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *APIError) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *APIError) UnmarshalBinary(b []byte) error {
	var res APIError
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
