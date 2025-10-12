package query

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/ui"
	"github.com/stretchr/testify/assert"
)

// MockTask для тестирования
type MockTask struct {
	title                 string
	done                  bool
	hasError              bool
	err                   error
	stopOnErr             bool
	preserveErrorNewLines bool
}

func (m *MockTask) Title() string            { return m.title }
func (m *MockTask) IsDone() bool             { return m.done }
func (m *MockTask) HasError() bool           { return m.hasError }
func (m *MockTask) Error() error             { return m.err }
func (m *MockTask) StopOnError() bool        { return m.stopOnErr }
func (m *MockTask) SetStopOnError(stop bool) { m.stopOnErr = stop }
func (m *MockTask) WithNewLinesInErrors(preserve bool) common.Task {
	m.preserveErrorNewLines = preserve
	return m
}
func (m *MockTask) Run() tea.Cmd {
	return func() tea.Msg { return nil }
}
func (m *MockTask) Update(msg tea.Msg) (common.Task, tea.Cmd) { return m, nil }
func (m *MockTask) View(width int) string                     { return m.title }
func (m *MockTask) FinalView(width int) string                { return m.title }

func NewMockTask(title string) *MockTask {
	return &MockTask{
		title:     title,
		done:      false,
		hasError:  false,
		stopOnErr: true,
	}
}

func (m *MockTask) CompleteSuccessfully() *MockTask {
	m.done = true
	m.hasError = false
	return m
}

func (m *MockTask) CompleteWithError(err error) *MockTask {
	m.done = true
	m.hasError = true
	m.err = err
	return m
}

// TestModelTaskStats проверяет правильность подсчета статистики задач
func TestModelTaskStats(t *testing.T) {
	// Создаем модель с несколькими задачами
	model := New("Тест статистики")

	// Создаем тестовые задачи
	task1 := NewMockTask("Задача 1").CompleteSuccessfully()
	task2 := NewMockTask("Задача 2").CompleteWithError(errors.New("ошибка"))
	task3 := NewMockTask("Задача 3").CompleteSuccessfully()

	tasks := []common.Task{task1, task2, task3}
	model.AddTasks(tasks)

	// Симулируем завершение всех задач
	model.current = len(tasks)
	model.updateTaskStats()

	// Проверяем статистику
	assert.Equal(t, 2, model.successCount, "Должно быть 2 успешные задачи")
	assert.Equal(t, 1, model.errorCount, "Должна быть 1 задача с ошибкой")
}

// TestFormatSummaryWithStats проверяет форматирование сводки со статистикой
func TestFormatSummaryWithStats(t *testing.T) {
	model := New("Тест форматирования")
	// Используем стандартную сводку модели

	// Случай 1: Все задачи успешны
	task1 := NewMockTask("Задача 1").CompleteSuccessfully()
	task2 := NewMockTask("Задача 2").CompleteSuccessfully()

	model.AddTasks([]common.Task{task1, task2})
	model.current = 2
	model.updateTaskStats()

	leftSummary, rightStatus := model.formatSummaryWithStats()

	assert.Contains(t, leftSummary, defaults.SummaryCompleted, "Левая часть должна содержать сводку")
	assert.Contains(t, leftSummary, "(2/2)", "Левая часть должна содержать статистику")
	assert.Equal(t, defaults.StatusSuccess, rightStatus, "Правая часть должна показывать УСПЕШНО")

	// Случай 2: Есть ошибки
	task3 := NewMockTask("Задача 3").CompleteWithError(errors.New("ошибка"))
	model.AddTasks([]common.Task{task3})
	model.current = 3
	model.updateTaskStats()

	leftSummary2, rightStatus2 := model.formatSummaryWithStats()

	assert.Contains(t, leftSummary2, defaults.SummaryCompleted, "Левая часть должна содержать сводку")
	assert.Contains(t, leftSummary2, "(2/3)", "Должно показывать 2 успешных из 3 всего")
	assert.Equal(t, defaults.StatusProblem, rightStatus2, "Правая часть должна показывать С ОШИБКАМИ")
}

func TestRenderSummaryFooterWrapsLongSummary(t *testing.T) {
	model := New("Тест длинной сводки")
	tasks := []common.Task{
		NewMockTask("Задача 1").CompleteSuccessfully(),
		NewMockTask("Задача 2").CompleteSuccessfully(),
		NewMockTask("Задача 3").CompleteSuccessfully(),
		NewMockTask("Задача 4").CompleteWithError(errors.New("ошибка")),
	}
	model.AddTasks(tasks)
	model.current = len(tasks)
	model.updateTaskStats()
	model.summary = "Сводка с\nпереносами   и    лишними    пробелами"

	leftSummary, rightStatus := model.formatSummaryWithStats()
	assert.NotContains(t, leftSummary, "\n", "Сводка не должна содержать исходные переносы строк")
	assert.Contains(t, leftSummary, "(3/4)", "Сводка должна содержать статистику 3/4")

	layoutWidth := common.CalculateLayoutWidth(40)
	summaryStyle := ui.GetErrorStatusStyle()
	rightStyle := ui.GetErrorStatusStyle()
	footer := model.renderSummaryFooter(leftSummary, rightStatus, layoutWidth, summaryStyle, rightStyle)

	lines := strings.Split(footer, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		assert.LessOrEqual(t, lipgloss.Width(line), layoutWidth, "Строки сводки не должны превышать ширину макета")
	}
}

// TestModelViewWithStats проверяет отображение View с новой статистикой
func TestModelViewWithStats(t *testing.T) {
	model := New("Тест отображения статистики")
	// Включаем сводку для этого теста
	model.WithSummary(true)

	// Создаем задачи с разными результатами
	task1 := NewMockTask("Успешная задача").CompleteSuccessfully()
	task2 := NewMockTask("Задача с ошибкой").CompleteWithError(errors.New("тестовая ошибка"))

	model.AddTasks([]common.Task{task1, task2})

	// Симулируем завершение всех задач
	model.current = 2
	model.updateTaskStats()

	// Получаем представление
	view := model.View()

	// Проверяем, что View содержит статистику
	assert.Contains(t, view, "(1/2)", "View должен содержать статистику (1/2)")
	assert.Contains(t, view, defaults.StatusProblem, "View должен показывать статус С ОШИБКАМИ")
	assert.Contains(t, view, defaults.SummaryCompleted, "View должен содержать оригинальную сводку")
}

// TestModelStatsProgression проверяет обновление статистики по мере выполнения задач
func TestModelStatsProgression(t *testing.T) {
	model := New("Тест прогрессии")

	task1 := NewMockTask("Задача 1")
	task2 := NewMockTask("Задача 2")
	task3 := NewMockTask("Задача 3")

	model.AddTasks([]common.Task{task1, task2, task3})

	// Начальное состояние
	model.current = 0
	model.updateTaskStats()
	assert.Equal(t, 0, model.successCount, "В начале успешных задач должно быть 0")
	assert.Equal(t, 0, model.errorCount, "В начале задач с ошибками должно быть 0")

	// Завершаем первую задачу успешно
	task1.CompleteSuccessfully()
	model.current = 1
	model.updateTaskStats()
	assert.Equal(t, 1, model.successCount, "После первой задачи должна быть 1 успешная")
	assert.Equal(t, 0, model.errorCount, "После первой задачи ошибок быть не должно")

	// Завершаем вторую задачу с ошибкой
	task2.CompleteWithError(errors.New("ошибка"))
	model.current = 2
	model.updateTaskStats()
	assert.Equal(t, 1, model.successCount, "После второй задачи должна остаться 1 успешная")
	assert.Equal(t, 1, model.errorCount, "После второй задачи должна быть 1 ошибка")

	// Завершаем третью задачу успешно
	task3.CompleteSuccessfully()
	model.current = 3
	model.updateTaskStats()
	assert.Equal(t, 2, model.successCount, "После третьей задачи должно быть 2 успешных")
	assert.Equal(t, 1, model.errorCount, "После третьей задачи должна остаться 1 ошибка")
}

// TestWithSummary проверяет функцию WithSummary
func TestWithSummary(t *testing.T) {
	model := New("Тест флага сводки")

	// По умолчанию сводка должна быть включена
	assert.True(t, model.showSummary, "По умолчанию флаг showSummary должен быть true")

	// Включаем сводку
	model.WithSummary(true)
	assert.True(t, model.showSummary, "После WithSummary(true) флаг должен быть true")

	// Выключаем сводку
	model.WithSummary(false)
	assert.False(t, model.showSummary, "После WithSummary(false) флаг должен быть false")
}

// TestViewWithSummaryFlag проверяет отображение View с учетом флага showSummary
func TestViewWithSummaryFlag(t *testing.T) {
	// Тест с отключенной сводкой
	model := New("Тест без сводки")
	model.WithSummary(false) // Отключаем сводку
	task1 := NewMockTask("Задача 1").CompleteSuccessfully()
	model.AddTasks([]common.Task{task1})
	model.current = 1 // Все задачи завершены

	view := model.View()

	// Проверяем, что View НЕ содержит сводку
	assert.NotContains(t, view, defaults.StatusSuccess, "При showSummary=false View не должен содержать статус")
	assert.NotContains(t, view, "(1/1)", "При showSummary=false View не должен содержать статистику")

	// Тест с включенной сводкой (по умолчанию)
	model2 := New("Тест со сводкой")
	task2 := NewMockTask("Задача 1").CompleteSuccessfully()
	model2.AddTasks([]common.Task{task2})
	model2.current = 1
	model2.updateTaskStats()
	view2 := model2.View()

	// Проверяем, что View содержит сводку
	assert.Contains(t, view2, defaults.SummaryCompleted, "При showSummary=true View должен содержать сводку")
	assert.Contains(t, view2, defaults.StatusSuccess, "При showSummary=true View должен содержать статус")
	assert.Contains(t, view2, "(1/1)", "При showSummary=true View должен содержать статистику")
}

// TestRemoveVerticalLinesBeforeTaskSymbols проверяет удаление вертикальных линий
func TestRemoveVerticalLinesBeforeTaskSymbols(t *testing.T) {
	// Создаем тестовый контент с вертикальными линиями
	var sb strings.Builder
	sb.WriteString("┌─────────────────────────────────────┐\n")
	sb.WriteString("│  Заголовок                          │\n")
	sb.WriteString("├─────────────────────────────────────┤\n")
	sb.WriteString("  ✓ Задача 1: завершено              \n")
	sb.WriteString("  │                                   \n")
	sb.WriteString("  ✓ Задача 2: завершено              \n")
	sb.WriteString("  │                                   \n") // Эта линия должна быть заменена на пробелы
	sb.WriteString("  │                                   \n") // Эта тоже

	// Применяем функцию
	removeVerticalLinesBeforeTaskSymbols(&sb)

	result := sb.String()

	// Проверяем, что вертикальные линии после последней задачи заменены на пробелы
	lines := strings.Split(result, "\n")

	// Находим последнюю строку с символом задачи
	lastTaskLine := -1
	for i, line := range lines {
		if strings.Contains(line, "✓") {
			lastTaskLine = i
		}
	}

	assert.NotEqual(t, -1, lastTaskLine, "Должна быть найдена строка с символом задачи")

	// Проверяем строки после последней задачи - в них не должно быть вертикальных линий │
	for i := lastTaskLine + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, "│") {
			// Если есть │, это должна быть рамка, а не вертикальная линия отступа для задач
			assert.True(t, strings.HasPrefix(strings.TrimSpace(line), "│") ||
				strings.HasSuffix(strings.TrimSpace(line), "│"),
				"Вертикальные линии должны быть только в рамке, не в отступах задач в строке: %s", line)
		}
	}
}
