package report

import (
	"fmt"
	"strconv"

	"github.com/IT-Nick/tasks"
	"github.com/jung-kurt/gofpdf"
)

// ReportData содержит данные для формирования отчёта.
type ReportData struct {
	UserID            int64
	TelegramFirstName string
	TelegramUsername  string
	Role              string
	State             string
	Score             int
	TotalQuestions    int
	Answers           map[int]int
	TestTasks         []tasks.Task
}

// GeneratePDFReport генерирует PDF‑отчёт по данным ReportData и сохраняет его в файл.
// Отчёт формируется в виде непрерывного текста с переносами (без таблицы).
// Возвращает имя файла (например, "stepplex.pdf") и ошибку, если она произошла.
func GeneratePDFReport(r ReportData) (string, error) {
	// Создаем новый PDF документ формата A4.
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Регистрируем UTF-8 шрифты, поддерживающие кириллицу.
	pdf.AddUTF8Font("DejaVu", "", "report/fonts/DejaVuSans.ttf")
	pdf.AddUTF8Font("DejaVu", "B", "report/fonts/DejaVuSans-Bold.ttf")

	// Устанавливаем основной шрифт.
	pdf.SetFont("DejaVu", "", 14)
	pdf.AddPage()

	// Заголовок отчёта.
	pdf.SetFont("DejaVu", "B", 16)
	pdf.MultiCell(0, 10, "Отчет по тестированию", "", "L", false)
	pdf.Ln(4)

	// Информация о пользователе.
	pdf.SetFont("DejaVu", "", 12)
	info := fmt.Sprintf("Имя: %s\nUsername: %s\nUser ID: %d\nРоль: %s\nСостояние: %s\nРезультат: %d правильных ответов из %d\n",
		r.TelegramFirstName, r.TelegramUsername, r.UserID, r.Role, r.State, r.Score, r.TotalQuestions)
	pdf.MultiCell(0, 8, info, "", "L", false)
	pdf.Ln(4)

	// Для каждого вопроса выводим его данные.
	for i, t := range r.TestTasks {
		// Получаем выбранный ответ пользователя.
		userAnsIdx, ok := r.Answers[i]
		userAnsStr := ""
		if ok && userAnsIdx < len(t.Options) {
			userAnsStr = t.Options[userAnsIdx]
		}
		correctAnsStr := ""
		if t.Answer < len(t.Options) {
			correctAnsStr = t.Options[t.Answer]
		}

		// Формируем заголовок вопроса.
		qHeader := fmt.Sprintf("Вопрос %d:", i+1)
		pdf.SetFont("DejaVu", "B", 12)
		pdf.MultiCell(0, 8, qHeader, "", "L", false)

		// Выводим текст вопроса, автоматически перенося его.
		pdf.SetFont("DejaVu", "", 12)
		pdf.MultiCell(0, 8, t.Text, "", "L", false)
		pdf.Ln(2)

		// Выводим строку с ответами.
		answerLine := fmt.Sprintf("Ваш ответ: %s\nПравильный: %s\n", userAnsStr, correctAnsStr)
		pdf.MultiCell(0, 8, answerLine, "", "L", false)
		pdf.Ln(4)
	}

	// Формируем имя файла.
	filename := ""
	if r.TelegramUsername != "" {
		filename = r.TelegramUsername + ".pdf"
	} else {
		filename = "report_" + strconv.FormatInt(r.UserID, 10) + ".pdf"
	}

	if err := pdf.OutputFileAndClose(filename); err != nil {
		return "", err
	}
	return filename, nil
}
