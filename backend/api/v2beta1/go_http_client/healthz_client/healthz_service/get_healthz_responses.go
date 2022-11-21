// Code generated by go-swagger; DO NOT EDIT.

package healthz_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	healthz_model "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/healthz_model"
)

// GetHealthzReader is a Reader for the GetHealthz structure.
type GetHealthzReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetHealthzReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetHealthzOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewGetHealthzDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetHealthzOK creates a GetHealthzOK with default headers values
func NewGetHealthzOK() *GetHealthzOK {
	return &GetHealthzOK{}
}

/*GetHealthzOK handles this case with default header values.

A successful response.
*/
type GetHealthzOK struct {
	Payload *healthz_model.APIGetHealthzResponse
}

func (o *GetHealthzOK) Error() string {
	return fmt.Sprintf("[GET /apis/v2beta1/healthz][%d] getHealthzOK  %+v", 200, o.Payload)
}

func (o *GetHealthzOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(healthz_model.APIGetHealthzResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetHealthzDefault creates a GetHealthzDefault with default headers values
func NewGetHealthzDefault(code int) *GetHealthzDefault {
	return &GetHealthzDefault{
		_statusCode: code,
	}
}

/*GetHealthzDefault handles this case with default header values.

GetHealthzDefault get healthz default
*/
type GetHealthzDefault struct {
	_statusCode int

	Payload *healthz_model.APIStatus
}

// Code gets the status code for the get healthz default response
func (o *GetHealthzDefault) Code() int {
	return o._statusCode
}

func (o *GetHealthzDefault) Error() string {
	return fmt.Sprintf("[GET /apis/v2beta1/healthz][%d] GetHealthz default  %+v", o._statusCode, o.Payload)
}

func (o *GetHealthzDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(healthz_model.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
