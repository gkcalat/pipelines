// Copyright 2019 The Kubeflow Authors
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

package common

import (
	"strings"

	apiV1beta1 "github.com/kubeflow/pipelines/backend/api/v1beta1/go_client"
)

const (
	DefaultPipelineRunnerServiceAccount = "pipeline-runner"
	HasDefaultBucketEnvVar              = "HAS_DEFAULT_BUCKET"
	DefaultBucketNameEnvVar             = "BUCKET_NAME"
	ProjectIDEnvVar                     = "PROJECT_ID"
)

func GetNamespaceFromAPIResourceReferences(resourceRefs []*apiV1beta1.ResourceReference) string {
	namespace := ""
	for _, resourceRef := range resourceRefs {
		if resourceRef.Key.Type == apiV1beta1.ResourceType_NAMESPACE {
			namespace = resourceRef.Key.Id
			break
		}
	}
	return namespace
}

func GetExperimentIDFromAPIResourceReferences(resourceRefs []*apiV1beta1.ResourceReference) string {
	experimentID := ""
	for _, resourceRef := range resourceRefs {
		if resourceRef.Key.Type == apiV1beta1.ResourceType_EXPERIMENT {
			experimentID = resourceRef.Key.Id
			break
		}
	}
	return experimentID
}

// Mutate default values of specified pipeline spec.
// Args:
//
//	text: (part of) pipeline file in string.
func PatchPipelineDefaultParameter(text string) (string, error) {
	defaultBucket := GetStringConfig(DefaultBucketNameEnvVar)
	projectId := GetStringConfig(ProjectIDEnvVar)
	toPatch := map[string]string{
		"{{kfp-default-bucket}}": defaultBucket,
		"{{kfp-project-id}}":     projectId,
	}
	for key, value := range toPatch {
		text = strings.Replace(text, key, value, -1)
	}
	return text, nil
}
