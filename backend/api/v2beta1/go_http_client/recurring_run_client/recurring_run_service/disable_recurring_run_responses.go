// Code generated by go-swagger; DO NOT EDIT.

package recurring_run_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	recurring_run_model "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/recurring_run_model"
)

// DisableRecurringRunReader is a Reader for the DisableRecurringRun structure.
type DisableRecurringRunReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DisableRecurringRunReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewDisableRecurringRunOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewDisableRecurringRunDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewDisableRecurringRunOK creates a DisableRecurringRunOK with default headers values
func NewDisableRecurringRunOK() *DisableRecurringRunOK {
	return &DisableRecurringRunOK{}
}

/*DisableRecurringRunOK handles this case with default header values.

A successful response.
*/
type DisableRecurringRunOK struct {
	Payload interface{}
}

func (o *DisableRecurringRunOK) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/recurringruns/{recurring_run_id}:disable][%d] disableRecurringRunOK  %+v", 200, o.Payload)
}

func (o *DisableRecurringRunOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDisableRecurringRunDefault creates a DisableRecurringRunDefault with default headers values
func NewDisableRecurringRunDefault(code int) *DisableRecurringRunDefault {
	return &DisableRecurringRunDefault{
		_statusCode: code,
	}
}

/*DisableRecurringRunDefault handles this case with default header values.

DisableRecurringRunDefault disable recurring run default
*/
type DisableRecurringRunDefault struct {
	_statusCode int

	Payload *recurring_run_model.APIStatus
}

// Code gets the status code for the disable recurring run default response
func (o *DisableRecurringRunDefault) Code() int {
	return o._statusCode
}

func (o *DisableRecurringRunDefault) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/recurringruns/{recurring_run_id}:disable][%d] DisableRecurringRun default  %+v", o._statusCode, o.Payload)
}

func (o *DisableRecurringRunDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(recurring_run_model.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
