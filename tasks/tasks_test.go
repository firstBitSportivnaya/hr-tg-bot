package tasks

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

// 1. TestGetRandomTasks_SingleCandidate Обычный случай для одного кандидата.
// Проверяется, что вызывается ровно n задач, все они уникальны, и если задача была свободна, то её поле ReservedBy заполнено.
// 2. TestGetRandomTasks_MultipleCandidates Случай с несколькими кандидатами, когда свободных задач достаточно.
// Проверяется, что для двух разных кандидатов (например, cand1 и cand2) возвращаются наборы по n уникальных задач и никакая задача, полученная для одного кандидата, не повторяется у другого (то есть свободные задачи распределены эксклюзивно).
// 3. TestGetRandomTasks_InsufficientFree Случай, когда свободных задач недостаточно.
// Если, например, всего m задач, а один кандидат забирает почти все свободные, то для следующего кандидата свободных задач меньше, и в его набор могут попасть задачи, уже зарезервированные другим кандидатом. При этом в наборе задач для каждого кандидата они должны оставаться уникальными.
// 4. TestNewManager_FileLoad Проверка загрузки из файла.

// newTestManager создаёт тестового менеджера с переданными задачами.
func newTestManager(tasks []Task) *Manager {
	return &Manager{
		tasks: tasks,
	}
}

// TestGetRandomTasks_SingleCandidate проверяет случай для одного кандидата.
func TestGetRandomTasks_SingleCandidate(t *testing.T) {
	// Создаем набор тестовых задач.
	testTasks := []Task{
		{ID: 1, Text: "Вопрос 1", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
		{ID: 2, Text: "Вопрос 2", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 1},
		{ID: 3, Text: "Вопрос 3", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 2},
		{ID: 4, Text: "Вопрос 4", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
		{ID: 5, Text: "Вопрос 5", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 1},
		{ID: 6, Text: "Вопрос 6", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 2},
		{ID: 7, Text: "Вопрос 7", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
		{ID: 8, Text: "Вопрос 8", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 1},
		{ID: 9, Text: "Вопрос 9", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 2},
		{ID: 10, Text: "Вопрос 10", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
	}

	mgr := newTestManager(testTasks)
	candidateID := "candidate1"
	n := 5

	tasksSelected, err := mgr.GetRandomTasks(n, candidateID)
	if err != nil {
		t.Fatalf("GetRandomTasks вернул ошибку: %v", err)
	}

	if len(tasksSelected) != n {
		t.Errorf("Ожидалось %d задач, получено %d", n, len(tasksSelected))
	}

	// Проверяем, что все задачи уникальны по ID и что задачи, которые были свободны, теперь зарезервированы.
	ids := make(map[int]bool)
	for _, task := range tasksSelected {
		if ids[task.ID] {
			t.Errorf("Задача с ID %d повторяется в наборе", task.ID)
		}
		ids[task.ID] = true

		if task.ReservedBy == "" {
			t.Errorf("Задача с ID %d не была зарезервирована", task.ID)
		}
	}
}

// TestGetRandomTasks_MultipleCandidates проверяет, что для нескольких кандидатов (когда свободных задач достаточно)
// наборы задач для разных кандидатов не пересекаются.
func TestGetRandomTasks_MultipleCandidates(t *testing.T) {
	// Создаем набор тестовых задач – 10 задач.
	testTasks := []Task{
		{ID: 1, Text: "Вопрос 1", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
		{ID: 2, Text: "Вопрос 2", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 1},
		{ID: 3, Text: "Вопрос 3", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 2},
		{ID: 4, Text: "Вопрос 4", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
		{ID: 5, Text: "Вопрос 5", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 1},
		{ID: 6, Text: "Вопрос 6", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 2},
		{ID: 7, Text: "Вопрос 7", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
		{ID: 8, Text: "Вопрос 8", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 1},
		{ID: 9, Text: "Вопрос 9", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 2},
		{ID: 10, Text: "Вопрос 10", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
	}

	mgr := newTestManager(testTasks)
	n := 5

	candidateIDs := []string{"cand1", "cand2"}
	results := make(map[string][]Task)

	for _, cand := range candidateIDs {
		tasksSelected, err := mgr.GetRandomTasks(n, cand)
		if err != nil {
			t.Fatalf("Кандидат %s: GetRandomTasks вернул ошибку: %v", cand, err)
		}
		if len(tasksSelected) != n {
			t.Errorf("Кандидат %s: ожидалось %d задач, получено %d", cand, n, len(tasksSelected))
		}

		// Проверяем уникальность задач внутри набора кандидата.
		ids := make(map[int]bool)
		for _, task := range tasksSelected {
			if ids[task.ID] {
				t.Errorf("Кандидат %s: задача с ID %d повторяется", cand, task.ID)
			}
			ids[task.ID] = true
		}
		results[cand] = tasksSelected
	}

	// Если свободных задач достаточно, наборы задач для разных кандидатов не должны пересекаться.
	setCand1 := make(map[int]bool)
	for _, task := range results["cand1"] {
		// Проверяем, что задача в cand1 зарезервирована для cand1.
		if task.ReservedBy != "cand1" {
			t.Errorf("Кандидат cand1: задача с ID %d имеет ReservedBy=%s", task.ID, task.ReservedBy)
		}
		setCand1[task.ID] = true
	}
	for _, task := range results["cand2"] {
		if task.ReservedBy != "cand2" {
			t.Errorf("Кандидат cand2: задача с ID %d имеет ReservedBy=%s", task.ID, task.ReservedBy)
		}
		if setCand1[task.ID] {
			t.Errorf("Задача с ID %d получена и у cand1, и у cand2", task.ID)
		}
	}
}

// TestGetRandomTasks_InsufficientFree проверяет случай, когда свободных задач недостаточно.
// Если задач меньше, чем запрошено, для второго кандидата часть задач может уже быть зарезервирована другим.
func TestGetRandomTasks_InsufficientFree(t *testing.T) {
	// Создаем набор из 6 задач.
	testTasks := []Task{
		{ID: 1, Text: "Вопрос 1", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
		{ID: 2, Text: "Вопрос 2", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 1},
		{ID: 3, Text: "Вопрос 3", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 2},
		{ID: 4, Text: "Вопрос 4", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
		{ID: 5, Text: "Вопрос 5", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 1},
		{ID: 6, Text: "Вопрос 6", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 2},
	}

	mgr := newTestManager(testTasks)
	n := 5
	candidate1 := "cand1"
	candidate2 := "cand2"

	// Первый кандидат получает 5 задач (результат не сохраняем, чтобы избежать "unused variable").
	if _, err := mgr.GetRandomTasks(n, candidate1); err != nil {
		t.Fatalf("Кандидат %s: GetRandomTasks вернул ошибку: %v", candidate1, err)
	}
	// Второй кандидат запрашивает тоже 5 задач.
	tasksCand2, err := mgr.GetRandomTasks(n, candidate2)
	if err != nil {
		t.Fatalf("Кандидат %s: GetRandomTasks вернул ошибку: %v", candidate2, err)
	}

	// Проверяем, что кандидат2 получил ровно n задач.
	if len(tasksCand2) != n {
		t.Errorf("Кандидат %s: ожидалось %d задач, получено %d", candidate2, n, len(tasksCand2))
	}

	// Проверяем уникальность задач внутри набора кандидата2.
	idsCand2 := make(map[int]bool)
	for _, task := range tasksCand2 {
		if idsCand2[task.ID] {
			t.Errorf("Кандидат %s: задача с ID %d повторяется", candidate2, task.ID)
		}
		idsCand2[task.ID] = true
	}
}

// TestNewManager_FileLoad проверяет загрузку задач из JSON-файла.
func TestNewManager_FileLoad(t *testing.T) {
	tasksData := []Task{
		{ID: 1, Text: "Вопрос 1", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 0},
		{ID: 2, Text: "Вопрос 2", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 1},
		{ID: 3, Text: "Вопрос 3", Options: []string{"Да", "Нет", "Не уверен"}, Answer: 2},
	}
	data, err := json.Marshal(tasksData)
	if err != nil {
		t.Fatalf("Ошибка маршалинга JSON: %v", err)
	}
	tmpFile, err := ioutil.TempFile("", "tasks_*.json")
	if err != nil {
		t.Fatalf("Ошибка создания временного файла: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(data); err != nil {
		t.Fatalf("Ошибка записи во временный файл: %v", err)
	}
	tmpFile.Close()

	mgr, err := NewManager(tmpFile.Name())
	if err != nil {
		t.Fatalf("NewManager вернул ошибку: %v", err)
	}

	if len(mgr.tasks) != len(tasksData) {
		t.Errorf("Ожидалось %d задач, получено %d", len(tasksData), len(mgr.tasks))
	}

	// Проверяем, что задачи совпадают по содержимому.
	for i, task := range mgr.tasks {
		if !reflect.DeepEqual(task, tasksData[i]) {
			t.Errorf("Задачи не совпадают: ожидалось %+v, получено %+v", tasksData[i], task)
		}
	}
}
