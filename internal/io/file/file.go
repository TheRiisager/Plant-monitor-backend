package file

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadfromFile[T any](path string) (T, error) {
	var container T
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return container, err
	}

	json.Unmarshal(file, &container)
	return container, err
}

func WriteToFile[T any](path string, input *T) error {

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonString, jsonErr := json.Marshal(input)
	if jsonErr != nil || len(jsonString) == 0 {
		fmt.Println("error marshalling to json, or json result empty")
		return jsonErr
	}
	_, writeErr := file.Write(jsonString)
	if writeErr != nil {
		return writeErr
	}

	syncerr := file.Sync()
	if syncerr != nil {
		return syncerr
	}
	return nil
}
