package dto

type MCPCredentialSecret struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type MCPEnrollmentExport struct {
	MCPCredentialMutation
	Client string `json:"client"`
}

type MCPEnrollmentDocument struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}
