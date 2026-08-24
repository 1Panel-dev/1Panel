package publicshare

func IsAPI(path string) bool {
	switch path {
	case "/api/v2/files/share/info",
		"/api/v2/files/share/check",
		"/api/v2/files/share/download":
		return true
	default:
		return false
	}
}
