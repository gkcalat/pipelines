// Code generated by go-swagger; DO NOT EDIT.

package recurring_run_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// APITrigger Trigger defines what starts a pipeline run.
// swagger:model apiTrigger
type APITrigger struct {

	// cron schedule
	CronSchedule *APICronSchedule `json:"cron_schedule,omitempty"`

	// periodic schedule
	PeriodicSchedule *APIPeriodicSchedule `json:"periodic_schedule,omitempty"`
}

// Validate validates this api trigger
func (m *APITrigger) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCronSchedule(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePeriodicSchedule(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *APITrigger) validateCronSchedule(formats strfmt.Registry) error {

	if swag.IsZero(m.CronSchedule) { // not required
		return nil
	}

	if m.CronSchedule != nil {
		if err := m.CronSchedule.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("cron_schedule")
			}
			return err
		}
	}

	return nil
}

func (m *APITrigger) validatePeriodicSchedule(formats strfmt.Registry) error {

	if swag.IsZero(m.PeriodicSchedule) { // not required
		return nil
	}

	if m.PeriodicSchedule != nil {
		if err := m.PeriodicSchedule.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("periodic_schedule")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *APITrigger) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *APITrigger) UnmarshalBinary(b []byte) error {
	var res APITrigger
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
