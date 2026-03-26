package bark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type BarkMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func SendMessage(url string, title string, body string, transport *http.Transport) error {
	msg := BarkMessage{
		Title: title,
		Body:  body,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bark push failed with status code %d", resp.StatusCode)
	}

	return nil
}
