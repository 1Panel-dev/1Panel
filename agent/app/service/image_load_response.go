package service

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/docker/docker/pkg/jsonmessage"
)

// Docker may report load failures in a successful HTTP response's JSON stream.
func consumeImageLoadResponse(reader io.Reader, log func(string)) error {
	decoder := json.NewDecoder(reader)
	for {
		var message jsonmessage.JSONMessage
		if err := decoder.Decode(&message); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if message.Error != nil && message.Error.Message != "" {
			return message.Error
		}
		if message.ErrorMessage != "" {
			return errors.New(message.ErrorMessage)
		}
		if text := strings.TrimSpace(message.Stream); text != "" {
			log(text)
		}
	}
}
