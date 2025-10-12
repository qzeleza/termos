package query

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/ui"
	"github.com/stretchr/testify/assert"
)

// MockErrorTask для тестирования ошибок
type MockErrorTask struct {
	title                 string
	done                  bool
	err                   error
	preserveErrorNewLines bool
}

func (m *MockErrorTask) Title() string            { return m.title }
func (m *MockErrorTask) IsDone() bool             { return m.done }
func (m *MockErrorTask) HasError() bool           { return m.err != nil }
func (m *MockErrorTask) Error() error             { return m.err }
func (m *MockErrorTask) StopOnError() bool        { return false }
func (m *MockErrorTask) SetStopOnError(stop bool) {}
func (m *MockErrorTask) WithNewLinesInErrors(preserve bool) common.Task {
	m.preserveErrorNewLines = preserve
	return m
}
func (m *MockErrorTask) Run() tea.Cmd                              { return nil }
func (m *MockErrorTask) Update(msg tea.Msg) (common.Task, tea.Cmd) { return m, nil }
func (m *MockErrorTask) View(width int) string                     { return m.title }
func (m *MockErrorTask) FinalView(width int) string                { return m.title }

func NewMockErrorTask(title string, hasError bool) *MockErrorTask {
	task := &MockErrorTask{
		title: title,
		done:  true,
	}
	if hasError {
		task.err = errors.New("тестовая ошибка")
	}
	return task
}

// TestSetErrorColorInQueue проверяет работу SetErrorColor в контексте очереди
func TestSetErrorColorInQueue(t *testing.T) {
	// Сохраняем исходные стили
	ui.ResetErrorColors()
	originalMessageStyle := ui.GetErrorMessageStyle()
	originalStatusStyle := ui.GetErrorStatusStyle()

	// Создаем задачи с ошибками
	successTask := NewMockErrorTask("Успешная задача", false)
	errorTask := NewMockErrorTask("Задача с ошибкой", true)

	// Создаем модель очереди с кастомным цветом ошибок
	model := New("Тест цвета ошибок").
		WithSummary(true).
		SetErrorColor(Red)

	tasks := []common.Task{successTask, errorTask}
	model.AddTasks(tasks)

	// Симулируем завершение всех задач
	model.current = len(tasks)
	model.updateTaskStats()

	// Проверяем, что глобальные стили изменились
	assert.Equal(t, ui.ColorDarkRed, ui.GetErrorMessageStyle().GetForeground())
	assert.Equal(t, ui.ColorBrightRed, ui.GetErrorStatusStyle().GetForeground())

	// Получаем представление
	view := model.View()

	// Проверяем, что в представлении есть статус с ошибками
	assert.Contains(t, view, defaults.StatusProblem, "View должен содержать статус ошибки")
	assert.Contains(t, view, "(1/2)", "View должен показывать правильную статистику")

	// Восстанавливаем исходные стили
	ui.ErrorMessageStyle = originalMessageStyle
	ui.ErrorStatusStyle = originalStatusStyle
}

// TestSetErrorColorChaining проверяет цепочку вызовов методов
func TestSetErrorColorChaining(t *testing.T) {
	// Сбрасываем к значениям по умолчанию
	ui.ResetErrorColors()

	// Тестируем цепочку вызовов
	model := New("Тест цепочки").
		WithSummary(true).
		SetErrorColor(Red).
		WithTitleColor(ui.ColorBrightBlue, true).
		WithAppName("TestApp")

	// Проверяем, что модель создана правильно
	assert.NotNil(t, model)
	assert.True(t, model.showSummary)
	assert.Contains(t, model.appName, "TestApp")

	// Проверяем, что цвет ошибок установился
	assert.Equal(t, ui.ColorDarkRed, ui.GetErrorMessageStyle().GetForeground())
	assert.Equal(t, ui.ColorBrightRed, ui.GetErrorStatusStyle().GetForeground())

	// Сбрасываем стили
	ui.ResetErrorColors()
}

// TestMultipleErrorColorChanges проверяет множественные изменения цвета
func TestMultipleErrorColorChanges(t *testing.T) {
	ui.ResetErrorColors()

	// Создаем модель
	model := New("Тест множественных изменений").WithSummary(true)

	// Изменяем цвет несколько раз
	model.SetErrorColor(Red)
	assert.Equal(t, ui.ColorBrightRed, ui.GetErrorStatusStyle().GetForeground())

	model.SetErrorColor(Orange)
	assert.Equal(t, ui.ColorBrightOrange, ui.GetErrorStatusStyle().GetForeground())

	// Сбрасываем
	ui.ResetErrorColors()
	assert.Equal(t, ui.ColorBrightYellow, ui.GetErrorStatusStyle().GetForeground())
}

// TestErrorColorWithDifferentTasks проверяет работу с разными типами задач
func TestErrorColorWithDifferentTasks(t *testing.T) {
	ui.ResetErrorColors()

	// Создаем задачи разных типов
	tasks := []common.Task{
		NewMockErrorTask("Задача 1", false), // Успешная
		NewMockErrorTask("Задача 2", true),  // С ошибкой
		NewMockErrorTask("Задача 3", false), // Успешная
		NewMockErrorTask("Задача 4", true),  // С ошибкой
	}

	// Устанавливаем кастомный цвет
	model := New("Тест с разными задачами").
		WithSummary(true).
		SetErrorColor(Orange)

	model.AddTasks(tasks)
	model.current = len(tasks)
	model.updateTaskStats()

	// Проверяем статистику: 2 успешных из 4
	assert.Equal(t, 2, model.successCount)
	assert.Equal(t, 2, model.errorCount)

	// Получаем представление
	view := model.View()
	assert.Contains(t, view, "(2/4)", "Должно быть 2 успешных из 4")
	assert.Contains(t, view, defaults.StatusProblem, "Должен быть статус с ошибками")

	// Проверяем, что цвет установился правильно
	assert.Equal(t, ui.ColorBrightOrange, ui.GetErrorStatusStyle().GetForeground())

	// Сбрасываем
	ui.ResetErrorColors()
}
