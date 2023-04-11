# Copyright 2023 The Kubeflow Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from typing import Dict, List

from kfp.dsl import ConcatPlaceholder
from kfp.dsl import container_component
from kfp.dsl import ContainerSpec
from kfp.dsl import OutputPath


@container_component
def dataproc_create_spark_batch(
    project: str,
    gcp_resources: OutputPath(str),
    location: str = 'us-central1',
    batch_id: str = '',
    labels: Dict[str, str] = {},
    container_image: str = '',
    runtime_config_version: str = '',
    runtime_config_properties: Dict[str, str] = {},
    service_account: str = '',
    network_tags: List[str] = [],
    kms_key: str = '',
    network_uri: str = '',
    subnetwork_uri: str = '',
    metastore_service: str = '',
    spark_history_dataproc_cluster: str = '',
    main_jar_file_uri: str = '',
    main_class: str = '',
    jar_file_uris: List[str] = [],
    file_uris: List[str] = [],
    archive_uris: List[str] = [],
    args: List[str] = [],
):
  # fmt: off
  """Create a Dataproc Spark batch workload and wait for it to finish.

  Args:
      project (str):
        Required: Project to run the Dataproc batch workload.
      location (Optional[str]):
        Location of the Dataproc batch workload. If
        not set, default to `us-central1`.
      batch_id (Optional[str]):
        The ID to use for the batch, which will become
        the final component of the batch's resource name. If none is
        specified, a default name will be generated by the component.  This
        value must be 4-63 characters. Valid characters are /[a-z][0-9]-/.
      labels (Optional[dict]):
        The labels to associate with this batch. Label
        keys must contain 1 to 63 characters, and must conform to RFC 1035.
        Label values may be empty, but, if present, must contain 1 to 63
        characters, and must conform to RFC 1035. No more than 32 labels can
        be associated with a batch.  An object containing a list of "key":
        value pairs.
          Example: { "name": "wrench", "mass": "1.3kg", "count": "3" }.
      container_image (Optional[str]):
        Optional custom container image for the
        job runtime environment. If not specified, a default container image
        will be used.
      runtime_config_version (Optional[str]):
        Version of the batch runtime.
      runtime_config_properties (Optional[dict]):
        Runtime configuration for a
        workload.
      service_account (Optional[str]):
        Service account that used to execute
        workload.
      network_tags (Optional[Sequence]):
        Tags used for network traffic
        control.
      kms_key (Optional[str]):
        The Cloud KMS key to use for encryption.
      network_uri (Optional[str]):
        Network URI to connect workload to.
      subnetwork_uri (Optional[str]):
        Subnetwork URI to connect workload to.
      metastore_service (Optional[str]):
        Resource name of an existing Dataproc
        Metastore service.
      spark_history_dataproc_cluster (Optional[str]):
        The Spark History Server
        configuration for the workload.
      main_jar_file_uri (Optional[str]):
        The HCFS URI of the jar file that
        contains the main class.
      main_class (Optional[str]):
        The name of the driver main class. The jar
        file that contains the class must be in the classpath or specified in
        jar_file_uris.
      jar_file_uris (Optional[Sequence]):
        HCFS URIs of jar files to add to the
        classpath of the Spark driver and tasks.
      file_uris (Optional[Sequence]):
        HCFS URIs of files to be placed in the
        working directory of each executor.
      archive_uris (Optional[Sequence]):
        HCFS URIs of archives to be extracted
        into the working directory of each executor.
      args (Optional[Sequence]):
        The arguments to pass to the driver.

  Returns:
      gcp_resources (str):
          Serialized gcp_resources proto tracking the Dataproc batch workload.
          For more details, see
          https://github.com/kubeflow/pipelines/blob/master/components/google-cloud/google_cloud_pipeline_components/proto/README.md.
  """
  # fmt: on
  return ContainerSpec(
      image='gcr.io/ml-pipeline/google-cloud-pipeline-components:2.0.0b1',
      command=[
          'python3',
          '-u',
          '-m',
          'google_cloud_pipeline_components.container.v1.dataproc.create_spark_batch.launcher',
      ],
      args=[
          '--type',
          'DataprocSparkBatch',
          '--payload',
          ConcatPlaceholder([
              '{',
              '"labels": ',
              labels,
              ', "runtime_config": {',
              '"version": "',
              runtime_config_version,
              '"',
              ', "container_image": "',
              container_image,
              '"',
              ', "properties": ',
              runtime_config_properties,
              '}',
              ', "environment_config": {',
              '"execution_config": {',
              '"service_account": "',
              service_account,
              '"',
              ', "network_tags": ',
              network_tags,
              ', "kms_key": "',
              kms_key,
              '"',
              ', "network_uri": "',
              network_uri,
              '"',
              ', "subnetwork_uri": "',
              subnetwork_uri,
              '"',
              '}',
              ', "peripherals_config": {',
              '"metastore_service": "',
              metastore_service,
              '"',
              ', "spark_history_server_config": { ',
              '"dataproc_cluster": "',
              spark_history_dataproc_cluster,
              '"',
              '}',
              '}',
              '}',
              ', "spark_batch": {',
              '"main_jar_file_uri": "',
              main_jar_file_uri,
              '"',
              ', "main_class": "',
              main_class,
              '"',
              ', "jar_file_uris": ',
              jar_file_uris,
              ', "file_uris": ',
              file_uris,
              ', "archive_uris": ',
              archive_uris,
              ', "args": ',
              args,
              '}',
              '}',
          ]),
          '--project',
          project,
          '--location',
          location,
          '--batch_id',
          batch_id,
          '--gcp_resources',
          gcp_resources,
      ],
  )
