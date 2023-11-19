package networkmodel

// LifecycleAction defines a type of change request.
type LifecycleAction string

const (
	// Start defines that the server should be started if not online rn.
	Start LifecycleAction = "start"

	// Stop defines that the server should be stopped if not currently online.
	Stop LifecycleAction = "stop"

	// Restart simply restarts the server by stopping and then starting it.
	Restart LifecycleAction = "restart"

	// UpdateWithRestart updates a servers deployment by restarting the entire server.
	// For this, the server is stopped, the artefacts are updated and the server is started again.
	UpdateWithRestart LifecycleAction = "update+restart"

	// UpdateWithoutRestart updates a server without restarting it.
	// During this lifecycle action only artefacts are updated that do not require a restart.
	UpdateWithoutRestart LifecycleAction = "update-restart"
)

// KnownLifecycleChangeActionType computes if the passed change action is known by marauder.
func KnownLifecycleChangeActionType(changeActionType LifecycleAction) bool {
	switch changeActionType {
	case Start:
		fallthrough
	case Stop:
		fallthrough
	case Restart:
		fallthrough
	case UpdateWithRestart:
		fallthrough
	case UpdateWithoutRestart:
		return true
	}

	return false
}
