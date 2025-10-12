package task

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInputTaskRightArrowAppliesDefaultValue(t *testing.T) {
	task := NewInputTaskNew("Поле", "Введите значение")
	task.WithTimeout(time.Second, "значение по умолчанию")
	task.DisableTimeout()

	assert.Empty(t, task.textInput.Value(), "Поле ввода должно быть пустым до нажатия клавиши")

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyRight})
	updatedTask, ok := updated.(*InputTaskNew)
	if assert.True(t, ok, "Обновлённая задача должна быть типа *InputTaskNew") {
		assert.False(t, updatedTask.IsDone(), "После вставки значения по умолчанию задача не должна завершаться")
		assert.Equal(t, "значение по умолчанию", updatedTask.textInput.Value(), "Значение по умолчанию должно быть вставлено в поле ввода")
	}
}

func TestInputTaskEnterWithDefaultValueCompletes(t *testing.T) {
	task := NewInputTaskNew("Поле", "Введите значение")
	task.WithTimeout(time.Second, "значение по умолчанию")
	task.DisableTimeout()

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedTask, ok := updated.(*InputTaskNew)
	if assert.True(t, ok, "Обновлённая задача должна быть типа *InputTaskNew") {
		assert.True(t, updatedTask.IsDone(), "Задача должна завершиться после нажатия Enter с пустым полем и значением по умолчанию")
		assert.Equal(t, "значение по умолчанию", updatedTask.GetValue(), "Возвращаемое значение должно соответствовать значению по умолчанию")
		assert.NoError(t, updatedTask.Error(), "Задача должна завершиться без ошибки")
	}
}

func TestInputTaskDefaultValueSetsPlaceholder(t *testing.T) {
	task := NewInputTaskNew("Поле", "Введите значение")
	task.WithTimeout(time.Second, "авто")

	assert.Equal(t, "авто", task.textInput.Placeholder, "Placeholder должен наследовать значение по умолчанию")
}

func TestInputTaskCustomPlaceholderHasPriority(t *testing.T) {
	task := NewInputTaskNew("Поле", "Введите значение")
	task.WithPlaceholder("кастом")
	task.WithTimeout(time.Second, "авто")

	assert.Equal(t, "кастом", task.textInput.Placeholder, "Пользовательский placeholder должен иметь приоритет над значением по умолчанию")
}
