package ui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// TestConstants проверяет корректность объявленных констант
func TestConstants(t *testing.T) {
	assert.Equal(t, "─", HorizontalLineSymbol, "Символ горизонтальной линии должен быть корректным")
	assert.Equal(t, "│", VerticalLineSymbol, "Символ вертикальной линии должен быть корректным")
	assert.Equal(t, "└", CornerDownSymbol, "Угловой символ должен быть корректным")
	assert.Equal(t, ">", ArrowSymbol, "Символ стрелки должен быть корректным")
	assert.Equal(t, "├", BranchSymbol, "Символ ветки должен быть корректным")
	assert.Equal(t, "●", TaskCompletedSymbol, "Активный символ должен быть корректным")
	assert.Equal(t, 3, MessageIndentSpaces, "Отступ для сообщений должен быть 3")
	assert.Equal(t, 2, MainLeftIndent, "Основной отступ должен быть 2")
}

// TestStringIndents проверяет корректное формирование строковых отступов
func TestStringIndents(t *testing.T) {
	// Проверяем основные отступы новой системы
	assert.Equal(t, strings.Repeat(" ", MessageIndentSpaces), MessageIndent, "Отступ сообщений должен быть 3 пробела")

	// Проверяем функции префиксов
	expectedCurrent := strings.Repeat(" ", MainLeftIndent) + TaskInProgressSymbol + "  "
	assert.Equal(t, expectedCurrent, GetCurrentTaskPrefix(), "Префикс текущей задачи должен быть правильным")

	expectedCompletedSuccess := strings.Repeat(" ", MainLeftIndent) + TaskCompletedSymbol
	assert.Equal(t, expectedCompletedSuccess, GetCompletedTaskPrefix(true), "Префикс успешной задачи должен быть правильным")

	expectedCompletedError := strings.Repeat(" ", MainLeftIndent) + TaskInProgressSymbol
	assert.Equal(t, expectedCompletedError, GetCompletedTaskPrefix(false), "Префикс неуспешной задачи должен быть правильным")

	expectedInputSuccess := strings.Repeat(" ", MainLeftIndent) + TaskCompletedSymbol
	assert.Equal(t, expectedInputSuccess, GetCompletedInputTaskPrefix(true), "Префикс успешной текстовой задачи должен быть правильным")

	expectedInputError := strings.Repeat(" ", MainLeftIndent) + TaskInProgressSymbol
	assert.Equal(t, expectedInputError, GetCompletedInputTaskPrefix(false), "Префикс неуспешной текстовой задачи должен быть правильным")
}

// TestIcons проверяет наличие и валидность иконок
func TestIcons(t *testing.T) {
	icons := []string{
		IconDone, IconError, IconCancelled, IconQuestion,
		IconSelected, IconRadioOn, IconCursor, IconUndone, IconRadioOff,
	}

	for _, icon := range icons {
		assert.NotEmpty(t, icon, "Иконка не должна быть пустой")
		assert.True(t, utf8.ValidString(icon), "Иконка должна быть валидной UTF-8 строкой")
	}
}

// TestAlignText проверяет функцию выравнивания текста
func TestAlignText(t *testing.T) {
	tests := []struct {
		name       string
		left       string
		right      string
		totalWidth int
		expected   string
	}{
		{
			name:       "Базовое выравнивание",
			left:       "Левый",
			right:      "Правый",
			totalWidth: 20,
			expected:   "Левый" + strings.Repeat(" ", 7) + "Правый  ",
		},
		{
			name:       "Недостаточная ширина",
			left:       "Очень длинный левый текст",
			right:      "Правый",
			totalWidth: 10,
			expected:   "Очень длинный левый текст Правый",
		},
		{
			name:       "Точная ширина",
			left:       "Лев",
			right:      "Прав",
			totalWidth: 7,
			expected:   "ЛевПрав",
		},
		{
			name:       "Пустые строки",
			left:       "",
			right:      "",
			totalWidth: 10,
			expected:   "          ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AlignTextToRight(tt.left, tt.right, tt.totalWidth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDrawLine проверяет функцию создания горизонтальной линии
func TestDrawLine(t *testing.T) {
	tests := []struct {
		width    int
		expected string
	}{
		{0, "\n"},
		{1, "─\n"},
		{5, "─────\n"},
		{10, "──────────\n"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := DrawLine(tt.width)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCleanMessage проверяет очистку сообщений от управляющих символов
func TestCleanMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Обычный текст",
			input:    "Обычное сообщение",
			expected: "Обычное сообщение",
		},
		{
			name:     "С переносами строк",
			input:    "Строка\nс\rпереносами\t",
			expected: "Строка с переносами",
		},
		{
			name:     "С эскейп последовательностями",
			input:    "Текст\\nс\\tэскейпами\\r",
			expected: "Текстсэскейпами",
		},
		{
			name:     "Множественные пробелы",
			input:    "Текст    с     множественными     пробелами",
			expected: "Текст с множественными пробелами",
		},
		{
			name:     "Пустая строка",
			input:    "",
			expected: "",
		},
		{
			name:     "Только пробелы",
			input:    "   \t\n\r   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCapitalizeFirst проверяет функцию капитализации первой буквы
func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Обычный текст",
			input:    "обычный текст",
			expected: "Обычный текст",
		},
		{
			name:     "Уже капитализирован",
			input:    "Уже капитализирован",
			expected: "Уже капитализирован",
		},
		{
			name:     "Кириллица",
			input:    "кириллический текст",
			expected: "Кириллический текст",
		},
		{
			name:     "Пустая строка",
			input:    "",
			expected: "",
		},
		{
			name:     "Один символ",
			input:    "a",
			expected: "A",
		},
		{
			name:     "Unicode символы",
			input:    "αβγ",
			expected: "Αβγ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CapitalizeFirst(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestWrapText проверяет функцию переноса текста
func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		expected []string
	}{
		{
			name:     "Короткий текст",
			input:    "Короткий",
			maxWidth: 20,
			expected: []string{"Короткий"},
		},
		{
			name:     "Текст точно по ширине",
			input:    "Точная ширина",
			maxWidth: 13,
			expected: []string{"Точная ширина"},
		},
		{
			name:     "Длинный текст с пробелами",
			input:    "Это очень длинный текст который нужно разбить",
			maxWidth: 20,
			expected: []string{"Это очень длинный", "текст который нужно", "разбить"},
		},
		{
			name:     "Текст без пробелов",
			input:    "Оченьдлинноесловобезпробелов",
			maxWidth: 10,
			expected: []string{"Оченьдлинн", "оесловобез", "пробелов"},
		},
		{
			name:     "Пустая строка",
			input:    "",
			maxWidth: 10,
			expected: []string{},
		},
		{
			name:     "Кириллица с пробелами",
			input:    "Кириллический текст с переносами строк",
			maxWidth: 15,
			expected: []string{"Кириллический", "текст с", "переносами", "строк"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.input, tt.maxWidth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFindOptimalCutPointRunes проверяет функцию поиска оптимальной точки разреза
func TestFindOptimalCutPointRunes(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		start    int
		maxWidth int
		expected int
	}{
		{
			name:     "Разрез по пробелу",
			text:     "Текст с пробелами",
			start:    0,
			maxWidth: 10,
			expected: 7, // до пробела после "Текст с"
		},
		{
			name:     "Разрез без пробела",
			text:     "Текстбезпробелов",
			start:    0,
			maxWidth: 5,
			expected: 5, // по максимальной ширине
		},
		{
			name:     "Начало не с нуля",
			text:     "Начальный текст с пробелами",
			start:    10,
			maxWidth: 8,
			expected: 7, // до пробела после "текст с"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			textRunes := []rune(tt.text)
			result := findOptimalCutPointRunes(textRunes, tt.start, tt.maxWidth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatErrorMessage проверяет функцию форматирования сообщений об ошибках
func TestFormatErrorMessage(t *testing.T) {
	tests := []struct {
		name           string
		errMsg         string
		layoutWidth    int
		expectedTokens []string
	}{
		{
			name:           "Обычная ошибка",
			errMsg:         "Произошла ошибка",
			layoutWidth:    50,
			expectedTokens: []string{"Произошла", "ошибка"},
		},
		{
			name:        "Пустое сообщение",
			errMsg:      "",
			layoutWidth: 50,
		},
		{
			name:           "Очень узкая ширина",
			errMsg:         "Ошибка",
			layoutWidth:    5,
			expectedTokens: []string{"Оши"},
		},
		{
			name:           "Длинное сообщение",
			errMsg:         "Это очень длинное сообщение об ошибке которое должно быть разбито на несколько строк",
			layoutWidth:    30,
			expectedTokens: []string{"Это", "очень", "длинное", "сообщение"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatErrorMessage(tt.errMsg, tt.layoutWidth, false)

			if tt.errMsg == "" {
				assert.Empty(t, result, "Результат должен быть пустым для пустого сообщения")
				return
			}

			normalized := strings.ReplaceAll(strings.Join(strings.Fields(strings.ReplaceAll(result, "\n", "")), ""), " ", "")
			normalized = strings.ReplaceAll(normalized, VerticalLineSymbol, "")
			for _, token := range tt.expectedTokens {
				sanitized := strings.ReplaceAll(token, " ", "")
				assert.Contains(t, normalized, sanitized, "Результат должен содержать ожидаемый текст: %s", token)
			}
			assert.True(t, strings.Contains(result, MessageIndent), "Результат должен содержать отступ")
		})
	}
}

// TestBuildFormattedMessage проверяет функцию построения форматированного сообщения
func TestBuildFormattedMessage(t *testing.T) {
	tests := []struct {
		name           string
		msg            string
		layoutWidth    int
		expectedTokens []string
		expectIndent   bool
	}{
		{
			name:           "Короткое сообщение",
			msg:            "Короткое сообщение",
			layoutWidth:    52,
			expectedTokens: []string{"Короткое", "сообщение"},
			expectIndent:   true,
		},
		{
			name:           "Длинное сообщение",
			msg:            "это очень длинное сообщение которое должно быть разбито на несколько строк для лучшей читаемости",
			layoutWidth:    22,
			expectedTokens: []string{"Это", "очень", "длинное", "сообщен", "ие"},
			expectIndent:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatErrorMessage(tt.msg, tt.layoutWidth, false)

			if tt.expectIndent {
				assert.Contains(t, result, MessageIndent, "Результат должен содержать отступ")
			}

			normalized := strings.ReplaceAll(strings.Join(strings.Fields(strings.ReplaceAll(result, "\n", "")), ""), " ", "")
			normalized = strings.ReplaceAll(normalized, VerticalLineSymbol, "")
			for _, token := range tt.expectedTokens {
				sanitized := strings.ReplaceAll(token, " ", "")
				assert.Contains(t, normalized, sanitized, "Результат должен содержать: %s", token)
			}
		})
	}
}

// TestUTF8Handling проверяет корректную работу с UTF-8 символами
func TestUTF8Handling(t *testing.T) {
	// Тест с эмодзи
	emoji := "🚀 Запуск приложения"
	result := wrapText(emoji, 10)
	assert.NotEmpty(t, result, "Должен корректно обрабатывать эмодзи")

	// Тест с китайскими символами
	chinese := "这是中文测试"
	result = wrapText(chinese, 5)
	assert.NotEmpty(t, result, "Должен корректно обрабатывать китайские символы")

	// Тест капитализации кириллицы
	cyrillic := "русский текст"
	capitalized := CapitalizeFirst(cyrillic)
	assert.Equal(t, "Русский текст", capitalized, "Должен корректно капитализировать кириллицу")
}

// BenchmarkWrapText проверяет производительность функции переноса текста
func BenchmarkWrapText(b *testing.B) {
	longText := strings.Repeat("Это тестовый текст для проверки производительности. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapText(longText, 50)
	}
}

// BenchmarkFormatErrorMessage проверяет производительность форматирования ошибок
func BenchmarkFormatErrorMessage(b *testing.B) {
	errorMsg := "Это длинное сообщение об ошибке которое используется для тестирования производительности форматирования."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatErrorMessage(errorMsg, 80, false)
	}
}
