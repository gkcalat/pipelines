// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	apiv1beta1 "github.com/kubeflow/pipelines/backend/api/v1beta1/go_client"
	apiv2beta1 "github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/kubeflow/pipelines/backend/src/apiserver/client"
	"github.com/kubeflow/pipelines/backend/src/apiserver/common"
	"github.com/kubeflow/pipelines/backend/src/apiserver/resource"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	commonApiJob = &apiv1beta1.Job{
		Name:           "job1",
		Enabled:        true,
		MaxConcurrency: 1,
		Trigger: &apiv1beta1.Trigger{
			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		PipelineSpec: &apiv1beta1.PipelineSpec{
			WorkflowManifest: testWorkflow.ToStringForStore(),
			Parameters:       []*apiv1beta1.Parameter{{Name: "param1", Value: "world"}},
		},
		ResourceReferences: []*apiv1beta1.ResourceReference{
			{
				Key:          &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_EXPERIMENT, Id: "123e4567-e89b-12d3-a456-426655440000"},
				Relationship: apiv1beta1.Relationship_OWNER,
			},
		},
	}

	commonExpectedJob = &apiv1beta1.Job{
		Id:             "123e4567-e89b-12d3-a456-426655440000",
		Name:           "job1",
		ServiceAccount: "pipeline-runner",
		Enabled:        true,
		MaxConcurrency: 1,
		Trigger: &apiv1beta1.Trigger{
			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		CreatedAt: &timestamp.Timestamp{Seconds: 2},
		UpdatedAt: &timestamp.Timestamp{Seconds: 2},
		Status:    "STATUS_UNSPECIFIED",
		PipelineSpec: &apiv1beta1.PipelineSpec{
			WorkflowManifest: testWorkflow.ToStringForStore(),
			Parameters:       []*apiv1beta1.Parameter{{Name: "param1", Value: "world"}},
		},
		ResourceReferences: []*apiv1beta1.ResourceReference{
			{
				Key:          &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_EXPERIMENT, Id: "123e4567-e89b-12d3-a456-426655440000"},
				Relationship: apiv1beta1.Relationship_OWNER,
			},
		},
	}

	commonApiRecurringRun = &apiv2beta1.RecurringRun{
		DisplayName:    "job1",
		Mode:           apiv2beta1.RecurringRun_ENABLE,
		MaxConcurrency: 1,
		Trigger: &apiv2beta1.Trigger{
			Trigger: &apiv2beta1.Trigger_CronSchedule{CronSchedule: &apiv2beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		PipelineSource: &apiv2beta1.RecurringRun_PipelineSpec{PipelineSpec: &structpb.Struct{}},
		ExperimentId:   "123e4567-e89b-12d3-a456-426655440000",
	}
)

// func TestValidateApiJob(t *testing.T) {
// 	clients, manager, _ := initWithExperiment(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: commonApiJob})
// 	assert.Nil(t, err)
// }

// func TestValidateApiJob_WithPipelineVersion(t *testing.T) {
// 	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	apiJob := &apiv1beta1.Job{
// 		Name:           "job1",
// 		Enabled:        true,
// 		MaxConcurrency: 1,
// 		Trigger: &apiv1beta1.Trigger{
// 			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
// 				StartTime: &timestamp.Timestamp{Seconds: 1},
// 				Cron:      "1 * * * *",
// 			}}},
// 		ResourceReferences: validReferencesOfExperimentAndPipelineVersion,
// 	}
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: apiJob})
// 	assert.Nil(t, err)
// }

// func TestValidateApiJob_ValidateNoExperimentResourceReferenceSucceeds(t *testing.T) {
// 	clients, manager, _ := initWithExperiment(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	apiJob := &apiv1beta1.Job{
// 		Name:           "job1",
// 		Enabled:        true,
// 		MaxConcurrency: 1,
// 		Trigger: &apiv1beta1.Trigger{
// 			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
// 				StartTime: &timestamp.Timestamp{Seconds: 1},
// 				Cron:      "1 * * * *",
// 			}}},
// 		PipelineSpec: &apiv1beta1.PipelineSpec{
// 			WorkflowManifest: testWorkflow.ToStringForStore(),
// 			Parameters:       []*apiv1beta1.Parameter{{Name: "param1", Value: "world"}},
// 		},
// 		// This job has no ResourceReferences, no experiment
// 	}
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: apiJob})
// 	assert.Nil(t, err)
// }

// func TestValidateApiJob_WithInvalidPipelineVersionReference(t *testing.T) {
// 	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	apiJob := &apiv1beta1.Job{
// 		Name:           "job1",
// 		Enabled:        true,
// 		MaxConcurrency: 1,
// 		Trigger: &apiv1beta1.Trigger{
// 			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
// 				StartTime: &timestamp.Timestamp{Seconds: 1},
// 				Cron:      "1 * * * *",
// 			}}},
// 		ResourceReferences: referencesOfExperimentAndInvalidPipelineVersion,
// 	}
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: apiJob})
// 	assert.Equal(t, codes.NotFound, err.(*util.UserError).ExternalStatusCode())
// 	assert.NotNil(t, err)
// 	assert.Contains(t, err.Error(), "Get pipelineVersionId failed")
// }

// func TestValidateApiJob_NoValidPipelineSpecOrPipelineVersion(t *testing.T) {
// 	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	apiJob := &apiv1beta1.Job{
// 		Name:           "job1",
// 		Enabled:        true,
// 		MaxConcurrency: 1,
// 		Trigger: &apiv1beta1.Trigger{
// 			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
// 				StartTime: &timestamp.Timestamp{Seconds: 1},
// 				Cron:      "1 * * * *",
// 			}}},
// 		ResourceReferences: validReference,
// 	}
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: apiJob})
// 	assert.Equal(t, codes.InvalidArgument, err.(*util.UserError).ExternalStatusCode())
// 	assert.NotNil(t, err)
// 	assert.Contains(t, err.Error(), "Please specify a pipeline by providing a (workflow manifest or pipeline manifest) or (pipeline id or/and pipeline version)")
// }

// func TestValidateApiJob_WorkflowManifestAndPipelineVersion(t *testing.T) {
// 	clients, manager, _ := initWithExperimentAndPipelineVersion(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	apiJob := &apiv1beta1.Job{
// 		Name:           "job1",
// 		Enabled:        true,
// 		MaxConcurrency: 1,
// 		Trigger: &apiv1beta1.Trigger{
// 			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
// 				StartTime: &timestamp.Timestamp{Seconds: 1},
// 				Cron:      "1 * * * *",
// 			}}},
// 		PipelineSpec: &apiv1beta1.PipelineSpec{
// 			WorkflowManifest: testWorkflow.ToStringForStore(),
// 			Parameters:       []*apiv1beta1.Parameter{{Name: "param2", Value: "world"}},
// 		},
// 		ResourceReferences: validReferencesOfExperimentAndPipelineVersion,
// 	}
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: apiJob})
// 	assert.Equal(t, codes.InvalidArgument, err.(*util.UserError).ExternalStatusCode())
// 	assert.NotNil(t, err)
// 	assert.Contains(t, err.Error(), "Please don't specify a pipeline version or pipeline ID when you specify a workflow manifest or pipeline manifest")
// }

// func TestValidateApiJob_ValidatePipelineSpecFailed(t *testing.T) {
// 	clients, manager, experiment := initWithExperiment(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	apiJob := &apiv1beta1.Job{
// 		Name:           "job1",
// 		Enabled:        true,
// 		MaxConcurrency: 1,
// 		Trigger: &apiv1beta1.Trigger{
// 			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
// 				StartTime: &timestamp.Timestamp{Seconds: 1},
// 				Cron:      "1 * * * *",
// 			}}},
// 		PipelineSpec: &apiv1beta1.PipelineSpec{
// 			PipelineId: "not_exist_pipeline",
// 			Parameters: []*apiv1beta1.Parameter{{Name: "param2", Value: "world"}},
// 		},
// 		ResourceReferences: []*apiv1beta1.ResourceReference{
// 			{Key: &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_EXPERIMENT, Id: experiment.UUID}, Relationship: apiv1beta1.Relationship_OWNER},
// 		},
// 	}
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: apiJob})
// 	assert.Equal(t, codes.NotFound, err.(*util.UserError).ExternalStatusCode())
// 	assert.NotNil(t, err)
// 	assert.Contains(t, err.Error(), "Pipeline not_exist_pipeline not found")
// }

// func TestValidateApiJob_InvalidCron(t *testing.T) {
// 	clients, manager, experiment := initWithExperiment(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	apiJob := &apiv1beta1.Job{
// 		Name:           "job1",
// 		Enabled:        true,
// 		MaxConcurrency: 1,
// 		Trigger: &apiv1beta1.Trigger{
// 			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
// 				StartTime: &timestamp.Timestamp{Seconds: 1},
// 				Cron:      "1 * * ",
// 			}}},
// 		PipelineSpec: &apiv1beta1.PipelineSpec{
// 			WorkflowManifest: testWorkflow.ToStringForStore(),
// 			Parameters:       []*apiv1beta1.Parameter{{Name: "param1", Value: "world"}},
// 		},
// 		ResourceReferences: []*apiv1beta1.ResourceReference{
// 			{Key: &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_EXPERIMENT, Id: experiment.UUID}, Relationship: apiv1beta1.Relationship_OWNER},
// 		},
// 	}
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: apiJob})
// 	assert.Equal(t, codes.InvalidArgument, err.(*util.UserError).ExternalStatusCode())
// 	assert.Contains(t, err.Error(), "Schedule cron is not a supported format")
// }

// func TestValidateApiJob_MaxConcurrencyOutOfRange(t *testing.T) {
// 	clients, manager, experiment := initWithExperiment(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	apiJob := &apiv1beta1.Job{
// 		Name:           "job1",
// 		Enabled:        true,
// 		MaxConcurrency: 0,
// 		Trigger: &apiv1beta1.Trigger{
// 			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
// 				StartTime: &timestamp.Timestamp{Seconds: 1},
// 				Cron:      "1 * * * *",
// 			}}},
// 		PipelineSpec: &apiv1beta1.PipelineSpec{
// 			WorkflowManifest: testWorkflow.ToStringForStore(),
// 			Parameters:       []*apiv1beta1.Parameter{{Name: "param1", Value: "world"}},
// 		},
// 		ResourceReferences: []*apiv1beta1.ResourceReference{
// 			{Key: &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_EXPERIMENT, Id: experiment.UUID}, Relationship: apiv1beta1.Relationship_OWNER},
// 		},
// 	}
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: apiJob})
// 	assert.Equal(t, codes.InvalidArgument, err.(*util.UserError).ExternalStatusCode())
// 	assert.Contains(t, err.Error(), "max concurrency of the job is out of range")
// }

// func TestValidateApiJob_NegativeIntervalSecond(t *testing.T) {
// 	clients, manager, experiment := initWithExperiment(t)
// 	defer clients.Close()
// 	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
// 	apiJob := &apiv1beta1.Job{
// 		Name:           "job1",
// 		Enabled:        true,
// 		MaxConcurrency: 0,
// 		Trigger: &apiv1beta1.Trigger{
// 			Trigger: &apiv1beta1.Trigger_PeriodicSchedule{PeriodicSchedule: &apiv1beta1.PeriodicSchedule{
// 				IntervalSecond: -1,
// 			}}},
// 		PipelineSpec: &apiv1beta1.PipelineSpec{
// 			WorkflowManifest: testWorkflow.ToStringForStore(),
// 			Parameters:       []*apiv1beta1.Parameter{{Name: "param1", Value: "world"}},
// 		},
// 		ResourceReferences: []*apiv1beta1.ResourceReference{
// 			{Key: &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_EXPERIMENT, Id: experiment.UUID}, Relationship: apiv1beta1.Relationship_OWNER},
// 		},
// 	}
// 	err := server.validateCreateJobRequest(&apiv1beta1.CreateJobRequest{Job: apiJob})
// 	assert.Equal(t, codes.InvalidArgument, err.(*util.UserError).ExternalStatusCode())
// 	assert.Contains(t, err.Error(), "The max concurrency of the job is out of range")
// }

func TestCreateJob(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
	job, err := server.CreateJob(nil, &apiv1beta1.CreateJobRequest{Job: commonApiJob})
	assert.Nil(t, err)
	matched := 0
	for _, resRef := range commonExpectedJob.GetResourceReferences() {
		for _, resRef2 := range job.GetResourceReferences() {
			if resRef.Key.Type == resRef2.Key.Type && resRef.Key.Id == resRef2.Key.Id && resRef.Relationship == resRef2.Relationship {
				matched++
			}
		}
	}
	assert.Equal(t, len(commonExpectedJob.GetResourceReferences()), matched)
	commonExpectedJob.ResourceReferences = job.GetResourceReferences()

	commonExpectedJob.PipelineSpec.PipelineId = job.GetPipelineSpec().GetPipelineId()
	commonExpectedJob.PipelineSpec.PipelineName = job.GetPipelineSpec().GetPipelineName()
	commonExpectedJob.PipelineSpec.PipelineManifest = job.GetPipelineSpec().GetPipelineManifest()
	commonExpectedJob.CreatedAt = job.GetCreatedAt()
	commonExpectedJob.UpdatedAt = job.GetUpdatedAt()
	assert.Equal(t, commonExpectedJob, job)
}

func TestCreateJob_V2(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	listParams := []interface{}{1, 2, 3}
	v2RuntimeListParams, _ := structpb.NewList(listParams)
	structParams := map[string]interface{}{"structParam1": "hello", "structParam2": 32}
	v2RuntimeStructParams, _ := structpb.NewStruct(structParams)

	// Test all parameters types converted to model.RuntimeConfig.Parameters, which is string type
	v2RuntimeParams := map[string]*structpb.Value{
		"param1": {Kind: &structpb.Value_StringValue{StringValue: "world"}},
		"param2": {Kind: &structpb.Value_BoolValue{BoolValue: true}},
		"param3": {Kind: &structpb.Value_ListValue{ListValue: v2RuntimeListParams}},
		"param4": {Kind: &structpb.Value_NumberValue{NumberValue: 12}},
		"param5": {Kind: &structpb.Value_StructValue{StructValue: v2RuntimeStructParams}},
	}

	apiJob_V2 := &apiv1beta1.Job{
		Name:           "job1",
		Enabled:        true,
		MaxConcurrency: 1,
		Trigger: &apiv1beta1.Trigger{
			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		PipelineSpec: &apiv1beta1.PipelineSpec{
			PipelineManifest: v2SpecHelloWorld,
			RuntimeConfig: &apiv1beta1.PipelineSpec_RuntimeConfig{
				Parameters:   v2RuntimeParams,
				PipelineRoot: "model-pipeline-root",
			},
		},
		ResourceReferences: []*apiv1beta1.ResourceReference{
			{
				Key:          &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_EXPERIMENT, Id: "123e4567-e89b-12d3-a456-426655440000"},
				Relationship: apiv1beta1.Relationship_OWNER,
			},
		},
	}

	expectedJob_V2 := &apiv1beta1.Job{
		Id:             "123e4567-e89b-12d3-a456-426655440000",
		Name:           "job1",
		ServiceAccount: "pipeline-runner",
		Enabled:        true,
		MaxConcurrency: 1,
		Trigger: &apiv1beta1.Trigger{
			Trigger: &apiv1beta1.Trigger_CronSchedule{CronSchedule: &apiv1beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		CreatedAt: &timestamp.Timestamp{Seconds: 2},
		UpdatedAt: &timestamp.Timestamp{Seconds: 2},
		Status:    "STATUS_UNSPECIFIED",
		PipelineSpec: &apiv1beta1.PipelineSpec{
			PipelineManifest: v2SpecHelloWorld,
			RuntimeConfig: &apiv1beta1.PipelineSpec_RuntimeConfig{
				Parameters:   v2RuntimeParams,
				PipelineRoot: "model-pipeline-root",
			},
		},
		ResourceReferences: []*apiv1beta1.ResourceReference{
			{
				Key:  &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_EXPERIMENT, Id: "123e4567-e89b-12d3-a456-426655440000"},
				Name: "exp1", Relationship: apiv1beta1.Relationship_OWNER,
			},
		},
	}
	job, err := server.CreateJob(nil, &apiv1beta1.CreateJobRequest{Job: apiJob_V2})
	assert.Nil(t, err)

	matched := 0
	for _, resRef := range expectedJob_V2.GetResourceReferences() {
		for _, resRef2 := range job.GetResourceReferences() {
			if resRef.Key.Type == resRef2.Key.Type && resRef.Key.Id == resRef2.Key.Id && resRef.Relationship == resRef2.Relationship {
				matched++
			}
		}
	}
	assert.Equal(t, len(expectedJob_V2.GetResourceReferences()), matched)
	expectedJob_V2.ResourceReferences = job.GetResourceReferences()

	expectedJob_V2.PipelineSpec.PipelineId = job.GetPipelineSpec().GetPipelineId()
	expectedJob_V2.PipelineSpec.PipelineName = job.GetPipelineSpec().GetPipelineName()
	expectedJob_V2.PipelineSpec.PipelineManifest = job.GetPipelineSpec().GetPipelineManifest()
	expectedJob_V2.CreatedAt = job.GetCreatedAt()
	expectedJob_V2.UpdatedAt = job.GetUpdatedAt()

	assert.Equal(t, expectedJob_V2, job)
}

func TestCreateJob_Unauthorized(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	userIdentity := "user@google.com"
	md := metadata.New(map[string]string{common.GoogleIAPUserIdentityHeader: common.GoogleIAPUserIdentityPrefix + userIdentity})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, _ := initWithExperiment_SubjectAccessReview_Unauthorized(t)
	defer clients.Close()

	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
	_, err := server.CreateJob(ctx, &apiv1beta1.CreateJobRequest{Job: commonApiJob})
	assert.NotNil(t, err)
	assert.Contains(
		t,
		err.Error(),
		"PermissionDenied: User 'user@google.com' is not authorized with reason",
	)
}

func TestGetJob_Unauthorized(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	userIdentity := "user@google.com"
	md := metadata.New(map[string]string{common.GoogleIAPUserIdentityHeader: common.GoogleIAPUserIdentityPrefix + userIdentity})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()

	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
	job, err := server.CreateJob(ctx, &apiv1beta1.CreateJobRequest{Job: commonApiJob})
	assert.Nil(t, err)

	clients.SubjectAccessReviewClientFake = client.NewFakeSubjectAccessReviewClientUnauthorized()
	manager = resource.NewResourceManager(clients, map[string]interface{}{"DefaultNamespace": "default", "ApiVersion": "v2beta1"})
	server = NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	_, err = server.GetJob(ctx, &apiv1beta1.GetJobRequest{Id: job.Id})
	assert.NotNil(t, err)
	assert.Contains(
		t,
		err.Error(),
		"PermissionDenied: User 'user@google.com' is not authorized with reason",
	)
}

func TestGetJob_Multiuser(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	md := metadata.New(map[string]string{common.GoogleIAPUserIdentityHeader: common.GoogleIAPUserIdentityPrefix + "user@google.com"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
	createdJob, err := server.CreateJob(ctx, &apiv1beta1.CreateJobRequest{Job: commonApiJob})
	assert.Nil(t, err)

	job, err := server.GetJob(ctx, &apiv1beta1.GetJobRequest{Id: createdJob.Id})
	assert.Nil(t, err)
	matched := 0
	for _, resRef := range commonExpectedJob.GetResourceReferences() {
		for _, resRef2 := range job.GetResourceReferences() {
			if resRef.Key.Type == resRef2.Key.Type && resRef.Key.Id == resRef2.Key.Id && resRef.Relationship == resRef2.Relationship {
				matched++
			}
		}
	}
	assert.Equal(t, len(commonExpectedJob.GetResourceReferences()), matched)
	commonExpectedJob.ResourceReferences = job.GetResourceReferences()

	commonExpectedJob.PipelineSpec.PipelineId = job.GetPipelineSpec().GetPipelineId()
	commonExpectedJob.PipelineSpec.PipelineName = job.GetPipelineSpec().GetPipelineName()
	commonExpectedJob.PipelineSpec.PipelineManifest = job.GetPipelineSpec().GetPipelineManifest()
	commonExpectedJob.CreatedAt = job.GetCreatedAt()
	commonExpectedJob.UpdatedAt = job.GetUpdatedAt()
	assert.Equal(t, commonExpectedJob, job)
}

func TestListJobs_Unauthorized(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	userIdentity := "user@google.com"
	md := metadata.New(map[string]string{common.GoogleIAPUserIdentityHeader: common.GoogleIAPUserIdentityPrefix + userIdentity})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, experiment := initWithExperiment_SubjectAccessReview_Unauthorized(t)
	defer clients.Close()

	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
	_, err := server.ListJobs(ctx, &apiv1beta1.ListJobsRequest{
		ResourceReferenceKey: &apiv1beta1.ResourceKey{
			Type: apiv1beta1.ResourceType_EXPERIMENT,
			Id:   experiment.UUID,
		},
	})
	assert.NotNil(t, err)
	assert.Contains(
		t,
		err.Error(),
		"PermissionDenied: User 'user@google.com' is not authorized with reason",
	)

	_, err = server.ListJobs(ctx, &apiv1beta1.ListJobsRequest{
		ResourceReferenceKey: &apiv1beta1.ResourceKey{
			Type: apiv1beta1.ResourceType_NAMESPACE,
			Id:   "ns1",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(
		t,
		err.Error(),
		"PermissionDenied: User 'user@google.com' is not authorized with reason",
	)
}

func TestListJobs_Multiuser(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	md := metadata.New(map[string]string{common.GoogleIAPUserIdentityHeader: common.GoogleIAPUserIdentityPrefix + "user@google.com"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
	_, err := server.CreateJob(ctx, &apiv1beta1.CreateJobRequest{Job: commonApiJob})
	assert.Nil(t, err)

	var expectedJobs []*apiv1beta1.Job
	commonExpectedJob.PipelineSpec.PipelineId = "123e4567-e89b-12d3-a456-426655440000"
	commonExpectedJob.PipelineSpec.PipelineName = "job1"
	commonExpectedJob.PipelineSpec.PipelineManifest = commonExpectedJob.PipelineSpec.WorkflowManifest
	commonExpectedJob.CreatedAt = &timestamp.Timestamp{Seconds: 4}
	commonExpectedJob.UpdatedAt = &timestamp.Timestamp{Seconds: 4}
	commonExpectedJob.ResourceReferences = []*apiv1beta1.ResourceReference{
		{Key: &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_NAMESPACE, Id: "ns1"}, Relationship: apiv1beta1.Relationship_OWNER},
		{Key: &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_EXPERIMENT, Id: DefaultFakePipelineId}, Relationship: apiv1beta1.Relationship_OWNER},
		{Key: &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_PIPELINE, Id: DefaultFakePipelineId}, Relationship: apiv1beta1.Relationship_CREATOR},
		{Key: &apiv1beta1.ResourceKey{Type: apiv1beta1.ResourceType_PIPELINE_VERSION, Id: DefaultFakePipelineId}, Relationship: apiv1beta1.Relationship_CREATOR},
	}
	expectedJobs = append(expectedJobs, commonExpectedJob)
	expectedJobsEmpty := []*apiv1beta1.Job{}

	tests := []struct {
		name         string
		request      *apiv1beta1.ListJobsRequest
		wantError    bool
		errorMessage string
		expectedJobs []*apiv1beta1.Job
	}{
		{
			"Valid - filter by experiment",
			&apiv1beta1.ListJobsRequest{
				ResourceReferenceKey: &apiv1beta1.ResourceKey{
					Type: apiv1beta1.ResourceType_EXPERIMENT,
					Id:   "123e4567-e89b-12d3-a456-426655440000",
				},
			},
			false,
			"",
			expectedJobs,
		},
		{
			"Valid - filter by namespace",
			&apiv1beta1.ListJobsRequest{
				ResourceReferenceKey: &apiv1beta1.ResourceKey{
					Type: apiv1beta1.ResourceType_NAMESPACE,
					Id:   "ns1",
				},
			},
			false,
			"",
			expectedJobs,
		},
		{
			"Vailid - filter by namespace - no result",
			&apiv1beta1.ListJobsRequest{
				ResourceReferenceKey: &apiv1beta1.ResourceKey{
					Type: apiv1beta1.ResourceType_NAMESPACE,
					Id:   "no-such-ns",
				},
			},
			false,
			"",
			expectedJobsEmpty,
		},
		{
			"Valid - no filter",
			&apiv1beta1.ListJobsRequest{},
			false,
			"",
			expectedJobsEmpty,
		},
		{
			"Inalid - invalid filter type",
			&apiv1beta1.ListJobsRequest{
				ResourceReferenceKey: &apiv1beta1.ResourceKey{
					Type: apiv1beta1.ResourceType_UNKNOWN_RESOURCE_TYPE,
					Id:   "unknown",
				},
			},
			true,
			"Unrecognized resource reference type",
			nil,
		},
	}

	for _, tc := range tests {
		response, err := server.ListJobs(ctx, tc.request)

		if tc.wantError {
			if err == nil {
				t.Errorf("TestListJobs_Multiuser(%v) expect error but got nil", tc.name)
			} else if !strings.Contains(err.Error(), tc.errorMessage) {
				t.Errorf("TestListJobs_Multiusert(%v) expect error containing: %v, but got: %v", tc.name, tc.errorMessage, err)
			}
		} else {
			if err != nil {
				t.Errorf("TestListJobs_Multiuser(%v) expect no error but got %v", tc.name, err)
			} else if !cmp.Equal(tc.expectedJobs, response.Jobs, cmpopts.EquateEmpty(), protocmp.Transform(), cmpopts.IgnoreFields(apiv1beta1.Job{}, "Trigger", "UpdatedAt", "CreatedAt"),
				cmpopts.IgnoreFields(apiv1beta1.Run{}, "CreatedAt")) {
				t.Errorf("TestListJobs_Multiuser(%v) expect (%+v) but got (%+v)", tc.name, tc.expectedJobs, response.Jobs)
			}
		}
	}
}

func TestEnableJob_Unauthorized(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	userIdentity := "user@google.com"
	md := metadata.New(map[string]string{common.GoogleIAPUserIdentityHeader: common.GoogleIAPUserIdentityPrefix + userIdentity})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
	job, err := server.CreateJob(ctx, &apiv1beta1.CreateJobRequest{Job: commonApiJob})
	assert.Nil(t, err)

	clients.SubjectAccessReviewClientFake = client.NewFakeSubjectAccessReviewClientUnauthorized()
	manager = resource.NewResourceManager(clients, map[string]interface{}{"DefaultNamespace": "default", "ApiVersion": "v2beta1"})
	server = NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	_, err = server.EnableJob(ctx, &apiv1beta1.EnableJobRequest{Id: job.Id})
	assert.NotNil(t, err)
	assert.Contains(
		t,
		err.Error(),
		" PermissionDenied: User 'user@google.com' is not authorized with reason",
	)
}

func TestEnableJob_Multiuser(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	md := metadata.New(map[string]string{common.GoogleIAPUserIdentityHeader: common.GoogleIAPUserIdentityPrefix + "user@google.com"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	job, err := server.CreateJob(ctx, &apiv1beta1.CreateJobRequest{Job: commonApiJob})
	assert.Nil(t, err)

	_, err = server.EnableJob(ctx, &apiv1beta1.EnableJobRequest{Id: job.Id})
	assert.Nil(t, err)
}

func TestDisableJob_Unauthorized(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	userIdentity := "user@google.com"
	md := metadata.New(map[string]string{common.GoogleIAPUserIdentityHeader: common.GoogleIAPUserIdentityPrefix + userIdentity})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
	job, err := server.CreateJob(ctx, &apiv1beta1.CreateJobRequest{Job: commonApiJob})
	assert.Nil(t, err)

	clients.SubjectAccessReviewClientFake = client.NewFakeSubjectAccessReviewClientUnauthorized()
	manager = resource.NewResourceManager(clients, map[string]interface{}{"DefaultNamespace": "default", "ApiVersion": "v2beta1"})
	server = NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	_, err = server.DisableJob(ctx, &apiv1beta1.DisableJobRequest{Id: job.Id})
	assert.NotNil(t, err)
	assert.Contains(
		t,
		err.Error(),
		" PermissionDenied: User 'user@google.com' is not authorized with reason",
	)
}

func TestDisableJob_Multiuser(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	md := metadata.New(map[string]string{common.GoogleIAPUserIdentityHeader: common.GoogleIAPUserIdentityPrefix + "user@google.com"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	job, err := server.CreateJob(ctx, &apiv1beta1.CreateJobRequest{Job: commonApiJob})
	assert.Nil(t, err)

	_, err = server.DisableJob(ctx, &apiv1beta1.DisableJobRequest{Id: job.Id})
	assert.Nil(t, err)
}

func TestListJobs_Unauthenticated(t *testing.T) {
	viper.Set(common.MultiUserMode, "true")
	defer viper.Set(common.MultiUserMode, "false")

	md := metadata.New(map[string]string{"no-identity-header": "user"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	clients, manager, experiment := initWithExperiment(t)
	defer clients.Close()

	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})
	_, err := server.ListJobs(ctx, &apiv1beta1.ListJobsRequest{
		ResourceReferenceKey: &apiv1beta1.ResourceKey{
			Type: apiv1beta1.ResourceType_EXPERIMENT,
			Id:   experiment.UUID,
		},
	})
	assert.NotNil(t, err)
	assert.Contains(
		t,
		err.Error(),
		"User identity is empty in the request header",
	)

	_, err = server.ListJobs(ctx, &apiv1beta1.ListJobsRequest{
		ResourceReferenceKey: &apiv1beta1.ResourceKey{
			Type: apiv1beta1.ResourceType_NAMESPACE,
			Id:   "ns1",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(
		t,
		err.Error(),
		"User identity is empty in the request header",
	)
}

func TestCreateRecurringRun(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	pipelineSpecStruct := &structpb.Struct{}
	yaml.Unmarshal([]byte(v2SpecHelloWorld), pipelineSpecStruct)

	apiRecurringRun := &apiv2beta1.RecurringRun{
		DisplayName:    "recurring_run_1",
		Mode:           apiv2beta1.RecurringRun_ENABLE,
		MaxConcurrency: 1,
		Trigger: &apiv2beta1.Trigger{
			Trigger: &apiv2beta1.Trigger_CronSchedule{CronSchedule: &apiv2beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		PipelineSource: &apiv2beta1.RecurringRun_PipelineSpec{PipelineSpec: pipelineSpecStruct},
		RuntimeConfig: &apiv2beta1.RuntimeConfig{
			PipelineRoot: "model-pipeline-root",
		},
		ExperimentId: "123e4567-e89b-12d3-a456-426655440000",
	}

	recurringRun, err := server.CreateRecurringRun(nil, &apiv2beta1.CreateRecurringRunRequest{RecurringRun: apiRecurringRun})
	assert.Nil(t, err)

	expectedRecurringRun := &apiv2beta1.RecurringRun{
		RecurringRunId: "123e4567-e89b-12d3-a456-426655440000",
		DisplayName:    "recurring_run_1",
		ServiceAccount: "pipeline-runner",
		Mode:           apiv2beta1.RecurringRun_ENABLE,
		Namespace:      "ns1",
		MaxConcurrency: 1,
		Trigger: &apiv2beta1.Trigger{
			Trigger: &apiv2beta1.Trigger_CronSchedule{CronSchedule: &apiv2beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		CreatedAt:      &timestamp.Timestamp{Seconds: 4},
		UpdatedAt:      &timestamp.Timestamp{Seconds: 4},
		Status:         apiv2beta1.RecurringRun_ENABLED,
		PipelineSource: &apiv2beta1.RecurringRun_PipelineVersionId{PipelineVersionId: recurringRun.GetPipelineVersionId()},
		RuntimeConfig: &apiv2beta1.RuntimeConfig{
			PipelineRoot: "model-pipeline-root",
			Parameters:   make(map[string]*structpb.Value),
		},
		ExperimentId: "123e4567-e89b-12d3-a456-426655440000",
	}
	assert.Equal(t, expectedRecurringRun, recurringRun)

}

func TestGetRecurringRun(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	pipelineSpecStruct := &structpb.Struct{}
	yaml.Unmarshal([]byte(v2SpecHelloWorld), pipelineSpecStruct)

	apiRecurringRun := &apiv2beta1.RecurringRun{
		DisplayName:    "recurring_run_1",
		Mode:           apiv2beta1.RecurringRun_ENABLE,
		MaxConcurrency: 1,
		Trigger: &apiv2beta1.Trigger{
			Trigger: &apiv2beta1.Trigger_CronSchedule{CronSchedule: &apiv2beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		PipelineSource: &apiv2beta1.RecurringRun_PipelineSpec{PipelineSpec: pipelineSpecStruct},
		RuntimeConfig: &apiv2beta1.RuntimeConfig{
			PipelineRoot: "model-pipeline-root",
		},
		ExperimentId: "123e4567-e89b-12d3-a456-426655440000",
	}

	createdRecurringRun, err := server.CreateRecurringRun(nil, &apiv2beta1.CreateRecurringRunRequest{RecurringRun: apiRecurringRun})
	assert.Nil(t, err)

	expectedRecurringRun := &apiv2beta1.RecurringRun{
		RecurringRunId: "123e4567-e89b-12d3-a456-426655440000",
		DisplayName:    "recurring_run_1",
		ServiceAccount: "pipeline-runner",
		Mode:           apiv2beta1.RecurringRun_ENABLE,
		Namespace:      "ns1",
		MaxConcurrency: 1,
		Trigger: &apiv2beta1.Trigger{
			Trigger: &apiv2beta1.Trigger_CronSchedule{CronSchedule: &apiv2beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		CreatedAt:      &timestamp.Timestamp{Seconds: 4},
		UpdatedAt:      &timestamp.Timestamp{Seconds: 4},
		Status:         apiv2beta1.RecurringRun_ENABLED,
		PipelineSource: &apiv2beta1.RecurringRun_PipelineVersionId{PipelineVersionId: createdRecurringRun.GetPipelineVersionId()},
		RuntimeConfig: &apiv2beta1.RuntimeConfig{
			PipelineRoot: "model-pipeline-root",
			Parameters:   make(map[string]*structpb.Value),
		},
		ExperimentId: "123e4567-e89b-12d3-a456-426655440000",
	}

	recurringRun, err := server.GetRecurringRun(nil, &apiv2beta1.GetRecurringRunRequest{RecurringRunId: createdRecurringRun.RecurringRunId})
	assert.Nil(t, err)
	assert.Equal(t, expectedRecurringRun, recurringRun)

}

func TestListRecurringRuns(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	pipelineSpecStruct := &structpb.Struct{}
	yaml.Unmarshal([]byte(v2SpecHelloWorld), pipelineSpecStruct)

	apiRecurringRun := &apiv2beta1.RecurringRun{
		DisplayName:    "recurring_run_1",
		Mode:           apiv2beta1.RecurringRun_ENABLE,
		MaxConcurrency: 1,
		Trigger: &apiv2beta1.Trigger{
			Trigger: &apiv2beta1.Trigger_CronSchedule{CronSchedule: &apiv2beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		PipelineSource: &apiv2beta1.RecurringRun_PipelineSpec{PipelineSpec: pipelineSpecStruct},
		RuntimeConfig: &apiv2beta1.RuntimeConfig{
			PipelineRoot: "model-pipeline-root",
		},
		ExperimentId: "123e4567-e89b-12d3-a456-426655440000",
	}

	createdRecurringRun, err := server.CreateRecurringRun(nil, &apiv2beta1.CreateRecurringRunRequest{RecurringRun: apiRecurringRun})
	assert.Nil(t, err)

	expectedRecurringRun := &apiv2beta1.RecurringRun{
		RecurringRunId: "123e4567-e89b-12d3-a456-426655440000",
		DisplayName:    "recurring_run_1",
		ServiceAccount: "pipeline-runner",
		Mode:           apiv2beta1.RecurringRun_ENABLE,
		Namespace:      "ns1",
		MaxConcurrency: 1,
		Trigger: &apiv2beta1.Trigger{
			Trigger: &apiv2beta1.Trigger_CronSchedule{CronSchedule: &apiv2beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		CreatedAt:      &timestamp.Timestamp{Seconds: 4},
		UpdatedAt:      &timestamp.Timestamp{Seconds: 4},
		PipelineSource: &apiv2beta1.RecurringRun_PipelineVersionId{PipelineVersionId: createdRecurringRun.GetPipelineVersionId()},
		RuntimeConfig: &apiv2beta1.RuntimeConfig{
			PipelineRoot: "model-pipeline-root",
			Parameters:   make(map[string]*structpb.Value),
		},
		Status:       apiv2beta1.RecurringRun_ENABLED,
		ExperimentId: "123e4567-e89b-12d3-a456-426655440000",
	}

	expectedRecurringRunsList := []*apiv2beta1.RecurringRun{expectedRecurringRun}

	actualRecurringRunsList, err := server.ListRecurringRuns(nil, &apiv2beta1.ListRecurringRunsRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(actualRecurringRunsList.RecurringRuns))

	actualRecurringRunsList2, err := server.ListRecurringRuns(nil, &apiv2beta1.ListRecurringRunsRequest{
		ExperimentId: "123e4567-e89b-12d3-a456-426655440000",
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(actualRecurringRunsList.RecurringRuns))
	assert.Equal(t, expectedRecurringRunsList, actualRecurringRunsList2.RecurringRuns)
}

func TestEnableRecurringRun(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	pipelineSpecStruct := &structpb.Struct{}
	yaml.Unmarshal([]byte(v2SpecHelloWorld), pipelineSpecStruct)

	apiRecurringRun := &apiv2beta1.RecurringRun{
		DisplayName:    "recurring_run_1",
		Mode:           apiv2beta1.RecurringRun_ENABLE,
		MaxConcurrency: 1,
		Trigger: &apiv2beta1.Trigger{
			Trigger: &apiv2beta1.Trigger_CronSchedule{CronSchedule: &apiv2beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		PipelineSource: &apiv2beta1.RecurringRun_PipelineSpec{PipelineSpec: pipelineSpecStruct},
		RuntimeConfig: &apiv2beta1.RuntimeConfig{
			PipelineRoot: "model-pipeline-root",
		},
		ExperimentId: "123e4567-e89b-12d3-a456-426655440000",
	}

	createdRecurringRun, err := server.CreateRecurringRun(nil, &apiv2beta1.CreateRecurringRunRequest{RecurringRun: apiRecurringRun})
	assert.Nil(t, err)

	_, err = server.EnableRecurringRun(nil, &apiv2beta1.EnableRecurringRunRequest{RecurringRunId: createdRecurringRun.RecurringRunId})
	assert.Nil(t, err)
}

func TestDisableRecurringRun(t *testing.T) {
	clients, manager, _ := initWithExperiment(t)
	defer clients.Close()
	server := NewJobServer(manager, &JobServerOptions{CollectMetrics: false})

	pipelineSpecStruct := &structpb.Struct{}
	yaml.Unmarshal([]byte(v2SpecHelloWorld), pipelineSpecStruct)

	apiRecurringRun := &apiv2beta1.RecurringRun{
		DisplayName:    "recurring_run_1",
		Mode:           apiv2beta1.RecurringRun_ENABLE,
		MaxConcurrency: 1,
		Trigger: &apiv2beta1.Trigger{
			Trigger: &apiv2beta1.Trigger_CronSchedule{CronSchedule: &apiv2beta1.CronSchedule{
				StartTime: &timestamp.Timestamp{Seconds: 1},
				Cron:      "1 * * * *",
			}}},
		PipelineSource: &apiv2beta1.RecurringRun_PipelineSpec{PipelineSpec: pipelineSpecStruct},
		RuntimeConfig: &apiv2beta1.RuntimeConfig{
			PipelineRoot: "model-pipeline-root",
		},
		ExperimentId: "123e4567-e89b-12d3-a456-426655440000",
	}

	createdRecurringRun, err := server.CreateRecurringRun(nil, &apiv2beta1.CreateRecurringRunRequest{RecurringRun: apiRecurringRun})
	assert.Nil(t, err)

	_, err = server.DisableRecurringRun(nil, &apiv2beta1.DisableRecurringRunRequest{RecurringRunId: createdRecurringRun.RecurringRunId})
	assert.Nil(t, err)
}
