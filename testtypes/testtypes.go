package testtypes

import (
	"encoding/json"
	"fmt"
	"os"
)

// TestType описывает тип теста с его описанием.
type TestType struct {
	Type        string `json:"type"`        // Например, "logic"
	Description string `json:"description"` // Описание типа, например, "Логические задачи"
}

// LoadTestTypes загружает список тестовых типов из указанного JSON-файла.
func LoadTestTypes(filename string) ([]TestType, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл с типами тестов: %v", err)
	}
	var testTypes []TestType
	if err := json.Unmarshal(data, &testTypes); err != nil {
		return nil, fmt.Errorf("не удалось разобрать JSON: %v", err)
	}
	return testTypes, nil
}
