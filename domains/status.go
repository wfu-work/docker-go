package domains

const (
	HostStatusOffline  = "offline"
	HostStatusOnline   = "online"
	HostStatusDisabled = "disabled"

	AgentStatusPending  = "pending"
	AgentStatusOnline   = "online"
	AgentStatusOffline  = "offline"
	AgentStatusDisabled = "disabled"

	TaskStatusPending    = "pending"
	TaskStatusDispatched = "dispatched"
	TaskStatusRunning    = "running"
	TaskStatusSuccess    = "success"
	TaskStatusFailed     = "failed"
	TaskStatusTimeout    = "timeout"
	TaskStatusCancelled  = "cancelled"
)

const (
	TaskTypeAgentPing    = "agent.ping"
	TaskTypeAgentUpgrade = "agent.upgrade"

	TaskTypeDockerContainerList    = "docker.container.list"
	TaskTypeDockerContainerStart   = "docker.container.start"
	TaskTypeDockerContainerStop    = "docker.container.stop"
	TaskTypeDockerContainerRestart = "docker.container.restart"
	TaskTypeDockerContainerRemove  = "docker.container.remove"
	TaskTypeDockerContainerLogs    = "docker.container.logs"
	TaskTypeDockerContainerStream  = "docker.container.logs.stream"

	TaskTypeDockerImageList = "docker.image.list"
	TaskTypeDockerImagePull = "docker.image.pull"

	TaskTypeDockerMetricsSnapshot = "docker.metrics.snapshot"

	TaskTypeDockerConfigValidate = "docker.config.validate"
	TaskTypeDockerConfigDeploy   = "docker.config.deploy"
	TaskTypeDockerComposeUp      = "docker.compose.up"
	TaskTypeDockerComposeDown    = "docker.compose.down"
	TaskTypeDockerComposeRestart = "docker.compose.restart"
	TaskTypeDockerComposePull    = "docker.compose.pull"
)

const (
	DeployConfigTypeCompose = "compose"

	DeployReleaseActionUp       = "up"
	DeployReleaseActionDown     = "down"
	DeployReleaseActionRestart  = "restart"
	DeployReleaseActionPull     = "pull"
	DeployReleaseActionRollback = "rollback"
)

const (
	RegistryCredentialStatusEnabled  = "enabled"
	RegistryCredentialStatusDisabled = "disabled"

	OperationApprovalStatusPending   = "pending"
	OperationApprovalStatusApproved  = "approved"
	OperationApprovalStatusRejected  = "rejected"
	OperationApprovalStatusCancelled = "cancelled"
	OperationApprovalStatusUsed      = "used"

	AgentUpgradePackageStatusEnabled  = "enabled"
	AgentUpgradePackageStatusDisabled = "disabled"
)
