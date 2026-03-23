package networkmodel

// ManagementToggleSaveBody is sent over the network to inform an operator to toggle the save state of a server.
type ManagementToggleSaveBody struct {
	ShouldSave bool `json:"shouldSave"`
}

// ManagementPlayer is a single player instance.
type ManagementPlayer struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
