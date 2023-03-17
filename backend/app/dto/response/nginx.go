package response

type NginxStatus struct {
	Active   string `json:"active"`
	Accepts  string `json:"accepts"`
	Handled  string `json:"handled"`
	Requests string `json:"requests"`
	Reading  string `json:"reading"`
	Writing  string `json:"writing"`
	Waiting  string `json:"waiting"`
}

type NginxParam struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}
