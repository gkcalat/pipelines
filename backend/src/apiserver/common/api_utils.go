// Copyright 2022 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"encoding/json"
	"fmt"
	"reflect"

	apiV1beta1 "github.com/kubeflow/pipelines/backend/api/v1beta1/go_client"
	apiV2beta1 "github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/kubeflow/pipelines/backend/src/common/util"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	structpb "google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// Generic Run interface for API v1 and v2
type ProtoRunInterface interface {
	Reset()
	String() string
	ProtoMessage()
	ProtoReflect() protoreflect.Message
	Descriptor() ([]byte, []int)
	GetDescription() string
	GetServiceAccount() string
	GetCreatedAt() *timestamppb.Timestamp
	GetScheduledAt() *timestamppb.Timestamp
	GetFinishedAt() *timestamppb.Timestamp
}

// Generic RuntimeConfig interface for API v1 and v2
type ProtoRuntimeConfigInterface interface {
	Reset()
	String() string
	ProtoMessage()
	ProtoReflect() protoreflect.Message
	Descriptor() ([]byte, []int)
	GetParameters() map[string]*structpb.Value
	GetPipelineRoot() string
}

// Fetches ResourceReferences from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetResourceReferenceFromRunInterface(r ProtoRunInterface) ([]*apiV1beta1.ResourceReference, error) {
	switch r.(type) {
	case *apiV1beta1.Run:
		return r.(*apiV1beta1.Run).GetResourceReferences(), nil
	case *apiV2beta1.Run:
		return nil, nil
	default:
		return nil, util.NewUnknownApiVersionError("GetResourceReferenceFromRunInterface()", fmt.Sprintf("ResourceReference from %T", reflect.TypeOf(r)))
	}
}

// Fetches a PipelineId from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetPipelineIdFromRunInterface(r ProtoRunInterface) (string, error) {
	switch runType := r.(type) {
	case *apiV1beta1.Run:
		return r.(*apiV1beta1.Run).GetPipelineSpec().GetPipelineId(), nil
	case *apiV2beta1.Run:
		pipelineId := r.(*apiV2beta1.Run).GetPipelineId()
		if pipelineId == "" {
			if pipelineIdValue, ok := r.(*apiV2beta1.Run).GetPipelineSpec().GetFields()["PipelineId"]; ok {
				pipelineId = pipelineIdValue.GetStringValue()
			} else {
				return "", util.NewResourceNotFoundError(fmt.Sprintf("PipelineId not found in %T. This could be because PipelineId is set to an empty string or invalid PipelineId field in PipelineSpec", runType), "")
			}
		}
		return pipelineId, nil
	default:
		return "", util.NewUnknownApiVersionError("GetPipelineIdFromRunInterface()", fmt.Sprintf("PipelineId from %T", reflect.TypeOf(r)))
	}
}

// Fetches a ExperimentId from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetExperimentIdFromRunInterface(r ProtoRunInterface) (string, error) {
	switch r.(type) {
	case *apiV1beta1.Run:
		return "", nil
	case *apiV2beta1.Run:
		return r.(*apiV2beta1.Run).GetExperimentId(), nil
	default:
		return "", util.NewUnknownApiVersionError("GetExperimentIdFromRunInterface()", fmt.Sprintf("ExperimentId from %T", reflect.TypeOf(r)))
	}
}

// Fetches a RuntimeState from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetStateFromRunInterface(r ProtoRunInterface) (string, error) {
	switch r.(type) {
	case *apiV1beta1.Run:
		return "", nil
	case *apiV2beta1.Run:
		return r.(*apiV2beta1.Run).GetState().String(), nil
	default:
		return "", util.NewUnknownApiVersionError("GetStateFromRunInterface()", fmt.Sprintf("State from %T", reflect.TypeOf(r)))
	}
}

// Fetches a RuntimeStateHistory from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetStateHistoryFromRunInterface(r ProtoRunInterface) (string, error) {
	switch r.(type) {
	case *apiV1beta1.Run:
		return "", nil
	case *apiV2beta1.Run:
		if serializedState, err := json.Marshal(r.(*apiV2beta1.Run).GetStateHistory()); err != nil {
			return "", err
		} else {
			return string(serializedState), err
		}
	default:
		return "", util.NewUnknownApiVersionError("GetStateHistoryFromRunInterface()", fmt.Sprintf("StateHistory from %T", reflect.TypeOf(r)))
	}
}

// Fetches a DisplayName from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetDisplayNameFromRunInterface(r ProtoRunInterface) (string, error) {
	switch r.(type) {
	case *apiV1beta1.Run:
		return r.(*apiV1beta1.Run).GetName(), nil
	case *apiV2beta1.Run:
		return r.(*apiV2beta1.Run).GetDisplayName(), nil
	default:
		return "", util.NewUnknownApiVersionError("GetDisplayNameFromRunInterface()", fmt.Sprintf("DisplayName from %T", reflect.TypeOf(r)))
	}
}

// // Converts any struct to a map[string] via json package
// func ConvertStructToMap(aStruct *interface{}) (aMap map[string]interface{}, anErr error) {
// 	if jsonBytes, anErr := json.Marshal(aStruct); anErr != nil {
// 		return nil, util.NewInvalidInputError("Error marshalling struct %T into a map[string]", aStruct)
// 	} else {
// 		anErr = json.Unmarshal(jsonBytes, &aMap)
// 		if anErr != nil {
// 			return nil, util.NewInvalidInputError("Error unmarshalling struct %T into a map[string]", aStruct)
// 		}
// 		return aMap, nil
// 	}
// 	return aMap, anErr
// }

// // Converts a slice of struct to a slice of map[string] via json package
// func ConvertSliceStructToMap(oldSliceStruct []*interface{}) (newSliceMap []map[string]interface{}, err error) {
// 	newSliceMap = make([]map[string]interface{}, len(oldSliceStruct))
// 	for i := range newSliceMap {
// 		newSliceMap[i], err = ConvertStructToMap(oldSliceStruct[i])
// 		if err != nil {
// 			return nil, util.NewInvalidInputError("Error converting struct %T into a map[string]", oldSliceStruct[i])
// 		}
// 	}
// 	return newSliceMap, err
// }

// Converts PipelineSpec (v1) into structpb.Struct (v2)
func ConvertPipelineSpecToProtoStruct(pipelineSpec *apiV1beta1.PipelineSpec) (*structpb.Struct, error) {
	mapValue := map[string]interface{}{
		"PipelineId":       pipelineSpec.GetPipelineId(),
		"PipelineName":     pipelineSpec.GetPipelineName(),
		"WorkflowManifest": pipelineSpec.GetWorkflowManifest(),
		"PipelineManifest": pipelineSpec.GetPipelineManifest(),
	}
	return structpb.NewStruct(mapValue)
}

// Fetches a PipelineSpec from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetPipelineSpecFromRunInterface(r ProtoRunInterface) (*structpb.Struct, error) {
	switch r.(type) {
	case *apiV1beta1.Run:
		return ConvertPipelineSpecToProtoStruct(r.(*apiV1beta1.Run).GetPipelineSpec())
	case *apiV2beta1.Run:
		return r.(*apiV2beta1.Run).GetPipelineSpec(), nil
	default:
		return nil, util.NewUnknownApiVersionError("GetPipelineSpecFromRunInterface()", fmt.Sprintf("PipelineSpec from %T", reflect.TypeOf(r)))
	}
}

// Fetches a PipelineRoot from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetPipelineRootFromRunInterface(r ProtoRunInterface) (string, error) {
	switch r.(type) {
	case *apiV1beta1.Run:
		return r.(*apiV1beta1.Run).GetPipelineSpec().GetRuntimeConfig().GetPipelineRoot(), nil
	case *apiV2beta1.Run:
		return r.(*apiV2beta1.Run).GetRuntimeConfig().GetPipelineRoot(), nil
	default:
		return "", util.NewUnknownApiVersionError("GetPipelineRootFromRunInterface()", fmt.Sprintf("PipelineRoot from %T", reflect.TypeOf(r)))
	}
}

// Fetches a RuntimeConfig from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetRuntimeConfigFromRunInterface(r ProtoRunInterface) (map[string]interface{}, error) {
	switch r.(type) {
	case *apiV1beta1.Run:
		// Fetch from parameters in V1 template
		var newParameters []map[string]*structpb.Value
		oldParameters := r.(*apiV1beta1.Run).GetPipelineSpec().GetParameters()
		if oldParameters == nil || len(oldParameters) == 0 {
			oldParams := r.(*apiV1beta1.Run).GetPipelineSpec().GetRuntimeConfig().GetParameters()
			for n, v := range oldParams {
				newParameters = append(
					newParameters,
					map[string]*structpb.Value{
						"Name":  structpb.NewStringValue(n),
						"Value": v,
					},
				)
			}
		} else {
			for i := range oldParameters {
				newParameters[i] = map[string]*structpb.Value{
					"Name":  structpb.NewStringValue(oldParameters[i].GetName()),
					"Value": structpb.NewStringValue(oldParameters[i].GetValue()),
				}
			}
		}
		// Convert RuntimeConfig
		newRuntimeConfig := map[string]interface{}{
			"Parameters":   newParameters,
			"PipelineRoot": r.(*apiV1beta1.Run).GetPipelineSpec().GetRuntimeConfig().GetPipelineRoot(),
		}
		return newRuntimeConfig, nil
	case *apiV2beta1.Run:
		newRuntimeConfig := map[string]interface{}{
			"Parameters":   r.(*apiV2beta1.Run).GetRuntimeConfig().GetParameters(),
			"PipelineRoot": r.(*apiV2beta1.Run).GetRuntimeConfig().GetPipelineRoot(),
		}
		return newRuntimeConfig, nil
	default:
		return nil, util.NewUnknownApiVersionError("GetRuntimeConfigFromRunInterface()", fmt.Sprintf("RuntimeConfig from %T", reflect.TypeOf(r)))
	}
}

// Fetches a RunDetails from a Run.
// This is not intended for validation.
// Raises error if an incompatible interface is used.
func GetRunDetailsFromRunInterface(r ProtoRunInterface) (string, error) {
	switch r.(type) {
	case *apiV1beta1.Run:
		return "", nil
	case *apiV2beta1.Run:
		return r.(*apiV2beta1.Run).GetRunDetails().String(), nil
	default:
		return "", util.NewUnknownApiVersionError("GetRunDetailsFromRunInterface()", fmt.Sprintf("RunDetails from %T", reflect.TypeOf(r)))
	}
}
