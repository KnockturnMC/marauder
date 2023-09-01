package networkmodel

// LifecycleChangeActionType defines a type of change request.
type LifecycleChangeActionType string

const (
	// Start defines that the server should be started if not online rn.
	Start LifecycleChangeActionType = "start"

	// Stop defines that the server should be stopped if not currently online.
	Stop LifecycleChangeActionType = "stop"

	// Restart simply restarts the server by stopping and then starting it.
	Restart LifecycleChangeActionType = "restart"

	// UpgradeDeployment updates a servers deployment.
	// For this, the server is stopped, the artefacts are updated and the server is started again.
	UpgradeDeployment LifecycleChangeActionType = "update"
)

// KnownLifecycleChangeActionType computes if the passed change action is known by marauder.
func KnownLifecycleChangeActionType(changeActionType LifecycleChangeActionType) bool {
	switch changeActionType {
	case Start:
		fallthrough
	case Stop:
		fallthrough
	case Restart:
		fallthrough
	case UpgradeDeployment:
		return true
	}

	return false
}
