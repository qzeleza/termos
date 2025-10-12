// task/base_task_test.go

package task

import (
	"errors"
	"strings"
	"testing"

	"github.com/qzeleza/ziva/internal/ui"
	"github.com/stretchr/testify/assert"
)

// TestBaseTaskCreation проверяет корректность создания BaseTask
func TestBaseTaskCreation(t *testing.T) {
	// Создаем базовую задачу
	title := "Тестовая задача"
	baseTask := NewBaseTask(title)

	// Проверяем, что задача создана корректно
	assert.Equal(t, title, baseTask.Title(), "Заголовок задачи должен соответствовать переданному значению")
	assert.False(t, baseTask.IsDone(), "Новая задача не должна быть отмечена как завершенная")
	assert.False(t, baseTask.HasError(), "Новая задача не должна содержать ошибок")
	assert.Nil(t, baseTask.Error(), "Ошибка новой задачи должна быть nil")
	assert.True(t, baseTask.StopOnError(), "По умолчанию stopOnError должен быть true")
}

// TestBaseTaskStopOnError проверяет работу флага StopOnError
func TestBaseTaskStopOnError(t *testing.T) {
	// Создаем базовую задачу
	baseTask := NewBaseTask("Тестовая задача")

	// Устанавливаем флаг StopOnError
	baseTask.SetStopOnError(true)
	assert.True(t, baseTask.StopOnError(), "stopOnError должен быть true после установки")

	// Сбрасываем флаг StopOnError
	baseTask.SetStopOnError(false)
	assert.False(t, baseTask.StopOnError(), "stopOnError должен быть false после сброса")
}

// TestBaseTaskView проверяет метод View
func TestBaseTaskView(t *testing.T) {
	// Создаем базовую задачу
	title := "Тестовая задача"
	baseTask := NewBaseTask(title)

	// Проверяем, что View содержит заголовок
	view := baseTask.View(80)
	assert.Contains(t, view, title, "View должен содержать заголовок задачи")
}

// TestBaseTaskFinalView проверяет метод FinalView
func TestBaseTaskFinalView(t *testing.T) {
	// Создаем базовую задачу
	title := "Тестовая задача"
	baseTask := NewBaseTask(title)

	// Проверяем FinalView для задачи без ошибок
	finalView := baseTask.FinalView(80)
	assert.Contains(t, finalView, title, "FinalView должен содержать заголовок задачи")

	// Создаем задачу с ошибкой
	errorTask := NewBaseTask(title)
	expectedError := errors.New("тестовая ошибка")
	errorTask.err = expectedError

	// Устанавливаем иконку ошибки и финальное значение
	errorTask.icon = ui.IconError // Иконка ошибки из пакета ui
	errorTask.finalValue = expectedError.Error()
	errorTask.done = true

	// Проверяем FinalView для задачи с ошибкой
	errorFinalView := errorTask.FinalView(80)
	assert.Contains(t, errorFinalView, title, "FinalView должен содержать заголовок задачи")
	assert.Contains(t, errorFinalView, "ОШИБКА", "FinalView должен содержать слово 'ОШИБКА'")
	assert.Contains(t, strings.ToLower(errorFinalView), strings.ToLower(expectedError.Error()), "FinalView должен содержать текст ошибки")
}
