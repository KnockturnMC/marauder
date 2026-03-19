package networkmodel

// ManagementToggleSaveBody is sent over the network to inform an operator to toggle the save state of a server.
type ManagementToggleSaveBody struct {
	ShouldSave bool `json:"shouldSave"`
}
