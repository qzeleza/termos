package ui

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"reflect"

	"github.com/charmbracelet/lipgloss"
	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/performance"
)

var (
	// NumberingEnabled определяет, включена ли нумерация задач
	NumberingEnabled bool
)

// Яркая цветовая палитра для UI
var (
	ColorBrightGreen    = lipgloss.Color("#00ff00") // Ярко-зелёный
	ColorBrightRed      = lipgloss.Color("#FF2104") // Ярко-красный
	ColorDarkRed        = lipgloss.Color("#B10C01") // Темно-красный
	ColorBrightYellow   = lipgloss.Color("#ffff00") // Ярко-жёлтый
	ColorDarkYellow     = lipgloss.Color("#D2BE88") // Темно-жёлтый
	ColorBrightOrange   = lipgloss.Color("#FFA500") // Ярко-оранжевый
	ColorDarkOrange     = lipgloss.Color("#D2691E") // Темно-оранжевый
	ColorBrightBlue     = lipgloss.Color("#0000ff") // Ярко-синий
	ColorDarkBlue       = lipgloss.Color("#01189B") // Темно-синий
	ColorBrightCyan     = lipgloss.Color("#00ffff") // Ярко-голубой
	ColorDarkCyan       = lipgloss.Color("#006875") // Темно-голубой
	ColorBrightMagenta  = lipgloss.Color("#ff00ff") // Ярко-фиолетовый
	ColorBrightWhite    = lipgloss.Color("#fff")    // Ярко-белый
	ColorBrightGray     = lipgloss.Color("#777")    // Светло-серый
	ColorDarkGray       = lipgloss.Color("#333")    // Темно-серый
	ColorLightBlue      = lipgloss.Color("#5DA9E9") // Светло-голубой
	ColorBlack          = lipgloss.Color("#000")    // Черный
	ColorDarkGreen      = lipgloss.Color("#008000") // Темно-зелёный
	ColorVeryDarkGray   = lipgloss.Color("#444")    // Очень темно-серый (для едва заметных элементов)
	ColorVeryDarkYellow = lipgloss.Color("#666633") // Очень темно-желтый (для приглушенных ошибок)
	ColorMutedGreen     = lipgloss.Color("#4a7c59") // Приглушенно-зеленый (для сводки при успехе)
)

// Яркие иконки
var (
	IconDone      = lipgloss.NewStyle().SetString("✔").Foreground(ColorBrightGreen).String()          // Галочка (ярко-зелёная)
	IconError     = lipgloss.NewStyle().SetString("✕").Foreground(ColorBrightRed).String()            // Крестик (ярко-красный)
	IconCancelled = lipgloss.NewStyle().SetString("✕").Foreground(ColorBrightYellow).String()         // Отмена (ярко-жёлтый)
	IconQuestion  = lipgloss.NewStyle().SetString("?").Foreground(ColorBrightGreen).String()          // Вопрос (ярко-жёлтый)
	IconSelected  = lipgloss.NewStyle().SetString("■").Foreground(ColorBrightGreen).String()          // Галочка выбора (ярко-синий)
	IconRadioOn   = lipgloss.NewStyle().SetString("●").Foreground(ColorLightBlue).String()            // Радио включено (ярко-синий)
	IconCursor    = lipgloss.NewStyle().SetString("➞").Foreground(ColorLightBlue).String()            // Курсор (ярко-синий)
	IconUndone    = lipgloss.NewStyle().SetString("◷").Foreground(ColorLightBlue).Bold(true).String() // Неактивный элемент (ярко-белый)
	IconRadioOff  = "○"
)

// Стили для текста
// Стили для текста с яркой палитрой
var (
	TitleStyle = lipgloss.NewStyle().Bold(true)

	// Динамические стили ошибок (могут быть изменены через SetErrorColor)
	ErrorMessageStyle    = lipgloss.NewStyle().Foreground(ColorDarkYellow)              // Ошибка: ярко-красный (без курсива)
	ErrorStatusStyle     = lipgloss.NewStyle().Foreground(ColorBrightYellow).Bold(true) // Вывод сообщений об ошибках: ярко-красный (курсив)
	CancelStyle          = lipgloss.NewStyle().Foreground(ColorBrightYellow)            // Вывод статуса ошибки: ярко-жёлтый
	SubtleStyle          = lipgloss.NewStyle().Foreground(ColorBrightGray)              // Подписи: светло-серый
	DisabledStyle        = lipgloss.NewStyle().Foreground(ColorBrightGray)              // Неактивные элементы: светло-серый
	HelpTextStyle        = lipgloss.NewStyle().Foreground(ColorLightBlue)               // Подсказки для элементов
	SelectionStyle       = lipgloss.NewStyle().Foreground(ColorBrightGreen)             // Выделение (Да): ярко-зелёный
	SelectionNoStyle     = lipgloss.NewStyle().Foreground(ColorBrightRed).Bold(true)    // Выделение (Нет): ярко-красный
	ActiveStyle          = lipgloss.NewStyle().Foreground(ColorLightBlue).Bold(true)    // Активный элемент: ярко-синий
	MenuExitItemStyle    = menuItemPlainStyle()                                         // По умолчанию кнопки выхода белые
	MenuBackItemStyle    = menuItemPlainStyle()                                         // По умолчанию кнопки возврата белые
	InputStyle           = lipgloss.NewStyle().Foreground(ColorLightBlue).Bold(true)    // Стиль для активного ввода
	SpinnerStyle         = lipgloss.NewStyle().Foreground(ColorLightBlue).Bold(true)    // Стиль для спиннера
	ActiveTitleStyle     = lipgloss.NewStyle().Foreground(ColorBrightGreen).Bold(true)  // Активный заголовок ввода
	ActiveTaskStyle      = lipgloss.NewStyle().Foreground(ColorBrightGreen)             // Стиль для активной задачи
	SuccessLabelStyle    = lipgloss.NewStyle().Foreground(ColorBrightGreen).Bold(true)  // Успешное завершение
	VerySubtleStyle      = lipgloss.NewStyle().Foreground(ColorVeryDarkGray)            // Едва заметные элементы
	VerySubtleErrorStyle = lipgloss.NewStyle().Foreground(ColorVeryDarkYellow)          // Едва заметные элементы ошибок

	FinishedLabelStyle     = lipgloss.NewStyle().Foreground(ColorBrightWhite).Bold(true) // Завершение
	SummaryLabelStyle      = lipgloss.NewStyle().Foreground(ColorBrightWhite).Bold(true) // Стиль для сводки
	SummarySuccessStyle    = lipgloss.NewStyle().Foreground(ColorMutedGreen).Bold(true)  // Приглушенно-зеленый стиль для успешной сводки
	TaskStatusSuccessStyle = lipgloss.NewStyle().Foreground(ColorBrightGreen).Bold(true) // Стиль для статуса успешных задач (соответствует SelectionStyle)
)

// Константы отступов (в пробелах)
const (
	// Основная система отступов с вертикальными линиями
	MainLeftIndent      = 2 // Основной отступ от левого края для всех задач
	MessageIndentSpaces = 3 // Отступ для вывода ошибок и сообщений при завершении задачи
)

// Строковые отступы
var (
	MessageIndent = strings.Repeat(" ", MessageIndentSpaces) // Отступ для сообщений об ошибках
)

// menuActionDefaultStyle возвращает стиль подсветки для специальных пунктов меню (яркий циан, жирный).
// По умолчанию подсветка пунктов меню имеет стиль - яркий циан, жирный.
func menuActionDefaultStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ColorBrightCyan).
		Bold(true)
}

func menuItemPlainStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorBrightWhite)
}

// MenuActionDefaultStyle возвращает стиль подсветки для специальных пунктов меню.
func MenuActionDefaultStyle() lipgloss.Style {
	return menuActionDefaultStyle()
}

// MenuItemPlainStyle возвращает стиль, совпадающий с обычными пунктами меню (белый текст).
func MenuItemPlainStyle() lipgloss.Style {
	return menuItemPlainStyle()
}

func isStyleEmpty(style lipgloss.Style) bool {
	return reflect.DeepEqual(style, lipgloss.Style{})
}

// Символы для новой системы отображения
const (
	HorizontalLineSymbol = "─" // Символ горизонтальной линии
	VerticalLineSymbol   = "│" // Символ вертикальной линии
	CornerDownSymbol     = "└" // Угловой символ вправо (устарело)
	CornerUpSymbol       = "┌" // Угловой символ вправо (устарело)
	ArrowSymbol          = ">" // Стрелка вправо (устарело)
	BranchSymbol         = "├" // Символ ветки для активных задач
	TaskCompletedSymbol  = "●" // Символ включенного радио
	TaskInProgressSymbol = "○" // Символ выключенного радио
	DownLineSymbol       = "┬" // Символ специальной линии
	UpLineSymbol         = "┴" // Символ специальной линии
	FinishSymbol         = "■" // Префикс для задач
	UpArrowSymbol        = "▲" // Символ стрелки вверх
	DownArrowSymbol      = "▼" // Символ стрелки вниз
)

// Функции для создания префиксов с новой системой отображения
// GetTaskBelowPrefix возвращает префикс для задачи ниже текущей выполняющейся задачи
// Формат: "  │" (отступ + ветка + линия)
func GetTaskBelowPrefix() string {
	return performance.FastConcat(
		performance.RepeatEfficient(" ", MainLeftIndent),
		VerticalLineSymbol,
	)
}

// GetCurrentTaskPrefix возвращает префикс для текущей выполняющейся задачи
// Формат: "   ○  " (отступ + символ и два пробела)
func GetCurrentTaskPrefix() string {
	return performance.FastConcat(
		performance.RepeatEfficient(" ", MainLeftIndent),
		TaskInProgressSymbol,
		"  ",
	)
}

// GetCurrentSelectTaskPrefix возвращает префикс для текущей выполняющейся задачи
// Формат: "└─>  " (отступ + угловой символ + линия + стрелка + два пробела)
func GetCurrentActiveTaskPrefix() string {
	return performance.FastConcat(
		performance.RepeatEfficient(" ", MainLeftIndent),
		CornerDownSymbol,
		HorizontalLineSymbol,
		ActiveStyle.Render(ArrowSymbol),
		" ",
	)
}

// GetCompletedTaskPrefix возвращает префикс для завершенной задачи
// Формат: "   ●" или "   ○"
func GetCompletedTaskPrefix(success bool) string {
	var icon string
	if success {
		icon = TaskCompletedSymbol
	} else {
		icon = TaskInProgressSymbol
	}

	return performance.FastConcat(
		performance.RepeatEfficient(" ", MainLeftIndent),
		icon,
	)
}

// GetCommentPrefix возвращает префикс для комментария
// Формат:
//
//	│   Комментарий
//	│
func GetCommentPrefix(value string) string {

	return performance.FastConcat(
		performance.RepeatEfficient(" ", MainLeftIndent),
		VerticalLineSymbol,
		GetResultIndentWhenNumberingEnabled(),
		SubtleStyle.Render(value),
		"\n",
		performance.RepeatEfficient(" ", MainLeftIndent),
		VerticalLineSymbol,
	)
}

// GetCompletedInputTaskPrefix возвращает префикс для завершенной задачи с текстовым вводом
// success = true: "  │ ●", success = false: "  │ ○"
func GetCompletedInputTaskPrefix(success bool) string {
	var icon string
	if success {
		// icon = IconDone
		icon = TaskCompletedSymbol
	} else {
		// icon = IconError
		icon = TaskInProgressSymbol
	}

	return performance.FastConcat(
		performance.RepeatEfficient(" ", MainLeftIndent),
		// TaskCompletedSymbol,
		// " ",
		icon,
	)
}

// GetSelectItemPrefix возвращает префикс для элементов в задачах выбора
// itemType: "active" - активный элемент, "above" - элемент выше активного, "below" - элемент ниже активного
func GetSelectItemPrefix(itemType string) string {
	switch itemType {

	case "above":
		// Элемент выше активного: "  |   "
		return performance.FastConcat(
			performance.RepeatEfficient(" ", MainLeftIndent),
			VerticalLineSymbol,
			performance.RepeatEfficient(" ", MainLeftIndent+1),
		)

	case "active":
		// Активный элемент: "  └─> "
		return GetCurrentActiveTaskPrefix()

	case "below":
		// Элемент ниже активного: "       " (отступ + 5 пробелов)
		return performance.RepeatEfficient(" ", MainLeftIndent+4)
	default:
		return performance.RepeatEfficient(" ", MainLeftIndent)
	}
}

// GetPendingTasksPlaceholder возвращает заглушку для отображения вместо невыполненных задач
// Показывает пустую строку и следом горизонтальную линию
func GetPendingTasksPlaceholder() string {
	return ""
	// return "\n" + performance.FastConcat(
	// 	performance.RepeatEfficient(" ", MainLeftIndent),
	// 	performance.RepeatEfficient(HorizontalLineSymbol, 30), // Горизонтальная линия
	// )
}

// AlignTextToRight выравнивает текст по правому краю по заданной ширине
func AlignTextToRight(left string, right string, totalWidth int) string {
	const rightMargin = 2

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)

	available := totalWidth - leftWidth - rightWidth
	if available < 0 {
		return performance.FastConcat(left, " ", right)
	}
	if available == 0 {
		return performance.FastConcat(left, right)
	}

	if available <= rightMargin {
		return performance.FastConcat(left, performance.RepeatEfficient(" ", available), right)
	}

	padding := performance.RepeatEfficient(" ", available-rightMargin)
	return performance.FastConcat(left, padding, right, "  ")
}

// CapitalizeFirst делает первую букву заглавной
func CapitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}

	return string(unicode.ToUpper(r)) + s[size:]
}

// WrapText разбивает текст на строки заданной максимальной длины.
// Экспортируется для повторного использования в других пакетах.
func WrapText(text string, maxWidth int) []string {
	return wrapText(text, maxWidth)
}

// wrapText разбивает текст на строки заданной максимальной длины
// Старается разбивать по словам, избегая разрыва слов посередине
// Корректно работает с UTF-8 символами
func wrapText(text string, maxWidth int) []string {
	if text == "" {
		return []string{}
	}

	textRunes := []rune(text)
	textLen := len(textRunes)

	if textLen <= maxWidth {
		return []string{text}
	}

	var lines []string
	start := 0

	for start < textLen {
		end := start + maxWidth
		if end >= textLen {
			// Последняя строка
			lines = append(lines, string(textRunes[start:]))
			break
		}

		// Ищем оптимальное место для разрыва
		cutPoint := findOptimalCutPointRunes(textRunes, start, maxWidth)

		// Добавляем строку
		lines = append(lines, string(textRunes[start:start+cutPoint]))

		// Обновляем позицию, пропуская ведущие пробелы
		start += cutPoint
		for start < textLen && textRunes[start] == ' ' {
			start++
		}
	}

	return lines
}

// findOptimalCutPointRunes находит оптимальное место для разрыва строки при работе с рунами
// Приоритет: последний пробел в пределах maxWidth, иначе - maxWidth
func findOptimalCutPointRunes(textRunes []rune, start, maxWidth int) int {
	textLen := len(textRunes)
	maxEnd := start + maxWidth

	if maxEnd >= textLen {
		return textLen - start
	}

	// Ищем последний пробел в пределах максимальной ширины
	for i := maxEnd - 1; i > start; i-- {
		if textRunes[i] == ' ' {
			return i - start
		}
	}

	// Если пробела нет, разрываем по максимальной ширине
	return maxWidth
}

// GetResultIndent возвращает отступ для результата, в зависимости от включенной нумерации задач
func GetResultIndentWhenNumberingEnabled() string {
	textIndent := performance.RepeatEfficient(" ", 3) // отступ по умолчанию
	if NumberingEnabled {
		textIndent = performance.RepeatEfficient(" ", 4) // больший отступ при включенной нумерации
	}
	return textIndent
}

// DrawSummaryLine рисует дополнительные строки с отступом
func DrawSummaryLine(text string, width int) string {
	prefix := performance.FastConcat(
		performance.RepeatEfficient(" ", MainLeftIndent),
		VerticalLineSymbol,
		GetResultIndentWhenNumberingEnabled(),
	)

	contentWidth := width - lipgloss.Width(prefix) - common.LayoutWrapMargin
	if contentWidth < 1 {
		contentWidth = 1
	}

	lines := WrapText(text, contentWidth)
	if len(lines) == 0 {
		lines = []string{""}
	}

	var builder strings.Builder
	for i, line := range lines {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(prefix)
		builder.WriteString(SubtleStyle.Render(line))
	}
	builder.WriteString("\n")
	return builder.String()
}

// DrawLine создает горизонтальную линию заданной ширины
// типа ───
func DrawLine(width int) string {
	return performance.FastConcat(performance.RepeatEfficient(HorizontalLineSymbol, width), "\n")
}

// DrawSpecialLine создает горизонтальную линию заданной ширины c угловой линией внизу
// типа ──┬─
func DrawSpecialHeaderLine(width int) string {
	return performance.FastConcat(
		performance.RepeatEfficient(" ", 2),
		// HorizontalLineSymbol,
		// HorizontalLineSymbol,
		CornerUpSymbol,
		performance.RepeatEfficient(HorizontalLineSymbol, width-3), "\n")
}

// SetErrorColor устанавливает цвет для стилей ошибок
// Изменяет цвета для ErrorMessageStyle и ErrorStatusStyle
func SetErrorColor(errorsColor lipgloss.TerminalColor, statusColor lipgloss.TerminalColor) {
	ErrorMessageStyle = ErrorMessageStyle.Foreground(errorsColor)
	ErrorStatusStyle = ErrorStatusStyle.Foreground(statusColor)
}

// SetMenuExitItemStyle обновляет стиль для пунктов меню выхода.
// Без аргументов подсвечивает пункты и возвращает яркий циан + жирный.
// Пустой стиль в аргументе сбрасывает оформление к белому тексту.
func SetMenuExitItemStyle(styles ...lipgloss.Style) {
	if len(styles) == 0 {
		MenuExitItemStyle = menuActionDefaultStyle()
		return
	}

	style := styles[0]
	if isStyleEmpty(style) {
		MenuExitItemStyle = menuItemPlainStyle()
		return
	}
	MenuExitItemStyle = style
}

// ResetMenuExitItemStyle сбрасывает стиль пункта выхода к белому тексту по умолчанию.
func ResetMenuExitItemStyle() {
	MenuExitItemStyle = menuItemPlainStyle()
}

// SetMenuBackItemStyle обновляет стиль для пунктов меню возврата.
// Без аргументов подсвечивает пункты и возвращает яркий циан + жирный.
// Пустой стиль в аргументе сбрасывает оформление к белому тексту.
func SetMenuBackItemStyle(styles ...lipgloss.Style) {
	if len(styles) == 0 {
		MenuBackItemStyle = menuActionDefaultStyle()
		return
	}

	style := styles[0]
	if isStyleEmpty(style) {
		MenuBackItemStyle = menuItemPlainStyle()
		return
	}
	MenuBackItemStyle = style
}

// ResetMenuBackItemStyle сбрасывает стиль пункта "Назад" к белому тексту по умолчанию.
func ResetMenuBackItemStyle() {
	MenuBackItemStyle = menuItemPlainStyle()
}

// ResetErrorColors сбрасывает цвета ошибок к значениям по умолчанию
func ResetErrorColors() {
	ErrorMessageStyle = lipgloss.NewStyle().Foreground(ColorDarkYellow)
	ErrorStatusStyle = lipgloss.NewStyle().Foreground(ColorBrightYellow).Bold(true)
}

// GetErrorMessageStyle возвращает текущий стиль сообщений об ошибках
func GetErrorMessageStyle() lipgloss.Style {
	return ErrorMessageStyle
}

// GetErrorStatusStyle возвращает текущий стиль статуса ошибки
func GetErrorStatusStyle() lipgloss.Style {
	return ErrorStatusStyle
}

// cleanMessage удаляет все новые строки, управляющие символы и эскейп последовательности
func cleanMessage(msg string) string {
	// Удаляем распространенные эскейп последовательности
	escapeSequences := regexp.MustCompile(`\\[nrtbfav\\'"?]`)
	msg = escapeSequences.ReplaceAllString(msg, "")

	// Удаляем ANSI эскейп последовательности (цветовые коды и т.д.)
	ansiSequences := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	msg = ansiSequences.ReplaceAllString(msg, "")

	// Удаляем все управляющие символы (включая \n, \r, \t и др.)
	controlChars := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	msg = controlChars.ReplaceAllString(msg, " ")

	// Нормализуем пробелы (убираем множественные пробелы и обрезаем)
	msg = performance.CleanWhitespaceEfficient(msg)

	return msg
}
