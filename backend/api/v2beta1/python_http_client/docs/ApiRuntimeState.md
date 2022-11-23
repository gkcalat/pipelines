# ApiRuntimeState

Describes the state of a runtime entity.   - RUNTIMESTATE_UNSPECIFIED: Default value. This value is not used.  - PENDING: Service is preparing to run an entity.  - RUNNING: Entity is in progress.  - SUCCEEDED: Entity completed successfully.  - SKIPPED: Entity has been skipped.  - FAILED: Entity failed.  - CANCELING: Entity is being canceled. From this state, an entity may only go to either SUCCEEDED, FAILED or CANCELED.  - CANCELED: Entity has been canceled.  - PAUSED: Entity has been stopped, and can be resumed.
## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


