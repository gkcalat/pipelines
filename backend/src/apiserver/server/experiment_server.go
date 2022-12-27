// Copyright 2018-2022 The Kubeflow Authors
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

	"github.com/golang/protobuf/ptypes/empty"
	apiv1beta1 "github.com/kubeflow/pipelines/backend/api/v1beta1/go_client"
	apiv2beta1 "github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/kubeflow/pipelines/backend/src/apiserver/common"
	"github.com/kubeflow/pipelines/backend/src/apiserver/model"
	"github.com/kubeflow/pipelines/backend/src/apiserver/resource"
	"github.com/kubeflow/pipelines/backend/src/common/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	authorizationv1 "k8s.io/api/authorization/v1"
)

// Metric variables. Please prefix the metric names with experiment_server_.
var (
	// Used to calculate the request rate.
	createExperimentRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "experiment_server_create_requests",
		Help: "The total number of CreateExperiment requests",
	})

	getExperimentRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "experiment_server_get_requests",
		Help: "The total number of GetExperiment requests",
	})

	listExperimentsV1Requests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "experiment_server_list_requests",
		Help: "The total number of ListExperimentsV1 requests",
	})

	deleteExperimentRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "experiment_server_delete_requests",
		Help: "The total number of DeleteExperiment requests",
	})

	archiveExperimentRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "experiment_server_archive_requests",
		Help: "The total number of ArchiveExperiment requests",
	})

	unarchiveExperimentRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "experiment_server_unarchive_requests",
		Help: "The total number of UnarchiveExperiment requests",
	})

	// TODO(jingzhang36): error count and success count.

	experimentCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "experiment_server_run_count",
		Help: "The current number of experiments in Kubeflow Pipelines instance",
	})
)

type ExperimentServerOptions struct {
	CollectMetrics bool
}

type ExperimentServer struct {
	resourceManager *resource.ResourceManager
	options         *ExperimentServerOptions
}

func (s *ExperimentServer) CreateExperimentV1(ctx context.Context, request *apiv1beta1.CreateExperimentRequest) (
	*apiv1beta1.Experiment, error) {
	if s.options.CollectMetrics {
		createExperimentRequests.Inc()
	}

	err := ValidateCreateExperimentRequestV1(request)
	if err != nil {
		return nil, util.Wrap(err, "Validate experiment request failed.")
	}

	resourceAttributes := &authorizationv1.ResourceAttributes{
		Namespace: GetNamespaceFromAPIResourceReferences(request.Experiment.ResourceReferences),
		Verb:      common.RbacResourceVerbCreate,
		Name:      request.Experiment.Name,
	}
	err = s.canAccessExperiment(ctx, "", resourceAttributes)
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}

	newExperiment, err := s.resourceManager.CreateExperiment(request.Experiment)
	if err != nil {
		return nil, util.Wrap(err, "Create experiment failed.")
	}

	if s.options.CollectMetrics {
		experimentCount.Inc()
	}
	return ToApiExperimentV1(newExperiment), nil
}

func (s *ExperimentServer) CreateExperiment(ctx context.Context, request *apiv2beta1.CreateExperimentRequest) (
	*apiv2beta1.Experiment, error) {
	if s.options.CollectMetrics {
		createExperimentRequests.Inc()
	}

	err := ValidateCreateExperimentRequest(request)
	if err != nil {
		return nil, util.Wrap(err, "Validate experiment request failed.")
	}

	resourceAttributes := &authorizationv1.ResourceAttributes{
		Namespace: request.Experiment.Namespace,
		Verb:      common.RbacResourceVerbCreate,
		Name:      request.Experiment.DisplayName,
	}
	err = s.canAccessExperiment(ctx, "", resourceAttributes)
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}

	newExperiment, err := s.resourceManager.CreateExperiment(request.Experiment)
	if err != nil {
		return nil, util.Wrap(err, "Create experiment failed.")
	}

	if s.options.CollectMetrics {
		experimentCount.Inc()
	}
	return ToApiExperiment(newExperiment), nil
}

func (s *ExperimentServer) GetExperimentV1(ctx context.Context, request *apiv1beta1.GetExperimentRequest) (
	*apiv1beta1.Experiment, error) {
	if s.options.CollectMetrics {
		getExperimentRequests.Inc()
	}

	err := s.canAccessExperiment(ctx, request.Id, &authorizationv1.ResourceAttributes{Verb: common.RbacResourceVerbGet})
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}

	experiment, err := s.resourceManager.GetExperiment(request.Id)
	if err != nil {
		return nil, util.Wrap(err, "Get experiment failed.")
	}
	return ToApiExperimentV1(experiment), nil
}

func (s *ExperimentServer) GetExperiment(ctx context.Context, request *apiv2beta1.GetExperimentRequest) (
	*apiv2beta1.Experiment, error) {
	if s.options.CollectMetrics {
		getExperimentRequests.Inc()
	}

	err := s.canAccessExperiment(ctx, request.ExperimentId, &authorizationv1.ResourceAttributes{Verb: common.RbacResourceVerbGet})
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}

	experiment, err := s.resourceManager.GetExperiment(request.ExperimentId)
	if err != nil {
		return nil, util.Wrap(err, "Get experiment failed.")
	}
	return ToApiExperiment(experiment), nil
}

func (s *ExperimentServer) ListExperimentsV1(ctx context.Context, request *apiv1beta1.ListExperimentsRequest) (
	*apiv1beta1.ListExperimentsResponse, error) {
	if s.options.CollectMetrics {
		listExperimentsV1Requests.Inc()
	}

	opts, err := validatedListOptions(&model.Experiment{}, request.PageToken, int(request.PageSize), request.SortBy, request.Filter)

	if err != nil {
		return nil, util.Wrap(err, "Failed to create list options")
	}

	filterContext, err := ValidateFilterV1(request.ResourceReferenceKey)
	if err != nil {
		return nil, util.Wrap(err, "Validating filter failed.")
	}

	refKey := filterContext.ReferenceKey
	if common.IsMultiUserMode() {
		if refKey == nil || refKey.Type != model.NamespaceResourceType {
			return nil, util.NewInvalidInputError("Invalid resource references for experiment. ListExperimentsV1 requires filtering by namespace.")
		}
		namespace := refKey.ID
		if len(namespace) == 0 {
			return nil, util.NewInvalidInputError("Invalid resource references for experiment. Namespace is empty.")
		}
		resourceAttributes := &authorizationv1.ResourceAttributes{
			Namespace: namespace,
			Verb:      common.RbacResourceVerbList,
		}
		err = s.canAccessExperiment(ctx, "", resourceAttributes)
		if err != nil {
			return nil, util.Wrap(err, "Failed to authorize with API")
		}
	} else {
		if refKey != nil && refKey.Type == model.NamespaceResourceType && len(refKey.ID) > 0 {
			return nil, util.NewInvalidInputError("In single-user mode, ListExperimentsV1 cannot filter by namespace.")
		}
		// In single user mode, apply filter with empty namespace for backward compatibile.
		filterContext = &model.FilterContext{
			ReferenceKey: &model.ReferenceKey{Type: model.NamespaceResourceType, ID: ""},
		}
	}

	experiments, total_size, nextPageToken, err := s.resourceManager.ListExperiments(filterContext, opts)
	if err != nil {
		return nil, util.Wrap(err, "List experiments failed.")
	}
	return &apiv1beta1.ListExperimentsResponse{
			Experiments:   ToApiExperimentsV1(experiments),
			TotalSize:     int32(total_size),
			NextPageToken: nextPageToken},
		nil
}

func (s *ExperimentServer) ListExperiments(ctx context.Context, request *apiv2beta1.ListExperimentsRequest) (
	*apiv2beta1.ListExperimentsResponse, error) {
	if s.options.CollectMetrics {
		listExperimentsV1Requests.Inc()
	}

	opts, err := validatedListOptions(&model.Experiment{}, request.PageToken, int(request.PageSize), request.SortBy, request.Filter)

	if err != nil {
		return nil, util.Wrap(err, "Failed to create list options")
	}

	filterContext := &model.FilterContext{}
	if common.IsMultiUserMode() {
		if request.Namespace == "" {
			return nil, util.NewInvalidInputError("Invalid ListExperiments request. No namespace provided in multi-user mode.")
		}
		resourceAttributes := &authorizationv1.ResourceAttributes{
			Namespace: request.Namespace,
			Verb:      common.RbacResourceVerbList,
		}
		err = s.canAccessExperiment(ctx, "", resourceAttributes)
		if err != nil {
			return nil, util.Wrap(err, "Failed to authorize with API")
		}
		// In multi-user mode, apply filter with the namespace provided.
		filterContext = &model.FilterContext{
			ReferenceKey: &model.ReferenceKey{Type: model.NamespaceResourceType, ID: request.Namespace},
		}
	} else {
		if request.Namespace != "" {
			return nil, util.NewInvalidInputError("Invalid ListExperiments request. Namespace should not be provided in single-user mode.")
		}
		// In single user mode, apply filter with empty namespace for backward compatibility.
		filterContext = &model.FilterContext{
			ReferenceKey: &model.ReferenceKey{Type: model.NamespaceResourceType, ID: ""},
		}
	}

	experiments, total_size, nextPageToken, err := s.resourceManager.ListExperiments(filterContext, opts)
	if err != nil {
		return nil, util.Wrap(err, "List experiments failed.")
	}
	return &apiv2beta1.ListExperimentsResponse{
			Experiments:   ToApiExperiments(experiments),
			TotalSize:     int32(total_size),
			NextPageToken: nextPageToken},
		nil
}

func (s *ExperimentServer) DeleteExperimentV1(ctx context.Context, request *apiv1beta1.DeleteExperimentRequest) (*empty.Empty, error) {
	if s.options.CollectMetrics {
		deleteExperimentRequests.Inc()
	}
	err := s.canAccessExperiment(ctx, request.Id, &authorizationv1.ResourceAttributes{Verb: common.RbacResourceVerbDelete})
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}

	err = s.resourceManager.DeleteExperiment(request.Id)
	if err != nil {
		return nil, err
	}

	if s.options.CollectMetrics {
		experimentCount.Dec()
	}
	return &empty.Empty{}, nil
}

func (s *ExperimentServer) DeleteExperiment(ctx context.Context, request *apiv2beta1.DeleteExperimentRequest) (*empty.Empty, error) {
	if s.options.CollectMetrics {
		deleteExperimentRequests.Inc()
	}
	err := s.canAccessExperiment(ctx, request.ExperimentId, &authorizationv1.ResourceAttributes{Verb: common.RbacResourceVerbDelete})
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}

	err = s.resourceManager.DeleteExperiment(request.ExperimentId)
	if err != nil {
		return nil, err
	}

	if s.options.CollectMetrics {
		experimentCount.Dec()
	}
	return &empty.Empty{}, nil
}

func ValidateCreateExperimentRequestV1(request *apiv1beta1.CreateExperimentRequest) error {
	if request.Experiment == nil || request.Experiment.Name == "" {
		return util.NewInvalidInputError("Experiment name is empty. Please specify a valid experiment name.")
	}

	resourceReferences := request.Experiment.GetResourceReferences()
	if common.IsMultiUserMode() {
		if len(resourceReferences) != 1 ||
			resourceReferences[0].Key.Type != apiv1beta1.ResourceType_NAMESPACE ||
			resourceReferences[0].Relationship != apiv1beta1.Relationship_OWNER {
			return util.NewInvalidInputError(
				"Invalid resource references for experiment. Expect one namespace type with owner relationship. Got: %v", resourceReferences)
		}
		namespace := GetNamespaceFromAPIResourceReferences(request.Experiment.ResourceReferences)
		if len(namespace) == 0 {
			return util.NewInvalidInputError("Invalid resource references for experiment. Namespace is empty.")
		}
	} else if len(resourceReferences) > 0 {
		return util.NewInvalidInputError("In single-user mode, CreateExperimentRequest shouldn't contain resource references.")
	}
	return nil
}

func ValidateCreateExperimentRequest(request *apiv2beta1.CreateExperimentRequest) error {
	if request.Experiment == nil {
		return util.NewInvalidInputError("Experiment is empty.")
	}
	if request.Experiment.DisplayName == "" {
		return util.NewInvalidInputError("Experiment display name is empty. Please specify a valid experiment name.")
	}
	if common.IsMultiUserMode() && request.Experiment.Namespace == "" {
		return util.NewInvalidInputError("In multi-user mode, experiment namespace is empty. Please specify a valid namespace.")
	} else if !common.IsMultiUserMode() && request.Experiment.Namespace != "" {
		return util.NewInvalidInputError("In single-user mode, CreateExperimentRequest shouldn't contain namespace.")
	}
	return nil
}

// TODO(chensun): consider refactoring the code to get rid of double-query of experiment.
func (s *ExperimentServer) canAccessExperiment(ctx context.Context, experimentID string, resourceAttributes *authorizationv1.ResourceAttributes) error {
	if !common.IsMultiUserMode() {
		// Skip authorization if not multi-user mode.
		return nil
	}

	if len(experimentID) > 0 {
		experiment, err := s.resourceManager.GetExperiment(experimentID)
		if err != nil {
			return util.Wrap(err, "Failed to authorize with the experiment ID.")
		}
		if len(resourceAttributes.Namespace) == 0 {
			if len(experiment.Namespace) == 0 {
				return util.NewInternalServerError(
					errors.New("Empty namespace"),
					"The experiment doesn't have a valid namespace.",
				)
			}
			resourceAttributes.Namespace = experiment.Namespace
		}
		if len(resourceAttributes.Name) == 0 {
			resourceAttributes.Name = experiment.Name
		}
	}

	resourceAttributes.Group = common.RbacPipelinesGroup
	resourceAttributes.Version = common.RbacPipelinesVersion
	resourceAttributes.Resource = common.RbacResourceTypeExperiments

	err := isAuthorized(s.resourceManager, ctx, resourceAttributes)
	if err != nil {
		return util.Wrap(err, "Failed to authorize with API")
	}
	return nil
}

func (s *ExperimentServer) ArchiveExperimentV1(ctx context.Context, request *apiv1beta1.ArchiveExperimentRequest) (*empty.Empty, error) {
	if s.options.CollectMetrics {
		archiveExperimentRequests.Inc()
	}

	err := s.canAccessExperiment(ctx, request.Id, &authorizationv1.ResourceAttributes{Verb: common.RbacResourceVerbArchive})
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}
	err = s.resourceManager.ArchiveExperiment(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ExperimentServer) ArchiveExperiment(ctx context.Context, request *apiv2beta1.ArchiveExperimentRequest) (*empty.Empty, error) {
	if s.options.CollectMetrics {
		archiveExperimentRequests.Inc()
	}

	err := s.canAccessExperiment(ctx, request.ExperimentId, &authorizationv1.ResourceAttributes{Verb: common.RbacResourceVerbArchive})
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}
	err = s.resourceManager.ArchiveExperiment(ctx, request.ExperimentId)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ExperimentServer) UnarchiveExperimentV1(ctx context.Context, request *apiv1beta1.UnarchiveExperimentRequest) (*empty.Empty, error) {
	if s.options.CollectMetrics {
		unarchiveExperimentRequests.Inc()
	}

	err := s.canAccessExperiment(ctx, request.Id, &authorizationv1.ResourceAttributes{Verb: common.RbacResourceVerbUnarchive})
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}
	err = s.resourceManager.UnarchiveExperiment(request.Id)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ExperimentServer) UnarchiveExperiment(ctx context.Context, request *apiv2beta1.UnarchiveExperimentRequest) (*empty.Empty, error) {
	if s.options.CollectMetrics {
		unarchiveExperimentRequests.Inc()
	}

	err := s.canAccessExperiment(ctx, request.ExperimentId, &authorizationv1.ResourceAttributes{Verb: common.RbacResourceVerbUnarchive})
	if err != nil {
		return nil, util.Wrap(err, "Failed to authorize the request")
	}
	err = s.resourceManager.UnarchiveExperiment(request.ExperimentId)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func NewExperimentServer(resourceManager *resource.ResourceManager, options *ExperimentServerOptions) *ExperimentServer {
	return &ExperimentServer{resourceManager: resourceManager, options: options}
}
