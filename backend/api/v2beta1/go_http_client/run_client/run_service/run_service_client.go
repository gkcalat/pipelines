// Code generated by go-swagger; DO NOT EDIT.

package run_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new run service API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for run service API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
ArchiveRunV1 archives a run
*/
func (a *Client) ArchiveRunV1(params *ArchiveRunV1Params, authInfo runtime.ClientAuthInfoWriter) (*ArchiveRunV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewArchiveRunV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "ArchiveRunV1",
		Method:             "POST",
		PathPattern:        "/apis/v2beta1/runs/{id}:archive",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ArchiveRunV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*ArchiveRunV1OK), nil

}

/*
CreateRunV1 creates a new run
*/
func (a *Client) CreateRunV1(params *CreateRunV1Params, authInfo runtime.ClientAuthInfoWriter) (*CreateRunV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateRunV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "CreateRunV1",
		Method:             "POST",
		PathPattern:        "/apis/v2beta1/runs",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CreateRunV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*CreateRunV1OK), nil

}

/*
DeleteRunV1 deletes a run
*/
func (a *Client) DeleteRunV1(params *DeleteRunV1Params, authInfo runtime.ClientAuthInfoWriter) (*DeleteRunV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteRunV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "DeleteRunV1",
		Method:             "DELETE",
		PathPattern:        "/apis/v2beta1/runs/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &DeleteRunV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*DeleteRunV1OK), nil

}

/*
GetRunV1 finds a specific run by ID
*/
func (a *Client) GetRunV1(params *GetRunV1Params, authInfo runtime.ClientAuthInfoWriter) (*GetRunV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetRunV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "GetRunV1",
		Method:             "GET",
		PathPattern:        "/apis/v2beta1/runs/{run_id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetRunV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*GetRunV1OK), nil

}

/*
ListRunsV1 finds all runs
*/
func (a *Client) ListRunsV1(params *ListRunsV1Params, authInfo runtime.ClientAuthInfoWriter) (*ListRunsV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListRunsV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "ListRunsV1",
		Method:             "GET",
		PathPattern:        "/apis/v2beta1/runs",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ListRunsV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*ListRunsV1OK), nil

}

/*
ReadArtifactV1 finds a run s artifact data
*/
func (a *Client) ReadArtifactV1(params *ReadArtifactV1Params, authInfo runtime.ClientAuthInfoWriter) (*ReadArtifactV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewReadArtifactV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "ReadArtifactV1",
		Method:             "GET",
		PathPattern:        "/apis/v2beta1/runs/{run_id}/nodes/{node_id}/artifacts/{artifact_name}:read",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ReadArtifactV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*ReadArtifactV1OK), nil

}

/*
ReportRunMetricsV1 reports run metrics reports metrics of a run each metric is reported in its own transaction so this API accepts partial failures metric can be uniquely identified by run id node id name duplicate reporting will be ignored by the API first reporting wins
*/
func (a *Client) ReportRunMetricsV1(params *ReportRunMetricsV1Params, authInfo runtime.ClientAuthInfoWriter) (*ReportRunMetricsV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewReportRunMetricsV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "ReportRunMetricsV1",
		Method:             "POST",
		PathPattern:        "/apis/v2beta1/runs/{run_id}:reportMetrics",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ReportRunMetricsV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*ReportRunMetricsV1OK), nil

}

/*
RetryRunV1 res initiates a failed or terminated run
*/
func (a *Client) RetryRunV1(params *RetryRunV1Params, authInfo runtime.ClientAuthInfoWriter) (*RetryRunV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewRetryRunV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "RetryRunV1",
		Method:             "POST",
		PathPattern:        "/apis/v2beta1/runs/{run_id}/retry",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &RetryRunV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*RetryRunV1OK), nil

}

/*
TerminateRunV1 terminates an active run
*/
func (a *Client) TerminateRunV1(params *TerminateRunV1Params, authInfo runtime.ClientAuthInfoWriter) (*TerminateRunV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewTerminateRunV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "TerminateRunV1",
		Method:             "POST",
		PathPattern:        "/apis/v2beta1/runs/{run_id}/terminate",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &TerminateRunV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*TerminateRunV1OK), nil

}

/*
UnarchiveRunV1 restores an archived run
*/
func (a *Client) UnarchiveRunV1(params *UnarchiveRunV1Params, authInfo runtime.ClientAuthInfoWriter) (*UnarchiveRunV1OK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUnarchiveRunV1Params()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "UnarchiveRunV1",
		Method:             "POST",
		PathPattern:        "/apis/v2beta1/runs/{id}:unarchive",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &UnarchiveRunV1Reader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*UnarchiveRunV1OK), nil

}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
