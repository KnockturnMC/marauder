package networkmodel

import (
	"errors"
	"fmt"

	"github.com/Goldziher/go-utils/maputils"
)

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

var (
	// ErrUnknownLifecycleAction is returned if the passed string is not a lifecycle action.
	ErrUnknownLifecycleAction = errors.New("unknown lifecycle action")

	// knownLifecycleActions contains a map access for all known lifecycle actions.
	//nolint: gochecknoglobals
	knownLifecycleActions = map[LifecycleAction]*struct{}{
		Start:                     nil,
		Stop:                      nil,
		Restart:                   nil,
		UpdateWithRestart:         nil,
		ForceUpdateWithRestart:    nil,
		UpdateWithoutRestart:      nil,
		ForceUpdateWithoutRestart: nil,
	}
)

// Type implemented for cobra flag.
func (l *LifecycleAction) Type() string {
	return "LifecycleAction"
}

// String implemented for cobra flag.
func (l *LifecycleAction) String() string {
	return string(*l)
}

// Set implemented for cobra flag.
func (l *LifecycleAction) Set(s string) error {
	if !KnownLifecycleChangeActionType(LifecycleAction(s)) {
		return fmt.Errorf("unknown %s: %w", s, ErrUnknownLifecycleAction)
	}

	*l = LifecycleAction(s)
	return nil
}

// KnownLifecycleActions yields a slice of all known lifecycle actions.
func KnownLifecycleActions() []LifecycleAction {
	return maputils.Keys(knownLifecycleActions)
}

// KnownLifecycleChangeActionType computes if the passed change action is known by marauder.
func KnownLifecycleChangeActionType(changeActionType LifecycleAction) bool {
	_, ok := knownLifecycleActions[changeActionType]
	return ok
}
