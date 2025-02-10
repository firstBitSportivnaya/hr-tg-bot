package report

import (
	"fmt"
	"strconv"

	"github.com/IT-Nick/tasks"
	"github.com/jung-kurt/gofpdf"
)

// ReportData содержит информацию о результатах тестирования, необходимую для формирования PDF-отчёта.
// Структура включает данные пользователя, его результаты, а также список вопросов и ответы.
type ReportData struct {
	UserID            int64        // Идентификатор пользователя.
	TelegramFirstName string       // Имя пользователя в Telegram.
	TelegramUsername  string       // Username пользователя в Telegram.
	Role              string       // Роль пользователя (например, "user", "hr", "admin").
	State             string       // Текущее состояние пользователя (например, "finished").
	Score             int          // Количество правильных ответов.
	TotalQuestions    int          // Общее количество вопросов в тесте.
	TestType          string       // Вид теста (например, "logic" или "math").
	Answers           map[int]int  // Карта, связывающая индекс вопроса с выбранным вариантом ответа.
	TestTasks         []tasks.Task // Список вопросов, представленных пользователю во время тестирования.
}

// sanitizeString заменяет все символы за пределами базовой многоязычной плоскости Unicode (BMP, U+0000..U+FFFF)
// на знак вопроса. Это предотвращает возможные проблемы с отображением и генерацией PDF-документа,
// если в строке присутствуют нестандартные символы (например, эмодзи или редкие иероглифы).
func sanitizeString(s string) string {
	var out []rune
	for _, r := range s {
		if r > 0xFFFF {
			out = append(out, '?')
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

// GeneratePDFReport формирует PDF-отчёт по результатам тестирования на основе переданных данных.
// Отчёт включает информацию о пользователе, его результатах и детальное описание каждого вопроса с
// указанными пользователем и правильными ответами. PDF-файл сохраняется в директории "reports" с именем,
// основанным на username или ID пользователя.
//
// Параметры:
//   - r: ReportData, содержащая все необходимые данные для формирования отчёта.
//
// Возвращает:
//   - string: имя (и путь) сгенерированного PDF-файла.
//   - error: ошибку, если процесс генерации или записи файла завершился неудачно.
func GeneratePDFReport(r ReportData) (string, error) {
	// Создаём новый PDF документ в портретном формате, единицы измерения – мм, формат страницы – A4.
	pdf := gofpdf.New("P", "mm", "A4", "")
	// Добавляем шрифты с поддержкой UTF-8, чтобы корректно отображались символы кириллицы и другие.
	pdf.AddUTF8Font("DejaVu", "", "report/fonts/DejaVuSans.ttf")
	pdf.AddUTF8Font("DejaVu", "B", "report/fonts/DejaVuSans-Bold.ttf")

	// Устанавливаем основной шрифт и размер текста для документа.
	pdf.SetFont("DejaVu", "", 14)
	pdf.AddPage()

	// Добавляем заголовок отчёта.
	pdf.SetFont("DejaVu", "B", 16)
	pdf.MultiCell(0, 10, "Отчет по тестированию", "", "L", false)
	pdf.Ln(4)

	// Устанавливаем шрифт для основного текста.
	pdf.SetFont("DejaVu", "", 12)

	// Приводим потенциально проблемные строки к безопасному виду,
	// заменяя символы вне диапазона BMP на знак вопроса.
	safeFirstName := sanitizeString(r.TelegramFirstName)
	safeUsername := sanitizeString(r.TelegramUsername)

	// Формируем информационный блок с данными о пользователе и результатами теста,
	// включая вид теста.
	info := fmt.Sprintf(
		"Имя: %s\nUsername: %s\nUser ID: %d\nРоль: %s\nСостояние: %s\nВид теста: %s\nРезультат: %d правильных ответов из %d\n",
		safeFirstName,
		safeUsername,
		r.UserID,
		r.Role,
		r.State,
		r.TestType,
		r.Score,
		r.TotalQuestions,
	)
	// Выводим информационный блок в документе.
	pdf.MultiCell(0, 8, info, "", "L", false)
	pdf.Ln(4)

	// Перебираем все вопросы теста и добавляем их в отчёт.
	for i, t := range r.TestTasks {
		// Получаем индекс и текст ответа, выбранного пользователем (если имеется).
		userAnsIdx, ok := r.Answers[i]
		userAnsStr := ""
		if ok && userAnsIdx < len(t.Options) {
			userAnsStr = t.Options[userAnsIdx]
		}
		// Получаем правильный ответ для вопроса.
		correctAnsStr := ""
		if t.Answer < len(t.Options) {
			correctAnsStr = t.Options[t.Answer]
		}

		// "Очищаем" текст вопроса и ответы от символов, способных вызвать проблемы при генерации PDF.
		qText := sanitizeString(t.Text)
		userAnsStr = sanitizeString(userAnsStr)
		correctAnsStr = sanitizeString(correctAnsStr)

		// Формируем заголовок для вопроса.
		qHeader := fmt.Sprintf("Вопрос %d:", i+1)
		pdf.SetFont("DejaVu", "B", 12)
		pdf.MultiCell(0, 8, qHeader, "", "L", false)

		// Выводим текст вопроса.
		pdf.SetFont("DejaVu", "", 12)
		pdf.MultiCell(0, 8, qText, "", "L", false)
		pdf.Ln(2)

		// Формируем строку с ответами: выбранный пользователем и правильный ответ.
		answerLine := fmt.Sprintf("Ваш ответ: %s\nПравильный: %s\n", userAnsStr, correctAnsStr)
		pdf.MultiCell(0, 8, answerLine, "", "L", false)
		pdf.Ln(4)
	}

	// Определяем имя файла для сохранения отчёта.
	var filename string
	if r.TelegramUsername != "" {
		filename = "reports/" + r.TelegramUsername + ".pdf"
	} else {
		filename = "reports/" + "report_" + strconv.FormatInt(r.UserID, 10) + ".pdf"
	}

	// Сохраняем PDF-файл и закрываем документ.
	if err := pdf.OutputFileAndClose(filename); err != nil {
		return "", err
	}
	return filename, nil
}
