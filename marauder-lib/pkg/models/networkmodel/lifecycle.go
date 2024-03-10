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

	// UpdateWithoutRestart updates a server without restarting it.
	// During this lifecycle action only artefacts are updated that do not require a restart.
	UpdateWithoutRestart LifecycleAction = "update-restart"

	// ForceUpdateWithoutRestart updates a servers deployment without restarting it.
	// During this lifecycle action only artefacts are updated that do not require a restart.
	// Force updates do not fail if the local files were changed from the currently tracked server state.
	ForceUpdateWithoutRestart LifecycleAction = "force+update-restart"

	// UpdateWithRestart updates a servers deployment by restarting the entire server.
	// For this, the server is stopped, the artefacts are updated and the server is started again.
	UpdateWithRestart LifecycleAction = "update+restart"

	// ForceUpdateWithRestart updates a servers deployment by restarting the entire server.
	// For this, the server is stopped, the artefacts are updated and the server is started again.
	// Force updates do not fail if the local files were changed from the currently tracked server state.
	ForceUpdateWithRestart LifecycleAction = "force+update+restart"
)

// KnownLifecycleChangeActionType computes if the passed change action is known by marauder.
func KnownLifecycleChangeActionType(changeActionType LifecycleAction) bool {
	switch changeActionType {
	case Start,
		Stop,
		Restart,
		UpdateWithRestart,
		ForceUpdateWithRestart,
		UpdateWithoutRestart,
		ForceUpdateWithoutRestart:
		return true
	}

	return false
}
