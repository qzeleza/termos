// task/multiselect_task_test.go

package task

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeMultiItems(values []string) []Item {
	result := make([]Item, len(values))
	for i, v := range values {
		result[i] = Item{Key: v, Name: v}
	}
	return result
}

// TestMultiSelectTaskCreation проверяет корректность создания задачи MultiSelectTask
func TestMultiSelectTaskCreation(t *testing.T) {
	// Создаем задачу MultiSelectTask
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}

	// Создаем задачу
	multiSelectTask := NewMultiSelectTask(title, makeMultiItems(options))

	// Проверяем, что задача создана корректно
	assert.NotNil(t, multiSelectTask, "Задача не должна быть nil")
	assert.Equal(t, title, multiSelectTask.Title(), "Заголовок задачи должен соответствовать переданному значению")
	assert.False(t, multiSelectTask.IsDone(), "Новая задача не должна быть отмечена как завершенная")

	// Создаем еще одну задачу
	multiSelectTaskEmpty := NewMultiSelectTask(title, makeMultiItems(options))

	// Проверяем, что задача создана корректно
	assert.NotNil(t, multiSelectTaskEmpty, "Задача не должна быть nil")
	assert.Equal(t, title, multiSelectTaskEmpty.Title(), "Заголовок задачи должен соответствовать переданному значению")
	assert.False(t, multiSelectTaskEmpty.IsDone(), "Новая задача не должна быть отмечена как завершенная")
}

// TestMultiSelectTaskUpdate проверяет обработку различных клавиш в методе Update
func TestMultiSelectTaskUpdate(t *testing.T) {
	// Создаем задачу MultiSelectTask
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}
	multiSelectTask := NewMultiSelectTask(title, makeMultiItems(options))

	// Проверяем обработку клавиши 'down'
	updatedTask, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	multiSelectTaskAfterDown, ok := updatedTask.(*MultiSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *MultiSelectTask")
	assert.False(t, multiSelectTaskAfterDown.IsDone(), "Задача не должна быть отмечена как завершенная после нажатия 'down'")

	// Проверяем обработку клавиши 'up'
	updatedTask, _ = multiSelectTaskAfterDown.Update(tea.KeyMsg{Type: tea.KeyUp})
	multiSelectTaskAfterUp, ok := updatedTask.(*MultiSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *MultiSelectTask")
	assert.False(t, multiSelectTaskAfterUp.IsDone(), "Задача не должна быть отмечена как завершенная после нажатия 'up'")

	// Проверяем обработку клавиши 'space' для выбора опции
	updatedTask, _ = multiSelectTaskAfterUp.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTaskAfterSpace, ok := updatedTask.(*MultiSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *MultiSelectTask")
	assert.False(t, multiSelectTaskAfterSpace.IsDone(), "Задача не должна быть отмечена как завершенная после нажатия 'space'")

	// Проверяем обработку клавиши 'enter' для завершения задачи
	updatedTask, _ = multiSelectTaskAfterSpace.Update(tea.KeyMsg{Type: tea.KeyEnter})
	multiSelectTaskAfterEnter, ok := updatedTask.(*MultiSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *MultiSelectTask")
	assert.True(t, multiSelectTaskAfterEnter.IsDone(), "Задача должна быть отмечена как завершенная после нажатия 'enter'")
}

// TestMultiSelectTaskView проверяет отображение задачи в активном состоянии
func TestMultiSelectTaskView(t *testing.T) {
	// Создаем задачу MultiSelectTask
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}
	multiSelectTask := NewMultiSelectTask(title, makeMultiItems(options))

	// Проверяем, что View содержит заголовок и опции
	view := multiSelectTask.View(80)
	assert.Contains(t, view, title, "View должен содержать заголовок")
	for _, option := range options {
		assert.Contains(t, view, option, "View должен содержать опцию")
	}

	// Сначала выбираем опцию с помощью пробела
	updatedTask1, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTaskWithSelection, _ := updatedTask1.(*MultiSelectTask)

	// Затем завершаем задачу с помощью Enter
	updatedTask2, _ := multiSelectTaskWithSelection.Update(tea.KeyMsg{Type: tea.KeyEnter})
	multiSelectTaskDone, _ := updatedTask2.(*MultiSelectTask)

	// Проверяем, что задача завершена
	assert.True(t, multiSelectTaskDone.IsDone(), "Задача должна быть завершена после нажатия Enter")
}

// TestMultiSelectTaskFinalValue проверяет корректность финального значения
func TestMultiSelectTaskFinalValue(t *testing.T) {
	// Создаем задачу MultiSelectTask
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}
	multiSelectTask := NewMultiSelectTask(title, makeMultiItems(options))

	// Выбираем первую опцию с помощью пробела
	updatedTask1, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTaskWithSelection, _ := updatedTask1.(*MultiSelectTask)

	// Завершаем задачу с помощью Enter
	updatedTask2, _ := multiSelectTaskWithSelection.Update(tea.KeyMsg{Type: tea.KeyEnter})
	multiSelectTaskDone, _ := updatedTask2.(*MultiSelectTask)

	// Проверяем, что задача завершена и финальное значение содержит выбранную опцию
	assert.True(t, multiSelectTaskDone.IsDone(), "Задача должна быть завершена после нажатия Enter")

	// Проверяем финальное представление
	finalView := multiSelectTaskDone.FinalView(80)
	assert.Contains(t, finalView, "Опция 1", "FinalView должен содержать выбранную опцию")
}

// TestMultiSelectTaskBoundaries проверяет обработку граничных случаев
func TestMultiSelectTaskBoundaries(t *testing.T) {
	// Создаем задачу MultiSelectTask
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}
	multiSelectTask := NewMultiSelectTask(title, makeMultiItems(options))

	// Движемся вниз до последнего элемента
	for i := 0; i < len(options)-1; i++ {
		updatedTask, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
		multiSelectTask, _ = updatedTask.(*MultiSelectTask)
	}
	assert.Equal(t, len(options)-1, multiSelectTask.cursor, "Курсор должен находиться на последнем элементе")

	// Следующее нажатие вниз возвращает курсор к первому элементу
	updatedTask, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	multiSelectTask, _ = updatedTask.(*MultiSelectTask)
	assert.Equal(t, 0, multiSelectTask.cursor, "После достижения конца список должен зациклиться к началу")

	// Стрелка вверх с первого элемента переносит на конец
	updatedTask, _ = multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyUp})
	multiSelectTask, _ = updatedTask.(*MultiSelectTask)
	assert.Equal(t, len(options)-1, multiSelectTask.cursor, "Стрелка вверх на первом элементе должна переносить на последний")

	// Выбираем последнюю опцию и завершаем задачу
	updatedTask, _ = multiSelectTask.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTask, _ = updatedTask.(*MultiSelectTask)
	updatedTask, _ = multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyEnter})
	multiSelectTaskDone, _ := updatedTask.(*MultiSelectTask)
	assert.True(t, multiSelectTaskDone.IsDone(), "Задача должна быть завершена")
	finalView := multiSelectTaskDone.FinalView(80)
	assert.Contains(t, finalView, "Опция 3", "FinalView должен содержать выбранную опцию")

	// Сбрасываем задачу и проверяем перенос вверх → вниз
	multiSelectTask = NewMultiSelectTask(title, makeMultiItems(options))
	updatedTask, _ = multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyUp})
	multiSelectTask, _ = updatedTask.(*MultiSelectTask)
	assert.Equal(t, len(options)-1, multiSelectTask.cursor, "Первое нажатие вверх должно переносить на последний элемент")

	updatedTask, _ = multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	multiSelectTask, _ = updatedTask.(*MultiSelectTask)
	assert.Equal(t, 0, multiSelectTask.cursor, "После возвращения вниз курсор должен оказаться на первом элементе")

	// Выбираем первую опцию и завершаем задачу
	updatedTask, _ = multiSelectTask.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTask, _ = updatedTask.(*MultiSelectTask)
	updatedTask, _ = multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyEnter})
	multiSelectTaskDone, _ = updatedTask.(*MultiSelectTask)
	assert.True(t, multiSelectTaskDone.IsDone(), "Задача должна быть завершена")
	view := multiSelectTaskDone.FinalView(80)
	assert.Contains(t, view, "Опция 1", "FinalView должен содержать выбранную опцию")
}

// TestMultiSelectTaskWithSelectAll проверяет функциональность "Выбрать все"
func TestMultiSelectTaskWithSelectAll(t *testing.T) {
	// Создаем задачу MultiSelectTask с опцией "Выбрать все"
	title := "Выберите компоненты"
	options := []string{"API", "Frontend", "Database", "Worker"}
	multiSelectTask := NewMultiSelectTask(title, makeMultiItems(options)).WithSelectAll()

	// Проверяем, что опция "Выбрать все" включена
	assert.True(t, multiSelectTask.hasSelectAll, "Опция 'Выбрать все' должна быть включена")
	assert.Equal(t, -1, multiSelectTask.cursor, "Курсор должен быть на опции 'Выбрать все'")
	assert.Equal(t, defaults.SelectAllDefaultText, multiSelectTask.selectAllEnableText, "Текст по умолчанию должен совпадать с локализацией")
	assert.Equal(t, defaults.SelectAllDisableDefaultText, multiSelectTask.selectAllDisableText, "Текст отключения по умолчанию должен совпадать с локализацией")
	assert.Equal(t, ui.MenuActionDefaultStyle(), multiSelectTask.selectAllStyle, "Стиль по умолчанию должен совпадать с ожидаемым")

	// Проверяем, что View содержит опцию "Выбрать все"
	view := multiSelectTask.View(80)
	assert.Contains(t, view, defaults.SelectAllDefaultText, "View должен содержать опцию 'Выбрать все'")

	// Проверяем циклическую навигацию с учётом опции "Выбрать все"
	updatedTaskNav, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyUp})
	multiSelectTask, _ = updatedTaskNav.(*MultiSelectTask)
	assert.Equal(t, len(options)-1, multiSelectTask.cursor, "Стрелка вверх на 'Выбрать все' должна переносить на последний элемент")
	updatedTaskNav, _ = multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	multiSelectTask, _ = updatedTaskNav.(*MultiSelectTask)
	assert.Equal(t, -1, multiSelectTask.cursor, "Стрелка вниз на последнем элементе должна возвращать к опции 'Выбрать все'")

	// Выбираем "Выбрать все" с помощью пробела
	updatedTask1, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTaskAfterSelectAll, _ := updatedTask1.(*MultiSelectTask)

	// Проверяем, что все элементы выбраны
	assert.True(t, multiSelectTaskAfterSelectAll.isAllSelected(), "Все элементы должны быть выбраны")
	assert.Equal(t, len(options), multiSelectTaskAfterSelectAll.selected.Count(), "Количество выбранных элементов должно равняться количеству опций")
	selectAllView := stripANSI(multiSelectTaskAfterSelectAll.View(80))
	assert.Contains(t, selectAllView, defaults.SelectAllDisableDefaultText, "После выбора всех элементов текст должен переключаться на режим отключения")

	// Повторно выбираем "Выбрать все" чтобы снять выбор со всех
	updatedTask2, _ := multiSelectTaskAfterSelectAll.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTaskAfterUnselectAll, _ := updatedTask2.(*MultiSelectTask)

	// Проверяем, что все элементы не выбраны
	assert.False(t, multiSelectTaskAfterUnselectAll.isAllSelected(), "Все элементы должны быть не выбраны")
	assert.Equal(t, 0, multiSelectTaskAfterUnselectAll.selected.Count(), "Количество выбранных элементов должно быть равно 0")
	selectNoneView := stripANSI(multiSelectTaskAfterUnselectAll.View(80))
	assert.Contains(t, selectNoneView, defaults.SelectAllDefaultText, "После отключения всех элементов должен отображаться текст включения")
}

func TestMultiSelectTaskDividerRecognition(t *testing.T) {
	divider := SingleSelectDividerItem
	items := []Item{
		{Key: "first", Name: "Первый"},
		divider,
		{Key: "second", Name: "Второй пункт меню"},
	}

	task := NewMultiSelectTask("С разделителем", items)

	assert.True(t, task.isDisabled(1), "Разделитель должен быть недоступным для выбора")
	assert.Equal(t, 0, task.cursor, "Курсор должен указывать на первый доступный элемент")

	task.toggleSelection(1)
	assert.False(t, task.isSelected(1), "Разделитель не должен попадать в выбранные")

	moved := task.moveCursorForward()
	assert.True(t, moved, "Курсор должен переключиться на следующий доступный элемент")
	assert.Equal(t, 2, task.cursor, "Курсор должен пропускать разделитель")

	assert.Equal(t, -1, task.choiceIndex(divider.Name), "Разделитель не должен сопоставляться по имени")
}

func TestMultiSelectTaskDividerViewRendering(t *testing.T) {
	divider := SingleSelectDividerItem
	items := []Item{
		{Key: "first", Name: "Первый пункт"},
		divider,
		{Key: "second", Name: "Вариант два"},
		{Key: "third", Name: "Очень длинный пункт для проверки"},
	}

	task := NewMultiSelectTask("С разделителем", items)

	view := stripANSI(task.View(100))
	lines := strings.Split(view, "\n")
	var dividerLine string
	for _, line := range lines {
		if strings.Contains(line, "─") && !strings.Contains(line, "[") {
			dividerLine = strings.TrimSpace(line)
			break
		}
	}
	if assert.NotEmpty(t, dividerLine, "Вывод должен содержать строку разделителя") {
		expectedLength := longestItemNameLength(items) + 5
		dividerText := strings.TrimLeft(dividerLine, "│")
		dividerText = strings.TrimSpace(dividerText)
		assert.True(t, strings.Trim(dividerText, "─") == "", "Разделитель должен состоять только из символов '─'")
		assert.GreaterOrEqual(t, strings.Count(dividerText, "─"), expectedLength, "Разделитель должен быть не короче самого длинного пункта плюс 5 символов")
		assert.NotContains(t, dividerText, "[", "Разделитель не должен содержать индикатор выбора")
	}
}

func TestMultiSelectTaskDisabledItems(t *testing.T) {
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}

	task := NewMultiSelectTask(title, makeMultiItems(options))
	task = task.WithItemsDisabled([]int{1})
	assert.Equal(t, 0, task.cursor, "Курсор должен начинаться на первом доступном элементе")

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyDown})
	task, _ = updated.(*MultiSelectTask)
	assert.Equal(t, 2, task.cursor, "Курсор должен перепрыгивать через отключённые элементы")

	selectedBefore := task.isSelected(1)
	task.toggleSelection(1)
	assert.Equal(t, selectedBefore, task.isSelected(1), "Отключённый элемент не должен переключать состояние выбора")

	task = task.WithItemsDisabled(nil)
	updated, _ = task.Update(tea.KeyMsg{Type: tea.KeyUp})
	task, _ = updated.(*MultiSelectTask)
	assert.Equal(t, 1, task.cursor, "После включения элемента курсор должен уметь на него переходить")
}

func TestMultiSelectTaskDisabledItemsWithSelectAll(t *testing.T) {
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}

	task := NewMultiSelectTask(title, makeMultiItems(options)).WithItemsDisabled([]int{1}).WithSelectAll()
	assert.Equal(t, -1, task.cursor, "Курсор должен быть на опции 'Выбрать все'")

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeySpace})
	task, _ = updated.(*MultiSelectTask)
	assert.True(t, task.isSelected(0), "Должен быть выбран первый доступный элемент")
	assert.False(t, task.isSelected(1), "Отключённый элемент не должен выбираться")
	assert.True(t, task.isSelected(2), "Последний элемент должен быть выбран")
	assert.True(t, task.isAllSelected(), "Все доступные элементы должны считаться выбранными")

	updated, _ = task.Update(tea.KeyMsg{Type: tea.KeySpace})
	task, _ = updated.(*MultiSelectTask)
	assert.False(t, task.isSelected(0), "После повторного выбора все элементы должны быть сняты")
	assert.False(t, task.isSelected(2), "После повторного выбора все элементы должны быть сняты")
}

func TestMultiSelectTaskViewportIndicators(t *testing.T) {
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2", "Опция 3", "Опция 4"}

	task := NewMultiSelectTask(title, makeMultiItems(options)).WithViewport(2)
	// Перемещаем курсор вниз, чтобы появился индикатор сверху
	for i := 0; i < 3; i++ {
		updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyDown})
		task, _ = updated.(*MultiSelectTask)
	}
	viewWithCounters := task.View(80)
	assert.Contains(t, viewWithCounters, "▲", "Индикатор должен содержать символ стрелки")
	assert.Contains(t, viewWithCounters, "выше", "Индикатор должен указывать на элементы выше")

	task = NewMultiSelectTask(title, makeMultiItems(options)).WithViewport(2, false)
	for i := 0; i < 3; i++ {
		updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyDown})
		task, _ = updated.(*MultiSelectTask)
	}
	viewWithoutCounters := task.View(80)
	assert.Contains(t, viewWithoutCounters, "▲", "Индикатор должен содержать символ стрелки")
	assert.NotContains(t, viewWithoutCounters, "above", "При отключении счётчиков текст не должен отображаться")
	assert.NotContains(t, viewWithoutCounters, "выше", "При отключении счётчиков текст не должен отображаться")
}

func TestMultiSelectTaskHelpTagRendering(t *testing.T) {
	items := []Item{
		{Key: "Опция 1", Name: "Опция 1", Description: "подсказка 1"},
		{Key: "Опция 2", Name: "Опция 2", Description: "подсказка 2"},
	}
	task := NewMultiSelectTask("Выбор", items)
	view := task.View(80)
	assert.Contains(t, view, "подсказка 1", "Под активным элементом должна отображаться подсказка")

	// Перемещаемся на следующий элемент, чтобы убедиться, что подсказка сменяется
	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyDown})
	task, _ = updated.(*MultiSelectTask)
	view = task.View(80)
	assert.Contains(t, view, "подсказка 2", "Подсказка должна отображаться для активного элемента")
	assert.NotContains(t, view, "подсказка 1", "Старая подсказка не должна отображаться")
}

func TestMultiSelectTaskWithDefaultItems(t *testing.T) {
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}

	task := NewMultiSelectTask(title, makeMultiItems(options)).WithDefaultItems([]int{0, 2})

	assert.True(t, task.isSelected(0), "Элемент с индексом 0 должен быть выбран")
	assert.False(t, task.isSelected(1), "Элемент с индексом 1 не должен быть выбран")
	assert.True(t, task.isSelected(2), "Элемент с индексом 2 должен быть выбран")
	assert.Equal(t, 0, task.cursor, "Курсор должен быть установлен на первый выбранный элемент")

	task.WithDefaultItems([]string{"Опция 2"})
	assert.False(t, task.isSelected(0), "После переинициализации элемент 0 не должен быть выбран")
	assert.True(t, task.isSelected(1), "Элемент 'Опция 2' должен быть выбран")
	assert.False(t, task.isSelected(2), "Элемент 2 не должен быть выбран")
	assert.Equal(t, 1, task.cursor, "Курсор должен указывать на выбранный элемент")
}

func TestMultiSelectTaskLeftCancels(t *testing.T) {
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2"}

	task := NewMultiSelectTask(title, makeMultiItems(options))

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyLeft})
	canceledTask, ok := updated.(*MultiSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *MultiSelectTask")
	assert.True(t, canceledTask.IsDone(), "Задача должна завершиться после нажатия ←")
	assert.NoError(t, canceledTask.Error(), "После нажатия ← не должно быть ошибки")
	assert.Equal(t, defaults.DefaultSuccessLabel, canceledTask.finalValue, "Завершение должно соответствовать поведению Enter без выбора")
}

func TestMultiSelectTaskRightToggles(t *testing.T) {
	title := "Выберите опции"
	options := []string{"Опция 1", "Опция 2"}

	task := NewMultiSelectTask(title, makeMultiItems(options))

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyRight})
	withSelection, ok := updated.(*MultiSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *MultiSelectTask")
	assert.False(t, withSelection.IsDone(), "Задача не должна завершаться после нажатия →")
	assert.True(t, withSelection.isSelected(0), "Первый элемент должен быть выбран после нажатия →")
}

// TestMultiSelectTaskWithCustomSelectAllText проверяет кастомный текст для "Выбрать все"
func TestMultiSelectTaskWithCustomSelectAllText(t *testing.T) {
	// Создаем задачу MultiSelectTask с кастомным текстом для "Выбрать все"
	title := "Выберите модули"
	options := []string{"Модуль A", "Модуль B", "Модуль C"}
	customText := "Выделить всё"
	multiSelectTask := NewMultiSelectTask(title, makeMultiItems(options)).WithSelectAll(customText)

	// Проверяем, что кастомный текст установлен
	assert.Equal(t, customText, multiSelectTask.selectAllEnableText, "Кастомный текст должен быть установлен")
	assert.Equal(t, defaults.SelectAllDisableDefaultText, multiSelectTask.selectAllDisableText, "Текст отключения должен остаться значением по умолчанию")
	assert.Equal(t, ui.MenuActionDefaultStyle(), multiSelectTask.selectAllStyle, "При отсутствии кастомного стиля должен применяться стиль по умолчанию")

	// Проверяем, что View содержит кастомный текст
	view := multiSelectTask.View(80)
	assert.Contains(t, view, customText, "View должен содержать кастомный текст")
	assert.NotContains(t, view, "Выбрать все", "View не должен содержать текст по умолчанию")
}

func TestMultiSelectTaskWithCustomSelectAllDisableText(t *testing.T) {
	title := "Выберите пакеты"
	options := []string{"Пакет 1", "Пакет 2"}
	enableText := "Включить все"
	disableText := "Снять весь выбор"
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff9f1c")).Bold(true)
	task := NewMultiSelectTask(title, makeMultiItems(options)).WithSelectAll(enableText, disableText, customStyle)

	assert.Equal(t, enableText, task.selectAllEnableText, "Кастомный текст включения должен быть установлен")
	assert.Equal(t, disableText, task.selectAllDisableText, "Кастомный текст отключения должен быть установлен")
	assert.Equal(t, customStyle, task.selectAllStyle, "Должен использоваться кастомный стиль для опции")

	initialView := stripANSI(task.View(80))
	assert.Contains(t, initialView, enableText, "Начальный вид должен содержать текст включения")
	assert.NotContains(t, initialView, disableText, "Начальный вид не должен содержать текст отключения")

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeySpace})
	taskSelected, _ := updated.(*MultiSelectTask)
	selectedView := stripANSI(taskSelected.View(80))
	assert.Contains(t, selectedView, disableText, "После выбора всех элементов должен отображаться текст отключения")
	assert.NotContains(t, selectedView, enableText, "После выбора всех элементов не должен отображаться текст включения")
}

// TestMultiSelectTaskNavigationWithSelectAll проверяет навигацию с опцией "Выбрать все"
func TestMultiSelectTaskNavigationWithSelectAll(t *testing.T) {
	// Создаем задачу MultiSelectTask с опцией "Выбрать все"
	title := "Выберите элементы"
	options := []string{"Элемент 1", "Элемент 2", "Элемент 3"}
	multiSelectTask := NewMultiSelectTask(title, makeMultiItems(options)).WithSelectAll()

	// Изначально курсор должен быть на "Выбрать все"
	assert.Equal(t, -1, multiSelectTask.cursor, "Курсор должен быть на опции 'Выбрать все'")

	// Нажимаем "down" - переходим к первому элементу списка
	updatedTask1, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	multiSelectTask1, _ := updatedTask1.(*MultiSelectTask)
	assert.Equal(t, 0, multiSelectTask1.cursor, "Курсор должен быть на первом элементе")

	// Нажимаем "up" - возвращаемся к "Выбрать все"
	updatedTask2, _ := multiSelectTask1.Update(tea.KeyMsg{Type: tea.KeyUp})
	multiSelectTask2, _ := updatedTask2.(*MultiSelectTask)
	assert.Equal(t, -1, multiSelectTask2.cursor, "Курсор должен вернуться на опцию 'Выбрать все'")

	// Нажимаем "up" еще раз - должны остаться на "Выбрать все"
	updatedTask3, _ := multiSelectTask2.Update(tea.KeyMsg{Type: tea.KeyUp})
	multiSelectTask3, _ := updatedTask3.(*MultiSelectTask)
	assert.Equal(t, len(options)-1, multiSelectTask3.cursor, "При повторном нажатии 'up' курсор должен перейти к последнему элементу")
}

// TestMultiSelectTaskToggleSelectAllLogic проверяет логику переключения "Выбрать все"
func TestMultiSelectTaskToggleSelectAllLogic(t *testing.T) {
	// Создаем задачу MultiSelectTask с опцией "Выбрать все"
	title := "Выберите пункты"
	options := []string{"Пункт 1", "Пункт 2", "Пункт 3", "Пункт 4"}
	multiSelectTask := NewMultiSelectTask(title, makeMultiItems(options)).WithSelectAll()

	// Выбираем некоторые элементы вручную
	// Переходим к первому элементу
	updatedTask1, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	multiSelectTask, _ = updatedTask1.(*MultiSelectTask)

	// Выбираем первый элемент
	updatedTask2, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTask, _ = updatedTask2.(*MultiSelectTask)

	// Переходим ко второму элементу
	updatedTask3, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	multiSelectTask, _ = updatedTask3.(*MultiSelectTask)

	// Выбираем второй элемент
	updatedTask4, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTask, _ = updatedTask4.(*MultiSelectTask)

	// Проверяем, что выбрано 2 элемента из 4
	assert.Equal(t, 2, multiSelectTask.selected.Count(), "Должно быть выбрано 2 элемента")
	assert.False(t, multiSelectTask.isAllSelected(), "Не все элементы должны быть выбраны")

	// Возвращаемся к опции "Выбрать все"
	for multiSelectTask.cursor != -1 {
		updatedTask, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyUp})
		multiSelectTask, _ = updatedTask.(*MultiSelectTask)
	}

	// Нажимаем "Выбрать все" - должны выбраться все элементы
	updatedTask5, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTask, _ = updatedTask5.(*MultiSelectTask)

	// Проверяем, что все элементы выбраны
	assert.Equal(t, len(options), multiSelectTask.selected.Count(), "Все элементы должны быть выбраны")
	assert.True(t, multiSelectTask.isAllSelected(), "Все элементы должны быть выбраны")

	// Нажимаем "Выбрать все" еще раз - должны сняться все выборы
	updatedTask6, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTask, _ = updatedTask6.(*MultiSelectTask)

	// Проверяем, что все элементы не выбраны
	assert.Equal(t, 0, multiSelectTask.selected.Count(), "Все элементы должны быть не выбраны")
	assert.False(t, multiSelectTask.isAllSelected(), "Все элементы должны быть не выбраны")
}

// TestMultiSelectTaskAllowsEmptySelection проверяет, что по умолчанию можно завершить задачу без выбора
func TestMultiSelectTaskAllowsEmptySelection(t *testing.T) {
	options := []string{"Элемент 1", "Элемент 2", "Элемент 3"}
	multiSelectTask := NewMultiSelectTask("Выберите элементы", makeMultiItems(options))

	updatedTask, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyEnter})
	completedTask, _ := updatedTask.(*MultiSelectTask)

	assert.True(t, completedTask.IsDone(), "Задача должна завершаться даже при пустом выборе")
	assert.False(t, completedTask.HasError(), "При пустом выборе не должно быть ошибки")
	assert.False(t, completedTask.showHelpMessage, "Подсказка не должна отображаться по умолчанию")
	assert.Equal(t, defaults.DefaultSuccessLabel, completedTask.finalValue, "Финальный статус должен соответствовать успеху")
}

// TestMultiSelectTaskRequireSelection проверяет поведение при обязательном выборе хотя бы одного пункта
func TestMultiSelectTaskRequireSelection(t *testing.T) {
	options := []string{"Элемент 1", "Элемент 2", "Элемент 3"}
	multiSelectTask := NewMultiSelectTask("Выберите элементы", makeMultiItems(options)).WithRequireSelection(true)

	updatedTask1, _ := multiSelectTask.Update(tea.KeyMsg{Type: tea.KeyEnter})
	multiSelectTaskAfterEnter, _ := updatedTask1.(*MultiSelectTask)

	assert.False(t, multiSelectTaskAfterEnter.IsDone(), "Задача не должна завершиться при пустом выборе, когда выбор обязателен")
	assert.False(t, multiSelectTaskAfterEnter.HasError(), "Не должно быть ошибки при пустом выборе")
	assert.True(t, multiSelectTaskAfterEnter.showHelpMessage, "Должно показываться сообщение-подсказка")
	assert.Equal(t, defaults.NeedSelectAtLeastOne, multiSelectTaskAfterEnter.helpMessage, "Сообщение подсказки должно быть локализованным текстом")

	// Выбираем элемент и завершаем задачу
	updatedTask2, _ := multiSelectTaskAfterEnter.Update(tea.KeyMsg{Type: tea.KeySpace})
	multiSelectTaskAfterSpace, _ := updatedTask2.(*MultiSelectTask)
	updatedTask3, _ := multiSelectTaskAfterSpace.Update(tea.KeyMsg{Type: tea.KeyEnter})
	completedTask, _ := updatedTask3.(*MultiSelectTask)

	assert.True(t, completedTask.IsDone(), "Задача должна завершиться после выбора элемента")
	assert.False(t, completedTask.HasError(), "Не должно быть ошибки при успешном завершении")
}

func TestMultiSelectDependenciesDisableAndEnable(t *testing.T) {
	items := []Item{
		{Key: "logging", Name: "Логирование"},
		{Key: "debug", Name: "Отладка"},
		{Key: "metrics", Name: "Метрики"},
	}

	task := NewMultiSelectTask("Опции", items)
	task.WithDependencies(map[string]MultiSelectDependencyRule{
		"logging": {
			OnSelect: MultiSelectDependencyActions{
				Enable: []string{"debug"},
			},
			OnDeselect: MultiSelectDependencyActions{
				Disable:    []string{"debug"},
				ForceClear: []string{"debug"},
			},
		},
	})

	assert.True(t, task.isDisabled(1))
	assert.False(t, task.isSelected(1))
	task.toggleSelection(1)
	assert.False(t, task.isSelected(1))

	task.toggleSelection(0)
	assert.False(t, task.isDisabled(1))
	assert.False(t, task.isSelected(1))

	task.toggleSelection(1)
	assert.True(t, task.isSelected(1))

	task.toggleSelection(0)
	assert.False(t, task.isSelected(1))
	assert.True(t, task.isDisabled(1))
	assert.Equal(t, 0, task.cursor)
}

func TestMultiSelectTaskDescriptionWrapping(t *testing.T) {
	description := "Описание элемента списка, которое должно автоматически разбиваться на несколько строк при отображении, чтобы текст подсказки не выходил за пределы доступной ширины."
	items := []Item{
		{Key: "first", Name: "Первый", Description: description},
		{Key: "second", Name: "Второй"},
	}

	task := NewMultiSelectTask("Выберите элементы", items)
	view := stripANSI(task.View(80))

	descIndex := strings.Index(view, "Описание элемента списка")
	require.Greater(t, descIndex, -1, "Описание должно присутствовать в выводе задачи")

	helpIndex := strings.Index(view, defaults.MultiSelectHelp)
	require.Greater(t, helpIndex, -1, "Подсказка навигации должна присутствовать в выводе задачи")
	require.Less(t, descIndex, helpIndex, "Описание должно отображаться перед подсказкой навигации")

	descSegment := view[descIndex:helpIndex]
	lines := strings.Split(descSegment, "\n")

	var nonEmpty []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}

	assert.Greater(t, len(nonEmpty), 1, "Описание должно разбиваться на несколько строк")
	for _, line := range nonEmpty {
		assert.True(t, strings.HasPrefix(line, "  "), "Каждая строка описания должна содержать базовый отступ")
	}
}
