// task/base.go

package task

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/performance"
	"github.com/qzeleza/ziva/internal/ui"
)

// Task - псевдоним интерфейса common.Task для обратной совместимости
type Task = common.Task

// BaseTask contains common fields for all tasks.
type BaseTask struct {
	title       string
	done        bool
	icon        string // Icon to show when done (e.g., check or cross)
	finalValue  string // The final value to display (e.g., "Yes", "Option 1")
	err         error  // Ошибка, если задача завершилась с ошибкой
	stopOnError bool   // Флаг, указывающий, нужно ли останавливать очередь при ошибке
	// customCompletedPrefix позволяет очереди подменять префикс завершённой задачи (например, на номер)
	completedPrefix string
	// inProgressPrefix используется очередью для нумерации активных задач
	inProgressPrefix string

	// Флаг отображения разделительной линии в активных задачах выбора
	showSelectionSeparator bool

	// Флаг, указывающий, нужно ли сохранять переносы строк в сообщениях об ошибках
	preserveErrorNewLines bool

	// Поля для управления тайм-аутом
	timeoutManager *TimeoutManager // Менеджер тайм-аута
	timeoutEnabled bool            // Флаг, указывающий, включен ли тайм-аут
	defaultValue   interface{}     // Значение по умолчанию, которое будет выбрано при тайм-ауте
	showTimeout    bool            // Флаг, указывающий, нужно ли отображать оставшееся время
}

func NewBaseTask(title string) BaseTask {
	return BaseTask{
		title:                  title,
		stopOnError:            true, // По умолчанию останавливаем очередь при ошибке
		preserveErrorNewLines:  true, // По умолчанию сохраняем переносы строк в ошибках - печатаем "как есть".
		timeoutEnabled:         false,
		showTimeout:            true, // По умолчанию отображаем оставшееся время
		showSelectionSeparator: true,
	}
}

func (t *BaseTask) Title() string { return t.title }
func (t *BaseTask) IsDone() bool  { return t.done }

func (t *BaseTask) Run() tea.Cmd {
	// Если тайм-аут включен, запускаем таймер
	if t.timeoutEnabled && t.timeoutManager != nil {
		return t.timeoutManager.StartTimeout()
	}
	return nil
}

func (t *BaseTask) Update(msg tea.Msg) (Task, tea.Cmd) {
	// Базовый Update не обрабатывает сообщения - это делают конкретные задачи
	// Возвращаем задачу без изменений
	return t, nil
}

// HasError возвращает true, если при выполнении задачи произошла ошибка.
func (t *BaseTask) HasError() bool { return t.err != nil }

// Error возвращает ошибку, если она есть.
func (t *BaseTask) Error() error { return t.err }

// StopOnError возвращает true, если при возникновении ошибки в этой задаче
// нужно остановить выполнение всей очереди задач.
func (t *BaseTask) StopOnError() bool { return t.stopOnError }

// SetStopOnError устанавливает флаг остановки очереди при ошибке.
func (t *BaseTask) SetStopOnError(stop bool) { t.stopOnError = stop }

// WithNewLinesInErrors устанавливает режим сохранения переносов строк в сообщениях об ошибках.
// Если preserve=true, то оригинальные переносы строк сохраняются.
// Если preserve=false, то переносы строк удаляются и текст переформатируется.
// @return Интерфейс Task для цепочки вызовов
func (t *BaseTask) WithNewLinesInErrors(preserve bool) common.Task {
	t.preserveErrorNewLines = preserve
	return t
}

// SetError устанавливает ошибку для задачи
func (t *BaseTask) SetError(err error) { t.err = err }

// View provides a defauilt implementation for active tasks.
func (t *BaseTask) View(_ int) string {
	// Most active tasks manage their own view, so this is a fallback.
	return t.title
}

// WithTimeout устанавливает тайм-аут для задачи
// @param duration Длительность тайм-аута
// @param defaultValue Значение, которое будет выбрано при тайм-ауте
// @return Указатель на текущую задачу для цепочки вызовов
func (t *BaseTask) WithTimeout(duration time.Duration, defaultValue interface{}) *BaseTask {
	t.timeoutManager = NewTimeoutManager(duration)
	t.timeoutEnabled = true
	t.defaultValue = defaultValue
	return t
}

// DisableTimeout отключает тайм-аут
func (t *BaseTask) DisableTimeout() *BaseTask {
	if t.timeoutManager != nil {
		t.timeoutManager.StopTimeout()
	}
	t.timeoutEnabled = false
	return t
}

// ShowTimeout устанавливает флаг отображения оставшегося времени
// @param show true - отображать, false - не отображать
// @return Указатель на текущую задачу для цепочки вызовов
func (t *BaseTask) ShowTimeout(show bool) *BaseTask {
	t.showTimeout = show
	return t
}

// SetSelectionSeparatorEnabled управляет отображением разделительной линии в активных задачах выбора.
func (t *BaseTask) SetSelectionSeparatorEnabled(enabled bool) {
	t.showSelectionSeparator = enabled
}

// SelectionSeparatorEnabled сообщает, активна ли разделительная линия в задачах выбора.
func (t *BaseTask) SelectionSeparatorEnabled() bool {
	return t.showSelectionSeparator
}

// GetRemainingTime возвращает оставшееся время в секундах
func (t *BaseTask) GetRemainingTime() int {
	if t.timeoutManager != nil && t.timeoutEnabled {
		return t.timeoutManager.RemainingTime()
	}
	return 0
}

// GetRemainingTimeFormatted возвращает оставшееся время в формате MM:SS
func (t *BaseTask) GetRemainingTimeFormatted() string {
	if t.timeoutManager != nil && t.timeoutEnabled {
		return t.timeoutManager.RemainingTimeFormatted()
	}
	return ""
}

// RenderTimer возвращает отформатированную строку с таймером для отображения справа от заголовка
// Если таймер не активен или showTimeout = false, возвращает пустую строку
func (t *BaseTask) RenderTimer() string {
	if !t.timeoutEnabled || !t.showTimeout || t.timeoutManager == nil || !t.timeoutManager.IsActive() {
		return ""
	}

	remaining := t.GetRemainingTimeFormatted()
	if remaining == "" {
		return ""
	}

	return ui.SubtleStyle.Render(fmt.Sprintf("[%s]", remaining))
}

// FinalView handles right-alignment for all tasks and formats error messages.
//
// @param width Ширина макета для выравнивания текста
// @return Отформатированное представление задачи с выравниванием
func (t *BaseTask) FinalView(width int) string {
	// Используем константы из пакета common для расчета оптимальной ширины
	// если переданная ширина меньше минимальной
	if width < common.DefaultWidth {
		width = common.DefaultWidth
	}

	// Определяем успешность выполнения задачи
	success := !t.HasError() && t.finalValue != defaults.TaskStatusCancelled

	// Определяем тип задачи для выбора правильного префикса
	isTextInputTask := IsTextInputTask(t)

	// Создаем префикс для завершенной задачи с новой системой отображения
	var prefix string
	if isTextInputTask {
		prefix = ui.GetCompletedInputTaskPrefix(success)
	} else {
		prefix = ui.GetCompletedTaskPrefix(success)
	}
	if t.completedPrefix != "" {
		prefix = t.completedPrefix
	}

	// Для простых значений Yes/No используем отдельные стили для "Да" и "Нет"
	if t.finalValue == defaults.DefaultYes || t.finalValue == defaults.DefaultNo {
		titleStyle := lipgloss.NewStyle()
		if t.finalValue == defaults.DefaultNo {
			titleStyle = ui.GetErrorStatusStyle()
		}

		var right string
		if t.finalValue == defaults.DefaultYes {
			right = ui.TaskStatusSuccessStyle.Render(t.finalValue)
		} else {
			right = ui.GetErrorStatusStyle().Render(t.finalValue)
		}
		return renderTitleWithWrap(prefix, t.title, titleStyle, right, width)
	}

	// Для ошибок выводим текст ошибки с отступом и слово "Ошибка" справа
	if t.icon == ui.IconError {
		titleStyle := ui.GetErrorStatusStyle()
		right := ui.GetErrorStatusStyle().Render(defaults.TaskStatusError)
		header := renderTitleWithWrap(prefix, t.title, titleStyle, right, width)

		var result strings.Builder
		result.WriteString(header)
		result.WriteString("\n")

		// Форматируем текст ошибки с отступом и переносами строк
		errText := ""
		// Получаем текст ошибки из finalValue, так как это уже отрендеренный текст
		if t.finalValue != "" {
			// Убираем стилизацию из текста ошибки
			errText = strings.ReplaceAll(t.finalValue, ui.IconError, "")
			errText = strings.TrimSpace(errText)
		}

		// Добавляем отформатированный текст ошибки
		// Используем параметр preserveErrorNewLines для управления форматированием
		errorMsg := ui.FormatErrorMessage(errText, common.CalculateLayoutWidth(width), t.preserveErrorNewLines)
		result.WriteString(errorMsg)

		return result.String()
	}

	// Для обычных задач используем стандартное форматирование с новым префиксом
	if t.finalValue != "" && !strings.Contains(t.finalValue, t.title) {
		titleStyle := lipgloss.NewStyle()
		if !success {
			titleStyle = ui.GetErrorStatusStyle()
		}

		// Формируем статус
		statusLabel := strings.ToUpper(defaults.DefaultSuccessLabel)
		statusStyle := ui.TaskStatusSuccessStyle
		if t.icon == ui.IconCancelled {
			statusLabel = defaults.ErrorTypeUserCancel
			statusStyle = ui.ErrorMessageStyle
		} else if !success {
			statusLabel = defaults.TaskStatusError
			statusStyle = ui.GetErrorStatusStyle()
		}

		// Формируем правую часть заголовка
		right := statusStyle.Render(statusLabel)
		// Формируем заголовок
		header := renderTitleWithWrap(prefix, t.title, titleStyle, right, width)

		// Формируем основной текст
		var result strings.Builder
		result.WriteString(header)
		// Добавляем основной текст
		if t.icon == ui.IconCancelled {
			// Убираем стилизацию из текста ошибки
			trimmedValue := strings.TrimSpace(t.finalValue)
			if trimmedValue != "" {
				result.WriteString("\n")
				// Формируем префикс для основного текста
				valuePrefix := performance.FastConcat(
					performance.RepeatEfficient(" ", ui.MainLeftIndent),
					ui.VerticalLineSymbol,
					ui.GetResultIndentWhenNumberingEnabled(),
				)
				// Переносим текст ошибки
				wrapped := wrapTextWithPrefix(valuePrefix, trimmedValue, width, lipgloss.NewStyle())
				for i, line := range wrapped {
					if i > 0 {
						result.WriteString("\n")
					}
					result.WriteString(line)
				}
				result.WriteString("\n")
			}
		}

		return result.String()
	}

	// Если finalValue уже содержит полное форматирование, возвращаем как есть
	if t.finalValue != "" {
		return t.finalValue
	}

	// Запасной вариант - просто отображаем заголовок с префиксом
	titleStyle := lipgloss.NewStyle()
	if !success {
		titleStyle = ui.GetErrorStatusStyle()
	}
	return renderTitleWithWrap(prefix, t.title, titleStyle, "", width)
}

// renderTitleWithWrap отображает заголовок с префиксом и правым текстом
// @param prefix - префикс заголовка
// @param title - заголовок
// @param titleStyle - стиль заголовка
// @param right - правый текст
// @param width - ширина
// @return string - отформатированный заголовок
func renderTitleWithWrap(prefix, title string, titleStyle lipgloss.Style, right string, width int) string {
	titlePrefix := ensurePrefixSpacing(prefix)                      // добавляет отступы к префиксу
	continuationPrefix := buildContinuationTitlePrefix(titlePrefix) // создает префикс для продолжения заголовка

	// вычисляет эффективную ширину
	effectiveWidth := width - common.LayoutWrapMargin
	if effectiveWidth < 1 {
		effectiveWidth = 1
	}

	// вычисляет ширину префикса
	prefixWidth := lipgloss.Width(titlePrefix)
	if prefixWidth >= effectiveWidth {
		prefixWidth = effectiveWidth - 1
	}

	// вычисляет ширину правого текста
	rightWidth := lipgloss.Width(right)
	const minGap = 2

	// вычисляет ширину первой строки
	firstWidth := effectiveWidth - prefixWidth
	if right != "" {
		firstWidth = effectiveWidth - prefixWidth - rightWidth - minGap
	}
	if firstWidth < 1 {
		firstWidth = 1
	}

	// вычисляет ширину остальных строк
	otherWidth := effectiveWidth - prefixWidth
	if otherWidth < 1 {
		otherWidth = 1
	}

	// оборачивает заголовок в несколько строк
	wrapped := wrapTitleText(title, firstWidth, otherWidth)
	if len(wrapped) == 0 {
		wrapped = []string{""}
	}

	// формирует заголовок
	firstLeft := titlePrefix + titleStyle.Render(wrapped[0])
	header := ui.AlignTextToRight(firstLeft, right, width) // выравнивает заголовок по правому краю

	if len(wrapped) == 1 {
		return header
	}

	// формирует заголовок
	var builder strings.Builder
	builder.WriteString(header)
	// добавляет остальные строки
	for i := 1; i < len(wrapped); i++ {
		builder.WriteString("\n")
		builder.WriteString(continuationPrefix)
		builder.WriteString(titleStyle.Render(wrapped[i]))
	}

	return builder.String()
}

// ensurePrefixSpacing добавляет отступы к префиксу
// @param prefix - префикс
// @return string - префикс с отступами
func ensurePrefixSpacing(prefix string) string {
	switch {
	case strings.HasSuffix(prefix, "  "):
		return prefix
	case strings.HasSuffix(prefix, " "):
		return prefix + " "
	default:
		return prefix + "  "
	}
}

func wrapTitleText(text string, firstWidth, otherWidth int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{""}
	}

	runes := []rune(text)
	var lines []string
	start := 0
	currentWidth := firstWidth

	for start < len(runes) {
		if currentWidth < 1 {
			currentWidth = 1
		}

		if start+currentWidth >= len(runes) {
			line := strings.TrimSpace(string(runes[start:]))
			if line == "" {
				line = string(runes[start:])
			}
			lines = append(lines, line)
			break
		}

		cut := start + currentWidth
		for cut > start && runes[cut-1] != ' ' {
			cut--
		}
		if cut == start {
			cut = start + currentWidth
		}

		line := strings.TrimSpace(string(runes[start:cut]))
		if line == "" {
			line = string(runes[start:cut])
		}
		lines = append(lines, line)

		start = cut
		for start < len(runes) && runes[start] == ' ' {
			start++
		}

		currentWidth = otherWidth
	}

	if len(lines) == 0 {
		lines = append(lines, "")
	}

	return lines
}

func wrapTextWithPrefix(prefix, text string, width int, style lipgloss.Style) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	contentWidth := width - lipgloss.Width(prefix) - common.LayoutWrapMargin
	if contentWidth < 1 {
		contentWidth = 1
	}

	lines := ui.WrapText(text, contentWidth)
	if len(lines) == 0 {
		lines = []string{""}
	}

	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = prefix + style.Render(line)
	}
	return result
}

func buildContinuationTitlePrefix(firstPrefix string) string {
	width := lipgloss.Width(firstPrefix)
	if width <= 0 {
		width = lipgloss.Width(ensurePrefixSpacing(""))
	}

	verticalWidth := lipgloss.Width(ui.VerticalLineSymbol)
	suffixWidth := lipgloss.Width("  ")
	spacesCount := width - verticalWidth - suffixWidth
	if spacesCount < 0 {
		spacesCount = 0
	}

	return performance.FastConcat(
		performance.RepeatEfficient(" ", spacesCount),
		ui.VerticalLineSymbol,
		performance.RepeatEfficient(" ", suffixWidth),
	)
}

// SetCompletedPrefix позволяет переопределить префикс завершённой задачи (используется очередью)
func (t *BaseTask) SetCompletedPrefix(prefix string) {
	t.completedPrefix = prefix
}

// CompletedPrefix возвращает текущий переопределённый префикс (если установлен)
func (t *BaseTask) CompletedPrefix() string {
	return t.completedPrefix
}

// SetInProgressPrefix позволяет очереди переопределять префикс активной задачи
func (t *BaseTask) SetInProgressPrefix(prefix string) {
	t.inProgressPrefix = prefix
}

// InProgressPrefix возвращает текущий префикс активной задачи (с учётом значения по умолчанию)
func (t *BaseTask) InProgressPrefix() string {
	if strings.TrimSpace(t.inProgressPrefix) != "" {
		return t.inProgressPrefix + " "
	}
	return ui.GetCurrentTaskPrefix()
}

// IsTextInputTask определяет, является ли задача текстовой задачей ввода
// (не задачей выбора SingleSelect/MultiSelect)
func IsTextInputTask(task Task) bool {
	// Проверяем по названию типа через рефлексию
	switch task.(type) {
	case *SingleSelectTask, *MultiSelectTask:
		return false
	default:
		// Все остальные задачи (InputTaskNew, YesNoTask, FuncTask) являются текстовыми
		return true
	}
}
