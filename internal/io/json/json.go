package json

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadfromFile[T any](path string) ([]T, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var container []T
	json.Unmarshal(file, &container)
	return container, err
}

func AppendToFile(path string, input any) error {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	jsonString, jsonErr := json.Marshal(input)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return jsonErr
	}
	file.Write(jsonString)
	return nil
}
