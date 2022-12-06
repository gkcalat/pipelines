// Code generated by go-swagger; DO NOT EDIT.

package run_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewUnarchiveRunParams creates a new UnarchiveRunParams object
// with the default values initialized.
func NewUnarchiveRunParams() *UnarchiveRunParams {
	var ()
	return &UnarchiveRunParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewUnarchiveRunParamsWithTimeout creates a new UnarchiveRunParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewUnarchiveRunParamsWithTimeout(timeout time.Duration) *UnarchiveRunParams {
	var ()
	return &UnarchiveRunParams{

		timeout: timeout,
	}
}

// NewUnarchiveRunParamsWithContext creates a new UnarchiveRunParams object
// with the default values initialized, and the ability to set a context for a request
func NewUnarchiveRunParamsWithContext(ctx context.Context) *UnarchiveRunParams {
	var ()
	return &UnarchiveRunParams{

		Context: ctx,
	}
}

// NewUnarchiveRunParamsWithHTTPClient creates a new UnarchiveRunParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewUnarchiveRunParamsWithHTTPClient(client *http.Client) *UnarchiveRunParams {
	var ()
	return &UnarchiveRunParams{
		HTTPClient: client,
	}
}

/*UnarchiveRunParams contains all the parameters to send to the API endpoint
for the unarchive run operation typically these are written to a http.Request
*/
type UnarchiveRunParams struct {

	/*ExperimentID
	  The ID of the parent experiment.

	*/
	ExperimentID string
	/*RunID
	  The ID of the run to be restored.

	*/
	RunID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the unarchive run params
func (o *UnarchiveRunParams) WithTimeout(timeout time.Duration) *UnarchiveRunParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the unarchive run params
func (o *UnarchiveRunParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the unarchive run params
func (o *UnarchiveRunParams) WithContext(ctx context.Context) *UnarchiveRunParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the unarchive run params
func (o *UnarchiveRunParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the unarchive run params
func (o *UnarchiveRunParams) WithHTTPClient(client *http.Client) *UnarchiveRunParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the unarchive run params
func (o *UnarchiveRunParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithExperimentID adds the experimentID to the unarchive run params
func (o *UnarchiveRunParams) WithExperimentID(experimentID string) *UnarchiveRunParams {
	o.SetExperimentID(experimentID)
	return o
}

// SetExperimentID adds the experimentId to the unarchive run params
func (o *UnarchiveRunParams) SetExperimentID(experimentID string) {
	o.ExperimentID = experimentID
}

// WithRunID adds the runID to the unarchive run params
func (o *UnarchiveRunParams) WithRunID(runID string) *UnarchiveRunParams {
	o.SetRunID(runID)
	return o
}

// SetRunID adds the runId to the unarchive run params
func (o *UnarchiveRunParams) SetRunID(runID string) {
	o.RunID = runID
}

// WriteToRequest writes these params to a swagger request
func (o *UnarchiveRunParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param experiment_id
	if err := r.SetPathParam("experiment_id", o.ExperimentID); err != nil {
		return err
	}

	// path param run_id
	if err := r.SetPathParam("run_id", o.RunID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
