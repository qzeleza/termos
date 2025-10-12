package query

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/performance"
	"github.com/qzeleza/ziva/internal/ui"
)

type ErrorColor int

const (
	Yellow ErrorColor = iota
	Red
	Orange
)

// Константы для управления памятью на embedded устройствах
var (
	// Значения могут быть переопределены через переменные окружения
	MaxCompletedTasks       int    = 50               // Максимальное количество завершенных задач в памяти
	MemoryPressureThreshold uint64 = 64 * 1024 * 1024 // 64MB - порог для запуска очистки памяти
)

// init инициализирует константы из переменных окружения
// при запуске программы
//
// @return void
func init() {
	// Устанавливаем максимальное количество завершенных задач
	// ZIVA_MAX_COMPLETED_TASKS=int
	if v := strings.TrimSpace(os.Getenv("ZIVA_MAX_COMPLETED_TASKS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			MaxCompletedTasks = n
		}
	}

	// Устанавливаем порог для запуска очистки памяти
	// ZIVA_MEMORY_PRESSURE_THRESHOLD=64MB/65536KB/67108864/64MiB
	if v := strings.TrimSpace(os.Getenv("ZIVA_MEMORY_PRESSURE_THRESHOLD")); v != "" {
		if bytes, err := parseMemoryEnv(v); err == nil && bytes > 0 {
			MemoryPressureThreshold = bytes
		}
	} else if v := strings.TrimSpace(os.Getenv("GOMEMLIMIT")); v != "" {
		if bytes, err := parseMemoryEnv(v); err == nil && bytes > 0 {
			// По умолчанию порог в 0.8 от лимита памяти рантайма
			MemoryPressureThreshold = uint64(float64(bytes) * 0.8)
		}
	}
}

// parseMemoryEnv — локальный парсер для значений памяти с суффиксами
//
// @param s Строка с значением памяти
// @return uint64 - значение памяти в байтах
func parseMemoryEnv(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	up := strings.ToUpper(s)
	var mult uint64 = 1
	num := s
	switch {
	case strings.HasSuffix(up, "GIB"):
		mult = 1024 * 1024 * 1024
		num = s[:len(s)-3]
	case strings.HasSuffix(up, "MIB"):
		mult = 1024 * 1024
		num = s[:len(s)-3]
	case strings.HasSuffix(up, "KIB"):
		mult = 1024
		num = s[:len(s)-3]
	case strings.HasSuffix(up, "GB"):
		mult = 1024 * 1024 * 1024
		num = s[:len(s)-2]
	case strings.HasSuffix(up, "MB"):
		mult = 1024 * 1024
		num = s[:len(s)-2]
	case strings.HasSuffix(up, "KB"):
		mult = 1024
		num = s[:len(s)-2]
	case strings.HasSuffix(up, "B"):
		mult = 1
		num = s[:len(s)-1]
	}
	num = strings.TrimSpace(num)
	cleanNum, err := normalizeMemoryNumber(num)
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseUint(cleanNum, 10, 64)
	if err != nil {
		return 0, err
	}
	return n * mult, nil
}

func normalizeMemoryNumber(num string) (string, error) {
	if !strings.Contains(num, "_") {
		return num, nil
	}

	parts := strings.Split(num, "_")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid numeric format")
	}

	for i, part := range parts {
		if part == "" {
			return "", fmt.Errorf("invalid numeric format")
		}
		if i == 0 {
			if len(part) == 0 || len(part) > 3 {
				return "", fmt.Errorf("invalid numeric format")
			}
		} else if len(part) != 3 {
			return "", fmt.Errorf("invalid numeric format")
		}
	}

	return strings.Join(parts, ""), nil
}

// Используем константы из пакета common для расчета ширины макета

// Модель очереди задач.
// Модель представляет собой очередь задач, которые выполняются последовательно.
type Model struct {
	tasks      []common.Task  // Список задач.
	current    int            // Индекс текущей задачи.
	title      string         // Заголовок очереди.
	titleStyle lipgloss.Style // Стиль заголовка.
	summary    string         // Сводка по выполненным задачам.
	// summaryStyle   lipgloss.Style // Стиль сводки.
	width           int            // Ширина экрана.
	quitting        bool           // Флаг завершения работы.
	stoppedOnError  bool           // Флаг прерывания очереди из-за ошибки.
	appName         string         // Название приложения.
	appVersion      string         // Версия приложения.
	appNameStyle    lipgloss.Style // Стиль названия приложения.
	appVersionStyle lipgloss.Style // Стиль версии приложения.
	errorTask       common.Task    // Задача, вызвавшая прерывание очереди.
	clearScreen     bool           // Флаг очистки экрана перед запуском
	hardClearScreen bool           // Флаг твердой очистки экрана перед запуском

	// Счетчики для подсчета результатов выполнения
	successCount int  // Количество успешно выполненных задач
	errorCount   int  // Количество задач с ошибками
	showSummary  bool // Флаг отображения сводки

	// Параметры отображения префиксов завершённых задач
	numberCompletedTasks bool   // Включает отображение номеров вместо символа завершения
	keepFirstSymbol      bool   // Если true, первая завершённая задача сохраняет символ
	numberFormat         string // Строка формата для отображения номера задачи

	// Параметры форматирования вывода результатов
	resultFormattingEnabled bool   // Включает форматирование результатов с разделительными линиями
	resultLinePrefix        string // Префикс для разделительной линии (по умолчанию "  │  ")
	resultLineLength        int    // Количество символов "─" в разделительной линии
}

type selectionSeparatorSetter interface {
	SetSelectionSeparatorEnabled(bool)
}

const defauiltNumberFormat = "[%02d]" // формат по умолчанию для отображения номеров задач

// New создает новую модель очереди с заданным заголовком и задачами.
func New(title string) *Model {
	return &Model{
		title:           title,
		summary:         defaults.SummaryCompleted,
		width:           common.DefaultWidth, // Начальная ширина
		showSummary:     true,                // По умолчанию сводка отображается
		clearScreen:     false,               // По умолчанию экран не очищается
		titleStyle:      lipgloss.NewStyle().Foreground(ui.ColorBrightWhite).Bold(true),
		appNameStyle:    lipgloss.NewStyle().Foreground(ui.ColorDarkGray).Background(ui.ColorBrightWhite).Bold(false),
		appVersionStyle: lipgloss.NewStyle().Foreground(ui.ColorBrightGray).Bold(false),
		numberFormat:    defauiltNumberFormat,
		hardClearScreen: true, // По умолчанию используется твердая очистка экрана (вместе с буфером)
		// Инициализация параметров форматирования результатов
		resultFormattingEnabled: true,                           // По умолчанию отключено
		resultLinePrefix:        "  │  ",                        // Префикс по умолчанию с символом │
		resultLineLength:        common.DefaultWidth * 93 / 100, // Длина линии перед выводом результатов задачи по умолчанию
	}
}

// Добавляет список задач для выполнения.
func (m *Model) AddTasks(tasks []common.Task) {
	// Создаем новый срез, куда будем добавлять только валидные (не nil) задачи.
	// Это более эффективно, чем многократно вызывать append для m.tasks в цикле.
	validTasks := make([]common.Task, 0, len(tasks))

	// Проходим по всем задачам, которые пришли в функцию.
	for _, task := range tasks {
		// Проверяем, не является ли задача nil.
		if task != nil {
			// Если задача не nil, добавляем ее в срез валидных задач.
			validTasks = append(validTasks, task)
		}
	}

	// Добавляем все валидные задачи в основной срез m.tasks одним вызовом append.
	m.tasks = append(m.tasks, validTasks...)

	if len(validTasks) > 0 {
		m.applySelectionSeparatorFlag(validTasks)
	}
}

// WithTasksNumbered включает нумерацию и задаёт формат представления числа в префиксе.
// Формат передаётся как строка для fmt.Sprintf (например, "[%02d]", "(%d)", "[0%d]").
// Если формат пустой, используется значение по умолчанию.
func (m *Model) WithTasksNumbered(enable bool, keepFirstSymbol bool, numberFormat string) *Model {
	m.numberCompletedTasks = enable
	ui.NumberingEnabled = enable
	m.keepFirstSymbol = keepFirstSymbol
	if strings.TrimSpace(numberFormat) == "" {
		m.numberFormat = defauiltNumberFormat
	} else {
		m.numberFormat = numberFormat
	}
	return m
}

// updateTaskStats обновляет статистику выполнения задач
func (m *Model) updateTaskStats() {
	m.successCount = 0
	m.errorCount = 0

	// Подсчитываем все задачи - просматриваем все до текущей позиции или все задачи если завершены
	tasksToCheck := m.current
	if m.current >= len(m.tasks) {
		tasksToCheck = len(m.tasks)
	}

	for i := 0; i < tasksToCheck; i++ {
		task := m.tasks[i]
		if task.IsDone() {
			if task.HasError() {
				m.errorCount++
			} else {
				m.successCount++
			}
		}
	}

	// Если очередь была остановлена из-за ошибки, и задача с ошибкой сохранена,
	// проверяем, учтена ли она в счетчике ошибок
	if m.stoppedOnError && m.errorTask != nil {
		// Проверяем, является ли задача с ошибкой текущей задачей
		if m.current < len(m.tasks) && m.tasks[m.current] == m.errorTask && m.errorTask.HasError() {
			// Если текущая задача с ошибкой не была учтена в цикле выше, учитываем её
			if m.current >= tasksToCheck {
				m.errorCount++
			}
		}
	}
}

// formatSummaryWithStats форматирует сводку с учетом статистики
func (m *Model) formatSummaryWithStats() (string, string) {
	totalTasks := len(m.tasks)
	completedTasks := m.successCount + m.errorCount

	// Формируем левую часть: summary + (успешных/всего)
	summaryLabel := sanitizeSummaryText(m.summary)
	if summaryLabel == "" {
		summaryLabel = defaults.SummaryCompleted
	}

	leftSummary := performance.FastConcat(
		summaryLabel,
		" (",
		performance.IntToString(m.successCount),
		"/",
		performance.IntToString(totalTasks),
		")",
	)

	// Формируем правую часть: УСПЕШНО или С ОШИБКАМИ
	var rightStatus string
	if m.errorCount > 0 {
		rightStatus = defaults.StatusProblem
	} else if completedTasks == totalTasks && completedTasks > 0 {
		rightStatus = defaults.StatusSuccess
	} else {
		// Для состояния "В ПРОЦЕССЕ" или когда нет завершенных задач
		rightStatus = defaults.StatusInProgress
	}

	return leftSummary, rightStatus
}

// renderSummaryFooter рендерит футер сводки
// @param leftSummary Левая часть сводки
// @param rightStatus Правая часть сводки
// @param layoutWidth Ширина макета
// @param summaryStyle Стиль для левой части
// @param rightStyle Стиль для правой части
// @return string
func (m *Model) renderSummaryFooter(leftSummary, rightStatus string, layoutWidth int, summaryStyle, rightStyle lipgloss.Style) string {
	cleanSummary := sanitizeSummaryText(leftSummary)
	if cleanSummary == "" {
		cleanSummary = defaults.SummaryCompleted
	}

	rightRendered := rightStyle.Render(rightStatus)
	actionSymbol := ui.FinishedLabelStyle.Render(ui.TaskCompletedSymbol)
	prefix := performance.FastConcat("  ", actionSymbol, "  ")
	prefixWidth := lipgloss.Width(prefix)
	continuationPrefix := performance.FastConcat(
		"  ",
		performance.RepeatEfficient(" ", lipgloss.Width(actionSymbol)),
		"  ",
	)

	const minGap = 2
	contentWidth := layoutWidth - common.LayoutWrapMargin
	if contentWidth < 1 {
		contentWidth = 1
	}

	rightWidth := lipgloss.Width(rightRendered)
	availableFirstWidth := contentWidth - prefixWidth - rightWidth - minGap
	if availableFirstWidth < 1 {
		availableFirstWidth = contentWidth - prefixWidth
	}
	if availableFirstWidth < 1 {
		availableFirstWidth = 1
	}

	subsequentWidth := contentWidth - prefixWidth
	if subsequentWidth < 1 {
		subsequentWidth = 1
	}

	lines := wrapSummaryText(cleanSummary, availableFirstWidth, subsequentWidth)
	if len(lines) == 0 {
		lines = []string{""}
	}

	var builder strings.Builder

	firstLeft := performance.FastConcat(prefix, summaryStyle.Render(lines[0]))
	builder.WriteString(ui.AlignTextToRight(firstLeft, rightRendered, layoutWidth))
	builder.WriteString("\n")

	for i := 1; i < len(lines); i++ {
		builder.WriteString(continuationPrefix)
		builder.WriteString(summaryStyle.Render(lines[i]))
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	builder.WriteString(ui.DrawLine(layoutWidth))
	builder.WriteString("\n")

	return builder.String()
}

// sanitizeSummaryText удаляет лишние пробелы и переносы строк
// @param summary Сводка
// @return string
func sanitizeSummaryText(summary string) string {
	if summary == "" {
		return ""
	}
	normalized := strings.ReplaceAll(summary, "\n", " ")
	fields := strings.Fields(normalized)
	return strings.Join(fields, " ")
}

// wrapSummaryText обрезает текст сводки
// @param text Текст сводки
// @param firstWidth Ширина первой строки
// @param otherWidth Ширина последующих строк
// @return []string
func wrapSummaryText(text string, firstWidth, otherWidth int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	if firstWidth < 1 {
		firstWidth = 1
	}
	if otherWidth < 1 {
		otherWidth = 1
	}

	runes := []rune(text)
	length := len(runes)
	var lines []string

	start := 0
	currentWidth := firstWidth

	for start < length {
		if currentWidth < 1 {
			currentWidth = 1
		}

		if start+currentWidth >= length {
			line := strings.TrimSpace(string(runes[start:]))
			if line == "" {
				line = string(runes[start:])
			}
			lines = append(lines, line)
			break
		}

		cut := summaryCutPoint(runes, start, currentWidth)
		if cut <= 0 {
			cut = currentWidth
		}

		line := strings.TrimSpace(string(runes[start : start+cut]))
		if line == "" {
			line = string(runes[start : start+cut])
		}
		lines = append(lines, line)
		start += cut

		for start < length && runes[start] == ' ' {
			start++
		}

		currentWidth = otherWidth
	}

	return lines
}

// summaryCutPoint находит оптимальную точку обрезания текста
// @param runes Срез символов текста
// @param start Начальная позиция
// @param width Ширина
// @return int
func summaryCutPoint(runes []rune, start, width int) int {
	end := start + width
	if end >= len(runes) {
		return len(runes) - start
	}
	for i := end; i > start; i-- {
		if runes[i-1] == ' ' {
			return i - start
		}
	}
	return width
}

func splitHeaderAndResult(lines []string, title string) ([]string, []string) {
	if len(lines) == 0 {
		return []string{""}, nil
	}

	separatorIdx := -1
	for i, line := range lines {
		if isSeparatorLine(line) {
			separatorIdx = i
			break
		}
	}

	if separatorIdx != -1 {
		header := append([]string(nil), lines[:separatorIdx]...)
		if len(header) == 0 {
			header = []string{""}
		}

		var result []string
		if separatorIdx+1 < len(lines) {
			result = append(result, lines[separatorIdx+1:]...)
		}
		return header, result
	}

	titlePlain := strings.TrimSpace(stripANSISequences(title))
	header := []string{lines[0]}
	resultStart := 1

	for i := 1; i < len(lines); i++ {
		raw := stripANSISequences(lines[i])
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			header = append(header, lines[i])
			resultStart = i + 1
			continue
		}

		normalized := strings.TrimLeft(trimmed, " "+ui.VerticalLineSymbol)
		if normalized == "" {
			header = append(header, lines[i])
			resultStart = i + 1
			continue
		}

		if titlePlain != "" && strings.Contains(titlePlain, normalized) {
			header = append(header, lines[i])
			resultStart = i + 1
			continue
		}

		resultStart = i
		break
	}

	var result []string
	if resultStart < len(lines) {
		result = append(result, lines[resultStart:]...)
	}

	if len(header) == 0 {
		header = []string{""}
	}

	return header, result
}

func isSeparatorLine(line string) bool {
	clean := stripANSISequences(line)
	clean = strings.TrimSpace(clean)
	if clean == "" {
		return false
	}

	for _, r := range clean {
		if string(r) != ui.HorizontalLineSymbol {
			return false
		}
	}

	return true
}

func stripANSISequences(s string) string {
	var builder strings.Builder
	builder.Grow(len(s))

	inEscape := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inEscape {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inEscape = false
			}
			continue
		}

		if c == 0x1b {
			inEscape = true
			continue
		}

		builder.WriteByte(c)
	}

	return builder.String()
}

// Run запускает очередь задач
// @return error
func (m *Model) Run() error {
	// Если установлен флаг очистки экрана, очищаем экран перед запуском
	if m.clearScreen {
		// Используем ANSI-последовательность для очистки экрана
		if m.hardClearScreen {
			fmt.Print(defaults.HardClearScreen) // Очищаем весь экран и его буфер
		} else {
			fmt.Print(defaults.ClearScreen) // Очищаем только область вывода
		}
	}

	_, err := tea.NewProgram(m).Run()
	return err
}

// WithTitleColor устанавливает цвет заголовка.
// @param titleColor Цвет заголовка
// @param bold Жирный стиль
// @return Указатель на очередь задач
func (m *Model) WithTitleColor(titleColor lipgloss.TerminalColor, bold bool) *Model {
	m.titleStyle = lipgloss.NewStyle().Foreground(titleColor).Bold(bold)
	return m
}

// WithAppName устанавливает название приложения.
// @param appName Название приложения
// @param version Необязательная строка версии приложения
// @return Указатель на очередь задач
func (m *Model) WithAppName(appName string, version ...string) *Model {
	// Устанавливаем название приложения
	m.appName = "  " + appName + "  "

	// Устанавливаем версию приложения
	if len(version) > 0 {
		trimmedVersion := strings.TrimSpace(version[0])
		if trimmedVersion != "" {
			m.appVersion = trimmedVersion
		} else {
			m.appVersion = ""
		}
	} else {
		m.appVersion = ""
	}

	return m
}

// WithAppNameStyle устанавливает стиль названия приложения.
// @param textColor Цвет текста
// @param bold Флаг, указывающий, нужно ли сделать текст курсивом
// @return Указатель на очередь задач
func (m *Model) WithAppNameColor(textColor lipgloss.TerminalColor, bold bool) *Model {
	m.appNameStyle = lipgloss.NewStyle().Foreground(textColor).Bold(bold).Background(ui.ColorBrightWhite)
	return m
}

// WithSummary устанавливает флаг отображения сводки.
// @param show Флаг, указывающий, нужно ли отображать сводку
// @return Указатель на очередь задач
func (m *Model) WithSummary(show bool) *Model {
	m.showSummary = show
	return m
}

// WithClearScreen устанавливает флаг очистки экрана перед запуском очереди задач.
// @param clear Флаг, указывающий, нужно ли очищать экран перед запуском очереди задач
// @param hardClear Флаг, указывающий, нужно ли использовать твердую очистку экрана (true) или нет (false)
// @return Указатель на очередь задач
func (m *Model) WithClearScreen(clear bool, hardClear ...bool) *Model {
	m.clearScreen = clear
	if len(hardClear) > 0 {
		m.hardClearScreen = hardClear[0]
	}
	return m
}

// WithResultFormatting включает форматирование результатов задач с разделительными линиями.
// Если enabled=true, то перед каждым результатом задачи будет добавляться разделительная линия
// из префикса и указанного количества символов "─".
// При enabled=false поведение остается как и раньше - результаты выводятся сразу после строки с задачей.
// @param enabled - включить/выключить форматирование
func (m *Model) WithResultFormatting(enabled bool) *Model {
	m.resultFormattingEnabled = enabled
	m.applySelectionSeparatorFlag(m.tasks)
	return m
}

// applySelectionSeparatorFlag применяет флаг разделительной линии для задач
// @param tasks Список задач
// @return void
func (m *Model) applySelectionSeparatorFlag(tasks []common.Task) {
	for _, task := range tasks {
		if setter, ok := task.(selectionSeparatorSetter); ok {
			setter.SetSelectionSeparatorEnabled(m.resultFormattingEnabled)
		}
	}
}

// SetErrorColor устанавливает цвет для отображения ошибок в очереди.
// @param color Цвет ошибки
// @return Указатель на очередь задач
func (m *Model) SetErrorColor(color ErrorColor) *Model {

	switch color {
	case Yellow:
		ui.SetErrorColor(ui.ColorDarkYellow, ui.ColorBrightYellow)
	case Red:
		ui.SetErrorColor(ui.ColorDarkRed, ui.ColorBrightRed)
	case Orange:
		ui.SetErrorColor(ui.ColorDarkOrange, ui.ColorBrightOrange)
	}
	return m
}

// layoutWidth вычисляет ширину для рендеринга задач.
// Использует функцию из пакета common.
// @return Ширина для рендеринга задач
func (m *Model) layoutWidth() int {
	return common.CalculateLayoutWidth(m.width)
}

// Init запускает первую задачу.
// @return Команда для запуска первой задачи
func (m *Model) Init() tea.Cmd {
	if len(m.tasks) > 0 {
		return m.tasks[0].Run()
	}
	return tea.Quit
}

// setTitle прорисовывает заголовок
// @param width Ширина окна
// @return Строка заголовка
func (m *Model) setTitle(width int) string {
	var result string
	if m.appName != "" {
		// Если есть версия, добавляем её в левую часть заголовка
		rightParts := make([]string, 0, 3)
		if m.appVersion != "" {
			rightParts = append(rightParts, m.appVersionStyle.Render(m.appVersion), " ")
		}
		// Добавляем отступ в конце
		rightParts = append(rightParts, m.appNameStyle.Render(m.appName))
		right := performance.FastConcat(rightParts...)
		title := m.titleStyle.Render(m.title)
		line := ui.AlignTextToRight(" "+title, right, width+1)
		line = strings.TrimRight(line, " ")
		result = line + "\n"
	} else {
		result = "  " + m.title + "\n"
	}
	return result
}

// Update обрабатывает сообщения и делегирует их задачам.
// @param msg Сообщение
// @return Обновленная модель и команда
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Обработка обновлений размера окна.
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
	}

	if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "Ctrl+c" {
		m.quitting = true
		return m, tea.Quit
	}

	if m.current >= len(m.tasks) {
		return m, tea.Quit
	}

	currentTask := m.tasks[m.current]
	var cmd tea.Cmd

	// Обновляем текущую задачу и сохраняем её обратно
	updatedTask, cmd := currentTask.Update(msg)
	m.tasks[m.current] = updatedTask

	// Проверяем завершение уже ОБНОВЛЁННОЙ задачи
	if m.tasks[m.current].IsDone() {
		// Обновляем статистику выполненных задач
		m.updateTaskStats()

		// Проверяем давление памяти и выполняем очистку если необходимо
		m.checkMemoryPressure()

		// Проверяем, есть ли ошибка в текущей задаче
		if updatedTask.HasError() && updatedTask.StopOnError() {
			// Если есть ошибка и флаг StopOnError установлен,
			// прекращаем выполнение очереди и сохраняем информацию об ошибке
			m.stoppedOnError = true
			m.errorTask = updatedTask

			// Явно увеличиваем счетчик ошибок для текущей задачи
			m.errorCount++

			// Финальное обновление статистики перед остановкой из-за ошибки
			m.updateTaskStats()
			return m, tea.Quit
		}

		// Иначе переходим к следующей задаче
		m.current++
		if m.current < len(m.tasks) {
			nextCmd := m.tasks[m.current].Run()
			return m, tea.Batch(cmd, nextCmd)
		}
		// Финальное обновление статистики при завершении всех задач
		m.updateTaskStats()
		return m, tea.Quit
	}

	return m, cmd
}

// View отображает список задач.
// @return string - отображаемый список задач
func (m *Model) View() string {

	// Если очередь завершена, отображаем просто надпись о завершении
	// без прорисовки задач
	// if m.quitting {
	// 	return ui.CancelStyle.Render(task.DefaultCancelLabel) + "\n"
	// }

	var sb strings.Builder

	layoutWidth := m.layoutWidth()

	// Используем настроенный заголовок.
	sb.WriteString(ui.DrawLine(layoutWidth))
	sb.WriteString(m.setTitle(layoutWidth))
	sb.WriteString(ui.DrawLine(layoutWidth) + "\n")

	lastTaskIndex := len(m.tasks) - 1
	for i, t := range m.tasks {
		// Если задача завершена, отображаем её с форматированием
		if i < m.current {
			// Проверяем, есть ли ошибка в задаче
			hasError := t.HasError()
			stripPrefixes := !m.showSummary && lastTaskIndex >= 0 && i == lastTaskIndex
			// Применяем префикс завершённой задачи
			m.applyCompletedTaskPrefix(t, i, hasError)
			// Завершенные задачи: отображаем их с форматированием (или без, если отключено)
			sb.WriteString(m.formatTaskResult(t, layoutWidth, stripPrefixes))
			// Добавляем префикс для следующей задачи в виде пустой строки с префиксом "  │"
			sb.WriteString(ui.GetTaskBelowPrefix() + "\n")

		} else if i == m.current {
			// Если задача не завершена, отображаем её в интерактивном виде
			hasError := t.HasError()
			// Если задача завершена, отображаем её с форматированием
			if t.IsDone() {
				stripPrefixes := !m.showSummary && (i == lastTaskIndex || m.stoppedOnError)
				m.applyCompletedTaskPrefix(t, i, hasError)
				sb.WriteString(m.formatTaskResult(t, layoutWidth, stripPrefixes) + "\n")
			} else {
				// Если задача не завершена, отображаем её в интерактивном виде
				m.applyInProgressTaskPrefix(t, i, hasError)
				// Активная задача: отображаем ее интерактивный вид.
				// Обрезаем только завершающие символы новой строки, сохраняя ведущие пробелы
				// view := strings.TrimRight(t.View(layoutWidth), "\n")
				view := t.View(layoutWidth)
				sb.WriteString(view + "\n")
			}
			// Добавляем разделитель, если есть ожидающие задачи и нет ошибки
			if i+1 < len(m.tasks) && (!m.stoppedOnError && !hasError) {
				sb.WriteString(ui.DrawLine(layoutWidth))
			}
		}
	}

	// Убираем крайнюю линию, если она есть
	removeDuplicateLines(&sb)

	// Добавляем финальную разделительную линию
	// Если есть активная задача, добавляем обычную линию
	// Иначе добавляем специальную линию
	if m.current < len(m.tasks) && !m.stoppedOnError {
		sb.WriteString(ui.DrawLine(layoutWidth))
	} else {

		// Отображаем сводку только если включен флаг showSummary
		if m.showSummary {
			ensureTrailingPrefixLine(&sb)
			// Получаем форматированную сводку с статистикой
			leftSummary, rightStatus := m.formatSummaryWithStats()

			// Определяем стиль для правой части в зависимости от статуса
			var rightStyle lipgloss.Style
			switch rightStatus {
			case defaults.StatusSuccess:
				rightStyle = ui.SuccessLabelStyle
			case defaults.StatusProblem:
				rightStyle = ui.GetErrorStatusStyle()
			case defaults.StatusInProgress:
				rightStyle = ui.SubtleStyle
			default:
				rightStyle = ui.SubtleStyle
			}

			// Определяем стиль для левой части (summary) - используем те же стили что и для правой части
			var summaryStyle lipgloss.Style
			switch rightStatus {
			case defaults.StatusSuccess:
				summaryStyle = ui.SuccessLabelStyle // Тот же стиль что и для правой части "SUCCESS"
			case defaults.StatusProblem:
				summaryStyle = ui.GetErrorStatusStyle() // Тот же стиль что и для правой части при ошибках
			case defaults.StatusInProgress:
				// Для состояния "В ПРОЦЕССЕ" используем тот же стиль что и справа
				summaryStyle = ui.SubtleStyle
			default:
				summaryStyle = ui.SubtleStyle
			}

			footer := m.renderSummaryFooter(leftSummary, rightStatus, layoutWidth, summaryStyle, rightStyle)
			sb.WriteString(footer)
		} else {
			// Убираем висящий префикс вертикальной линии перед пустой строкой
			removeTrailingTaskBelowPrefix(&sb)
			// Заменяем вертикальные линии перед символами задач ПЕРЕД добавлением финальных элементов
			removeVerticalLinesBeforeTaskSymbols(&sb)
			ensureSingleBlankLine(&sb)
			sb.WriteString(ui.DrawLine(layoutWidth))
		}
	}

	return sb.String()
}

// applyCompletedTaskPrefix настраивает префикс завершённой задачи в зависимости от параметров модели
// @param task - задача
// @param index - индекс задачи
// @param hasError - есть ли ошибка в задаче
func (m *Model) applyCompletedTaskPrefix(task common.Task, index int, hasError bool) {
	setter, ok := task.(interface{ SetCompletedPrefix(string) })
	if !ok {
		return
	}

	if !m.numberCompletedTasks {
		setter.SetCompletedPrefix("")
		return
	}

	if m.keepFirstSymbol && index == 0 && !hasError {
		setter.SetCompletedPrefix("")
		return
	}

	number := index + 1
	if m.keepFirstSymbol && !hasError {
		number = index
		if number <= 0 {
			number = 1
		}
	}
	setter.SetCompletedPrefix(buildCompletedPrefix(number, m.numberFormat))
}

// applyInProgressTaskPrefix настраивает префикс активной задачи в зависимости от параметров модели
// @param task - задача
// @param index - индекс задачи
// @param hasError - есть ли ошибка в задаче
func (m *Model) applyInProgressTaskPrefix(task common.Task, index int, hasError bool) {
	setter, ok := task.(interface{ SetInProgressPrefix(string) })
	if !ok {
		return
	}

	if !m.numberCompletedTasks {
		setter.SetInProgressPrefix("")
		return
	}

	if m.keepFirstSymbol && index == 0 && !hasError {
		setter.SetInProgressPrefix("")
		return
	}

	number := index + 1
	if m.keepFirstSymbol && !hasError {
		number = index
		if number <= 0 {
			number = 1
		}
	}

	setter.SetInProgressPrefix(buildInProgressPrefix(number, m.numberFormat))
}

// calculateResultLinePrefix вычисляет префикс для разделительной линии и строк результата
// в зависимости от настроек нумерации
func (m *Model) calculateResultLinePrefix() string {
	if m.numberCompletedTasks {
		// При включенной нумерации нужно учесть ширину номера
		// Например, для "[1]  " нужно "     │  " (5 пробелов + │ + 2 пробела)
		// Вычисляем ширину номера для максимального номера задачи
		maxNumber := len(m.tasks)
		sampleNumber := fmt.Sprintf(m.numberFormat, maxNumber)
		numberWidth := len(sampleNumber)

		// Добавляем 2 пробела после номера (как в основном префиксе) + пробел перед │
		totalSpaces := numberWidth - 1 // +2 после номера, +1 перед │
		return performance.FastConcat(
			performance.RepeatEfficient(" ", totalSpaces),
			ui.VerticalLineSymbol,
			performance.RepeatEfficient(" ", 3), // 2 пробела после │
		)
	} else {
		// При отключенной нумерации используем стандартный префикс
		return m.resultLinePrefix
	}
}

// formatTaskResult форматирует результат задачи с разделительной линией
// Создает линию из префикса и символов "─", затем выводит результат задачи с новой строки
// stripVerticalPrefixes управляет заменой вертикальной линии на пробел для последнего блока без сводки
func (m *Model) formatTaskResult(task common.Task, width int, stripVerticalPrefixes bool) string {
	if !m.resultFormattingEnabled {
		// Если форматирование отключено, возвращаем обычное представление
		view := task.FinalView(width)
		if stripVerticalPrefixes {
			return stripResultPrefixes(view)
		}
		return view
	}

	var result strings.Builder

	// Получаем обычное представление задачи
	taskView := task.FinalView(width)

	// Разделяем на строки
	lines := strings.Split(taskView, "\n")
	if len(lines) == 0 {
		return taskView
	}

	// Отделяем заголовок от дополнительных строк результата
	headerLines, resultLines := splitHeaderAndResult(lines, task.Title())

	// Добавляем заголовок полностью (включая перенесенные строки)
	result.WriteString(strings.Join(headerLines, "\n"))

	// Если есть дополнительные строки (результаты), добавляем разделительную линию
	if len(resultLines) > 0 {
		result.WriteString("\n")

		// Определяем стиль линии на основе типа результата задачи
		var lineStyle lipgloss.Style
		if task.HasError() {
			// Для ошибок используем очень приглушенный желтый цвет (более приглушенный чем текст ошибки)
			lineStyle = ui.VerySubtleErrorStyle
		} else {
			// Для успешных результатов используем очень приглушенный стиль (едва заметный)
			lineStyle = ui.VerySubtleStyle
		}

		// Создаем стилизованную разделительную линию той же ширины, что и у активной задачи
		dynamicPrefix := m.calculateResultLinePrefix()
		lineLength := width - lipgloss.Width(dynamicPrefix) - 1
		if lineLength < 0 {
			lineLength = 0
		}
		separatorContent := performance.RepeatEfficient(ui.HorizontalLineSymbol, lineLength)
		styledSeparator := lineStyle.Render(separatorContent)

		// Используем динамический префикс для разделительной линии
		separatorLine := performance.FastConcat(dynamicPrefix, styledSeparator)
		result.WriteString(separatorLine + "\n")

		// Добавляем остальные строки результата
		for i, line := range resultLines {
			if strings.TrimSpace(line) != "" {
				result.WriteString(line)
				if i < len(resultLines)-1 {
					result.WriteString("\n")
				}
			}
		}
	}

	finalResult := result.String()

	// Определяем, была ли задача отменена пользователем
	cancelDetected := false
	cancelIndicators := []string{
		defaults.TaskCancelledByUser,
		defaults.ErrorMsgCanceled,
		defaults.CancelShort,
	}
	// Проверяем наличие индикаторов отмены
	for _, indicator := range cancelIndicators {
		if indicator != "" && strings.Contains(finalResult, indicator) {
			cancelDetected = true
			break
		}
	}
	if cancelDetected {
		finalResult = ensureTrailingSingleNewline(finalResult)
	}

	if stripVerticalPrefixes {
		finalResult = stripResultPrefixes(finalResult)
		if cancelDetected {
			finalResult = ensureTrailingSingleNewline(finalResult)
		}
	}

	return finalResult
}

/**
 * hasHiddenResultLine проверяет, есть ли среди задач завершённые элементы без строки результата.
 *
 * @param tasks Список задач очереди.
 * @return true, если найдена хотя бы одна завершённая задача с отключённой строкой результата.
 */
func hasHiddenResultLine(tasks []common.Task) bool {
	for _, task := range tasks {
		visibleProvider, ok := task.(interface{ ResultLineVisible() bool })
		if !ok {
			continue
		}
		if !visibleProvider.ResultLineVisible() && task.IsDone() {
			return true
		}
	}
	return false
}

// buildCompletedPrefix возвращает префикс завершённой задачи, учитывая формат номера.
func buildCompletedPrefix(number int, format string) string {
	indent := ui.MainLeftIndent - 1
	if indent < 0 {
		indent = 0
	}
	return performance.FastConcat(
		performance.RepeatEfficient(" ", indent),
		formatTaskNumber(number, format),
	)
}

// buildInProgressPrefix возвращает префикс активной задачи с учётом формата номера.
func buildInProgressPrefix(number int, format string) string {
	indent := ui.MainLeftIndent - 1
	if indent < 0 {
		indent = 0
	}
	return performance.FastConcat(
		performance.RepeatEfficient(" ", indent),
		formatTaskNumber(number, format),
		" ",
	)
}

// formatTaskNumber форматирует номер задачи на основе заданного шаблона fmt.Sprintf.
// Если переданный формат пустой, используется формат по умолчанию.
func formatTaskNumber(number int, format string) string {
	if number <= 0 {
		number = 1
	}
	if strings.TrimSpace(format) == "" {
		format = defauiltNumberFormat
	}
	if !strings.Contains(format, "%") {
		format = format + "%d"
	}
	return fmt.Sprintf(format, number)
}

// Убираем из потока вывода крайнюю линию, если она есть
func removeDuplicateLines(sb *strings.Builder) {

	// Обработка финальной разделительной линии
	// Получаем текущее содержимое Builder для анализа
	content := sb.String()

	// Находим последний перенос строки для выделения последней строки
	lastNewlineIndex := strings.LastIndex(content, "\n")
	if lastNewlineIndex != -1 {
		// Проверяем, не является ли последняя строка пустой (только перенос строки)
		if lastNewlineIndex == len(content)-1 {
			// Если последняя строка пустая, ищем предыдущий перенос строки
			lastNewlineIndex = strings.LastIndex(content[:lastNewlineIndex], "\n")
		}

		if lastNewlineIndex != -1 && lastNewlineIndex < len(content)-1 {
			// Получаем последнюю строку (или предпоследнюю, если последняя пустая)
			lastLine := content[lastNewlineIndex+1:]
			// Убираем возможный перенос строки в конце
			lastLine = strings.TrimSuffix(lastLine, "\n")

			// Проверяем, является ли строка горизонтальной линией
			// (состоит только из символов HorizontalLineSymbol)
			isHorizontalLine := strings.Contains(lastLine, ui.HorizontalLineSymbol)

			// Если строка - горизонтальная линия и не пустая, удаляем её
			if isHorizontalLine && len(lastLine) > 0 {
				// Очищаем Builder и записываем контент без последней линии-разделителя
				sb.Reset()
				sb.WriteString(content[:lastNewlineIndex+1])
			}
		}
	}
}

// removeTrailingTaskBelowPrefix убирает последнюю строку, состоящую только из префикса вертикальной линии
// Это необходимо когда итоговая сводка отключена и перед финальной линией должна быть пустая строка
func removeTrailingTaskBelowPrefix(sb *strings.Builder) {
	if sb == nil {
		return
	}

	content := sb.String()
	suffix := ui.GetTaskBelowPrefix() + "\n"
	if strings.HasSuffix(content, suffix) {
		sb.Reset()
		sb.WriteString(content[:len(content)-len(suffix)])
	}
}

// stripResultPrefixes заменяет вертикальную линию на пробел в строках результатов
// Используется для последнего блока при отключенной сводке, чтобы не рисовать хвост линии
func stripResultPrefixes(block string) string {
	if block == "" {
		return block
	}

	lines := strings.Split(block, "\n")
	for i := 1; i < len(lines)-1; i++ {
		lines[i] = replaceFirstVerticalSymbol(lines[i])
	}
	return strings.Join(lines, "\n")
}

func replaceFirstVerticalSymbol(line string) string {
	idx := strings.Index(line, ui.VerticalLineSymbol)
	if idx == -1 {
		return line
	}
	return performance.FastConcat(line[:idx], " ", line[idx+len(ui.VerticalLineSymbol):])
}

func ensureTrailingSingleNewline(value string) string {
	if value == "" {
		return value
	}
	trimmed := strings.TrimRight(value, "\n")
	return trimmed + "\n"
}

func ensureSingleBlankLine(sb *strings.Builder) {
	if sb == nil {
		return
	}
	content := sb.String()
	trimmed := strings.TrimRight(content, "\n")
	sb.Reset()
	sb.WriteString(trimmed)
	sb.WriteString("\n\n")
}

func ensureTrailingPrefixLine(sb *strings.Builder) {
	if sb == nil {
		return
	}
	content := sb.String()
	trimmed := strings.TrimRight(content, "\n")
	prefix := ui.GetTaskBelowPrefix()
	var lines []string
	if trimmed != "" {
		lines = strings.Split(trimmed, "\n")
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	hasPrefix := len(lines) > 0 && lines[len(lines)-1] == prefix
	sb.Reset()
	if len(lines) > 0 {
		sb.WriteString(strings.Join(lines, "\n"))
		sb.WriteString("\n")
	}
	if !hasPrefix {
		sb.WriteString(prefix)
		sb.WriteString("\n")
	}
}

// removeVerticalLinesBeforeTaskSymbols убирает вертикальные линии, ведущие к последнему (самому нижнему)
// символу задачи (TaskCompletedSymbol/TaskInProgressSymbol). Это позволяет «обрубить» вертикальные
// ветки, чтобы они не тянулись до конца списка задач.
//
// Алгоритм:
//  1. Ищем строку и колонку (в рунах) последнего символа задачи
//  2. Ищем самый нижний вертикальный сегмент в той же колонке и меняем его на пробел
func removeVerticalLinesBeforeTaskSymbols(sb *strings.Builder) {
	// Конвертируем буфер в строки
	content := sb.String()
	lines := strings.Split(content, "\n")

	// 1. Находим строку и колонку (в рунах) последнего символа задачи
	lastLine := -1
	col := -1
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		completed := strings.Index(line, ui.TaskCompletedSymbol)
		progress := strings.Index(line, ui.TaskInProgressSymbol)
		if completed == -1 && progress == -1 {
			continue
		}
		bytePos := completed
		if bytePos == -1 || (progress != -1 && progress < completed) {
			bytePos = progress
		}
		lastLine = i
		col = utf8.RuneCountInString(line[:bytePos])
		break
	}

	if lastLine == -1 {
		return // символы задач не найдены
	}

	// 2. Ищем все вертикальные сегменты в той же колонке после последней задачи и заменяем их на пробелы
	for i := lastLine + 1; i < len(lines); i++ {
		runes := []rune(lines[i])
		if col < len(runes) && string(runes[col]) == ui.VerticalLineSymbol {
			// Заменяем вертикальную линию на пробел
			runes[col] = ' '
			lines[i] = string(runes)
		}
	}

	// 3. Обновляем builder
	sb.Reset()
	sb.WriteString(strings.Join(lines, "\n"))
}

// DrawSpecialLine создает горизонтальную линию заданной ширины c угловой линией вверху
// типа ──┴─
func DrawFooterLine(width int) string {
	return performance.FastConcat(
		ui.HorizontalLineSymbol,
		ui.FinishedLabelStyle.Render(" "+ui.TaskCompletedSymbol+" "),
		performance.RepeatEfficient(ui.HorizontalLineSymbol, width-4), "\n")
}

// checkMemoryPressure проверяет использование памяти и выполняет очистку при необходимости
func (m *Model) checkMemoryPressure() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// Если использование памяти превышает порог, выполняем экстренную очистку
	if ms.Sys > MemoryPressureThreshold {
		m.emergencyCleanup()
		// Принудительный сборщик мусора для embedded устройств
		runtime.GC()
	}
}

// emergencyCleanup выполняет экстренную очистку памяти
func (m *Model) emergencyCleanup() {
	// Ограничиваем количество завершенных задач
	m.cleanupOldTasks()

	// Очищаем буферные пулы через performance пакет
	performance.EmergencyPoolCleanup()

	// Очищаем кэш интернирования строк
	ui.ClearInternCache()
}

// cleanupOldTasks ограничивает количество завершенных задач в памяти
func (m *Model) cleanupOldTasks() {
	if m.current <= MaxCompletedTasks {
		return // Нет необходимости в очистке
	}

	// Сохраняем только последние MaxCompletedTasks завершенных задач
	// плюс все активные/незавершенные задачи
	keepFrom := m.current - MaxCompletedTasks
	if keepFrom < 0 {
		keepFrom = 0
	}

	// Создаем новый срез с ограниченным количеством задач
	newTasks := make([]common.Task, len(m.tasks)-keepFrom)
	copy(newTasks, m.tasks[keepFrom:])

	// Обновляем индекс текущей задачи
	m.current -= keepFrom
	m.tasks = newTasks
}
