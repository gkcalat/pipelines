// Code generated by go-swagger; DO NOT EDIT.

package visualization_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	visualization_model "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/visualization_model"
)

// CreateVisualizationV1Reader is a Reader for the CreateVisualizationV1 structure.
type CreateVisualizationV1Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateVisualizationV1Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewCreateVisualizationV1OK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewCreateVisualizationV1Default(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewCreateVisualizationV1OK creates a CreateVisualizationV1OK with default headers values
func NewCreateVisualizationV1OK() *CreateVisualizationV1OK {
	return &CreateVisualizationV1OK{}
}

/*CreateVisualizationV1OK handles this case with default header values.

A successful response.
*/
type CreateVisualizationV1OK struct {
	Payload *visualization_model.APIVisualization
}

func (o *CreateVisualizationV1OK) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/visualizations/{namespace}][%d] createVisualizationV1OK  %+v", 200, o.Payload)
}

func (o *CreateVisualizationV1OK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(visualization_model.APIVisualization)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateVisualizationV1Default creates a CreateVisualizationV1Default with default headers values
func NewCreateVisualizationV1Default(code int) *CreateVisualizationV1Default {
	return &CreateVisualizationV1Default{
		_statusCode: code,
	}
}

/*CreateVisualizationV1Default handles this case with default header values.

CreateVisualizationV1Default create visualization v1 default
*/
type CreateVisualizationV1Default struct {
	_statusCode int

	Payload *visualization_model.APIStatus
}

// Code gets the status code for the create visualization v1 default response
func (o *CreateVisualizationV1Default) Code() int {
	return o._statusCode
}

func (o *CreateVisualizationV1Default) Error() string {
	return fmt.Sprintf("[POST /apis/v2beta1/visualizations/{namespace}][%d] CreateVisualizationV1 default  %+v", o._statusCode, o.Payload)
}

func (o *CreateVisualizationV1Default) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(visualization_model.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
