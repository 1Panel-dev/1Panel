package csv

import (
	"encoding/csv"
	"os"

	"github.com/1Panel-dev/1Panel/core/i18n"
)

type CommandTemplate struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

func ExportCommands(filename string, commands []CommandTemplate) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{
		i18n.GetMsgByKey("Name"),
		i18n.GetMsgByKey("Command"),
	}); err != nil {
		return err
	}

	for _, log := range commands {
		record := []string{
			log.Name,
			log.Command,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
