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

	run_model "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/run_model"
)

// NewReportRunMetricsParams creates a new ReportRunMetricsParams object
// with the default values initialized.
func NewReportRunMetricsParams() *ReportRunMetricsParams {
	var ()
	return &ReportRunMetricsParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewReportRunMetricsParamsWithTimeout creates a new ReportRunMetricsParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewReportRunMetricsParamsWithTimeout(timeout time.Duration) *ReportRunMetricsParams {
	var ()
	return &ReportRunMetricsParams{

		timeout: timeout,
	}
}

// NewReportRunMetricsParamsWithContext creates a new ReportRunMetricsParams object
// with the default values initialized, and the ability to set a context for a request
func NewReportRunMetricsParamsWithContext(ctx context.Context) *ReportRunMetricsParams {
	var ()
	return &ReportRunMetricsParams{

		Context: ctx,
	}
}

// NewReportRunMetricsParamsWithHTTPClient creates a new ReportRunMetricsParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewReportRunMetricsParamsWithHTTPClient(client *http.Client) *ReportRunMetricsParams {
	var ()
	return &ReportRunMetricsParams{
		HTTPClient: client,
	}
}

/*ReportRunMetricsParams contains all the parameters to send to the API endpoint
for the report run metrics operation typically these are written to a http.Request
*/
type ReportRunMetricsParams struct {

	/*Body*/
	Body *run_model.BackendReportRunMetricsRequest
	/*ExperimentID
	  The ID of the parent experiment.

	*/
	ExperimentID string
	/*RunID
	  Required. The parent run ID of the metric.

	*/
	RunID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the report run metrics params
func (o *ReportRunMetricsParams) WithTimeout(timeout time.Duration) *ReportRunMetricsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the report run metrics params
func (o *ReportRunMetricsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the report run metrics params
func (o *ReportRunMetricsParams) WithContext(ctx context.Context) *ReportRunMetricsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the report run metrics params
func (o *ReportRunMetricsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the report run metrics params
func (o *ReportRunMetricsParams) WithHTTPClient(client *http.Client) *ReportRunMetricsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the report run metrics params
func (o *ReportRunMetricsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the report run metrics params
func (o *ReportRunMetricsParams) WithBody(body *run_model.BackendReportRunMetricsRequest) *ReportRunMetricsParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the report run metrics params
func (o *ReportRunMetricsParams) SetBody(body *run_model.BackendReportRunMetricsRequest) {
	o.Body = body
}

// WithExperimentID adds the experimentID to the report run metrics params
func (o *ReportRunMetricsParams) WithExperimentID(experimentID string) *ReportRunMetricsParams {
	o.SetExperimentID(experimentID)
	return o
}

// SetExperimentID adds the experimentId to the report run metrics params
func (o *ReportRunMetricsParams) SetExperimentID(experimentID string) {
	o.ExperimentID = experimentID
}

// WithRunID adds the runID to the report run metrics params
func (o *ReportRunMetricsParams) WithRunID(runID string) *ReportRunMetricsParams {
	o.SetRunID(runID)
	return o
}

// SetRunID adds the runId to the report run metrics params
func (o *ReportRunMetricsParams) SetRunID(runID string) {
	o.RunID = runID
}

// WriteToRequest writes these params to a swagger request
func (o *ReportRunMetricsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

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
