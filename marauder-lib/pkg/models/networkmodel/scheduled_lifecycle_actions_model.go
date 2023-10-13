package networkmodel

import (
	"time"

	"github.com/google/uuid"
)

// A ScheduledLifecycleAction holds a lifecycle action that is to be executed on a specific server instance at a specific time.
type ScheduledLifecycleAction struct {
	UUID            uuid.UUID       `db:"uuid"              json:"uuid"`
	ServerUUID      uuid.UUID       `db:"server"            json:"-"`
	Server          ServerModel     `db:"-"                 json:"server"`
	LifecycleAction LifecycleAction `db:"action"            json:"lifecycleAction"`
	TimeOfExecution time.Time       `db:"time_of_execution" json:"timeOfExecution"`
}
