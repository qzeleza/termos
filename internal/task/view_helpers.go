package task

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/performance"
	"github.com/qzeleza/ziva/internal/ui"
)

// renderSelectionSeparator формирует разделитель между заголовком и списком пунктов
// @param width Ширина макета для отображения
// @param enabled Флаг, указывающий, нужно ли отображать разделительную линию
// @param inProgressPrefix Префикс активной задачи
// @return Строка с отформатированным разделителем
func renderSelectionSeparator(width int, enabled bool, inProgressPrefix string) string {
	if !enabled {
		return ""
	}
	// Если префикс активной задачи не указан, используем текущий префикс
	if strings.TrimSpace(inProgressPrefix) == "" {
		inProgressPrefix = ui.GetCurrentTaskPrefix()
	}

	// Формируем базовый префикс
	basePrefix := performance.FastConcat(
		performance.RepeatEfficient(" ", ui.MainLeftIndent),
		ui.VerticalLineSymbol,
		"  ",
	)

	// Вычисляем ширину префикса активной задачи и базового префикса
	targetWidth := lipgloss.Width(inProgressPrefix)
	baseWidth := lipgloss.Width(basePrefix)
	if targetWidth < baseWidth {
		targetWidth = baseWidth
	}

	// Вычисляем количество дополнительных пробелов
	extraSpaces := targetWidth - baseWidth
	if extraSpaces < 0 {
		extraSpaces = 0
	}

	// Формируем префикс
	prefix := performance.FastConcat(
		basePrefix,
		performance.RepeatEfficient(" ", extraSpaces),
	)

	// Вычисляем доступную ширину для горизонтальной линии
	available := width - lipgloss.Width(prefix)
	if available > 0 {
		available--
	}
	if available < 0 {
		available = 0
	}

	// Формируем горизонтальную линию с бледным серым оттенком
	horizontal := ui.VerySubtleStyle.Render(performance.RepeatEfficient(ui.HorizontalLineSymbol, available))

	// Формируем разделитель
	return performance.FastConcat(
		prefix,
		horizontal,
		"\n",
	)
}

// formatNavigationHelpText подготавливает строку подсказки по навигации с учётом ширины макета.
// @param helpText Строка подсказки по навигации
// @param width Ширина макета для отображения
// @return Строка с отформатированной подсказкой
func formatNavigationHelpText(helpText string, width int) string {
	if strings.TrimSpace(helpText) == "" {
		return helpText
	}

	layoutWidth := common.CalculateLayoutWidth(width)
	available := layoutWidth - 4
	if available <= 0 {
		available = layoutWidth
	}
	if available <= 0 {
		return helpText
	}

	formatted := helpText
	if utf8.RuneCountInString(formatted) > available {
		if idx := strings.LastIndex(formatted, ", "); idx != -1 {
			first := strings.TrimRight(formatted[:idx+1], " ")
			var second string
			if idx+2 < len(formatted) {
				second = formatted[idx+2:]
			}
			second = strings.TrimSpace(second)
			if second != "" {
				formatted = performance.FastConcat(first, "\n", second)
			} else {
				formatted = truncateRunes(first, available)
			}
		} else {
			formatted = truncateRunes(formatted, available)
		}
	}

	lines := strings.Split(formatted, "\n")
	for i, line := range lines {
		if utf8.RuneCountInString(line) > available {
			lines[i] = truncateRunes(line, available)
		}
	}
	return strings.Join(lines, "\n")
}

// indentLines добавляет отступ перед каждой строкой текста.
func indentLines(text, indent string) string {
	if text == "" {
		return ""
	}
	if indent == "" {
		return text
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = performance.FastConcat(indent, line)
	}
	return strings.Join(lines, "\n")
}

// truncateRunes обрезает строку по количеству рун.
func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
