package json

import (
	"encoding/json"
	"fmt"
	"os"
)

// TestType описывает тип теста с его описанием.
type TestType struct {
	Type          string `json:"type"`           // Например, "logic"
	Description   string `json:"description"`    // Описание типа, например, "Логические задачи"
	TestQuestions int    `json:"test_questions"` // Количество вопросов для этого теста
	TestDuration  int    `json:"test_duration"`  // Время прохождения теста в минутах
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

// GetTestTypeSettings возвращает настройки для заданного типа теста.
func GetTestTypeSettings(testType string) (*TestType, error) {
	types, err := LoadTestTypes("data/test_types.json")
	if err != nil {
		return nil, err
	}
	for _, t := range types {
		if t.Type == testType {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("настройки для теста типа %s не найдены", testType)
}
