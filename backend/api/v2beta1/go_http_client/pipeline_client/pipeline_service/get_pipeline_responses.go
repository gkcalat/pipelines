// Code generated by go-swagger; DO NOT EDIT.

package pipeline_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	pipeline_model "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_model"
)

// GetPipelineReader is a Reader for the GetPipeline structure.
type GetPipelineReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetPipelineReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetPipelineOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewGetPipelineDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetPipelineOK creates a GetPipelineOK with default headers values
func NewGetPipelineOK() *GetPipelineOK {
	return &GetPipelineOK{}
}

/*GetPipelineOK handles this case with default header values.

A successful response.
*/
type GetPipelineOK struct {
	Payload *pipeline_model.BackendPipeline
}

func (o *GetPipelineOK) Error() string {
	return fmt.Sprintf("[GET /apis/v2beta1/pipelines/{pipeline_id}][%d] getPipelineOK  %+v", 200, o.Payload)
}

func (o *GetPipelineOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(pipeline_model.BackendPipeline)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPipelineDefault creates a GetPipelineDefault with default headers values
func NewGetPipelineDefault(code int) *GetPipelineDefault {
	return &GetPipelineDefault{
		_statusCode: code,
	}
}

/*GetPipelineDefault handles this case with default header values.

GetPipelineDefault get pipeline default
*/
type GetPipelineDefault struct {
	_statusCode int

	Payload *pipeline_model.BackendStatus
}

// Code gets the status code for the get pipeline default response
func (o *GetPipelineDefault) Code() int {
	return o._statusCode
}

func (o *GetPipelineDefault) Error() string {
	return fmt.Sprintf("[GET /apis/v2beta1/pipelines/{pipeline_id}][%d] GetPipeline default  %+v", o._statusCode, o.Payload)
}

func (o *GetPipelineDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(pipeline_model.BackendStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
