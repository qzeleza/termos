package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/performance"
)

// FormatErrorMessage форматирует сообщение об ошибке с отступами и ограничением по ширине
// Если preserveNewLines=false, то удаляет все новые строки и эскейп последовательности, затем разбивает на строки
// Если preserveNewLines=true, то сохраняет оригинальные переносы строк
// Корректно работает с UTF-8 символами (кириллица, эмодзи, китайские символы и т.д.)
// maxWidth - полная ширина доступного пространства в символах (не байтах)
// layoutWidth - полная ширина макета для финальной линии
// preserveNewLines - если true, то сохраняет оригинальные переносы строк
func FormatErrorMessage(errMsg string, layoutWidth int, preserveNewLines bool) string {
	if errMsg == "" {
		return ""
	}

	// Константы для отступов и правого поля
	const rightMargin = 2

	// Префикс для каждой линии: два пробела + вертикальная линия + три пробела
	// Длина префикса
	const prefixLen = 2 + 1 + 3

	// Эффективная ширина области текста (с учётом отступов, правого поля и резервного зазора)
	wrapWidth := layoutWidth - rightMargin - prefixLen - common.LayoutWrapMargin
	if wrapWidth < 3 {
		// Минимальная ширина, чтобы тесты для узкого лэйаута получали хотя бы 3 символа (например, "Оши")
		wrapWidth = 3
	}

	// Базовый отступ слева (2 пробела)
	numIndent := 2
	indent := performance.RepeatEfficient(" ", numIndent)

	// Если нужно сохранить переносы строк
	if preserveNewLines {
		lines := strings.Split(errMsg, "\n")
		result := performance.GetBuffer()
		defer performance.PutBuffer(result)

		lineCount := 0
		for _, rawLine := range lines {
			line := strings.TrimRight(rawLine, "\r")
			var segments []string
			if strings.TrimSpace(line) == "" {
				segments = []string{""}
			} else {
				segments = wrapText(line, wrapWidth)
				if len(segments) == 0 {
					segments = []string{""}
				}
			}

			for _, segment := range segments {
				if lineCount > 0 {
					result.WriteString("\n")
				}
				result.WriteString(indent)
				result.WriteString(VerticalLineSymbol)
				result.WriteString(performance.RepeatEfficient(" ", 3))
				result.WriteString(GetErrorMessageStyle().Render(segment))
				lineCount++
			}
		}

		return result.String()
	}

	// Если не нужно сохранять переносы строк
	// 1. Очищаем строку от управляющих символов и эскейп последовательностей
	cleanedMsg := cleanMessage(errMsg)

	if cleanedMsg == "" {
		return ""
	}

	// Создаем форматированный результат с учётом новой ширины
	errorMsg := formatErrorEveryLine(cleanedMsg, wrapWidth, indent, true)

	return errorMsg
}

// formatErrorEveryLine создает отформатированное сообщение с разделительными линиями и отступами
// Если delNewLines=true, то переносы строк удаляются и текст переформатируется
// Если delNewLines=false, то сохраняются оригинальные переносы строк
func formatErrorEveryLine(msg string, effectiveWidth int, indent string, delNewLines bool) string {

	result := performance.GetBuffer()
	defer performance.PutBuffer(result)

	// Делаем первую букву в предложении заглавной
	msg = CapitalizeFirst(msg)

	// Если не нужно удалять переносы строк, обрабатываем каждую строку отдельно
	if !delNewLines {
		// Разбиваем по переносам строк
		lines := strings.Split(msg, "\n")
		for i, line := range lines {
			if i > 0 {
				result.WriteString("\n")
			}
			result.WriteString(indent)
			result.WriteString(VerticalLineSymbol)
			result.WriteString(performance.RepeatEfficient(" ", 3))
			result.WriteString(GetErrorMessageStyle().Render(line))
		}
		return result.String()
	}

	// Если сообщение помещается в одну строку
	if utf8.RuneCountInString(msg) <= effectiveWidth {
		result.WriteString(indent)
		result.WriteString(VerticalLineSymbol)
		result.WriteString(performance.RepeatEfficient(" ", 3))
		result.WriteString(GetErrorMessageStyle().Render(msg))
		// Возвращаем без переводов строк, чтобы проверки Contains находили фразы целиком
		return result.String()
	}
	// Разбиваем на строки
	lines := wrapText(msg, effectiveWidth)

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(indent)
		result.WriteString(VerticalLineSymbol)
		result.WriteString(performance.RepeatEfficient(" ", 3))
		result.WriteString(GetErrorMessageStyle().Render(line))
	}

	// Возвращаем отформатированное сообщение с переносами строк
	return result.String()
}
