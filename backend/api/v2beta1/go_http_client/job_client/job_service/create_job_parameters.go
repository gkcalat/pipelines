// Code generated by go-swagger; DO NOT EDIT.

package job_service

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

	job_model "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/job_model"
)

// NewCreateJobParams creates a new CreateJobParams object
// with the default values initialized.
func NewCreateJobParams() *CreateJobParams {
	var ()
	return &CreateJobParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewCreateJobParamsWithTimeout creates a new CreateJobParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewCreateJobParamsWithTimeout(timeout time.Duration) *CreateJobParams {
	var ()
	return &CreateJobParams{

		timeout: timeout,
	}
}

// NewCreateJobParamsWithContext creates a new CreateJobParams object
// with the default values initialized, and the ability to set a context for a request
func NewCreateJobParamsWithContext(ctx context.Context) *CreateJobParams {
	var ()
	return &CreateJobParams{

		Context: ctx,
	}
}

// NewCreateJobParamsWithHTTPClient creates a new CreateJobParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewCreateJobParamsWithHTTPClient(client *http.Client) *CreateJobParams {
	var ()
	return &CreateJobParams{
		HTTPClient: client,
	}
}

/*CreateJobParams contains all the parameters to send to the API endpoint
for the create job operation typically these are written to a http.Request
*/
type CreateJobParams struct {

	/*Body
	  The job to be created

	*/
	Body *job_model.APIJob

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the create job params
func (o *CreateJobParams) WithTimeout(timeout time.Duration) *CreateJobParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the create job params
func (o *CreateJobParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the create job params
func (o *CreateJobParams) WithContext(ctx context.Context) *CreateJobParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the create job params
func (o *CreateJobParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the create job params
func (o *CreateJobParams) WithHTTPClient(client *http.Client) *CreateJobParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the create job params
func (o *CreateJobParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the create job params
func (o *CreateJobParams) WithBody(body *job_model.APIJob) *CreateJobParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the create job params
func (o *CreateJobParams) SetBody(body *job_model.APIJob) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *CreateJobParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
