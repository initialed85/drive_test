package file_writer

import (
	"encoding/json"
	"log"
	"os"
)

func WriteIndentedJSONToFile(object interface{}, outputPath string) error {
	objectJSON, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return err
	}

	log.Println(string(objectJSON) + "\n")

	f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	_, err = f.WriteString(string(objectJSON) + "\n")
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
