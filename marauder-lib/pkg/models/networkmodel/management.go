package networkmodel

import "time"

// ManagementToggleSaveBody is sent over the network to inform an operator to toggle the save state of a server.
type ManagementToggleSaveBody struct {
	ShouldSave bool `json:"shouldSave"`
}

// ManagementPlayersResponse is returned by the operator when querying the players on a server.
type ManagementPlayersResponse struct {
	LastPlayerLoginTimestamp  time.Time          `json:"lastPlayerLoginTimestamp"`
	LastPlayerLogoutTimestamp time.Time          `json:"lastPlayerLogoutTimestamp"`
	Players                   []ManagementPlayer `json:"players"`
}

// ManagementPlayer is a single player instance.
type ManagementPlayer struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
