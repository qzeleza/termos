// task/func_task_test.go

package task

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/stretchr/testify/assert"
)

// TestFuncTaskCreation проверяет корректность создания задачи FuncTask
func TestFuncTaskCreation(t *testing.T) {
	// Создаем задачу FuncTask с успешным выполнением
	title := "Тестовая функция"
	funcTask := NewFuncTask(title, func() error {
		return nil
	})

	// Проверяем, что задача создана корректно
	assert.NotNil(t, funcTask, "Задача не должна быть nil")
	assert.Equal(t, title, funcTask.Title(), "Заголовок задачи должен соответствовать переданному значению")
	assert.False(t, funcTask.IsDone(), "Новая задача не должна быть отмечена как завершенная")

	// Проверяем значение флага stopOnError по умолчанию
	assert.True(t, funcTask.StopOnError(), "По умолчанию stopOnError должен быть true")

	// Создаем задачу FuncTask с указанием флага stopOnError = false
	funcTaskNoStop := NewFuncTask(title, func() error {
		return nil
	}, WithStopOnError(false))

	// Проверяем значение флага stopOnError
	assert.False(t, funcTaskNoStop.StopOnError(), "stopOnError должен быть false")
}

// TestFuncTaskWithSuccessLabel проверяет метод WithSuccessLabel
func TestFuncTaskWithSuccessLabel(t *testing.T) {
	// Создаем задачу FuncTask
	title := "Тестовая функция"
	funcTask := NewFuncTask(title, func() error {
		return nil
	})

	// Устанавливаем метку успешного завершения
	successLabel := "Готово"
	funcTask = funcTask.WithSuccessLabel(successLabel)

	// Запускаем задачу
	cmd := funcTask.Run()
	assert.NotNil(t, cmd, "Команда запуска задачи не должна быть nil")

	// Симулируем выполнение команды
	msg := funcTaskCompleteMsg{}
	updatedTask, _ := funcTask.Update(msg)

	// Проверяем, что задача завершена успешно
	assert.True(t, updatedTask.IsDone(), "Задача должна быть отмечена как завершенная")
	// Проверяем, что финальное представление содержит метку успеха
	finalView := updatedTask.FinalView(80)
	assert.Contains(t, finalView, "ГОТОВО", "Финальное представление должно содержать метку успешного завершения")
}

// TestFuncTaskWithoutSuccessLabel проверяет отключение метки успешного завершения
func TestFuncTaskWithoutSuccessLabel(t *testing.T) {
	funcTask := NewFuncTask("Без статуса", func() error {
		return nil
	})

	funcTask = funcTask.WithoutSuccessLabel()

	updatedTask, _ := funcTask.Update(funcTaskCompleteMsg{})
	finalView := updatedTask.FinalView(80)

	assert.NotContains(t, finalView, strings.ToUpper(defaults.DefaultSuccessLabel), "Финальное представление не должно содержать метку успешного завершения")
}

// TestFuncTaskSuccessfulExecution проверяет успешное выполнение задачи
func TestFuncTaskSuccessfulExecution(t *testing.T) {
	// Создаем задачу FuncTask с успешным выполнением
	funcTask := NewFuncTask("Успешная задача", func() error {
		// В реальном сценарии здесь выполняется какая-то работа
		return nil
	})

	// Запускаем задачу
	cmd := funcTask.Run()
	assert.NotNil(t, cmd, "Команда запуска задачи не должна быть nil")

	// Симулируем выполнение функции и отправку сообщения
	msg := funcTaskCompleteMsg{}
	updatedTask, _ := funcTask.Update(msg)

	// Проверяем, что задача завершена успешно
	assert.True(t, updatedTask.IsDone(), "Задача должна быть отмечена как завершенная")
	assert.False(t, updatedTask.HasError(), "Задача не должна содержать ошибок")
	assert.Nil(t, updatedTask.Error(), "Ошибка задачи должна быть nil")

	// Проверяем, что функция была выполнена при запуске команды
	// В реальном сценарии функция выполняется асинхронно, но в тесте мы симулируем её выполнение
	// Поэтому здесь мы не можем проверить executed напрямую, так как мы симулируем сообщение
}

// TestFuncTaskErrorExecution проверяет выполнение задачи с ошибкой
func TestFuncTaskErrorExecution(t *testing.T) {
	// Создаем задачу FuncTask с ошибкой
	expectedError := errors.New("тестовая ошибка")
	funcTask := NewFuncTask("Задача с ошибкой", func() error {
		return expectedError
	})

	// Запускаем задачу
	cmd := funcTask.Run()
	assert.NotNil(t, cmd, "Команда запуска задачи не должна быть nil")

	// Симулируем получение ошибки от функции
	updatedTask, _ := funcTask.Update(expectedError)

	// Проверяем, что задача завершена с ошибкой
	assert.True(t, updatedTask.IsDone(), "Задача должна быть отмечена как завершенная")
	assert.True(t, updatedTask.HasError(), "Задача должна содержать ошибку")
	assert.Equal(t, expectedError, updatedTask.Error(), "Ошибка задачи должна соответствовать ожидаемой")
}

// TestFuncTaskUpdate проверяет обработку сообщений в методе Update
func TestFuncTaskUpdate(t *testing.T) {
	// Создаем задачу FuncTask
	funcTask := NewFuncTask("Тестовая задача", func() error {
		return nil
	})

	// Запускаем задачу
	cmd := funcTask.Run()
	assert.NotNil(t, cmd, "Команда запуска задачи не должна быть nil")

	// Проверяем, что метод Update не изменяет состояние задачи при нажатии клавиш
	updatedTask, cmd := funcTask.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, funcTask, updatedTask, "Метод Update не должен изменять задачу при нажатии клавиш")
	assert.Nil(t, cmd, "Команда должна быть nil при нажатии клавиш")
}

// TestFuncTaskView проверяет отображение задачи в активном состоянии
func TestFuncTaskView(t *testing.T) {
	// Создаем задачу FuncTask
	title := "Тестовая задача"
	funcTask := NewFuncTask(title, func() error {
		return nil
	})

	// Проверяем, что View содержит заголовок
	view := funcTask.View(80)
	assert.Contains(t, view, title, "View должен содержать заголовок задачи")

	// Запускаем задачу и симулируем завершение
	cmd := funcTask.Run()
	assert.NotNil(t, cmd, "Команда запуска задачи не должна быть nil")

	// Симулируем успешное завершение
	updatedTask, _ := funcTask.Update(funcTaskCompleteMsg{})

	// Проверяем, что View для завершенной задачи возвращает FinalView
	view = updatedTask.View(80)
	assert.Contains(t, view, "ГОТОВО", "View для завершенной задачи должен содержать метку успешного завершения")
}
