// Code generated by go-swagger; DO NOT EDIT.

package run_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// APIRun api run
// swagger:model apiRun
type APIRun struct {

	// Output. The time the run was created.
	// Format: date-time
	CreatedAt strfmt.DateTime `json:"created_at,omitempty"`

	// Optional input field. Describing the purpose of the Run.
	Description string `json:"description,omitempty"`

	// Required input field. Name provided by user,
	// or auto generated if Run is created by scheduled job. Not unique.
	DisplayName string `json:"display_name,omitempty"`

	// In case any error happens retrieving a run field, only run ID
	// and the error message is returned. Client has the flexibility of choosing
	// how to handle error. This is especially useful during listing call.
	Error *APIError `json:"error,omitempty"`

	// Id of the experiment this run belongs to.
	ExperimentID string `json:"experiment_id,omitempty"`

	// Output. The time this run is finished.
	// Format: date-time
	FinishedAt strfmt.DateTime `json:"finished_at,omitempty"`

	// Namespace this run belongs to.
	Namespace string `json:"namespace,omitempty"`

	// The ID of the pipeline user uploaded before.
	PipelineID string `json:"pipeline_id,omitempty"`

	// The pipeline spec.
	PipelineSpec interface{} `json:"pipeline_spec,omitempty"`

	// Output. The runtime details of a Run.
	RunDetails *APIRunDetails `json:"run_details,omitempty"`

	// Output. Unique Run ID. Generated by API server.
	RunID string `json:"run_id,omitempty"`

	// Runtime config of the pipeline.
	RuntimeConfig *APIRuntimeConfig `json:"runtime_config,omitempty"`

	// Output. When this run is scheduled to run. This could be different from
	// created_at. For example, if a run is from a backfilling job that was
	// supposed to run 2 month ago, the scheduled_at is 2 month ago,
	// v.s. created_at is the current time.
	// Format: date-time
	ScheduledAt strfmt.DateTime `json:"scheduled_at,omitempty"`

	// Optional input field. Specifies which Kubernetes service account this run uses.
	ServiceAccount string `json:"service_account,omitempty"`

	// Output. State of a Run.
	State APIRuntimeState `json:"state,omitempty"`

	// Output. A list of Run statuses. This field keeps a record of status
	// evolving over time.
	// Being discussed. Planned as a P1 feature.
	StateHistory []*APIRuntimeStatus `json:"state_history"`

	// Output. Specifies whether this run is in archived or available mode.
	StorageState APIStorageState `json:"storage_state,omitempty"`
}

// Validate validates this api run
func (m *APIRun) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCreatedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateError(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateFinishedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRunDetails(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRuntimeConfig(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateScheduledAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateState(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStateHistory(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStorageState(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *APIRun) validateCreatedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.CreatedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("created_at", "body", "date-time", m.CreatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *APIRun) validateError(formats strfmt.Registry) error {

	if swag.IsZero(m.Error) { // not required
		return nil
	}

	if m.Error != nil {
		if err := m.Error.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("error")
			}
			return err
		}
	}

	return nil
}

func (m *APIRun) validateFinishedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.FinishedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("finished_at", "body", "date-time", m.FinishedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *APIRun) validateRunDetails(formats strfmt.Registry) error {

	if swag.IsZero(m.RunDetails) { // not required
		return nil
	}

	if m.RunDetails != nil {
		if err := m.RunDetails.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("run_details")
			}
			return err
		}
	}

	return nil
}

func (m *APIRun) validateRuntimeConfig(formats strfmt.Registry) error {

	if swag.IsZero(m.RuntimeConfig) { // not required
		return nil
	}

	if m.RuntimeConfig != nil {
		if err := m.RuntimeConfig.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("runtime_config")
			}
			return err
		}
	}

	return nil
}

func (m *APIRun) validateScheduledAt(formats strfmt.Registry) error {

	if swag.IsZero(m.ScheduledAt) { // not required
		return nil
	}

	if err := validate.FormatOf("scheduled_at", "body", "date-time", m.ScheduledAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *APIRun) validateState(formats strfmt.Registry) error {

	if swag.IsZero(m.State) { // not required
		return nil
	}

	if err := m.State.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("state")
		}
		return err
	}

	return nil
}

func (m *APIRun) validateStateHistory(formats strfmt.Registry) error {

	if swag.IsZero(m.StateHistory) { // not required
		return nil
	}

	for i := 0; i < len(m.StateHistory); i++ {
		if swag.IsZero(m.StateHistory[i]) { // not required
			continue
		}

		if m.StateHistory[i] != nil {
			if err := m.StateHistory[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("state_history" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *APIRun) validateStorageState(formats strfmt.Registry) error {

	if swag.IsZero(m.StorageState) { // not required
		return nil
	}

	if err := m.StorageState.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("storage_state")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *APIRun) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *APIRun) UnmarshalBinary(b []byte) error {
	var res APIRun
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
