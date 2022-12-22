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

package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/cenkalti/backoff"
	"github.com/golang/glog"
	apiV1beta1 "github.com/kubeflow/pipelines/backend/api/v1beta1/go_client"

	// apiV2beta1 "github.com/kubeflow/pipelines/backend/api/v1beta1/go_client"
	"github.com/kubeflow/pipelines/backend/src/apiserver/archive"
	kfpauth "github.com/kubeflow/pipelines/backend/src/apiserver/auth"
	"github.com/kubeflow/pipelines/backend/src/apiserver/client"
	"github.com/kubeflow/pipelines/backend/src/apiserver/common"
	"github.com/kubeflow/pipelines/backend/src/apiserver/list"
	"github.com/kubeflow/pipelines/backend/src/apiserver/model"
	"github.com/kubeflow/pipelines/backend/src/apiserver/storage"
	"github.com/kubeflow/pipelines/backend/src/apiserver/template"
	exec "github.com/kubeflow/pipelines/backend/src/common"
	"github.com/kubeflow/pipelines/backend/src/common/util"
	scheduledworkflowclient "github.com/kubeflow/pipelines/backend/src/crd/pkg/client/clientset/versioned/typed/scheduledworkflow/v1beta1"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc/codes"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

// Metric variables. Please prefix the metric names with resource_manager_.
var (
	// Count the removed workflows due to garbage collection.
	workflowGCCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "resource_manager_workflow_gc",
		Help: "The number of gabarage-collected workflows",
	})
)

type ClientManagerInterface interface {
	ExperimentStore() storage.ExperimentStoreInterface
	PipelineStore() storage.PipelineStoreInterface
	JobStore() storage.JobStoreInterface
	RunStore() storage.RunStoreInterface
	TaskStore() storage.TaskStoreInterface
	ResourceReferenceStore() storage.ResourceReferenceStoreInterface
	DBStatusStore() storage.DBStatusStoreInterface
	DefaultExperimentStore() storage.DefaultExperimentStoreInterface
	ObjectStore() storage.ObjectStoreInterface
	ExecClient() util.ExecutionClient
	SwfClient() client.SwfClientInterface
	KubernetesCoreClient() client.KubernetesCoreInterface
	SubjectAccessReviewClient() client.SubjectAccessReviewInterface
	TokenReviewClient() client.TokenReviewInterface
	LogArchive() archive.LogArchiveInterface
	Time() util.TimeInterface
	UUID() util.UUIDGeneratorInterface
	Authenticators() []kfpauth.Authenticator
}

type ResourceManager struct {
	experimentStore           storage.ExperimentStoreInterface
	pipelineStore             storage.PipelineStoreInterface
	jobStore                  storage.JobStoreInterface
	runStore                  storage.RunStoreInterface
	taskStore                 storage.TaskStoreInterface
	resourceReferenceStore    storage.ResourceReferenceStoreInterface
	dBStatusStore             storage.DBStatusStoreInterface
	defaultExperimentStore    storage.DefaultExperimentStoreInterface
	objectStore               storage.ObjectStoreInterface
	execClient                util.ExecutionClient
	swfClient                 client.SwfClientInterface
	k8sCoreClient             client.KubernetesCoreInterface
	subjectAccessReviewClient client.SubjectAccessReviewInterface
	tokenReviewClient         client.TokenReviewInterface
	logArchive                archive.LogArchiveInterface
	time                      util.TimeInterface
	uuid                      util.UUIDGeneratorInterface
	authenticators            []kfpauth.Authenticator
}

func NewResourceManager(clientManager ClientManagerInterface) *ResourceManager {
	return &ResourceManager{
		experimentStore:           clientManager.ExperimentStore(),
		pipelineStore:             clientManager.PipelineStore(),
		jobStore:                  clientManager.JobStore(),
		runStore:                  clientManager.RunStore(),
		taskStore:                 clientManager.TaskStore(),
		resourceReferenceStore:    clientManager.ResourceReferenceStore(),
		dBStatusStore:             clientManager.DBStatusStore(),
		defaultExperimentStore:    clientManager.DefaultExperimentStore(),
		objectStore:               clientManager.ObjectStore(),
		execClient:                clientManager.ExecClient(),
		swfClient:                 clientManager.SwfClient(),
		k8sCoreClient:             clientManager.KubernetesCoreClient(),
		subjectAccessReviewClient: clientManager.SubjectAccessReviewClient(),
		tokenReviewClient:         clientManager.TokenReviewClient(),
		logArchive:                clientManager.LogArchive(),
		time:                      clientManager.Time(),
		uuid:                      clientManager.UUID(),
		authenticators:            clientManager.Authenticators(),
	}
}

func (r *ResourceManager) getWorkflowClient(namespace string) util.ExecutionInterface {
	return r.execClient.Execution(namespace)
}

func (r *ResourceManager) getScheduledWorkflowClient(namespace string) scheduledworkflowclient.ScheduledWorkflowInterface {
	return r.swfClient.ScheduledWorkflow(namespace)
}

func (r *ResourceManager) GetTime() util.TimeInterface {
	return r.time
}

func (r *ResourceManager) CreateExperiment(inputExperiment interface{}) (*model.Experiment, error) {
	experiment, err := r.ToModelExperiment(inputExperiment)
	if err != nil {
		return nil, util.Wrap(err, "Failed to convert experiment model")
	}
	return r.experimentStore.CreateExperiment(experiment)
}

func (r *ResourceManager) GetExperiment(experimentId string) (*model.Experiment, error) {
	return r.experimentStore.GetExperiment(experimentId)
}

func (r *ResourceManager) ListExperiments(filterContext *common.FilterContext, opts *list.Options) (
	experiments []*model.Experiment, total_size int, nextPageToken string, err error) {
	return r.experimentStore.ListExperiments(filterContext, opts)
}

func (r *ResourceManager) DeleteExperiment(experimentID string) error {
	_, err := r.experimentStore.GetExperiment(experimentID)
	if err != nil {
		return util.Wrap(err, "Delete experiment failed")
	}
	return r.experimentStore.DeleteExperiment(experimentID)
}

func (r *ResourceManager) ArchiveExperiment(ctx context.Context, experimentId string) error {
	// To archive an experiment
	// (1) update our persistent agent to disable CRDs of jobs in experiment
	// (2) update database to
	// (2.1) archive experiemnts
	// (2.2) archive runs
	// (2.3) disable jobs
	opts, err := list.NewOptions(&model.Job{}, 50, "name", nil)
	if err != nil {
		return util.NewInternalServerError(err,
			"Failed to create list jobs options when archiving experiment. ")
	}
	for {
		jobs, _, newToken, err := r.jobStore.ListJobs(&common.FilterContext{
			ReferenceKey: &common.ReferenceKey{Type: common.Experiment, ID: experimentId}}, opts)
		if err != nil {
			return util.NewInternalServerError(err,
				"Failed to list jobs of to-be-archived experiment. expID: %v", experimentId)
		}
		for _, job := range jobs {
			_, err = r.getScheduledWorkflowClient(job.Namespace).Patch(
				ctx,
				job.Name,
				types.MergePatchType,
				[]byte(fmt.Sprintf(`{"spec":{"enabled":%s}}`, strconv.FormatBool(false))))
			if err != nil {
				return util.NewInternalServerError(err,
					"Failed to disable job CR. jobID: %v", job.UUID)
			}
		}
		if newToken == "" {
			break
		} else {
			opts, err = list.NewOptionsFromToken(newToken, 50)
			if err != nil {
				return util.NewInternalServerError(err,
					"Failed to create list jobs options from page token when archiving experiment. ")
			}
		}
	}
	return r.experimentStore.ArchiveExperiment(experimentId)
}

func (r *ResourceManager) UnarchiveExperiment(experimentId string) error {
	return r.experimentStore.UnarchiveExperiment(experimentId)
}

// Creates a pipeline, but does not create a pipeline version.
// Call CreatePipelineVersion to create a pipeline version.
func (r *ResourceManager) CreatePipeline(p model.Pipeline) (*model.Pipeline, error) {
	// Create a record in KFP DB (only pipelines table)
	newPipeline, err := r.pipelineStore.CreatePipeline(&p)
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to create a pipeline in PipelineStore.")
	}

	newPipeline.Status = model.PipelineReady
	err = r.pipelineStore.UpdatePipelineStatus(
		newPipeline.UUID,
		newPipeline.Status,
	)
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to update status of a pipeline after creation.")
	}
	return newPipeline, nil
}

// Create a pipeline version.
// PipelineSpec is stored as a sting inside PipelineVersion in v2beta1.
func (r *ResourceManager) CreatePipelineVersion(pv model.PipelineVersion) (*model.PipelineVersion, error) {
	// Extract pipeline id
	pipelineId := pv.PipelineId
	if len(pipelineId) == 0 {
		return nil, util.NewInvalidInputError("ResourceManager: Failed to create a pipeline version due to missing pipeline id.")
	}

	// Fetch pipeline spec
	pipelineSpecBytes, pipelineSpecURI, err := r.FetchTemplateFromPipelineSpec(&pv)
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to create a pipeline version as template is broken.")
	}
	pv.PipelineSpec = string(pipelineSpecBytes)
	if pipelineSpecURI != "" {
		pv.PipelineSpecURI = pipelineSpecURI
	}

	// Create a template
	tmpl, err := template.New(pipelineSpecBytes)
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to create a pipeline version due to template creation error.")
	}
	if tmpl.IsV2() {
		pipeline, err := r.GetPipeline(pipelineId)
		if err != nil {
			return nil, util.Wrap(err, "ResourceManager: Failed to create a pipeline version as parent pipeline was not found.")
		}
		tmpl.OverrideV2PipelineName(pipeline.Name, pipeline.Namespace)
	}
	paramsJSON, err := tmpl.ParametersJSON()
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to create a pipeline version due to error converting parameters to json.")
	}
	pv.Parameters = paramsJSON
	pv.Status = model.PipelineVersionCreating

	// Create a record in DB
	version, err := r.pipelineStore.CreatePipelineVersion(pv)
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to create pipeline version in PipelineStore.")
	}

	// TODO (gkcalat): consider removing this after v2beta1 GA if we adopt storing PipelineSpec in DB.
	// Store the pipeline file
	err = r.objectStore.AddFile(tmpl.Bytes(), r.objectStore.GetPipelineKey(fmt.Sprint(version.UUID)))
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to create a pipeline version due to error saving PipelineSpec to ObjectStore.")
	}

	// After pipeline version being created in DB and pipeline file being
	// saved in minio server, set this pieline version to status ready.
	version.Status = model.PipelineVersionReady
	err = r.pipelineStore.UpdatePipelineVersionStatus(version.UUID, version.Status)
	if err != nil {
		return nil, util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to change the status of a new pipeline version with id %v.", version.UUID))
	}
	return version, nil
}

// Fetches PipelineSpec as []byte array and the corresponding URI from a PipelineVersion
func (r *ResourceManager) FetchTemplateFromPipelineSpec(pipelineVersion *model.PipelineVersion) ([]byte, string, error) {
	if len(pipelineVersion.PipelineSpec) != 0 {
		// Check pipeline spec string first
		bytes, err := util.MarshalJsonWithError(pipelineVersion.PipelineSpec)
		return bytes, pipelineVersion.PipelineSpecURI, err
	} else {
		// Try reading object store from pipeline_spec_uri
		template, errUri := r.objectStore.GetFile(pipelineVersion.PipelineSpecURI)
		if errUri != nil {
			// Try reading object store from pipeline_version_id
			template, errUUID := r.objectStore.GetFile(r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersion.PipelineSpecURI)))
			if errUri != nil {
				return nil, "", util.Wrap(
					util.Wrap(errUri, "ResourceManager: Failed to read a file from pipeline_spec_uri."),
					util.Wrap(errUUID, "ResourceManager: Failed to read a file from OS with pipeline_version_id.").Error(),
				)
			}
			return template, r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersion.PipelineSpecURI)), nil
		}
		return template, pipelineVersion.PipelineSpecURI, nil
	}
}

// Returns a pipeline.
func (r *ResourceManager) GetPipeline(pipelineId string) (*model.Pipeline, error) {
	if pipeline, err := r.pipelineStore.GetPipeline(pipelineId); err != nil {
		return nil, util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to get a pipeline with id %v.", pipelineId))
	} else {
		return pipeline, nil
	}
}

func (r *ResourceManager) GetPipelineVersion(versionId string) (*model.PipelineVersion, error) {
	return r.pipelineStore.GetPipelineVersion(versionId)
}

// Returns a pipeline specified by name and namespace.
func (r *ResourceManager) GetPipelineByNameAndNamespace(name string, namespace string) (*model.Pipeline, error) {
	if pipeline, err := r.pipelineStore.GetPipelineByNameAndNamespace(name, namespace); err != nil {
		return nil, util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to get a pipeline named %v in namespace %v.", name, namespace))
	} else {
		return pipeline, nil
	}
}

// Returns tge latest template for a specified pipeline id.
func (r *ResourceManager) GetPipelineLatestTemplate(pipelineId string) ([]byte, error) {
	// Verify pipeline exists
	_, err := r.pipelineStore.GetPipeline(pipelineId)
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to get a template as pipeline was not found.")
	}

	// Get the latest pipeline version
	latestPipelineVersion, err := r.pipelineStore.GetLatestPipelineVersion(pipelineId)
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to get the latest template for a pipeline.")
	}

	// Fetch template []byte array
	if bytes, _, err := r.FetchTemplateFromPipelineSpec(latestPipelineVersion); err != nil {
		return nil, util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to get the latest template for pipeline with id %v.", pipelineId))
	} else {
		return bytes, nil
	}
}

// Returns a template for a specified pipeline version id.
func (r *ResourceManager) GetPipelineVersionTemplate(versionId string) ([]byte, error) {
	// Verify pipeline version exist
	pipelineVersion, err := r.pipelineStore.GetPipelineVersion(versionId)
	if err != nil {
		return nil, util.Wrap(err, "ResourceManager: Failed to get pipeline version template as it was not found.")
	}

	// Fetch template []byte array
	if bytes, _, err := r.FetchTemplateFromPipelineSpec(pipelineVersion); err != nil {
		return nil, util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to get a template for pipeline version with id %v.", versionId))
	} else {
		return bytes, nil
	}
}

// Returns a list of pipelines.
func (r *ResourceManager) ListPipelines(filterContext *common.FilterContext, opts *list.Options) ([]*model.Pipeline, int, string, error) {
	pipelines, total_size, nextPageToken, err := r.pipelineStore.ListPipelines(filterContext, opts)
	if err != nil {
		err = util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to list pipelines: context %v, options %v.", filterContext, opts))
	}
	return pipelines, total_size, nextPageToken, err
}

func (r *ResourceManager) ListPipelineVersions(pipelineId string, opts *list.Options) (pipelines []*model.PipelineVersion, total_size int, nextPageToken string, err error) {
	return r.pipelineStore.ListPipelineVersions(pipelineId, opts)
}

// Updates the status of a pipeline.
func (r *ResourceManager) UpdatePipelineStatus(pipelineId string, status model.PipelineStatus) error {
	err := r.pipelineStore.UpdatePipelineStatus(pipelineId, status)
	if err != nil {
		return util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to update the status of pipeline %v to %v.", pipelineId, status))
	}
	return nil
}

// Updates the status of a pipeline version.
func (r *ResourceManager) UpdatePipelineVersionStatus(pipelineVersionId string, status model.PipelineVersionStatus) error {
	err := r.pipelineStore.UpdatePipelineVersionStatus(pipelineVersionId, status)
	if err != nil {
		return util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to update the status of pipeline version %v to %v.", pipelineVersionId, status))
	}
	return nil
}

// TODO: UPDATE THIS
// Deletes a pipeline.
// This has changed the behavior in v2beta1.
func (r *ResourceManager) DeletePipeline(pipelineId string) error {
	_, err := r.pipelineStore.GetPipeline(pipelineId)
	if err != nil {
		return util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to delete pipeline with id %v as it was not found.", pipelineId))
	}

	// Mark pipeline as deleting so it's not visible to user.
	err = r.pipelineStore.UpdatePipelineStatus(pipelineId, model.PipelineDeleting)
	if err != nil {
		return util.Wrap(err, fmt.Sprintf("ResourceManager: Failed to change pipeline status to DELETING for pipeline id %v.", pipelineId))
	}

	// Delete pipeline file and DB entry.
	// Not fail the request if this step failed. A background run will do the cleanup.
	// https://github.com/kubeflow/pipelines/issues/388
	// TODO(jingzhang36): For now (before exposing version API), we have only 1
	// file with both pipeline and version pointing to it;  so it is ok to do
	// the deletion as follows. After exposing version API, we can have multiple
	// versions and hence multiple files, and we shall improve performance by
	// either using async deletion in order for this method to be non-blocking
	// or or exploring other performance optimization tools provided by gcs.
	err = r.objectStore.DeleteFile(r.objectStore.GetPipelineKey(fmt.Sprint(pipelineId)))
	if err != nil {
		glog.Errorf("%v", errors.Wrapf(err, "ResourceManager: Failed to delete pipeline file for pipeline id %v.", pipelineId))
		return nil
	}
	err = r.objectStore.DeleteFile(r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersion.PipelineSpecURI)))
	if err != nil {
		glog.Errorf("%v", errors.Wrapf(err, "ResourceManager: Failed to delete pipeline file for pipeline id %v.", pipelineId))
		return nil
	}
	err = r.pipelineStore.DeletePipeline(pipelineId)
	if err != nil {
		glog.Errorf("%v", errors.Wrapf(err, "ResourceManager: Failed to delete pipeline DB entry for pipeline id %v.", pipelineId))
	}
	return nil
}

func (r *ResourceManager) DeletePipelineVersion(pipelineVersionId string) error {
	_, err := r.pipelineStore.GetPipelineVersion(pipelineVersionId)
	if err != nil {
		return util.Wrap(err, "Delete pipeline version failed")
	}

	// Mark pipeline as deleting so it's not visible to user.
	err = r.pipelineStore.UpdatePipelineVersionStatus(pipelineVersionId, model.PipelineVersionDeleting)
	if err != nil {
		return util.Wrap(err, "Delete pipeline version failed")
	}

	err = r.objectStore.DeleteFile(r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersionId)))
	if err != nil {
		glog.Errorf("%v", errors.Wrapf(err, "Failed to delete pipeline file for pipeline version %v", pipelineVersionId))
		return util.Wrap(err, "Delete pipeline version failed")
	}
	err = r.pipelineStore.DeletePipelineVersion(pipelineVersionId)
	if err != nil {
		glog.Errorf("%v", errors.Wrapf(err, "Failed to delete pipeline DB entry for pipeline %v", pipelineVersionId))
		return util.Wrap(err, "Delete pipeline version failed")
	}

	return nil
}

func (r *ResourceManager) CreateRun(ctx context.Context, apiRun *apiV1beta1.Run) (*model.RunDetail, error) {
	// Get manifest from either of the two places:
	// (1) raw manifest in pipeline_spec
	// (2) pipeline version in resource_references
	// And the latter takes priority over the former when the manifest is from pipeline_spec.pipeline_id
	// workflow/pipeline manifest and pipeline id/version will not exist at the same time, guaranteed by the validation phase
	manifestBytes, err := getManifestBytes(apiRun.PipelineSpec, &apiRun.ResourceReferences, r)
	if err != nil {
		return nil, err
	}

	uuid, err := r.uuid.NewRandom()
	if err != nil {
		return nil, util.NewInternalServerError(err, "Failed to generate run ID.")
	}
	runId := uuid.String()
	runAt := r.time.Now().Unix()

	tmpl, err := template.New(manifestBytes)
	if err != nil {
		return nil, err
	}
	runWorkflowOptions := template.RunWorkflowOptions{
		RunId: runId,
		RunAt: runAt,
	}
	executionSpec, err := tmpl.RunWorkflow(apiRun, runWorkflowOptions)
	if err != nil {
		return nil, util.NewInternalServerError(err, "failed to generate the ExecutionSpec.")
	}
	// Add a reference to the default experiment if run does not already have a containing experiment
	ref, err := r.getDefaultExperimentIfNoExperiment(apiRun.GetResourceReferences())
	if err != nil {
		return nil, err
	}
	if ref != nil {
		apiRun.ResourceReferences = append(apiRun.GetResourceReferences(), ref)
	}

	namespace, err := r.getNamespaceFromExperiment(apiRun.GetResourceReferences())
	if err != nil {
		return nil, err
	}

	err = executionSpec.Validate(false, false)
	if err != nil {
		return nil, util.NewInternalServerError(err, "Failed to validate workflow for (%+v)", executionSpec.ExecutionName())
	}
	// Create argo workflow CR resource
	newExecSpec, err := r.getWorkflowClient(namespace).Create(ctx, executionSpec, v1.CreateOptions{})
	if err != nil {
		return nil, util.NewInternalServerError(err, "Failed to create a workflow for (%s)", executionSpec.ExecutionName())
	}

	// Patched the default value to apiRun
	if common.GetBoolConfigWithDefault(common.HasDefaultBucketEnvVar, false) {
		for _, param := range apiRun.PipelineSpec.Parameters {
			var err error
			param.Value, err = common.PatchPipelineDefaultParameter(param.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to patch default value to pipeline. Error: %v", err)
			}
		}
	}

	// Store run metadata into database
	runDetail, err := r.ToModelRunDetail(apiRun, runId, newExecSpec, string(manifestBytes), tmpl.GetTemplateType())
	if err != nil {
		return nil, util.Wrap(err, "Failed to convert run model")
	}

	// Assign the create at time.
	runDetail.CreatedAtInSec = runAt

	// Assign the scheduled at time
	if !apiRun.ScheduledAt.AsTime().IsZero() {
		// if there is no scheduled time, then we assume this run is scheduled at the same time it is created
		runDetail.ScheduledAtInSec = runAt
	} else {
		runDetail.ScheduledAtInSec = apiRun.ScheduledAt.AsTime().Unix()
	}

	return r.runStore.CreateRun(runDetail)
}

func (r *ResourceManager) GetRun(runId string) (*model.RunDetail, error) {
	return r.runStore.GetRun(runId)
}

func (r *ResourceManager) ListRuns(filterContext *common.FilterContext,
	opts *list.Options) (runs []*model.Run, total_size int, nextPageToken string, err error) {
	return r.runStore.ListRuns(filterContext, opts)
}

func (r *ResourceManager) ArchiveRun(runId string) error {
	return r.runStore.ArchiveRun(runId)
}

func (r *ResourceManager) UnarchiveRun(runId string) error {
	experimentRef, err := r.resourceReferenceStore.GetResourceReference(runId, common.Run, common.Experiment)
	if err != nil {
		return util.Wrap(err, "Failed to retrieve resource reference")
	}

	experiment, err := r.GetExperiment(experimentRef.ReferenceUUID)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve experiment")
	}

	if experiment.StorageState == "ARCHIVED" {
		return util.NewFailedPreconditionError(errors.New("Unarchive the experiment first to allow the run to be restored"),
			fmt.Sprintf("Unarchive experiment with name `%s` first to allow run `%s` to be restored", experimentRef.ReferenceName, runId))
	}
	return r.runStore.UnarchiveRun(runId)
}

func (r *ResourceManager) DeleteRun(ctx context.Context, runID string) error {
	runDetail, err := r.checkRunExist(runID)
	if err != nil {
		return util.Wrap(err, "Delete run failed")
	}
	namespace, err := r.GetNamespaceFromRunID(runID)
	if err != nil {
		return util.Wrap(err, "Delete run failed")
	}
	err = r.getWorkflowClient(namespace).Delete(ctx, runDetail.Name, v1.DeleteOptions{})
	if err != nil {
		// API won't need to delete the workflow CR
		// once persistent agent sync the state to DB and set TTL for it.
		glog.Warningf("Failed to delete run %v. Error: %v", runDetail.Name, err.Error())
	}
	err = r.runStore.DeleteRun(runID)
	if err != nil {
		return util.Wrap(err, "Delete run failed")
	}
	return nil
}

func (r *ResourceManager) CreateTask(ctx context.Context, apiTask *apiV1beta1.Task) (*model.Task, error) {
	uuid, err := r.uuid.NewRandom()
	if err != nil {
		return nil, util.NewInternalServerError(err, "Failed to generate task ID.")
	}
	id := uuid.String()
	task := model.Task{
		UUID:              id,
		Namespace:         apiTask.Namespace,
		PipelineName:      apiTask.PipelineName,
		RunUUID:           apiTask.RunId,
		MLMDExecutionID:   apiTask.MlmdExecutionID,
		CreatedTimestamp:  apiTask.CreatedAt.AsTime().Unix(),
		FinishedTimestamp: apiTask.FinishedAt.AsTime().Unix(),
		Fingerprint:       apiTask.Fingerprint,
	}
	return r.taskStore.CreateTask(&task)
}

func (r *ResourceManager) ListTasks(filterContext *common.FilterContext,
	opts *list.Options) (tasks []*model.Task, total_size int, nextPageToken string, err error) {
	return r.taskStore.ListTasks(filterContext, opts)
}

func (r *ResourceManager) ListJobs(filterContext *common.FilterContext,
	opts *list.Options) (jobs []*model.Job, total_size int, nextPageToken string, err error) {
	return r.jobStore.ListJobs(filterContext, opts)
}

// TerminateWorkflow terminates a workflow by setting its activeDeadlineSeconds to 0
func TerminateWorkflow(ctx context.Context, wfClient util.ExecutionInterface, name string) error {
	patchObj := map[string]interface{}{
		"spec": map[string]interface{}{
			"activeDeadlineSeconds": 0,
		},
	}

	patch, err := json.Marshal(patchObj)
	if err != nil {
		return util.NewInternalServerError(err, "Unexpected error while marshalling a patch object.")
	}

	var operation = func() error {
		_, err = wfClient.Patch(ctx, name, types.MergePatchType, patch, v1.PatchOptions{})
		return err
	}
	var backoffPolicy = backoff.WithMaxRetries(backoff.NewConstantBackOff(100), 10)
	err = backoff.Retry(operation, backoffPolicy)
	return err
}

func (r *ResourceManager) TerminateRun(ctx context.Context, runId string) error {
	runDetail, err := r.checkRunExist(runId)
	if err != nil {
		return util.Wrap(err, "Terminate run failed")
	}

	namespace, err := r.GetNamespaceFromRunID(runId)
	if err != nil {
		return util.Wrap(err, "Terminate run failed")
	}

	err = r.runStore.TerminateRun(runId)
	if err != nil {
		return util.Wrap(err, "Terminate run failed")
	}

	err = TerminateWorkflow(ctx, r.getWorkflowClient(namespace), runDetail.Run.Name)
	if err != nil {
		return util.NewInternalServerError(err, "Failed to terminate the run")
	}
	return nil
}

func (r *ResourceManager) RetryRun(ctx context.Context, runId string) error {
	runDetail, err := r.checkRunExist(runId)
	if err != nil {
		return util.Wrap(err, "Retry run failed")
	}
	namespace, err := r.GetNamespaceFromRunID(runId)
	if err != nil {
		return util.Wrap(err, "Retry run failed")
	}

	if runDetail.WorkflowSpecManifest != "" && runDetail.WorkflowRuntimeManifest == "" {
		return util.NewBadRequestError(errors.New("workflow cannot be retried"), "Workflow must be Failed/Error to retry")
	}
	if runDetail.PipelineSpecManifest != "" {
		return util.NewBadRequestError(errors.New("workflow cannot be retried"), "Workflow must be with v1 mode to retry")
	}
	execSpec, err := util.NewExecutionSpecJSON(util.ArgoWorkflow, []byte(runDetail.WorkflowRuntimeManifest))
	if err != nil {
		return util.NewInternalServerError(err, "Failed to retrieve the runtime pipeline spec from the run")
	}

	if err := execSpec.Decompress(); err != nil {
		return util.NewInternalServerError(err, "Failed to decompress workflow")
	}

	if err := execSpec.CanRetry(); err != nil {
		return err
	}

	newExecSpec, podsToDelete, err := execSpec.GenerateRetryExecution()
	if err != nil {
		return util.Wrap(err, "Retry run failed.")
	}

	if err = deletePods(ctx, r.k8sCoreClient, podsToDelete, namespace); err != nil {
		return util.NewInternalServerError(err, "Retry run failed. Failed to clean up the failed pods from previous run.")
	}

	// First try to update workflow
	updateError := r.updateWorkflow(ctx, newExecSpec, namespace)
	if updateError != nil {
		// Remove resource version
		newExecSpec.SetVersion("")
		newCreatedWorkflow, createError := r.getWorkflowClient(namespace).Create(ctx, newExecSpec, v1.CreateOptions{})
		if createError != nil {
			return util.NewInternalServerError(createError,
				"Retry run failed. Failed to create or update the run. Update Error: %s, Create Error: %s",
				updateError.Error(), createError.Error())
		}
		newExecSpec = newCreatedWorkflow
	}
	err = r.runStore.UpdateRun(runId, string(newExecSpec.ExecutionStatus().Condition()), 0, newExecSpec.ToStringForStore())
	if err != nil {
		return util.NewInternalServerError(err, "Failed to update the database entry.")
	}
	return nil
}

func (r *ResourceManager) ReadLog(ctx context.Context, runId string, nodeId string, follow bool, dst io.Writer) error {
	run, err := r.checkRunExist(runId)
	if err != nil {
		return util.NewBadRequestError(errors.New("log cannot be read"), "Run does not exist")
	}

	err = r.readRunLogFromPod(ctx, run, nodeId, follow, dst)
	if err != nil && r.logArchive != nil {
		err = r.readRunLogFromArchive(run, nodeId, dst)
	}

	return err
}

func (r *ResourceManager) readRunLogFromPod(ctx context.Context, run *model.RunDetail, nodeId string, follow bool, dst io.Writer) error {
	logOptions := corev1.PodLogOptions{
		Container:  "main",
		Timestamps: false,
		Follow:     follow,
	}

	req := r.k8sCoreClient.PodClient(run.Namespace).GetLogs(nodeId, &logOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			glog.Errorf("Failed to access Pod log: %v", err)
		}
		return util.NewInternalServerError(err, "error in opening log stream")
	}
	defer podLogs.Close()

	_, err = io.Copy(dst, podLogs)
	if err != nil && err != io.EOF {
		return util.NewInternalServerError(err, "error in streaming the log")
	}

	return nil
}

func (r *ResourceManager) readRunLogFromArchive(run *model.RunDetail, nodeId string, dst io.Writer) error {
	if run.WorkflowRuntimeManifest == "" {
		return util.NewBadRequestError(errors.New("archived log cannot be read"), "Failed to retrieve the runtime workflow from the run")
	}

	execSpec, err := util.NewExecutionSpecJSON(util.ArgoWorkflow, []byte(run.WorkflowRuntimeManifest))
	if err != nil {
		return util.NewInternalServerError(err, "Failed to retrieve the runtime pipeline spec from the run")
	}

	logPath, err := r.logArchive.GetLogObjectKey(execSpec, nodeId)
	if err != nil {
		return err
	}

	logContent, err := r.objectStore.GetFile(logPath)
	if err != nil {
		return util.NewInternalServerError(err, "Failed to retrieve the log file from archive")
	}

	err = r.logArchive.CopyLogFromArchive(logContent, dst, archive.ExtractLogOptions{LogFormat: archive.LogFormatText, Timestamps: false})

	if err != nil {
		return util.NewInternalServerError(err, "error in streaming the log")
	}

	return nil
}

func (r *ResourceManager) updateWorkflow(ctx context.Context, newWorkflow util.ExecutionSpec, namespace string) error {
	// If fail to get the workflow, return error.
	latestWorkflow, err := r.getWorkflowClient(namespace).Get(ctx, newWorkflow.ExecutionName(), v1.GetOptions{})
	if err != nil {
		return err
	}
	// Update the workflow's resource version to latest.
	newWorkflow.SetVersion(latestWorkflow.Version())
	_, err = r.getWorkflowClient(namespace).Update(ctx, newWorkflow, v1.UpdateOptions{})
	return err
}

func (r *ResourceManager) GetJob(id string) (*model.Job, error) {
	return r.jobStore.GetJob(id)
}

func (r *ResourceManager) CreateJob(ctx context.Context, apiJob *apiV1beta1.Job) (*model.Job, error) {
	// Get workflow from either of the two places:
	// (1) raw pipeline manifest in pipeline_spec
	// (2) pipeline version in resource_references
	// 	And the latter takes priority over the former when the pipeline manifest is from pipeline_spec.pipeline_id
	// workflow manifest and pipeline id/version will not exist at the same time, guaranteed by the validation phase
	manifestBytes, err := getManifestBytes(apiJob.PipelineSpec, &apiJob.ResourceReferences, r)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(manifestBytes)
	if err != nil {
		return nil, err
	}

	scheduledWorkflow, err := tmpl.ScheduledWorkflow(apiJob)
	if err != nil {
		return nil, util.Wrap(err, "failed to generate the scheduledWorkflow.")
	}
	// Add a reference to the default experiment if run does not already have a containing experiment
	ref, err := r.getDefaultExperimentIfNoExperiment(apiJob.GetResourceReferences())
	if err != nil {
		return nil, err
	}
	if ref != nil {
		apiJob.ResourceReferences = append(apiJob.GetResourceReferences(), ref)
	}

	namespace, err := r.getNamespaceFromExperiment(apiJob.GetResourceReferences())
	if err != nil {
		return nil, err
	}

	newScheduledWorkflow, err := r.getScheduledWorkflowClient(namespace).Create(ctx, scheduledWorkflow)
	if err != nil {
		return nil, util.NewInternalServerError(err, "Failed to create a scheduled workflow for (%s)", scheduledWorkflow.Name)
	}

	job, err := r.ToModelJob(apiJob, util.NewScheduledWorkflow(newScheduledWorkflow), string(manifestBytes), tmpl.GetTemplateType())
	if err != nil {
		return nil, util.Wrap(err, "Create job failed")
	}

	now := r.time.Now().Unix()
	job.CreatedAtInSec = now
	job.UpdatedAtInSec = now
	return r.jobStore.CreateJob(job)
}

func (r *ResourceManager) EnableJob(ctx context.Context, jobID string, enabled bool) error {
	var job *model.Job
	var err error
	if enabled {
		job, err = r.checkJobExist(ctx, jobID)
	} else {
		// We can skip custom resource existence verification, because disabling
		// the job do not need to care about it.
		job, err = r.jobStore.GetJob(jobID)
	}
	if err != nil {
		return util.Wrap(err, "Enable/Disable job failed")
	}

	_, err = r.getScheduledWorkflowClient(job.Namespace).Patch(
		ctx,
		job.Name,
		types.MergePatchType,
		[]byte(fmt.Sprintf(`{"spec":{"enabled":%s}}`, strconv.FormatBool(enabled))))
	if err != nil {
		return util.NewInternalServerError(err,
			"Failed to enable/disable job CR. Enabled: %v, jobID: %v",
			enabled, jobID)
	}

	err = r.jobStore.EnableJob(jobID, enabled)
	if err != nil {
		return util.Wrapf(err, "Failed to enable/disable job. Enabled: %v, jobID: %v",
			enabled, jobID)
	}

	return nil
}

func (r *ResourceManager) DeleteJob(ctx context.Context, jobID string) error {
	job, err := r.jobStore.GetJob(jobID)
	if err != nil {
		return util.Wrap(err, "Delete job failed")
	}

	err = r.getScheduledWorkflowClient(job.Namespace).Delete(ctx, job.Name, &v1.DeleteOptions{})
	if err != nil {
		if !util.IsNotFound(err) {
			// For any error other than NotFound
			return util.NewInternalServerError(err, "Delete job CR failed")
		}

		// The ScheduledWorkflow was not found.
		glog.Infof("Deleting job '%v', but skipped deleting ScheduledWorkflow '%v' in namespace '%v' because it was not found. jobID: %v", job.Name, job.Name, job.Namespace, jobID)
		// Continue the execution, because we want to delete the
		// ScheduledWorkflow. We can skip deleting the ScheduledWorkflow
		// when it no longer exists.
	}
	err = r.jobStore.DeleteJob(jobID)
	if err != nil {
		return util.Wrap(err, "Delete job failed")
	}
	return nil
}

func (r *ResourceManager) ReportWorkflowResource(ctx context.Context, execSpec util.ExecutionSpec) error {
	objMeta := execSpec.ExecutionObjectMeta()
	execStatus := execSpec.ExecutionStatus()
	if _, ok := objMeta.Labels[util.LabelKeyWorkflowRunId]; !ok {
		// Skip reporting if the workflow doesn't have the run id label
		return util.NewInvalidInputError("Workflow[%s] missing the Run ID label", execSpec.ExecutionName())
	}
	runId := objMeta.Labels[util.LabelKeyWorkflowRunId]
	jobId := execSpec.ScheduledWorkflowUUIDAsStringOrEmpty()
	if len(execSpec.ExecutionNamespace()) == 0 {
		return util.NewInvalidInputError("Workflow missing namespace")
	}

	if execSpec.PersistedFinalState() {
		// If workflow's final state has being persisted, the workflow should be garbage collected.
		err := r.getWorkflowClient(execSpec.ExecutionNamespace()).Delete(ctx, execSpec.ExecutionName(), v1.DeleteOptions{})
		if err != nil {
			// A fix for kubeflow/pipelines#4484, persistence agent might have an outdated item in its workqueue, so it will
			// report workflows that no longer exist. It's important to return a not found error, so that persistence
			// agent won't retry again.
			if util.IsNotFound(err) {
				return util.NewNotFoundError(err, "Failed to delete the completed workflow for run %s", runId)
			} else {
				return util.NewInternalServerError(err, "Failed to delete the completed workflow for run %s", runId)
			}
		}
		// TODO(jingzhang36): find a proper way to pass collectMetricsFlag here.
		workflowGCCounter.Inc()
	}
	// If the run was Running and got terminated (activeDeadlineSeconds set to 0),
	// ignore its condition and mark it as such
	condition := execStatus.Condition()
	if execSpec.IsTerminating() {
		condition = exec.ExecutionPhase(model.RunTerminatingConditions)
	}
	if jobId == "" {
		// If a run doesn't have job ID, it's a one-time run created by Pipeline API server.
		// In this case the DB entry should already been created when argo workflow CR is created.
		if updateError := r.runStore.UpdateRun(runId, string(condition), execStatus.FinishedAt(), execSpec.ToStringForStore()); updateError != nil {
			if !util.IsUserErrorCodeMatch(updateError, codes.NotFound) {
				return util.Wrap(updateError, "Failed to update the run.")
			}
			// Handle run not found in run store error.
			// To avoid letting the workflow leak for ever, we need to GC it when its record does not exist in KFP DB.
			glog.Errorf("Cannot find reported workflow name=%q namespace=%q runId=%q in run store. "+
				"Deleting the workflow to avoid resource leaking. "+
				"This can be caused by installing two KFP instances that try to manage the same workflows "+
				"or an unknown bug. If you encounter this, recommend reporting more details in https://github.com/kubeflow/pipelines/issues/6189.",
				execSpec.ExecutionName(), execSpec.ExecutionNamespace(), runId)
			if err := r.getWorkflowClient(execSpec.ExecutionNamespace()).Delete(ctx, execSpec.ExecutionName(), v1.DeleteOptions{}); err != nil {
				if util.IsNotFound(err) {
					return util.NewNotFoundError(err, "Failed to delete the obsolete workflow for run %s", runId)
				}
				return util.NewInternalServerError(err, "Failed to delete the obsolete workflow for run %s", runId)
			}
			// TODO(jingzhang36): find a proper way to pass collectMetricsFlag here.
			workflowGCCounter.Inc()
			// Note, persistence agent will not retry reporting this workflow again, because updateError is a not found error.
			return util.Wrapf(updateError, "Failed to report workflow name=%q namespace=%q runId=%q", execSpec.ExecutionName(), execSpec.ExecutionNamespace(), runId)
		}
	} else {
		// Get the experiment resource reference for job.
		experimentRef, err := r.resourceReferenceStore.GetResourceReference(jobId, common.Job, common.Experiment)
		if err != nil {
			return util.Wrap(err, "Failed to retrieve the experiment ID for the job that created the run.")
		}
		jobName, err := r.getResourceName(common.Job, jobId)
		if err != nil {
			return util.Wrap(err, "Failed to retrieve the job name for the job that created the run.")
		}
		// Scheduled time equals created time if it is not specified
		var scheduledTimeInSec int64
		if execSpec.ScheduledAtInSecOr0() == 0 {
			scheduledTimeInSec = objMeta.CreationTimestamp.Unix()
		} else {
			scheduledTimeInSec = execSpec.ScheduledAtInSecOr0()
		}
		runDetail := &model.RunDetail{
			Run: model.Run{
				UUID:             runId,
				ExperimentUUID:   experimentRef.ReferenceUUID,
				DisplayName:      execSpec.ExecutionName(),
				Name:             execSpec.ExecutionName(),
				StorageState:     apiV1beta1.Run_STORAGESTATE_AVAILABLE.String(),
				Namespace:        execSpec.ExecutionNamespace(),
				CreatedAtInSec:   objMeta.CreationTimestamp.Unix(),
				ScheduledAtInSec: scheduledTimeInSec,
				FinishedAtInSec:  execStatus.FinishedAt(),
				Conditions:       string(condition),
				PipelineSpec: model.PipelineSpec{
					WorkflowSpecManifest: execSpec.GetExecutionSpec().ToStringForStore(),
				},
				ResourceReferences: []*model.ResourceReference{
					{
						ResourceUUID:  runId,
						ResourceType:  common.Run,
						ReferenceUUID: jobId,
						ReferenceName: jobName,
						ReferenceType: common.Job,
						Relationship:  common.Creator,
					},
					{
						ResourceUUID:  runId,
						ResourceType:  common.Run,
						ReferenceUUID: experimentRef.ReferenceUUID,
						ReferenceName: experimentRef.ReferenceName,
						ReferenceType: common.Experiment,
						Relationship:  common.Owner,
					},
				},
			},
			PipelineRuntime: model.PipelineRuntime{
				WorkflowRuntimeManifest: execSpec.ToStringForStore(),
			},
		}
		err = r.runStore.CreateOrUpdateRun(runDetail)
		if err != nil {
			return util.Wrap(err, "Failed to create or update the run.")
		}
	}

	if execStatus.IsInFinalState() {
		err := AddWorkflowLabel(ctx, r.getWorkflowClient(execSpec.ExecutionNamespace()), execSpec.ExecutionName(), util.LabelKeyWorkflowPersistedFinalState, "true")
		if err != nil {
			message := fmt.Sprintf("Failed to add PersistedFinalState label to workflow %s", execSpec.ExecutionName())
			// A fix for kubeflow/pipelines#4484, persistence agent might have an outdated item in its workqueue, so it will
			// report workflows that no longer exist. It's important to return a not found error, so that persistence
			// agent won't retry again.
			if util.IsNotFound(err) {
				return util.NewNotFoundError(err, message)
			} else {
				return util.Wrapf(err, message)
			}
		}
	}

	return nil
}

// AddWorkflowLabel add label for a workflow
func AddWorkflowLabel(ctx context.Context, wfClient util.ExecutionInterface, name string, labelKey string, labelValue string) error {
	patchObj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				labelKey: labelValue,
			},
		},
	}

	patch, err := json.Marshal(patchObj)
	if err != nil {
		return util.NewInternalServerError(err, "Unexpected error while marshalling a patch object.")
	}

	var operation = func() error {
		_, err = wfClient.Patch(ctx, name, types.MergePatchType, patch, v1.PatchOptions{})
		return err
	}
	var backoffPolicy = backoff.WithMaxRetries(backoff.NewConstantBackOff(100), 10)
	err = backoff.Retry(operation, backoffPolicy)
	return err
}

func (r *ResourceManager) ReportScheduledWorkflowResource(swf *util.ScheduledWorkflow) error {
	return r.jobStore.UpdateJob(swf)
}

// checkJobExist The Kubernetes API doesn't support CRUD by UID. This method
// retrieve the job metadata from the database, then retrieve the CR
// using the job name, and compare the given job id is same as the CR.
func (r *ResourceManager) checkJobExist(ctx context.Context, jobID string) (*model.Job, error) {
	job, err := r.jobStore.GetJob(jobID)
	if err != nil {
		return nil, util.Wrap(err, "Check job exist failed")
	}

	scheduledWorkflow, err := r.getScheduledWorkflowClient(job.Namespace).Get(ctx, job.Name, v1.GetOptions{})
	if err != nil {
		return nil, util.NewInternalServerError(err, "Check job exist failed")
	}
	if scheduledWorkflow == nil || string(scheduledWorkflow.UID) != jobID {
		return nil, util.NewResourceNotFoundError("job", job.Name)
	}
	return job, nil
}

// checkRunExist The Kubernetes API doesn't support CRUD by UID. This method
// retrieve the run metadata from the database, then retrieve the CR
// using the run name, and compare the given run id is same as the CR.
func (r *ResourceManager) checkRunExist(runID string) (*model.RunDetail, error) {
	runDetail, err := r.runStore.GetRun(runID)
	if err != nil {
		return nil, util.Wrap(err, "Check run exist failed")
	}
	return runDetail, nil
}

func (r *ResourceManager) getWorkflowSpecBytesFromPipelineSpec(spec *apiV1beta1.PipelineSpec) ([]byte, error) {
	if spec.GetWorkflowManifest() != "" {
		return []byte(spec.GetWorkflowManifest()), nil
	}
	return nil, util.NewInvalidInputError("Please provide a valid pipeline spec")
}

func (r *ResourceManager) getManifestBytesFromPipelineVersion(references []*apiV1beta1.ResourceReference) ([]byte, error) {
	var pipelineVersionId = ""
	for _, reference := range references {
		if reference.Key.Type == apiV1beta1.ResourceType_PIPELINE_VERSION && reference.Relationship == apiV1beta1.Relationship_CREATOR {
			pipelineVersionId = reference.Key.Id
		}
	}
	if len(pipelineVersionId) == 0 {
		return nil, util.NewInvalidInputError("No pipeline version.")
	}
	manifestBytes, err := r.objectStore.GetFile(r.objectStore.GetPipelineKey(pipelineVersionId))
	if err != nil {
		return nil, util.Wrap(err, "Get manifest bytes from PipelineVersion failed.")
	}

	return manifestBytes, nil
}

func getManifestBytes(pipelineSpec *apiV1beta1.PipelineSpec, resourceReferences *[]*apiV1beta1.ResourceReference, r *ResourceManager) ([]byte, error) {
	var manifestBytes []byte
	if pipelineSpec.GetWorkflowManifest() != "" {
		manifestBytes = []byte(pipelineSpec.GetWorkflowManifest())
	} else if pipelineSpec.GetPipelineManifest() != "" {
		manifestBytes = []byte(pipelineSpec.GetPipelineManifest())
	} else {
		err := convertPipelineIdToDefaultPipelineVersion(pipelineSpec, resourceReferences, r)
		if err != nil {
			return nil, util.Wrap(err, "Failed to find default version to create run with pipeline id.")
		}
		manifestBytes, err = r.getManifestBytesFromPipelineVersion(*resourceReferences)
		if err != nil {
			return nil, util.Wrap(err, "Failed to fetch manifest bytes.")
		}
	}
	return manifestBytes, nil
}

// Used to initialize the Experiment database with a default to be used for runs
func (r *ResourceManager) CreateDefaultExperiment() (string, error) {
	// First check that we don't already have a default experiment ID in the DB.
	defaultExperimentId, err := r.GetDefaultExperimentId()
	if err != nil {
		return "", fmt.Errorf("Failed to check if default experiment exists. Err: %v", err)
	}
	// If default experiment ID is already present, don't fail, simply return.
	if defaultExperimentId != "" {
		glog.Infof("Default experiment already exists! ID: %v", defaultExperimentId)
		return "", nil
	}

	// Create default experiment
	defaultExperiment := &apiV1beta1.Experiment{
		Name:        "Default",
		Description: "All runs created without specifying an experiment will be grouped here.",
	}
	experiment, err := r.CreateExperiment(defaultExperiment)
	if err != nil {
		return "", fmt.Errorf("Failed to create default experiment. Err: %v", err)
	}

	// Set default experiment ID in the DB
	err = r.SetDefaultExperimentId(experiment.UUID)
	if err != nil {
		return "", fmt.Errorf("Failed to set default experiment ID. Err: %v", err)
	}

	glog.Infof("Default experiment is set. ID is: %v", experiment.UUID)
	return experiment.UUID, nil
}

// getDefaultExperimentIfNoExperiment If the provided run does not include a reference to a containing
// experiment, then we fetch the default experiment's ID and create a reference to that.
func (r *ResourceManager) getDefaultExperimentIfNoExperiment(references []*apiV1beta1.ResourceReference) (*apiV1beta1.ResourceReference, error) {
	// First check if there is already a referenced experiment
	for _, ref := range references {
		if ref.Key.Type == apiV1beta1.ResourceType_EXPERIMENT && ref.Relationship == apiV1beta1.Relationship_OWNER {
			return nil, nil
		}
	}
	if common.IsMultiUserMode() {
		return nil, util.NewInvalidInputError("Experiment is required in resource references.")
	}
	return r.getDefaultExperimentResourceReference(references)
}

func (r *ResourceManager) getDefaultExperimentResourceReference(references []*apiV1beta1.ResourceReference) (*apiV1beta1.ResourceReference, error) {
	// Create reference to the default experiment
	defaultExperimentId, err := r.GetDefaultExperimentId()
	if err != nil {
		return nil, util.NewInternalServerError(err, "Failed to retrieve default experiment")
	}
	if defaultExperimentId == "" {
		glog.Info("No default experiment was found. Creating a new default experiment")
		defaultExperimentId, err = r.CreateDefaultExperiment()
		if defaultExperimentId == "" || err != nil {
			return nil, util.NewInternalServerError(err, "Failed to create new default experiment")
		}
	}
	defaultExperimentRef := &apiV1beta1.ResourceReference{
		Key: &apiV1beta1.ResourceKey{
			Id:   defaultExperimentId,
			Type: apiV1beta1.ResourceType_EXPERIMENT,
		},
		Relationship: apiV1beta1.Relationship_OWNER,
	}

	return defaultExperimentRef, nil
}

func (r *ResourceManager) ReportMetric(metric *apiV1beta1.RunMetric, runUUID string) error {
	return r.runStore.ReportMetric(r.ToModelRunMetric(metric, runUUID))
}

// ReadArtifact parses run's workflow to find artifact file path and reads the content of the file
// from object store.
func (r *ResourceManager) ReadArtifact(runID string, nodeID string, artifactName string) ([]byte, error) {
	run, err := r.runStore.GetRun(runID)
	if err != nil {
		return nil, err
	}
	if run.WorkflowRuntimeManifest == "" {
		return nil, util.NewInvalidInputError("read artifact from run with v2 IR spec is not supported")
	}
	execSpec, err := util.NewExecutionSpecJSON(util.ArgoWorkflow, []byte(run.WorkflowRuntimeManifest))
	if err != nil {
		// This should never happen.
		return nil, util.NewInternalServerError(
			err, "failed to unmarshal workflow '%s'", run.WorkflowRuntimeManifest)
	}
	artifactPath := execSpec.ExecutionStatus().FindObjectStoreArtifactKeyOrEmpty(nodeID, artifactName)
	if artifactPath == "" {
		return nil, util.NewResourceNotFoundError(
			"artifact", common.CreateArtifactPath(runID, nodeID, artifactName))
	}
	return r.objectStore.GetFile(artifactPath)
}

func (r *ResourceManager) GetDefaultExperimentId() (string, error) {
	return r.defaultExperimentStore.GetDefaultExperimentId()
}

func (r *ResourceManager) SetDefaultExperimentId(id string) error {
	return r.defaultExperimentStore.SetDefaultExperimentId(id)
}

func (r *ResourceManager) HaveSamplesLoaded() (bool, error) {
	return r.dBStatusStore.HaveSamplesLoaded()
}

func (r *ResourceManager) MarkSampleLoaded() error {
	return r.dBStatusStore.MarkSampleLoaded()
}

func (r *ResourceManager) AuthenticateRequest(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", util.NewUnauthenticatedError(errors.New("Request error: context is nil"), "Request error: context is nil.")
	}

	// If the request header contains the user identity, requests are authorized
	// based on the namespace field in the request.
	var errlist []error
	for _, auth := range r.authenticators {
		userIdentity, err := auth.GetUserIdentity(ctx)
		if err == nil {
			return userIdentity, nil
		}
		errlist = append(errlist, err)
	}
	return "", utilerrors.NewAggregate(errlist)
}

func (r *ResourceManager) IsRequestAuthorized(ctx context.Context, userIdentity string, resourceAttributes *authorizationv1.ResourceAttributes) error {
	result, err := r.subjectAccessReviewClient.Create(
		ctx,
		&authorizationv1.SubjectAccessReview{
			Spec: authorizationv1.SubjectAccessReviewSpec{
				ResourceAttributes: resourceAttributes,
				User:               userIdentity,
			},
		},
		v1.CreateOptions{},
	)
	if err != nil {
		return util.NewInternalServerError(
			err,
			"Failed to create SubjectAccessReview for user '%s' (request: %+v)",
			userIdentity,
			resourceAttributes,
		)
	}
	if !result.Status.Allowed {
		return util.NewPermissionDeniedError(
			errors.New("Unauthorized access"),
			"User '%s' is not authorized with reason: %s (request: %+v)",
			userIdentity,
			result.Status.Reason,
			resourceAttributes,
		)
	}
	return nil
}

func (r *ResourceManager) GetNamespaceFromExperimentID(experimentID string) (string, error) {
	experiment, err := r.GetExperiment(experimentID)
	if err != nil {
		return "", util.Wrap(err, "Failed to get namespace from experiment ID.")
	}
	return experiment.Namespace, nil
}

func (r *ResourceManager) GetNamespaceFromRunID(runId string) (string, error) {
	runDetail, err := r.GetRun(runId)
	if err != nil {
		return "", util.Wrap(err, "Failed to get namespace from run id.")
	}
	return runDetail.Namespace, nil
}

func (r *ResourceManager) GetNamespaceFromJobID(jobId string) (string, error) {
	job, err := r.GetJob(jobId)
	if err != nil {
		return "", util.Wrap(err, "Failed to get namespace from Job ID.")
	}
	return job.Namespace, nil
}

func (r *ResourceManager) GetNamespaceFromPipelineID(pipelineId string) (string, error) {
	pipeline, err := r.GetPipeline(pipelineId)
	if err != nil {
		return "", util.Wrap(err, "Failed to get namespace from Pipeline ID")
	}
	return pipeline.Namespace, nil
}

func (r *ResourceManager) GetNamespaceFromPipelineVersion(versionId string) (string, error) {
	pipelineVersion, err := r.GetPipelineVersion(versionId)
	if err != nil {
		return "", util.Wrap(err, "Failed to get namespace from versionId ID")
	}
	return r.GetNamespaceFromPipelineID(pipelineVersion.PipelineId)
}

func (r *ResourceManager) getNamespaceFromExperiment(references []*apiV1beta1.ResourceReference) (string, error) {
	experimentID := common.GetExperimentIDFromAPIResourceReferences(references)
	experiment, err := r.GetExperiment(experimentID)
	if err != nil {
		return "", util.NewInternalServerError(err, "Failed to get experiment.")
	}

	namespace := experiment.Namespace
	if len(namespace) == 0 {
		if common.IsMultiUserMode() {
			return "", util.NewInternalServerError(errors.New("Missing namespace"), "Experiment %v doesn't have a namespace.", experiment.Name)
		} else {
			namespace = common.GetPodNamespace()
		}
	}
	return namespace, nil
}
