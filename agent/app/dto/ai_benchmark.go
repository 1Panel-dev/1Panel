package dto

type AIBenchmarkSource struct {
	ID        string   `json:"id"`
	Kind      string   `json:"kind"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	BaseURL   string   `json:"baseUrl"`
	Endpoint  string   `json:"endpoint"`
	Models    []string `json:"models"`
	Tokenizer string   `json:"tokenizer"`
	Message   string   `json:"message"`
}
