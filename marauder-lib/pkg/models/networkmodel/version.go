package networkmodel

// The VersionResponseModel is returned by the /version endpoint.
type VersionResponseModel struct {
	Version string `json:"version"`
}
