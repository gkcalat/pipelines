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

// CreateRecurringRunReader is a Reader for the CreateRecurringRun structure.
type CreateRecurringRunReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateRecurringRunReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewCreateRecurringRunOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewCreateRecurringRunDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewCreateRecurringRunOK creates a CreateRecurringRunOK with default headers values
func NewCreateRecurringRunOK() *CreateRecurringRunOK {
	return &CreateRecurringRunOK{}
}

/*CreateRecurringRunOK handles this case with default header values.

A successful response.
*/
type CreateRecurringRunOK struct {
	Payload *recurring_run_model.APIRecurringRun
}

func (o *CreateRecurringRunOK) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/recurringruns][%d] createRecurringRunOK  %+v", 200, o.Payload)
}

func (o *CreateRecurringRunOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(recurring_run_model.APIRecurringRun)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateRecurringRunDefault creates a CreateRecurringRunDefault with default headers values
func NewCreateRecurringRunDefault(code int) *CreateRecurringRunDefault {
	return &CreateRecurringRunDefault{
		_statusCode: code,
	}
}

/*CreateRecurringRunDefault handles this case with default header values.

CreateRecurringRunDefault create recurring run default
*/
type CreateRecurringRunDefault struct {
	_statusCode int

	Payload *recurring_run_model.APIStatus
}

// Code gets the status code for the create recurring run default response
func (o *CreateRecurringRunDefault) Code() int {
	return o._statusCode
}

func (o *CreateRecurringRunDefault) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/recurringruns][%d] CreateRecurringRun default  %+v", o._statusCode, o.Payload)
}

func (o *CreateRecurringRunDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(recurring_run_model.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
