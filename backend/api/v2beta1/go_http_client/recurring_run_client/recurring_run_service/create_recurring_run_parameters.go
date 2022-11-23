// Code generated by go-swagger; DO NOT EDIT.

package recurring_run_service

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

	recurring_run_model "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/recurring_run_model"
)

// NewCreateRecurringRunParams creates a new CreateRecurringRunParams object
// with the default values initialized.
func NewCreateRecurringRunParams() *CreateRecurringRunParams {
	var ()
	return &CreateRecurringRunParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewCreateRecurringRunParamsWithTimeout creates a new CreateRecurringRunParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewCreateRecurringRunParamsWithTimeout(timeout time.Duration) *CreateRecurringRunParams {
	var ()
	return &CreateRecurringRunParams{

		timeout: timeout,
	}
}

// NewCreateRecurringRunParamsWithContext creates a new CreateRecurringRunParams object
// with the default values initialized, and the ability to set a context for a request
func NewCreateRecurringRunParamsWithContext(ctx context.Context) *CreateRecurringRunParams {
	var ()
	return &CreateRecurringRunParams{

		Context: ctx,
	}
}

// NewCreateRecurringRunParamsWithHTTPClient creates a new CreateRecurringRunParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewCreateRecurringRunParamsWithHTTPClient(client *http.Client) *CreateRecurringRunParams {
	var ()
	return &CreateRecurringRunParams{
		HTTPClient: client,
	}
}

/*CreateRecurringRunParams contains all the parameters to send to the API endpoint
for the create recurring run operation typically these are written to a http.Request
*/
type CreateRecurringRunParams struct {

	/*Body
	  The recurring run to be created.

	*/
	Body *recurring_run_model.APIRecurringRun

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the create recurring run params
func (o *CreateRecurringRunParams) WithTimeout(timeout time.Duration) *CreateRecurringRunParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the create recurring run params
func (o *CreateRecurringRunParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the create recurring run params
func (o *CreateRecurringRunParams) WithContext(ctx context.Context) *CreateRecurringRunParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the create recurring run params
func (o *CreateRecurringRunParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the create recurring run params
func (o *CreateRecurringRunParams) WithHTTPClient(client *http.Client) *CreateRecurringRunParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the create recurring run params
func (o *CreateRecurringRunParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the create recurring run params
func (o *CreateRecurringRunParams) WithBody(body *recurring_run_model.APIRecurringRun) *CreateRecurringRunParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the create recurring run params
func (o *CreateRecurringRunParams) SetBody(body *recurring_run_model.APIRecurringRun) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *CreateRecurringRunParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
