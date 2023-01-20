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
	serverOptions             map[string]interface{}
}

func NewResourceManager(clientManager ClientManagerInterface, opts map[string]interface{}) *ResourceManager {
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
		serverOptions:             opts,
	}
}

// Checks if sample pipelines have been loaded.
func (r *ResourceManager) HaveSamplesLoaded() (bool, error) {
	return r.dBStatusStore.HaveSamplesLoaded()
}

// Reports that sample pipelines have been loaded.
func (r *ResourceManager) MarkSampleLoaded() error {
	return r.dBStatusStore.MarkSampleLoaded()
}

func (r *ResourceManager) getWorkflowClient(namespace string) util.ExecutionInterface {
	return r.execClient.Execution(namespace)
}

func (r *ResourceManager) getScheduledWorkflowClient(namespace string) scheduledworkflowclient.ScheduledWorkflowInterface {
	return r.swfClient.ScheduledWorkflow(namespace)
}

// Fetches PipelineSpec as []byte array and a new URI of PipelineSpec.
// Returns empty string if PipelineSpec is found via PipelineSpecURI.
// It attempts to fetch PipelineSpec in the following order:
//  1. Directly read from pipeline versions's PipelineSpec field.
//  2. Fetch a yaml file from object store based on pipeline versions's PipelineSpecURI field.
//  3. Fetch a yaml file from object store based on pipeline versions's id.
//  4. Fetch a yaml file from object store based on pipeline's id.
func (r *ResourceManager) fetchTemplateFromPipelineVersion(pipelineVersion *model.PipelineVersion) ([]byte, string, error) {
	if len(pipelineVersion.PipelineSpec) != 0 {
		// Check pipeline spec string first
		bytes := []byte(pipelineVersion.PipelineSpec)
		return bytes, pipelineVersion.PipelineSpecURI, nil
	} else {
		// Try reading object store from pipeline_spec_uri
		template, errUri := r.objectStore.GetFile(pipelineVersion.PipelineSpecURI)
		if errUri != nil {
			// Try reading object store from pipeline_version_id
			template, errUUID := r.objectStore.GetFile(r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersion.UUID)))
			if errUUID != nil {
				// Try reading object store from pipeline_id
				template, errPipelineId := r.objectStore.GetFile(r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersion.PipelineId)))
				if errPipelineId != nil {
					return nil, "", util.Wrap(
						util.Wrap(
							util.Wrap(errUri, "Failed to read a file from pipeline_spec_uri"),
							util.Wrap(errUUID, "Failed to read a file from OS with pipeline_version_id").Error(),
						),
						util.Wrap(errPipelineId, "Failed to read a file from OS with pipeline_id").Error(),
					)
				}
				return template, r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersion.PipelineId)), nil
			}
			return template, r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersion.UUID)), nil
		}
		return template, "", nil
	}
}

// Fetches PipelineSpec's manifest as []byte array.
// It attempts to fetch PipelineSpec manifest in the following order:
//  1. Directly read from PipelineSpec's PipelineSpecManifest field.
//  2. Directly read from PipelineSpec's WorkflowSpecManifest field.
//  3. Fetch pipeline spec manifest from the pipeline version for PipelineSpec's PipelineVersionId field.
//  4. Fetch pipeline spec manifest from the latest pipeline version for PipelineSpec's PipelineId field.
func (r *ResourceManager) fetchTemplateFromPipelineSpec(p *model.PipelineSpec) ([]byte, error) {
	if p == nil {
		return nil, util.NewInvalidInputError("Failed to read pipeline spec manifest from nil")
	}
	if len(p.PipelineSpecManifest) != 0 {
		return []byte(p.PipelineSpecManifest), nil
	}
	if len(p.WorkflowSpecManifest) != 0 {
		return []byte(p.WorkflowSpecManifest), nil
	}
	var errPv, errP error
	if p.PipelineVersionId != "" {
		pv, errPv1 := r.GetPipelineVersion(p.PipelineVersionId)
		if errPv1 == nil {
			bytes, _, errPv2 := r.fetchTemplateFromPipelineVersion(pv)
			if errPv2 == nil {
				return bytes, nil
			} else {
				errPv = errPv2
			}
		} else {
			errPv = errPv1
		}
	}
	if p.PipelineId != "" {
		pv, errP1 := r.GetLatestPipelineVersion(p.PipelineId)
		if errP1 == nil {
			bytes, _, errP2 := r.fetchTemplateFromPipelineVersion(pv)
			if errP2 == nil {
				return bytes, nil
			} else {
				errP = errP2
			}
		} else {
			errP = errP1
		}
	}
	return nil, util.Wrap(
		util.Wrapf(errPv, "Failed to read pipeline spec for pipeline version id %v", p.PipelineVersionId),
		util.Wrapf(errP, "Failed to read pipeline spec for pipeline id %v", p.PipelineId).Error(),
	)
}

// Fetches PipelineSpec's manifest as []byte array from pipeline version's id.
func (r *ResourceManager) fetchTemplateFromPipelineVersionId(pipelineVersionId string) ([]byte, error) {
	if len(pipelineVersionId) == 0 {
		return nil, util.NewInvalidInputError("Failed to get manifest as pipeline version id is empty")
	}
	pv, err := r.GetPipelineVersion(pipelineVersionId)
	if err == nil {
		bytes, _, err := r.fetchTemplateFromPipelineVersion(pv)
		if err == nil {
			return bytes, nil
		}
	}
	return nil, util.Wrapf(err, "Failed to read pipeline spec for pipeline version id %v", pipelineVersionId)
}

// Fetches the default namespace for resources.
func (r *ResourceManager) GetDefaultNamespace() string {
	if namespace, ok := r.serverOptions["DefaultNamespace"]; ok {
		if namespace.(string) != "" && namespace.(string) != model.NoNamespace {
			return namespace.(string)
		}
	}
	return common.GetPodNamespace()
}

// Checks if the namespace is empty or equal to one of {`-`, `POD_NAMESPACE`, or the default value}.
func (r *ResourceManager) IsDefaultNamespace(namespace string) bool {
	if namespace == "" || namespace == model.NoNamespace {
		return true
	}
	if namespace == r.GetDefaultNamespace() {
		return true
	}
	return false
}

// Replaces the namespace if it is equal to one of {`-`, `POD_NAMESPACE`}.
// The new namespace will have the default value.
func (r *ResourceManager) ReplaceEmptyNamespace(namespace string) string {
	if namespace == "" || namespace == model.NoNamespace {
		return r.GetDefaultNamespace()
	}
	return namespace
}

// Fetches the default experiment id.
func (r *ResourceManager) GetDefaultExperimentId() (string, error) {
	return r.defaultExperimentStore.GetDefaultExperimentId()
}

// Sets the default experiment id.
func (r *ResourceManager) SetDefaultExperimentId(id string) error {
	return r.defaultExperimentStore.SetDefaultExperimentId(id)
}

// Fetches namespace that an experiment belongs to.
func (r *ResourceManager) GetNamespaceFromExperimentId(experimentId string) (string, error) {
	experiment, err := r.GetExperiment(experimentId)
	if err != nil {
		return "", util.Wrapf(err, "Failed to fetch namespace from experiment %v", experimentId)
	}
	if experiment.Namespace == "" {
		if common.IsMultiUserMode() {
			namespaceRef, err := r.resourceReferenceStore.GetResourceReference(experimentId, model.ExperimentResourceType, model.NamespaceResourceType)
			if err != nil {
				return "", util.Wrapf(err, "Failed to fetch namespace from experiment %v due to resource references fetching error", experimentId)
			}
			if namespaceRef == nil || namespaceRef.ReferenceUUID == "" {
				return "", util.NewInternalServerError(util.NewNotFoundError(errors.New("Namespace is empty"), "Experiment's namespace was not found"), "Failed to fetch a namespace for experiment %v in multi-user mode", experimentId)
			}
			experiment.Namespace = namespaceRef.ReferenceUUID
		} else {
			experiment.Namespace = common.GetPodNamespace()
		}
	}
	return experiment.Namespace, nil
}

// Fetches namespace that a pipeline belongs to.
func (r *ResourceManager) getNamespaceFromPipelineId(pipelineId string) (string, error) {
	pipeline, err := r.GetPipeline(pipelineId)
	if err != nil {
		return "", util.Wrapf(err, "Failed to fetch namespace from pipeline %v", pipelineId)
	}
	if pipeline.Namespace == "" {
		if common.IsMultiUserMode() {
			namespaceRef, err := r.resourceReferenceStore.GetResourceReference(pipelineId, model.PipelineResourceType, model.NamespaceResourceType)
			if err != nil {
				return "", util.Wrapf(err, "Failed to fetch namespace from pipeline %v due to resource references fetching error", pipelineId)
			}
			if namespaceRef == nil || namespaceRef.ReferenceUUID == "" {
				return "", util.NewInternalServerError(util.NewNotFoundError(errors.New("Namespace is empty"), "Pipeline's namespace was not found"), "Failed to fetch a namespace for pipeline %v in multi-user mode", pipelineId)
			}
			pipeline.Namespace = namespaceRef.ReferenceUUID
		} else {
			pipeline.Namespace = common.GetPodNamespace()
		}
	}
	return pipeline.Namespace, nil
}

// Fetches namespace that a pipeline version belongs to.
func (r *ResourceManager) getNamespaceFromPipelineVersionId(pipelineVersionId string) (string, error) {
	pipelineVersion, err := r.GetPipelineVersion(pipelineVersionId)
	if err != nil {
		return "", util.Wrapf(err, "Failed to fetch namespace from pipeline version %v due to fetching error", pipelineVersionId)
	}
	namespace, err := r.getNamespaceFromPipelineId(pipelineVersion.PipelineId)
	if err != nil {
		return "", util.Wrapf(err, "Failed to fetch namespace from pipeline version %v", pipelineVersionId)
	}
	return namespace, nil
}

// Fetches namespace that a run belongs to.
func (r *ResourceManager) getNamespaceFromRunId(runId string) (string, error) {
	run, err := r.GetRun(runId)
	if err != nil {
		return "", util.Wrapf(err, "Failed to fetch namespace from run %v due to fetching error", runId)
	}
	namespace, err := r.GetNamespaceFromExperimentId(run.ExperimentId)
	if err != nil {
		return "", util.Wrapf(err, "Failed to fetch namespace from run %v", runId)
	}
	return namespace, nil
}

// Returns parent namespace for a pipeline id.
func (r *ResourceManager) FetchNamespaceFromPipelineId(pipelineId string) (string, error) {
	pipeline, err := r.GetPipeline(pipelineId)
	if err != nil {
		return "", util.Wrapf(err, "Failed to get namespace for pipeline id %v", pipelineId)
	}
	return pipeline.Namespace, nil
}

// Returns parent namespace for a pipeline version id.
func (r *ResourceManager) FetchNamespaceFromPipelineVersionId(versionId string) (string, error) {
	pipelineVersion, err := r.GetPipelineVersion(versionId)
	if err != nil {
		return "", util.Wrapf(err, "Failed to get namespace for pipeline version id %v", versionId)
	}
	return r.FetchNamespaceFromPipelineId(pipelineVersion.PipelineId)
}

// Validates that the provided experiment belongs to the namespace. Returns error otherwise.
func (r *ResourceManager) ValidateExperimentNamespace(experimentId string, namespace string) error {
	if experimentId == "" || namespace == "" {
		return nil
	}
	experimentNamespace, err := r.GetNamespaceFromExperimentId(experimentId)
	if err != nil {
		return util.Wrapf(err, "Failed to validate the namespace of experiment %s", experimentId)
	}
	if experimentNamespace != "" && experimentNamespace != namespace {
		return util.NewInternalServerError(util.NewInvalidInputError("Experiment %s belongs to namespace '%s' (claimed a different namespace '%s')", experimentId, experimentNamespace, namespace), "Failed to validate the namespace of experiment %s", experimentId)
	}
	return nil
}

// Verifies whether the user identity, which is contained in the context object,
// can perform some action (verb) on a resource (resourceType/resourceName) living in the
// target namespace. If the returned error is nil, the authorization passes. Otherwise,
// authorization fails with a non-nil error.
func (r *ResourceManager) IsAuthorized(ctx context.Context, resourceAttributes *authorizationv1.ResourceAttributes) error {
	if !common.IsMultiUserMode() {
		// Skip authz if not multi-user mode.
		return nil
	}

	if common.IsMultiUserSharedReadMode() &&
		(resourceAttributes.Verb == common.RbacResourceVerbGet ||
			resourceAttributes.Verb == common.RbacResourceVerbList) {
		glog.Infof("Multi-user shared read mode is enabled. Request allowed: %+v", resourceAttributes)
		return nil
	}

	glog.Info("Getting user identity")
	if ctx == nil {
		return util.NewUnauthenticatedError(errors.New("Context is nil"), "Authentication request failed")
	}
	// If the request header contains the user identity, requests are authorized
	// based on the namespace field in the request.
	var errlist []error
	userIdentity := ""
	for _, auth := range r.authenticators {
		identity, err := auth.GetUserIdentity(ctx)
		if err == nil {
			userIdentity = identity
			break
		}
		errlist = append(errlist, err)
	}
	if userIdentity == "" {
		return util.NewUnauthenticatedError(utilerrors.NewAggregate(errlist), "Failed to check authorization. User identity is empty in the request header")
	}

	glog.Infof("User: %s, ResourceAttributes: %+v", userIdentity, resourceAttributes)
	glog.Info("Authorizing request")
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
		err = util.NewInternalServerError(
			err,
			"Failed to create SubjectAccessReview for user '%s' (request: %+v)",
			userIdentity,
			resourceAttributes,
		)
		glog.Info(err.Error())
		return err
	}
	if !result.Status.Allowed {
		err := util.NewPermissionDeniedError(
			errors.New("Unauthorized access"),
			"User '%s' is not authorized with reason: %s (request: %+v)",
			userIdentity,
			result.Status.Reason,
			resourceAttributes,
		)
		glog.Info(err.Error())
		return err
	}
	glog.Infof("Authorized user '%s': %+v", userIdentity, resourceAttributes)
	return nil
}

// Creates the default experiment entry.
func (r *ResourceManager) CreateDefaultExperiment() (string, error) {
	// First check that we don't already have a default experiment ID in the DB.
	defaultExperimentId, err := r.GetDefaultExperimentId()
	if err != nil {
		return "", util.Wrap(err, "Failed to check if default experiment exists")
	}
	// If default experiment ID is already present, don't fail, simply return.
	if defaultExperimentId != "" {
		glog.Infof("Default experiment already exists! ID: %v", defaultExperimentId)
		return "", nil
	}

	// TODO(gkcalat): consider moving the default namespace and experiment to server config.
	// Create default experiment
	defaultExperiment := &model.Experiment{
		Name:         "Default",
		Description:  "All runs created without specifying an experiment will be grouped here",
		Namespace:    r.GetDefaultNamespace(),
		StorageState: model.StorageStateAvailable,
	}
	experiment, err := r.CreateExperiment(defaultExperiment)
	if err != nil {
		return "", util.Wrap(err, "Failed to create the default experiment")
	}

	// Set default experiment ID in the DB
	err = r.SetDefaultExperimentId(experiment.UUID)
	if err != nil {
		return "", util.Wrap(err, "Failed to set default experiment ID")
	}

	glog.Infof("Default experiment is set. ID is: %v", experiment.UUID)
	return experiment.UUID, nil
}

// Creates a new experiment.
func (r *ResourceManager) CreateExperiment(experiment *model.Experiment) (*model.Experiment, error) {
	return r.experimentStore.CreateExperiment(experiment)
}

// Fetches an experiment with the given id.
func (r *ResourceManager) GetExperiment(experimentId string) (*model.Experiment, error) {
	return r.experimentStore.GetExperiment(experimentId)
}

// Fetches experiments with the given filtering and listing options.
func (r *ResourceManager) ListExperiments(filterContext *model.FilterContext, opts *list.Options) (
	experiments []*model.Experiment, total_size int, nextPageToken string, err error) {
	return r.experimentStore.ListExperiments(filterContext, opts)
}

// Archives the experiment with the given id.
func (r *ResourceManager) ArchiveExperiment(ctx context.Context, experimentId string) error {
	// To archive an experiment
	// (1) update our persistent agent to disable CRDs of jobs in experiment
	// (2) update database to
	// (2.1) archive experiments
	// (2.2) archive runs
	// (2.3) disable jobs
	opts, err := list.NewOptions(&model.Job{}, 50, "name", nil)
	if err != nil {
		return util.NewInternalServerError(err,
			"Failed to archive experiment %v. Report this backend error by creating a new issue on GitHub: https://github.com/kubeflow/pipelines/issues/new/choose", experimentId)
	}
	for {
		jobs, _, newToken, err := r.jobStore.ListJobs(&model.FilterContext{
			ReferenceKey: &model.ReferenceKey{Type: model.ExperimentResourceType, ID: experimentId}}, opts)
		if err != nil {
			return util.NewInternalServerError(err,
				"Failed to list jobs of to-be-archived experiment %v", experimentId)
		}
		for _, job := range jobs {
			_, err = r.getScheduledWorkflowClient(job.Namespace).Patch(
				ctx,
				job.K8SName,
				types.MergePatchType,
				[]byte(fmt.Sprintf(`{"spec":{"enabled":%s}}`, strconv.FormatBool(false))))
			if err != nil {
				return util.NewInternalServerError(err,
					"Failed to disable job %v while archiving experiment %v", job.UUID, experimentId)
			}
		}
		if newToken == "" {
			break
		} else {
			opts, err = list.NewOptionsFromToken(newToken, 50)
			if err != nil {
				return util.NewInternalServerError(err,
					"Failed to create list jobs options from page token when archiving experiment %v", experimentId)
			}
		}
	}
	return r.experimentStore.ArchiveExperiment(experimentId)
}

// Un-archives the experiment with the given id.
func (r *ResourceManager) UnarchiveExperiment(experimentId string) error {
	return r.experimentStore.UnarchiveExperiment(experimentId)
}

// Deletes the experiment with the given id.
func (r *ResourceManager) DeleteExperiment(experimentId string) error {
	_, err := r.experimentStore.GetExperiment(experimentId)
	if err != nil {
		return util.Wrapf(err, "Failed to delete experiment %v due to error fetching it", experimentId)
	}
	return r.experimentStore.DeleteExperiment(experimentId)
}

// Creates a pipeline, but does not create a pipeline version.
// Call CreatePipelineVersion to create a pipeline version.
func (r *ResourceManager) CreatePipeline(p *model.Pipeline) (*model.Pipeline, error) {
	// Assign the default namespace if it is empty
	if p.Namespace == "" {
		p.Namespace = r.serverOptions["DefaultNamespace"].(string)
	}

	// Create a record in KFP DB (only pipelines table)
	newPipeline, err := r.pipelineStore.CreatePipeline(p)
	if err != nil {
		return nil, util.Wrap(err, "Failed to create a pipeline in PipelineStore")
	}

	newPipeline.Status = model.PipelineReady
	err = r.pipelineStore.UpdatePipelineStatus(
		newPipeline.UUID,
		newPipeline.Status,
	)
	if err != nil {
		return nil, util.Wrap(err, "Failed to update status of a pipeline after creation")
	}
	return newPipeline, nil
}

// Create a pipeline version.
// PipelineSpec is stored as a sting inside PipelineVersion in v2beta1.
func (r *ResourceManager) CreatePipelineVersion(pv *model.PipelineVersion) (*model.PipelineVersion, error) {
	// Extract pipeline id
	pipelineId := pv.PipelineId
	if len(pipelineId) == 0 {
		return nil, util.NewInvalidInputError("Failed to create a pipeline version due to missing pipeline id")
	}

	// Fetch pipeline spec
	pipelineSpecBytes, pipelineSpecURI, err := r.fetchTemplateFromPipelineVersion(pv)
	if err != nil {
		return nil, util.Wrap(err, "Failed to create a pipeline version as template is broken")
	}
	pv.PipelineSpec = string(pipelineSpecBytes)
	if pipelineSpecURI != "" {
		pv.PipelineSpecURI = pipelineSpecURI
	}

	// Create a template
	tmpl, err := template.New(pipelineSpecBytes)
	if err != nil {
		return nil, util.Wrap(err, "Failed to create a pipeline version due to template creation error")
	}
	if tmpl.IsV2() {
		pipeline, err := r.GetPipeline(pipelineId)
		if err != nil {
			return nil, util.Wrap(err, "Failed to create a pipeline version as parent pipeline was not found")
		}
		tmpl.OverrideV2PipelineName(pipeline.Name, pipeline.Namespace)
	}
	paramsJSON, err := tmpl.ParametersJSON()
	if err != nil {
		return nil, util.Wrap(err, "Failed to create a pipeline version due to error converting parameters to json")
	}
	pv.Parameters = paramsJSON
	pv.Status = model.PipelineVersionCreating
	pv.PipelineSpec = string(tmpl.Bytes())

	// Create a record in DB
	version, err := r.pipelineStore.CreatePipelineVersion(pv)
	if err != nil {
		return nil, util.Wrap(err, "Failed to create pipeline version in PipelineStore")
	}

	// TODO(gkcalat): consider removing this after v2beta1 GA if we adopt storing PipelineSpec in DB.
	// Store the pipeline file
	err = r.objectStore.AddFile(tmpl.Bytes(), r.objectStore.GetPipelineKey(fmt.Sprint(version.UUID)))
	if err != nil {
		return nil, util.Wrap(err, "Failed to create a pipeline version due to error saving PipelineSpec to ObjectStore")
	}

	// After pipeline version being created in DB and pipeline file being
	// saved in minio server, set this pieline version to status ready.
	version.Status = model.PipelineVersionReady
	err = r.pipelineStore.UpdatePipelineVersionStatus(version.UUID, version.Status)
	if err != nil {
		return nil, util.Wrapf(err, "Failed to change the status of a new pipeline version with id %v", version.UUID)
	}
	return version, nil
}

// Returns a pipeline.
func (r *ResourceManager) GetPipeline(pipelineId string) (*model.Pipeline, error) {
	if pipeline, err := r.pipelineStore.GetPipeline(pipelineId); err != nil {
		return nil, util.Wrapf(err, "Failed to get a pipeline with id %v", pipelineId)
	} else {
		return pipeline, nil
	}
}

// Returns a pipeline version.
func (r *ResourceManager) GetPipelineVersion(pipelineVersionId string) (*model.PipelineVersion, error) {
	if pipelineVersion, err := r.pipelineStore.GetPipelineVersion(pipelineVersionId); err != nil {
		return nil, util.Wrapf(err, "Failed to get a pipeline version with id %v", pipelineVersionId)
	} else {
		return pipelineVersion, nil
	}
}

// Returns a pipeline specified by name and namespace.
func (r *ResourceManager) GetPipelineByNameAndNamespace(name string, namespace string) (*model.Pipeline, error) {
	if pipeline, err := r.pipelineStore.GetPipelineByNameAndNamespace(name, namespace); err != nil {
		return nil, util.Wrapf(err, "Failed to get a pipeline named %v in namespace %v", name, namespace)
	} else {
		return pipeline, nil
	}
}

// TODO(gkcalat): consider removing after KFP v2 GA if users are not affected.
// Returns a pipeline specified by name and namespace using LEFT JOIN on SQL query.
// This could be more performant for a large number of pipeline versions.
func (r *ResourceManager) GetPipelineByNameAndNamespaceV1(name string, namespace string) (*model.Pipeline, *model.PipelineVersion, error) {
	if pipeline, pipelineVersion, err := r.pipelineStore.GetPipelineByNameAndNamespaceV1(name, namespace); err != nil {
		return nil, nil, util.Wrapf(err, "ResourceManager (v1beta1): Failed to get a pipeline named %v in namespace %v", name, namespace)
	} else {
		return pipeline, pipelineVersion, nil
	}
}

// Returns the latest template for a specified pipeline id.
func (r *ResourceManager) GetPipelineLatestTemplate(pipelineId string) ([]byte, error) {
	// Verify pipeline exists
	_, err := r.pipelineStore.GetPipeline(pipelineId)
	if err != nil {
		return nil, util.Wrap(err, "Failed to get the latest template as pipeline was not found")
	}

	// Get the latest pipeline version
	latestPipelineVersion, err := r.pipelineStore.GetLatestPipelineVersion(pipelineId)
	if err != nil {
		return nil, util.Wrap(err, "Failed to get the latest template for a pipeline")
	}

	// Fetch template []byte array
	if bytes, _, err := r.fetchTemplateFromPipelineVersion(latestPipelineVersion); err != nil {
		return nil, util.Wrapf(err, "Failed to get the latest template for pipeline with id %v", pipelineId)
	} else {
		return bytes, nil
	}
}

// Returns the latest pipeline version for a specified pipeline id.
func (r *ResourceManager) GetLatestPipelineVersion(pipelineId string) (*model.PipelineVersion, error) {
	// Verify pipeline exists
	_, err := r.pipelineStore.GetPipeline(pipelineId)
	if err != nil {
		return nil, util.Wrap(err, "Failed to get the latest pipeline version as pipeline was not found")
	}

	// Get the latest pipeline version
	latestPipelineVersion, err := r.pipelineStore.GetLatestPipelineVersion(pipelineId)
	if err != nil {
		return nil, util.Wrap(err, "Failed to get the latest pipeline version for a pipeline")
	}
	return latestPipelineVersion, nil
}

// Returns a template for a specified pipeline version id.
func (r *ResourceManager) GetPipelineVersionTemplate(pipelineVersionId string) ([]byte, error) {
	// Verify pipeline version exist
	pipelineVersion, err := r.pipelineStore.GetPipelineVersion(pipelineVersionId)
	if err != nil {
		return nil, util.Wrapf(err, "Failed to get pipeline version template as pipeline version id %v was not found", pipelineVersionId)
	}

	// Fetch template []byte array
	if bytes, _, err := r.fetchTemplateFromPipelineVersion(pipelineVersion); err != nil {
		return nil, util.Wrapf(err, "Failed to get a template for pipeline version with id %v", pipelineVersionId)
	} else {
		return bytes, nil
	}
}

// Returns a list of pipelines.
func (r *ResourceManager) ListPipelines(filterContext *model.FilterContext, opts *list.Options) ([]*model.Pipeline, int, string, error) {
	pipelines, total_size, nextPageToken, err := r.pipelineStore.ListPipelines(filterContext, opts)
	if err != nil {
		err = util.Wrapf(err, "Failed to list pipelines with context %v, options %v", filterContext, opts)
	}
	return pipelines, total_size, nextPageToken, err
}

// TODO(gkcalat): consider removing after KFP v2 GA if users are not affected.
// Returns a list of pipelines using LEFT JOIN on SQL query.
// This could be more performant for a large number of pipeline versions.
func (r *ResourceManager) ListPipelinesV1(filterContext *model.FilterContext, opts *list.Options) ([]*model.Pipeline, []*model.PipelineVersion, int, string, error) {
	pipelines, pipelineVersions, total_size, nextPageToken, err := r.pipelineStore.ListPipelinesV1(filterContext, opts)
	if err != nil {
		err = util.Wrapf(err, "ResourceManager (v1beta1): Failed to list pipelines with context %v, options %v", filterContext, opts)
	}
	return pipelines, pipelineVersions, total_size, nextPageToken, err
}

// Returns a list of pipeline versions.
func (r *ResourceManager) ListPipelineVersions(pipelineId string, opts *list.Options) ([]*model.PipelineVersion, int, string, error) {
	pipelineVersions, total_size, nextPageToken, err := r.pipelineStore.ListPipelineVersions(pipelineId, opts)
	if err != nil {
		err = util.Wrapf(err, "Failed to list pipeline versions with pipeline id %v, options %v", pipelineId, opts)
	}
	return pipelineVersions, total_size, nextPageToken, err
}

// Updates the status of a pipeline.
func (r *ResourceManager) UpdatePipelineStatus(pipelineId string, status model.PipelineStatus) error {
	err := r.pipelineStore.UpdatePipelineStatus(pipelineId, status)
	if err != nil {
		return util.Wrapf(err, "Failed to update the status of pipeline id %v to %v", pipelineId, status)
	}
	return nil
}

// Updates the status of a pipeline version.
func (r *ResourceManager) UpdatePipelineVersionStatus(pipelineVersionId string, status model.PipelineVersionStatus) error {
	err := r.pipelineStore.UpdatePipelineVersionStatus(pipelineVersionId, status)
	if err != nil {
		return util.Wrapf(err, "Failed to update the status of pipeline version id %v to %v", pipelineVersionId, status)
	}
	return nil
}

// Deletes a pipeline that does not have any pipeline versions. Does not delete pipeline spec.
// This has changed the behavior in v2beta1.
func (r *ResourceManager) DeletePipeline(pipelineId string) error {
	// Check if pipeline exists
	_, err := r.pipelineStore.GetPipeline(pipelineId)
	if err != nil {
		return util.Wrapf(err, "Failed to delete pipeline with id %v as it was not found", pipelineId)
	}

	// Check if it has no pipeline versions in Ready state
	latestPipelineVersion, err := r.pipelineStore.GetLatestPipelineVersion(pipelineId)
	if latestPipelineVersion != nil {
		return util.NewInvalidInputError("Failed to delete pipeline with id %v as it has existing pipeline versions (e.g. %v)", pipelineId, latestPipelineVersion.UUID)
	} else if err.(*util.UserError).ExternalStatusCode() != codes.NotFound {
		return util.Wrapf(err, "Failed to delete pipeline with id %v as it failed to check existing pipeline versions", pipelineId)
	}

	// Mark pipeline as deleting so it's not visible to user.
	err = r.pipelineStore.UpdatePipelineStatus(pipelineId, model.PipelineDeleting)
	if err != nil {
		return util.Wrapf(err, "Failed to change the status of pipeline id %v to DELETING", pipelineId)
	}

	// Delete a pipeline.
	err = r.pipelineStore.DeletePipeline(pipelineId)
	if err != nil {
		return util.Wrapf(err, "Failed to delete pipeline DB entry for pipeline id %v", pipelineId)
	}
	return nil
}

// Deletes a pipeline version and the corresponding PipelineSpec.
func (r *ResourceManager) DeletePipelineVersion(pipelineVersionId string) error {
	// Check if pipeline version exists
	pipelineVersion, err := r.pipelineStore.GetPipelineVersion(pipelineVersionId)
	if err != nil {
		return util.Wrapf(err, "Failed to delete pipeline version with id %v as it was not found", pipelineVersionId)
	}

	// Mark pipeline as deleting so it's not visible to user.
	err = r.pipelineStore.UpdatePipelineVersionStatus(pipelineVersionId, model.PipelineVersionDeleting)
	if err != nil {
		return util.Wrapf(err, "Failed to change the status of pipeline version id %v to DELETING", pipelineVersionId)
	}

	// Delete pipeline spec file and DB entry.
	// Not fail the request if this step failed. A background run will do the cleanup.
	// https://github.com/kubeflow/pipelines/issues/388
	// TODO(jingzhang36): For now (before exposing version API), we have only 1
	// file with both pipeline and version pointing to it;  so it is ok to do
	// the deletion as follows. After exposing version API, we can have multiple
	// versions and hence multiple files, and we shall improve performance by
	// either using async deletion in order for this method to be non-blocking
	// or or exploring other performance optimization tools provided by gcs.
	//
	// TODO(gkcalat): consider removing this if we switch to storing PipelineSpec in DB.
	// DeleteObject always responds with http '204' even for
	// objects which do not exist. The err below will be nil.
	//
	// Delete based on pipeline spec URI
	pipelineSpecRemoved := false
	var osErr error
	err = r.objectStore.DeleteFile(pipelineVersion.PipelineSpecURI)
	if err != nil {
		glog.Errorf("%v", util.Wrapf(err, "Failed to delete pipeline spec for pipeline version id %v with URI %v", pipelineVersionId, pipelineVersion.PipelineSpecURI))
		osErr = util.Wrapf(err, "Failed to delete pipeline spec for pipeline version id %v with URI %v", pipelineVersionId, pipelineVersion.PipelineSpecURI)
	} else {
		pipelineSpecRemoved = true
	}
	// Delete based on pipeline version id
	err = r.objectStore.DeleteFile(r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersionId)))
	if err != nil {
		glog.Errorf("%v", util.Wrapf(err, "Failed to delete pipeline spec for pipeline version id %v", pipelineVersionId))
		err = util.Wrapf(err, "Failed to delete pipeline spec for pipeline version id %v", pipelineVersionId)
		osErr = util.Wrap(osErr, err.Error())
	} else {
		pipelineSpecRemoved = true
	}
	// Delete based on pipeline id
	err = r.objectStore.DeleteFile(r.objectStore.GetPipelineKey(fmt.Sprint(pipelineVersion.PipelineId)))
	if err != nil {
		glog.Errorf("%v", util.Wrapf(err, "Failed to delete pipeline spec for pipeline version id %v using pipeline id %v", pipelineVersionId, pipelineVersion.PipelineId))
		err = util.Wrapf(err, "Failed to delete pipeline spec for pipeline version id %v using pipeline id %v", pipelineVersionId, pipelineVersion.PipelineId)
		osErr = util.Wrap(osErr, err.Error())
	} else {
		pipelineSpecRemoved = true
	}
	if !pipelineSpecRemoved {
		return util.Wrap(osErr, "Failed to delete a pipeline spec")
	}
	// Delete the DB entry
	err = r.pipelineStore.DeletePipelineVersion(pipelineVersionId)
	if err != nil {
		glog.Errorf("%v", util.Wrapf(err, "Failed to delete a DB entry for pipeline version id %v", pipelineVersionId))
		return util.Wrapf(err, "Failed to delete a DB entry for pipeline version id %v", pipelineVersionId)
	}
	return nil
}

// Creates a run and schedule a workflow CR.
// Note: when creating a run from a manifest, this triggers creation of
// a new pipeline and pipeline version that share the name, description, and namespace.
// If run.ExperimentId is not specified, it is set to the default experiment.
// If run.Namespace  is no specified, it gets inferred from the parent experiment.
// Manifest's namespace gets overwritten with the run.Namespace.
// Creating a run from recurring run prioritizes recurring run's pipeline spec over the run's one.
func (r *ResourceManager) CreateRun(ctx context.Context, run *model.Run) (*model.Run, error) {
	if run.ExperimentId == "" {
		defaultExperimentId, err := r.GetDefaultExperimentId()
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a run with empty experiment id. Specify experiment id for the run or check if the default experiment exists")
		}
		run.ExperimentId = defaultExperimentId
	}

	// Validate namespace
	if run.Namespace == "" {
		namespace, err := r.GetNamespaceFromExperimentId(run.ExperimentId)
		if err != nil {
			return nil, util.Wrap(err, "Failed to create a run")
		}
		run.Namespace = namespace
	}
	if common.IsMultiUserMode() {
		if run.Namespace == "" {
			return nil, util.NewInternalServerError(util.NewInvalidInputError("Run cannot have an empty namespace in multi-user mode"), "Failed to create a run")
		}
	}
	if err := r.ValidateExperimentNamespace(run.ExperimentId, run.Namespace); err != nil {
		return nil, util.Wrap(err, "Failed to create a run")
	}

	// Validate pipeline version id
	var pipelineVersion *model.PipelineVersion
	if run.PipelineSpec.PipelineVersionId != "" {
		pv, err := r.GetPipelineVersion(run.PipelineSpec.PipelineVersionId)
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a run with an invalid pipeline version id: %v", run.PipelineSpec.PipelineVersionId)
		}
		pipelineVersion = pv
	} else if run.PipelineSpec.PipelineId != "" {
		pv, err := r.GetLatestPipelineVersion(run.PipelineSpec.PipelineId)
		if err != nil {
			return nil, util.Wrap(err, "Failed to create a run with an empty pipeline version id")
		}
		pipelineVersion = pv
	} else if run.PipelineSpec.PipelineSpecManifest != "" || run.PipelineSpec.WorkflowSpecManifest != "" {
		manifest := run.PipelineSpec.PipelineSpecManifest
		if manifest == "" {
			manifest = run.PipelineSpec.WorkflowSpecManifest
		}
		pipeline := &model.Pipeline{
			Name:        run.DisplayName,
			Description: run.Description,
			Namespace:   run.Namespace,
		}
		pipeline, err := r.CreatePipeline(pipeline)
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a run with pipeline spec manifests due to error creating a new pipeline")
		}
		pipelineVersion = &model.PipelineVersion{
			PipelineId:   pipeline.UUID,
			Name:         run.DisplayName,
			PipelineSpec: manifest,
			Description:  run.Description,
		}
		pipelineVersion, err = r.CreatePipelineVersion(pipelineVersion)
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a run with pipeline spec manifests due to error creating a new pipeline version")
		}
	} else if run.PipelineSpec.PipelineName != "" {
		resourceNames := common.ParseResourceIdsFromFullName(run.PipelineSpec.PipelineName)
		if resourceNames["PipelineVersionId"] == "" && resourceNames["PipelineId"] == "" {
			return nil, util.Wrap(util.NewInvalidInputError("Pipeline spec source is missing"), "Failed to create a run with an empty pipeline spec source")
		}
		if resourceNames["PipelineVersionId"] != "" {
			pv, err := r.GetPipelineVersion(resourceNames["PipelineVersionId"])
			if err != nil {
				return nil, util.Wrap(err, "Failed to create a run with an empty pipeline spec source")
			}
			pipelineVersion = pv
		} else {
			pv, err := r.GetLatestPipelineVersion(resourceNames["PipelineId"])
			if err != nil {
				return nil, util.Wrap(err, "Failed to create a run with an empty pipeline spec source")
			}
			pipelineVersion = pv
		}
	} else {
		return nil, util.Wrap(util.NewInvalidInputError("Pipeline spec source is missing"), "Failed to create a run with an empty pipeline spec source")
	}
	run.PipelineSpec.PipelineId = pipelineVersion.PipelineId
	run.PipelineSpec.PipelineVersionId = pipelineVersion.UUID
	run.PipelineSpec.PipelineName = pipelineVersion.Name
	run.PipelineSpec.PipelineSpecManifest = pipelineVersion.PipelineSpec

	// Get manifest from either of the two places:
	// (1) raw manifest in pipeline_spec
	// (2) pipeline version in resource_references
	// And the latter takes priority over the former when the manifest is from pipeline_spec.pipeline_id
	// workflow/pipeline manifest and pipeline id/version will not exist at the same time, guaranteed by the validation phase
	manifestBytes, err := r.fetchTemplateFromPipelineSpec(&run.PipelineSpec)
	if err != nil {
		tempBytes, err2 := r.fetchTemplateFromPipelineVersionId(run.PipelineSpec.PipelineId)
		if err2 != nil {
			return nil, util.Wrap(util.Wrap(err, err2.Error()), "Failed to create a run with an empty pipeline spec manifest")
		}
		manifestBytes = tempBytes
	}
	run.PipelineSpec.PipelineSpecManifest = string(manifestBytes)

	// TODO(gkcalat): consider changing the flow. Other resource UUIDs are assigned by their respective stores (DB).
	// Proposed flow:
	// 1. Create an entry and assign creation timestamp and uuid.
	// 2. Create a workflow CR.
	// 3. Update a record in the DB with scheduled timestamp, state, etc.
	// 4. Persistance agent will call apiserver to update the records later.
	if run.UUID == "" {
		uuid, err := r.uuid.NewRandom()
		if err != nil {
			return nil, util.NewInternalServerError(err, "Failed to generate run ID")
		}
		run.UUID = uuid.String()
	}
	run.RunDetails.CreatedAtInSec = r.time.Now().Unix()

	tmpl, err := template.New(manifestBytes)
	if err != nil {
		return nil, err
	}
	runWorkflowOptions := template.RunWorkflowOptions{
		RunId: run.UUID,
		RunAt: run.RunDetails.CreatedAtInSec,
	}
	executionSpec, err := tmpl.RunWorkflow(run, runWorkflowOptions)
	if err != nil {
		return nil, util.NewInternalServerError(err, "failed to generate the ExecutionSpec")
	}
	err = executionSpec.Validate(false, false)
	if err != nil {
		return nil, util.NewInternalServerError(err, "Failed to validate workflow for (%+v)", executionSpec.ExecutionName())
	}
	executionSpec.SetExecutionNamespace(run.Namespace)

	// Create argo workflow CR resource
	newExecSpec, err := r.getWorkflowClient(run.Namespace).Create(ctx, executionSpec, v1.CreateOptions{})
	if err != nil {
		return nil, util.NewInternalServerError(err, "Failed to create a workflow for (%s)", executionSpec.ExecutionName())
	}

	// Update the run with the new scheduled workflow
	run.K8SName = newExecSpec.ExecutionName()
	// run.Namespace = newExecSpec.ExecutionNamespace()
	run.ServiceAccount = newExecSpec.ServiceAccount()
	run.RunDetails.State = model.RuntimeState(string(newExecSpec.ExecutionStatus().Condition())).ToV2()
	run.RunDetails.Conditions = string(run.RunDetails.State.ToV1())
	// TODO(gkcalat): consider to avoid updating runtime manifest at create time and let
	// persistence agent update the runtime data.
	if tmpl.GetTemplateType() == template.V1 && run.RunDetails.WorkflowRuntimeManifest == "" {
		run.RunDetails.WorkflowRuntimeManifest = newExecSpec.ToStringForStore()
	}

	// Assign the scheduled at time
	if run.RunDetails.ScheduledAtInSec == 0 {
		// if there is no scheduled time, then we assume this run is scheduled at the same time it is created
		run.RunDetails.ScheduledAtInSec = run.RunDetails.CreatedAtInSec
	}

	newRun, err := r.runStore.CreateRun(run)
	if err != nil {
		return nil, util.Wrap(err, "Failed to create a run")
	}
	return newRun, nil
}

// Fetches a run with a given id.
func (r *ResourceManager) GetRun(runId string) (*model.Run, error) {
	run, err := r.runStore.GetRun(runId)
	if err != nil {
		return nil, util.Wrapf(err, "Failed to fetch run %v", runId)
	}
	return run, nil
}

// Fetches runs with a given set of filtering and listing options.
func (r *ResourceManager) ListRuns(filterContext *model.FilterContext, opts *list.Options) ([]*model.Run, int, string, error) {
	runs, totalSize, nextPageToken, err := r.runStore.ListRuns(filterContext, opts)
	if err != nil {
		return nil, 0, "", util.Wrap(err, "Failed to list runs")
	}
	return runs, totalSize, nextPageToken, nil
}

// Archives a run with a given id.
func (r *ResourceManager) ArchiveRun(runId string) error {
	if err := r.runStore.ArchiveRun(runId); err != nil {
		return util.Wrapf(err, "Failed to archive run %v", runId)
	}
	return nil
}

// Un-archives a run with a given id.
func (r *ResourceManager) UnarchiveRun(runId string) error {
	run, err := r.GetRun(runId)
	if err != nil {
		return util.Wrapf(err, "Failed to unarchive run %v as it does not exist", runId)
	}
	if run.ExperimentId == "" {
		experimentRef, err := r.resourceReferenceStore.GetResourceReference(runId, model.RunResourceType, model.ExperimentResourceType)
		if err != nil {
			return util.Wrapf(err, "Failed to unarchive run %v due to resource references fetching error", runId)
		}
		run.ExperimentId = experimentRef.ReferenceUUID
	}

	experiment, err := r.GetExperiment(run.ExperimentId)
	if err != nil {
		return util.Wrapf(err, "Failed to unarchive run %v due to experiment fetching error", runId)
	}
	if experiment.StorageState.ToV2() == model.StorageStateArchived {
		return util.NewFailedPreconditionError(
			errors.New("Unarchive the experiment first to allow the run to be restored"),
			fmt.Sprintf("Failed to unarchive run %v as experiment %v must be un-archived first", runId, run.ExperimentId),
		)
	}
	if err := r.runStore.UnarchiveRun(runId); err != nil {
		return util.Wrapf(err, "Failed to unarchive run %v", runId)
	}
	return nil
}

// Terminates a workflow by setting its activeDeadlineSeconds to 0.
func TerminateWorkflow(ctx context.Context, wfClient util.ExecutionInterface, name string) error {
	patchObj := map[string]interface{}{
		"spec": map[string]interface{}{
			"activeDeadlineSeconds": 0,
		},
	}
	patch, err := json.Marshal(patchObj)
	if err != nil {
		return util.NewInternalServerError(err, "Failed to terminate workflow %s due to error parsing the patch", name)
	}
	var operation = func() error {
		_, err = wfClient.Patch(ctx, name, types.MergePatchType, patch, v1.PatchOptions{})
		return util.Wrapf(err, "Failed to terminate workflow %s due to patching error", name)
	}
	var backoffPolicy = backoff.WithMaxRetries(backoff.NewConstantBackOff(100), 10)
	err = backoff.Retry(operation, backoffPolicy)
	if err != nil {
		return util.Wrapf(err, "Failed to terminate workflow %s due to patching error after multiple retries", name)
	}
	return nil
}

// Terminates a running run and the corresponding workflow.
func (r *ResourceManager) TerminateRun(ctx context.Context, runId string) error {
	run, err := r.GetRun(runId)
	if err != nil {
		return util.Wrapf(err, "Failed to terminate run %s due to error fetching the run", runId)
	}
	// TODO(gkcalat): consider using run.Namespace after migration logic will be available.
	namespace, err := r.getNamespaceFromRunId(runId)
	if err != nil {
		return util.Wrapf(err, "Failed to terminate run %s due to error fetching its namespace", runId)
	}

	err = r.runStore.TerminateRun(runId)
	if err != nil {
		return util.Wrapf(err, "Failed to terminate run %s", runId)
	}

	err = TerminateWorkflow(ctx, r.getWorkflowClient(namespace), run.K8SName)
	if err != nil {
		return util.NewInternalServerError(err, "Failed to terminate run %s due to error terminating its workflow", runId)
	}
	return nil
}

// Retries a run given its id.
func (r *ResourceManager) RetryRun(ctx context.Context, runId string) error {
	run, err := r.GetRun(runId)
	if err != nil {
		return util.Wrapf(err, "Failed to retry run %s due to error fetching the run", runId)
	}
	// TODO(gkcalat): consider using run.Namespace after migration logic will be available.
	namespace, err := r.getNamespaceFromRunId(runId)
	if err != nil {
		return util.Wrapf(err, "Failed to retry run %s due to error fetching its namespace", runId)
	}

	if run.PipelineSpec.WorkflowSpecManifest != "" && run.RunDetails.WorkflowRuntimeManifest == "" {
		return util.NewBadRequestError(util.NewInvalidInputError("Workflow manifest cannot be empty"), "Failed to retry run %s due to error fetching workflow manifest", runId)
	}
	workflowManifest := run.RunDetails.WorkflowRuntimeManifest
	if workflowManifest == "" {
		workflowManifest = run.PipelineSpec.WorkflowSpecManifest
	}
	execSpec, err := util.NewExecutionSpecJSON(util.ArgoWorkflow, []byte(workflowManifest))
	if err != nil {
		return util.NewInternalServerError(err, "Failed to retry run %s due to error parsing the workflow manifest", runId)
	}

	if err := execSpec.Decompress(); err != nil {
		return util.NewInternalServerError(err, "Failed to retry run %s due to error decompressing execution spec", runId)
	}

	if err := execSpec.CanRetry(); err != nil {
		return util.NewInternalServerError(err, "Failed to retry run %s as it does not allow reties", runId)
	}

	newExecSpec, podsToDelete, err := execSpec.GenerateRetryExecution()
	if err != nil {
		return util.Wrapf(err, "Failed to retry run %s", runId)
	}

	if err = deletePods(ctx, r.k8sCoreClient, podsToDelete, namespace); err != nil {
		return util.NewInternalServerError(err, "Failed to retry run %s due to error cleaning up the failed pods from the previous attempt", runId)
	}

	// First try to update workflow
	// If fail to get the workflow, return error.
	latestWorkflow, updateError := r.getWorkflowClient(namespace).Get(ctx, newExecSpec.ExecutionName(), v1.GetOptions{})
	if updateError == nil {
		// Update the workflow's resource version to latest.
		newExecSpec.SetVersion(latestWorkflow.Version())
		_, updateError = r.getWorkflowClient(namespace).Update(ctx, newExecSpec, v1.UpdateOptions{})
	}
	if updateError != nil {
		// Remove resource version
		newExecSpec.SetVersion("")
		newCreatedWorkflow, createError := r.getWorkflowClient(namespace).Create(ctx, newExecSpec, v1.CreateOptions{})
		if createError != nil {
			return util.Wrap(
				util.NewInternalServerError(updateError, "Failed to retry run %s due to error updating the old workflow", runId),
				util.NewInternalServerError(createError, "Failed to retry run %s due to error creating a new workflow", runId).Error(),
			)
		}
		newExecSpec = newCreatedWorkflow
	}
	condition := string(newExecSpec.ExecutionStatus().Condition())
	state := model.RuntimeState(condition).ToV2().ToString()
	err = r.runStore.UpdateRun(runId, condition, 0, newExecSpec.ToStringForStore(), state)
	if err != nil {
		return util.NewInternalServerError(err, "Failed to retry run %s due to error updating entry", runId)
	}
	return nil
}

// Deletes a run entry with a given id.
func (r *ResourceManager) DeleteRun(ctx context.Context, runId string) error {
	run, err := r.GetRun(runId)
	if err != nil {
		return util.Wrapf(err, "Failed to delete run %v as it does not exist", runId)
	}
	if run.Namespace == "" {
		namespace, err := r.GetNamespaceFromExperimentId(run.ExperimentId)
		if err != nil {
			return util.Wrapf(err, "Failed to delete a run %v due to namespace fetching error", runId)
		}
		run.Namespace = namespace
	}
	err = r.getWorkflowClient(run.Namespace).Delete(ctx, run.K8SName, v1.DeleteOptions{})
	if err != nil {
		// API won't need to delete the workflow CR
		// once persistent agent sync the state to DB and set TTL for it.
		glog.Warningf("Failed to delete run %v. Error: %v", run.K8SName, err.Error())
	}
	err = r.runStore.DeleteRun(runId)
	if err != nil {
		return util.Wrapf(err, "Failed to delete a run %v", runId)
	}
	return nil
}

// Creates a task entry.
func (r *ResourceManager) CreateTask(t *model.Task) (*model.Task, error) {
	run, err := r.GetRun(t.RunId)
	if err != nil {
		return nil, util.Wrapf(err, "Failed to create a task for run %v", t.RunId)
	}
	if run.ExperimentId == "" {
		defaultExperimentId, err := r.GetDefaultExperimentId()
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a task in run %v. Specify experiment id for the run or check if the default experiment exists", t.RunId)
		}
		run.ExperimentId = defaultExperimentId
	}

	// Validate namespace
	if t.Namespace == "" {
		namespace, err := r.GetNamespaceFromExperimentId(run.ExperimentId)
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a task in run %v", t.RunId)
		}
		t.Namespace = namespace
	}
	if common.IsMultiUserMode() {
		if t.Namespace == "" {
			return nil, util.NewInternalServerError(util.NewInvalidInputError("Task cannot have an empty namespace in multi-user mode"), "Failed to create a task in run %v", t.RunId)
		}
	}
	if err := r.ValidateExperimentNamespace(run.ExperimentId, t.Namespace); err != nil {
		return nil, util.Wrapf(err, "Failed to create a task in run %v", t.RunId)
	}

	newTask, err := r.taskStore.CreateTask(t)
	if err != nil {
		return nil, util.Wrapf(err, "Failed to create a task in run %v", t.RunId)
	}
	return newTask, nil
}

// Creates a task entry.
func (r *ResourceManager) GetTask(taskId string) (*model.Task, error) {
	task, err := r.taskStore.GetTask(taskId)
	if err != nil {
		return nil, util.Wrapf(err, "Failed to fetch task %v", taskId)
	}
	return task, nil
}

// Fetches tasks with a given set of filtering and listing options.
func (r *ResourceManager) ListTasks(filterContext *model.FilterContext, opts *list.Options) ([]*model.Task, int, string, error) {
	tasks, totalSize, nextPageToken, err := r.taskStore.ListTasks(filterContext, opts)
	if err != nil {
		return nil, 0, "", util.Wrap(err, "Failed to list tasks")
	}
	return tasks, totalSize, nextPageToken, nil
}

// Creates a run metric entry.
func (r *ResourceManager) ReportMetric(metric *model.RunMetric) error {
	err := r.runStore.CreateMetric(metric)
	if err != nil {
		return util.Wrap(err, "Failed to report a run metric")
	}
	return nil
}

// Fetches run metric entries for a given run id.
func (r *ResourceManager) GetRunMetrics(runId string) ([]*model.RunMetric, error) {
	metrics, err := r.runStore.GetMetrics(runId)
	if err != nil {
		return nil, util.Wrapf(err, "Failed to fetch run metrics for run %s", runId)
	}
	return metrics, nil
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

// Fetches execution logs and writes to the destination.
// 1. Attempts to read logs directly from pod.
// 2. Attempts to read logs from archive if reading from pod fails.
func (r *ResourceManager) ReadLog(ctx context.Context, runId string, nodeId string, follow bool, dst io.Writer) error {
	run, err := r.GetRun(runId)
	if err != nil {
		return util.NewBadRequestError(err, "Failed to read logs for run %v due to run fetching error", runId)
	}
	// TODO(gkcalat): consider using run.Namespace after migration logic will be available.
	namespace, err := r.getNamespaceFromRunId(runId)
	if err != nil {
		return util.NewBadRequestError(err, "Failed to read logs for run %v due to namespace fetching error", runId)
	}
	err = r.readRunLogFromPod(ctx, namespace, nodeId, follow, dst)
	if err != nil && r.logArchive != nil {
		err = r.readRunLogFromArchive(run.WorkflowRuntimeManifest, nodeId, dst)
		if err != nil {
			return util.NewBadRequestError(err, "Failed to read logs for run %v", runId)
		}
	}
	if err != nil {
		return util.NewBadRequestError(err, "Failed to read logs for run %v", runId)
	}
	return nil
}

// Fetches execution logs from a pod.
func (r *ResourceManager) readRunLogFromPod(ctx context.Context, namespace string, nodeId string, follow bool, dst io.Writer) error {
	logOptions := corev1.PodLogOptions{
		Container:  "main",
		Timestamps: false,
		Follow:     follow,
	}

	req := r.k8sCoreClient.PodClient(namespace).GetLogs(nodeId, &logOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			glog.Errorf("Failed to read logs from pod %v: %v", nodeId, err)
		}
		return util.NewInternalServerError(err, "Failed to read logs from pod %v due to error opening log stream", nodeId)
	}
	defer podLogs.Close()

	_, err = io.Copy(dst, podLogs)
	if err != nil && err != io.EOF {
		return util.NewInternalServerError(err, "Failed to read logs from pod %v due to error in streaming the log", nodeId)
	}
	return nil
}

// Fetches execution logs from a archived pod logs.
func (r *ResourceManager) readRunLogFromArchive(workflowManifest string, nodeId string, dst io.Writer) error {
	if workflowManifest == "" {
		return util.NewInternalServerError(util.NewInvalidInputError("Runtime workflow manifest cannot empty"), "Failed to read logs from archive %v due to empty runtime workflow manifest", nodeId)
	}

	execSpec, err := util.NewExecutionSpecJSON(util.ArgoWorkflow, []byte(workflowManifest))
	if err != nil {
		return util.NewInternalServerError(err, "Failed to read logs from archive %v due error reading execution spec", nodeId)
	}

	logPath, err := r.logArchive.GetLogObjectKey(execSpec, nodeId)
	if err != nil {
		return util.NewInternalServerError(err, "Failed to read logs from archive %v", nodeId)
	}

	logContent, err := r.objectStore.GetFile(logPath)
	if err != nil {
		return util.NewInternalServerError(err, "Failed to read logs from archive %v due to error fetching the log file", nodeId)
	}

	err = r.logArchive.CopyLogFromArchive(logContent, dst, archive.ExtractLogOptions{LogFormat: archive.LogFormatText, Timestamps: false})
	if err != nil {
		return util.NewInternalServerError(err, "Failed to read logs from archive %v due to error copying the log file", nodeId)
	}
	return nil
}

// Creates a recurring run.
// Note: when creating a recurring run from a manifest, this triggers creation of
// a new pipeline and pipeline version that share the name, description, and namespace.
// Manifest's namespace gets overwritten with the job.Namespace if the later is non-empty.
// Otherwise, job.Namespace gets overwritten by the manifest.
func (r *ResourceManager) CreateJob(ctx context.Context, job *model.Job) (*model.Job, error) {
	if job.ExperimentId == "" {
		defaultExperimentId, err := r.GetDefaultExperimentId()
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a recurring run with empty experiment id. Specify experiment id for the recurring run or check if the default experiment exists")
		}
		job.ExperimentId = defaultExperimentId
	}
	// Validate namespace
	if job.Namespace == "" {
		namespace, err := r.GetNamespaceFromExperimentId(job.ExperimentId)
		if err != nil {
			return nil, util.Wrap(err, "Failed to create a recurring run")
		}
		job.Namespace = namespace
	}
	if common.IsMultiUserMode() {
		if job.Namespace == "" {
			return nil, util.NewInternalServerError(util.NewInvalidInputError("Recurring run cannot have an empty namespace in multi-user mode"), "Failed to create a recurring run")
		}
	}
	if err := r.ValidateExperimentNamespace(job.ExperimentId, job.Namespace); err != nil {
		return nil, util.Wrap(err, "Failed to create a recurring run")
	}

	// Validate pipeline version id
	var pipelineVersion *model.PipelineVersion
	if job.PipelineSpec.PipelineVersionId != "" {
		pv, err := r.GetPipelineVersion(job.PipelineSpec.PipelineVersionId)
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a recurring run with an invalid pipeline version id: %v", job.PipelineSpec.PipelineVersionId)
		}
		pipelineVersion = pv
	} else if job.PipelineSpec.PipelineId != "" {
		pv, err := r.GetLatestPipelineVersion(job.PipelineSpec.PipelineId)
		if err != nil {
			return nil, util.Wrap(err, "Failed to create a recurring run with an empty pipeline version id")
		}
		pipelineVersion = pv
	} else if job.PipelineSpec.PipelineSpecManifest != "" || job.PipelineSpec.WorkflowSpecManifest != "" {
		manifest := job.PipelineSpec.PipelineSpecManifest
		if manifest == "" {
			manifest = job.PipelineSpec.WorkflowSpecManifest
		}
		pipeline := &model.Pipeline{
			Name:        job.DisplayName,
			Description: job.Description,
			Namespace:   job.Namespace,
		}
		pipeline, err := r.CreatePipeline(pipeline)
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a recurring run with pipeline spec manifests due to error creating a new pipeline")
		}
		pipelineVersion = &model.PipelineVersion{
			PipelineId:   pipeline.UUID,
			Name:         job.DisplayName,
			PipelineSpec: manifest,
			Description:  job.Description,
		}
		pipelineVersion, err = r.CreatePipelineVersion(pipelineVersion)
		if err != nil {
			return nil, util.Wrapf(err, "Failed to create a recurring run with pipeline spec manifests due to error creating a new pipeline version")
		}
	} else if job.PipelineSpec.PipelineName != "" {
		resourceNames := common.ParseResourceIdsFromFullName(job.PipelineSpec.PipelineName)
		if resourceNames["PipelineVersionId"] == "" && resourceNames["PipelineId"] == "" {
			return nil, util.Wrap(util.NewInvalidInputError("Pipeline spec source is missing"), "Failed to create a recurring run with an empty pipeline spec source")
		}
		if resourceNames["PipelineVersionId"] != "" {
			pv, err := r.GetPipelineVersion(resourceNames["PipelineVersionId"])
			if err != nil {
				return nil, util.Wrap(err, "Failed to create a recurring run with an empty pipeline spec source")
			}
			pipelineVersion = pv
		} else {
			pv, err := r.GetLatestPipelineVersion(resourceNames["PipelineId"])
			if err != nil {
				return nil, util.Wrap(err, "Failed to create a recurring run with an empty pipeline spec source")
			}
			pipelineVersion = pv
		}
	} else {
		return nil, util.Wrap(util.NewInvalidInputError("Pipeline spec source is missing"), "Failed to create a recurring run with an empty pipeline spec source")
	}
	job.PipelineSpec.PipelineId = pipelineVersion.PipelineId
	job.PipelineSpec.PipelineVersionId = pipelineVersion.UUID
	job.PipelineSpec.PipelineName = pipelineVersion.Name

	manifestBytes, err := r.fetchTemplateFromPipelineSpec(&job.PipelineSpec)
	if err != nil {
		tempBytes, err2 := r.fetchTemplateFromPipelineVersionId(job.PipelineSpec.PipelineId)
		if err2 != nil {
			return nil, util.Wrap(util.Wrap(err, err2.Error()), "Failed to create a recurring run with an empty pipeline spec manifest")
		}
		manifestBytes = tempBytes
	}
	job.PipelineSpec.PipelineSpecManifest = string(manifestBytes)

	tmpl, err := template.New(manifestBytes)
	if err != nil {
		return nil, util.Wrap(err, "Failed to create a recurring run during template creation")
	}

	// TODO(gkcalat): consider changing the flow. Other resource UUIDs are assigned by their respective stores (DB).
	// Convert modelJob into scheduledWorkflow.
	scheduledWorkflow, err := tmpl.ScheduledWorkflow(job)
	if err != nil {
		return nil, util.Wrap(err, "Failed to create a recurring run during scheduled workflow creation")
	}

	// Create a new ScheduledWorkflow at the ScheduledWorkflow client.
	newScheduledWorkflow, err := r.getScheduledWorkflowClient(job.Namespace).Create(ctx, scheduledWorkflow)
	if err != nil {
		return nil, util.Wrap(err, "Failed to create a recurring run during scheduling a workflow")
	}

	// Complete modelJob with info coming back from ScheduledWorkflow client.
	swf := util.NewScheduledWorkflow(newScheduledWorkflow)
	job.UUID = string(swf.UID)
	job.K8SName = swf.Name
	job.Namespace = swf.Namespace
	job.Conditions = model.StatusState(swf.ConditionSummary()).ToString()
	for _, modelRef := range job.ResourceReferences {
		modelRef.ResourceUUID = string(swf.UID)
	}

	// Get the service account
	serviceAccount := ""
	if swf.Spec.Workflow != nil {
		execSpec, err := util.ScheduleSpecToExecutionSpec(util.ArgoWorkflow, swf.Spec.Workflow)
		if err == nil {
			serviceAccount = execSpec.ServiceAccount()
		}
	}
	job.ServiceAccount = serviceAccount

	return r.jobStore.CreateJob(job)
}

// Fetches a recurring run with given id.
func (r *ResourceManager) GetJob(id string) (*model.Job, error) {
	return r.jobStore.GetJob(id)
}

// Enables or disables a recurring run with given id.
func (r *ResourceManager) ChangeJobMode(ctx context.Context, jobId string, enable bool) error {
	job, err := r.GetJob(jobId)
	if err != nil {
		return util.Wrapf(err, "Failed to change recurring run's mode to enable:%v. Check if recurring run %v exists", enable, jobId)
	}
	if enable {
		scheduledWorkflow, err := r.getScheduledWorkflowClient(job.Namespace).Get(ctx, job.K8SName, v1.GetOptions{})
		if err != nil {
			return util.NewInternalServerError(err, "Failed to enable recurring run %v. Check if the scheduled workflow exists", jobId)
		}
		if scheduledWorkflow == nil || string(scheduledWorkflow.UID) != jobId {
			return util.Wrapf(util.NewResourceNotFoundError("recurring run", job.K8SName), "Failed to enable recurring run %v. Check if its k8s resource exists", jobId)
		}
	}

	_, err = r.getScheduledWorkflowClient(job.Namespace).Patch(
		ctx,
		job.K8SName,
		types.MergePatchType,
		[]byte(fmt.Sprintf(`{"spec":{"enabled":%s}}`, strconv.FormatBool(enable))),
	)
	if err != nil {
		return util.NewInternalServerError(err, "Failed to change recurring run's %v mode to enable:%v", jobId, enable)
	}

	err = r.jobStore.ChangeJobMode(jobId, enable)
	if err != nil {
		return util.Wrapf(err, "Failed to change recurring run's %v mode to enable:%v", jobId, enable)
	}
	return nil
}

// Fetches recurring runs with given filtering and listing options.
func (r *ResourceManager) ListJobs(filterContext *model.FilterContext,
	opts *list.Options) (jobs []*model.Job, total_size int, nextPageToken string, err error) {
	return r.jobStore.ListJobs(filterContext, opts)
}

// Deletes a recurring run with given id.
func (r *ResourceManager) DeleteJob(ctx context.Context, jobId string) error {
	job, err := r.GetJob(jobId)
	if err != nil {
		return util.Wrapf(err, "Failed to delete recurring run %v. Check if exists", jobId)
	}

	err = r.getScheduledWorkflowClient(job.Namespace).Delete(ctx, job.K8SName, &v1.DeleteOptions{})
	if err != nil {
		if !util.IsNotFound(err) {
			return util.NewInternalServerError(err, "Failed to delete recurring run %v. Check if the scheduled workflow exists", jobId)
		}
		// The ScheduledWorkflow was not found.
		glog.Infof("Deleting recurring run '%v', but skipped deleting ScheduledWorkflow '%v' in namespace '%v' because it was not found", jobId, job.K8SName, job.Namespace)
		// Continue the execution, because we want to delete the
		// ScheduledWorkflow. We can skip deleting the ScheduledWorkflow
		// when it no longer exists.
	}
	err = r.jobStore.DeleteJob(jobId)
	if err != nil {
		return util.Wrapf(err, "Failed to delete recurring run %v", jobId)
	}
	return nil
}

// Updates a recurring run with a scheduled workflow CR.
func (r *ResourceManager) ReportScheduledWorkflowResource(swf *util.ScheduledWorkflow) error {
	// Verify the job exists
	if _, err := r.GetJob(string(swf.UID)); err != nil {
		return util.Wrapf(err, "Failed to report scheduled workflow due to error retrieving recurring run %s", string(swf.UID))
	}
	return r.jobStore.UpdateJob(swf)
}

// Reports a workflow CR.
// This is called by the persistence agent to update runs.
func (r *ResourceManager) ReportWorkflowResource(ctx context.Context, execSpec util.ExecutionSpec) error {
	objMeta := execSpec.ExecutionObjectMeta()
	execStatus := execSpec.ExecutionStatus()
	if _, ok := objMeta.Labels[util.LabelKeyWorkflowRunId]; !ok {
		// Skip reporting if the workflow doesn't have the run id label
		return util.NewInvalidInputError("Workflow[%s] missing the Run ID label", execSpec.ExecutionName())
	}
	runId := objMeta.Labels[util.LabelKeyWorkflowRunId]
	jobId := execSpec.ScheduledWorkflowUUIDAsStringOrEmpty()
	// TODO(gkcalat): consider adding namespace validation to catch mismatch in the namespaces and release resources.
	if len(execSpec.ExecutionNamespace()) == 0 {
		return util.NewInvalidInputError("Failed to report a workflow. Namespace is empty")
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
	state := model.RuntimeState(string(execStatus.Condition())).ToV2()
	if execSpec.IsTerminating() {
		state = model.RuntimeState(string(exec.ExecutionPhase(model.RunTerminatingConditionsV1))).ToV2()
	}
	// If run already exists, simply update it
	if _, err := r.GetRun(runId); err == nil {
		if updateError := r.runStore.UpdateRun(runId, string(state.ToV1()), execStatus.FinishedAt(), execSpec.ToStringForStore(), state.ToString()); updateError != nil {
			return util.Wrapf(err, "Failed to report a workflow for existing run %s during updating the run. Check if the run entry is corrupted", runId)
		}
	}
	if jobId == "" {
		// If a run doesn't have job ID, it's a one-time run created by Pipeline API server.
		// In this case the DB entry should already been created when argo workflow CR is created.
		// TODO(gkcalat): consider removing UpdateRUn call as it fail anyways
		if updateError := r.runStore.UpdateRun(runId, string(state.ToV1()), execStatus.FinishedAt(), execSpec.ToStringForStore(), state.ToString()); updateError != nil {
			if !util.IsUserErrorCodeMatch(updateError, codes.NotFound) {
				return util.Wrap(updateError, "Failed to update the run")
			}
			// Handle run not found in run store error.
			// To avoid letting the workflow leak for ever, we need to GC it when its record does not exist in KFP DB.
			glog.Errorf("Cannot find reported workflow name=%q namespace=%q runId=%q in run store. "+
				"Deleting the workflow to avoid resource leaking. "+
				"This can be caused by installing two KFP instances that try to manage the same workflows "+
				"or an unknown bug. If you encounter this, recommend reporting more details in https://github.com/kubeflow/pipelines/issues/6189",
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
		// TODO(gkcalat): consider adding manifest validation to catch mismatch, as runs should have the same pipeline spec as parent recurring run.
		// Try to fetch the job.
		existingJob, err := r.GetJob(jobId)
		if err != nil {
			return util.Wrapf(err, "Failed to report a workflow for run %s due to error retrieving recurring run %s", runId, jobId)
		}
		experimentId := existingJob.ExperimentId
		namespace := existingJob.Namespace
		pipelineSpec := existingJob.PipelineSpec
		pipelineSpec.WorkflowSpecManifest = execSpec.GetExecutionSpec().ToStringForStore()

		// Try to fetch experiment id from resource references if it is missing.
		if experimentId == "" {
			experimentRef, err := r.resourceReferenceStore.GetResourceReference(jobId, model.JobResourceType, model.ExperimentResourceType)
			if err != nil {
				return util.Wrapf(err, "Failed to retrieve the experiment ID for the job %v that created the run", jobId)
			}
			experimentId = experimentRef.ReferenceUUID
			if namespace == "" {
				if namespaceRef, err := r.resourceReferenceStore.GetResourceReference(jobId, model.JobResourceType, model.NamespaceResourceType); err == nil {
					namespace = namespaceRef.ReferenceUUID
				}
			}
		}
		if experimentId == "" {
			experimentId, err = r.GetDefaultExperimentId()
			if err != nil {
				return util.Wrapf(err, "Failed to report workflow for run %s. Fetching default experiment returned error. Check if you have experiment assigned for job %s", runId, jobId)
			}
		}
		// TODO(gkcalat): consider adding namespace validation to catch mismatch in the namespaces and release resources.
		if namespace == "" {
			namespace, err = r.GetNamespaceFromExperimentId(experimentId)
			if err != nil {
				return util.Wrapf(err, "Failed to report workflow for run %s. Fetching namespace for experiment %s returned error. Check if you have namespace assigned for job %s", runId, experimentId, jobId)
			}
		}
		if namespace == "" {
			namespace = execSpec.ExecutionNamespace()
		}
		// Scheduled time equals created time if it is not specified
		var scheduledTimeInSec int64
		if execSpec.ScheduledAtInSecOr0() == 0 {
			scheduledTimeInSec = objMeta.CreationTimestamp.Unix()
		} else {
			scheduledTimeInSec = execSpec.ScheduledAtInSecOr0()
		}
		run := &model.Run{
			UUID:           runId,
			ExperimentId:   experimentId,
			RecurringRunId: jobId,
			DisplayName:    execSpec.ExecutionName(),
			K8SName:        execSpec.ExecutionName(),
			StorageState:   model.StorageStateAvailable,
			Namespace:      namespace,
			PipelineSpec:   pipelineSpec,
			RunDetails: model.RunDetails{
				WorkflowRuntimeManifest: execSpec.ToStringForStore(),
				CreatedAtInSec:          objMeta.CreationTimestamp.Unix(),
				ScheduledAtInSec:        scheduledTimeInSec,
				FinishedAtInSec:         execStatus.FinishedAt(),
				Conditions:              string(state.ToV1()),
				State:                   state,
			},
		}
		_, err = r.runStore.CreateRun(run)
		if err != nil {
			return util.Wrapf(err, "Failed to report a workflow due to error creating run %s", runId)
		}
	}

	if execStatus.IsInFinalState() {
		err := addWorkflowLabel(ctx, r.getWorkflowClient(execSpec.ExecutionNamespace()), execSpec.ExecutionName(), util.LabelKeyWorkflowPersistedFinalState, "true")
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

// Adds a label for a workflow.
func addWorkflowLabel(ctx context.Context, wfClient util.ExecutionInterface, name string, labelKey string, labelValue string) error {
	patchObj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				labelKey: labelValue,
			},
		},
	}

	patch, err := json.Marshal(patchObj)
	if err != nil {
		return util.NewInternalServerError(err, "Unexpected error while marshalling a patch object")
	}

	var operation = func() error {
		_, err = wfClient.Patch(ctx, name, types.MergePatchType, patch, v1.PatchOptions{})
		return err
	}
	var backoffPolicy = backoff.WithMaxRetries(backoff.NewConstantBackOff(100), 10)
	err = backoff.Retry(operation, backoffPolicy)
	return err
}

// TODO(gkcalat): consider removing before v2beta1 GA as default version is deprecated. This requires changes to v1beta1 proto.
// Updates default pipeline version for a given pipeline.
// Supports v1beta1 behavior.
func (r *ResourceManager) UpdatePipelineDefaultVersion(pipelineId string, versionId string) error {
	return r.pipelineStore.UpdatePipelineDefaultVersion(pipelineId, versionId)
}
