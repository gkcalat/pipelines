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

// NewDeleteRunParams creates a new DeleteRunParams object
// with the default values initialized.
func NewDeleteRunParams() *DeleteRunParams {
	var ()
	return &DeleteRunParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewDeleteRunParamsWithTimeout creates a new DeleteRunParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewDeleteRunParamsWithTimeout(timeout time.Duration) *DeleteRunParams {
	var ()
	return &DeleteRunParams{

		timeout: timeout,
	}
}

// NewDeleteRunParamsWithContext creates a new DeleteRunParams object
// with the default values initialized, and the ability to set a context for a request
func NewDeleteRunParamsWithContext(ctx context.Context) *DeleteRunParams {
	var ()
	return &DeleteRunParams{

		Context: ctx,
	}
}

// NewDeleteRunParamsWithHTTPClient creates a new DeleteRunParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewDeleteRunParamsWithHTTPClient(client *http.Client) *DeleteRunParams {
	var ()
	return &DeleteRunParams{
		HTTPClient: client,
	}
}

/*DeleteRunParams contains all the parameters to send to the API endpoint
for the delete run operation typically these are written to a http.Request
*/
type DeleteRunParams struct {

	/*RunID
	  The ID of the run to be deleted.

	*/
	RunID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the delete run params
func (o *DeleteRunParams) WithTimeout(timeout time.Duration) *DeleteRunParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete run params
func (o *DeleteRunParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the delete run params
func (o *DeleteRunParams) WithContext(ctx context.Context) *DeleteRunParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete run params
func (o *DeleteRunParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete run params
func (o *DeleteRunParams) WithHTTPClient(client *http.Client) *DeleteRunParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete run params
func (o *DeleteRunParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithRunID adds the runID to the delete run params
func (o *DeleteRunParams) WithRunID(runID string) *DeleteRunParams {
	o.SetRunID(runID)
	return o
}

// SetRunID adds the runId to the delete run params
func (o *DeleteRunParams) SetRunID(runID string) {
	o.RunID = runID
}

// WriteToRequest writes these params to a swagger request
func (o *DeleteRunParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param run_id
	if err := r.SetPathParam("run_id", o.RunID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
