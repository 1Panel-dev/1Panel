package service

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// Docker may report load failures in a successful HTTP response's JSON stream.
func consumeImageLoadResponse(reader io.Reader, log func(string)) error {
	decoder := json.NewDecoder(reader)
	for {
		var message struct {
			Stream      string `json:"stream"`
			Error       string `json:"error"`
			ErrorDetail struct {
				Message string `json:"message"`
			} `json:"errorDetail"`
		}
		if err := decoder.Decode(&message); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if message.Error != "" {
			return errors.New(message.Error)
		}
		if message.ErrorDetail.Message != "" {
			return errors.New(message.ErrorDetail.Message)
		}
		if text := strings.TrimSpace(message.Stream); text != "" {
			log(text)
		}
	}
}
