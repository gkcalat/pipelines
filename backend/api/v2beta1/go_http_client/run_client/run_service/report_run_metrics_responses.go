// Code generated by go-swagger; DO NOT EDIT.

package run_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	run_model "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/run_model"
)

// ReportRunMetricsReader is a Reader for the ReportRunMetrics structure.
type ReportRunMetricsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ReportRunMetricsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewReportRunMetricsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewReportRunMetricsDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewReportRunMetricsOK creates a ReportRunMetricsOK with default headers values
func NewReportRunMetricsOK() *ReportRunMetricsOK {
	return &ReportRunMetricsOK{}
}

/*ReportRunMetricsOK handles this case with default header values.

A successful response.
*/
type ReportRunMetricsOK struct {
	Payload *run_model.BackendReportRunMetricsResponse
}

func (o *ReportRunMetricsOK) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/experiments/{experiment_id}/runs/{run_id}:reportMetrics][%d] reportRunMetricsOK  %+v", 200, o.Payload)
}

func (o *ReportRunMetricsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(run_model.BackendReportRunMetricsResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewReportRunMetricsDefault creates a ReportRunMetricsDefault with default headers values
func NewReportRunMetricsDefault(code int) *ReportRunMetricsDefault {
	return &ReportRunMetricsDefault{
		_statusCode: code,
	}
}

/*ReportRunMetricsDefault handles this case with default header values.

ReportRunMetricsDefault report run metrics default
*/
type ReportRunMetricsDefault struct {
	_statusCode int

	Payload *run_model.BackendStatus
}

// Code gets the status code for the report run metrics default response
func (o *ReportRunMetricsDefault) Code() int {
	return o._statusCode
}

func (o *ReportRunMetricsDefault) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/experiments/{experiment_id}/runs/{run_id}:reportMetrics][%d] ReportRunMetrics default  %+v", o._statusCode, o.Payload)
}

func (o *ReportRunMetricsDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(run_model.BackendStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
