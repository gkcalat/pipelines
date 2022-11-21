// Code generated by go-swagger; DO NOT EDIT.

package job_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	job_model "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/job_model"
)

// EnableJobReader is a Reader for the EnableJob structure.
type EnableJobReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *EnableJobReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewEnableJobOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewEnableJobDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewEnableJobOK creates a EnableJobOK with default headers values
func NewEnableJobOK() *EnableJobOK {
	return &EnableJobOK{}
}

/*EnableJobOK handles this case with default header values.

A successful response.
*/
type EnableJobOK struct {
	Payload interface{}
}

func (o *EnableJobOK) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/jobs/{id}/enable][%d] enableJobOK  %+v", 200, o.Payload)
}

func (o *EnableJobOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewEnableJobDefault creates a EnableJobDefault with default headers values
func NewEnableJobDefault(code int) *EnableJobDefault {
	return &EnableJobDefault{
		_statusCode: code,
	}
}

/*EnableJobDefault handles this case with default header values.

EnableJobDefault enable job default
*/
type EnableJobDefault struct {
	_statusCode int

	Payload *job_model.APIStatus
}

// Code gets the status code for the enable job default response
func (o *EnableJobDefault) Code() int {
	return o._statusCode
}

func (o *EnableJobDefault) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/jobs/{id}/enable][%d] EnableJob default  %+v", o._statusCode, o.Payload)
}

func (o *EnableJobDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(job_model.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
