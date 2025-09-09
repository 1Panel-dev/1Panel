package response

type ComponentInfo struct {
	Exists  bool   `json:"exists"`
	Version string `json:"version"`
	Path    string `json:"path"`
	Error   string `json:"error"`
}
