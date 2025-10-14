// task/singleselect_task_test.go

package task

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(input string) string {
	return ansiRegexp.ReplaceAllString(input, "")
}

func makeTestItems(values []string) []Item {
	result := make([]Item, len(values))
	for i, v := range values {
		result[i] = Item{Key: v, Name: v}
	}
	return result
}

func longestItemNameLength(items []Item) int {
	maxLen := 0
	for _, item := range items {
		if isDividerChoice(item.Key, item.Name) {
			continue
		}
		trimmed := strings.TrimSpace(item.Name)
		if trimmed == "" {
			continue
		}
		length := utf8.RuneCountInString(trimmed)
		if length > maxLen {
			maxLen = length
		}
	}
	return maxLen
}

// TestSingleSelectTaskCreation проверяет корректность создания задачи SingleSelectTask
func TestSingleSelectTaskCreation(t *testing.T) {
	// Создаем задачу SingleSelectTask
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}

	// Без указания индекса по умолчанию
	selectTask := NewSingleSelectTask(title, makeTestItems(options))

	// Проверяем, что задача создана корректно
	assert.NotNil(t, selectTask, "Задача не должна быть nil")
	assert.Equal(t, title, selectTask.Title(), "Заголовок задачи должен соответствовать переданному значению")
	assert.False(t, selectTask.IsDone(), "Новая задача не должна быть отмечена как завершенная")

	// Создаем еще одну задачу
	selectTaskWithDefault := NewSingleSelectTask(title, makeTestItems(options))

	// Проверяем, что задача создана корректно
	assert.NotNil(t, selectTaskWithDefault, "Задача не должна быть nil")
	assert.Equal(t, title, selectTaskWithDefault.Title(), "Заголовок задачи должен соответствовать переданному значению")
	assert.False(t, selectTaskWithDefault.IsDone(), "Новая задача не должна быть отмечена как завершенная")
}

// TestSingleSelectTaskUpdate проверяет обработку различных клавиш в методе Update
func TestSingleSelectTaskUpdate(t *testing.T) {
	// Создаем задачу SingleSelectTask
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}
	selectTask := NewSingleSelectTask(title, makeTestItems(options))

	// Проверяем обработку клавиши 'down'
	updatedTask, _ := selectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	selectTaskAfterDown, ok := updatedTask.(*SingleSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *SingleSelectTask")
	assert.False(t, selectTaskAfterDown.IsDone(), "Задача не должна быть отмечена как завершенная после нажатия 'down'")

	// Проверяем обработку клавиши 'up'
	updatedTask, _ = selectTaskAfterDown.Update(tea.KeyMsg{Type: tea.KeyUp})
	selectTaskAfterUp, ok := updatedTask.(*SingleSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *SingleSelectTask")
	assert.False(t, selectTaskAfterUp.IsDone(), "Задача не должна быть отмечена как завершенная после нажатия 'up'")

	// Проверяем обработку клавиши 'enter'
	updatedTask, _ = selectTaskAfterUp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	selectTaskAfterEnter, ok := updatedTask.(*SingleSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *SingleSelectTask")
	assert.True(t, selectTaskAfterEnter.IsDone(), "Задача должна быть отмечена как завершенная после нажатия 'enter'")

	// Проверяем, что выбрана правильная опция
	finalView := selectTaskAfterEnter.FinalView(80)
	assert.Contains(t, finalView, options[0], "Значение задачи должно содержать выбранную опцию")
}

// TestSingleSelectTaskView проверяет отображение задачи в активном состоянии
func TestSingleSelectTaskView(t *testing.T) {
	// Создаем задачу SingleSelectTask
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}
	selectTask := NewSingleSelectTask(title, makeTestItems(options))

	// Проверяем, что View содержит заголовок и опции
	view := selectTask.View(80)
	assert.Contains(t, view, title, "View должен содержать заголовок")
	for _, option := range options {
		assert.Contains(t, view, option, "View должен содержать опцию")
	}

	// Проверяем, что после завершения задачи View возвращает пустую строку
	updatedTask, _ := selectTask.Update(tea.KeyMsg{Type: tea.KeyEnter})
	selectTaskDone, _ := updatedTask.(*SingleSelectTask)
	assert.Equal(t, selectTaskDone.FinalView(80), selectTaskDone.View(80), "View завершенной задачи должен совпадать с FinalView")
}

// TestSingleSelectTaskWithDefaultIndex проверяет работу с выбором определенного индекса
func TestSingleSelectTaskWithDefaultIndex(t *testing.T) {
	// Создаем задачу SingleSelectTask
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}
	defauiltIndex := 1
	selectTask := NewSingleSelectTask(title, makeTestItems(options))

	// Устанавливаем курсор на нужный индекс
	// Нажимаем 'down' один раз, чтобы перейти к опции с индексом 1
	updatedTask, _ := selectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	selectTask, _ = updatedTask.(*SingleSelectTask)

	// Нажимаем Enter для завершения задачи
	updatedTask, _ = selectTask.Update(tea.KeyMsg{Type: tea.KeyEnter})
	selectTaskDone, _ := updatedTask.(*SingleSelectTask)
	assert.True(t, selectTaskDone.IsDone(), "Задача должна быть отмечена как завершенная после нажатия 'enter'")

	// Проверяем, что выбрана правильная опция
	finalView := selectTaskDone.FinalView(80)
	assert.Contains(t, finalView, options[defauiltIndex], "Значение задачи должно содержать опцию с выбранным индексом")
}

func TestSingleSelectTaskWithDefaultItemByIndex(t *testing.T) {
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}

	task := NewSingleSelectTask(title, makeTestItems(options)).WithDefaultItem(2)

	assert.Equal(t, 2, task.cursor, "Курсор должен указывать на элемент с индексом 2")
	assert.Equal(t, "Опция 3", task.GetSelected(), "Выбранным значением по умолчанию должна быть 'Опция 3'")
}

func TestSingleSelectTaskWithDefaultItemByValue(t *testing.T) {
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}

	task := NewSingleSelectTask(title, makeTestItems(options)).WithDefaultItem("Опция 2")

	assert.Equal(t, 1, task.cursor, "Курсор должен указывать на элемент с индексом 1")
	assert.Equal(t, "Опция 2", task.GetSelected(), "Выбранным значением по умолчанию должна быть 'Опция 2'")
}

func TestSingleSelectTaskLeftCancels(t *testing.T) {
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2"}

	task := NewSingleSelectTask(title, makeTestItems(options))

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyLeft})
	canceledTask, ok := updated.(*SingleSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *SingleSelectTask")
	assert.True(t, canceledTask.IsDone(), "Задача должна завершиться после нажатия ←")
	assert.NoError(t, canceledTask.Error(), "После нажатия ← не должно быть ошибки")
	assert.Equal(t, options[0], canceledTask.GetSelected(), "Должна быть выбрана текущая опция")
}

func TestSingleSelectTaskRightSelects(t *testing.T) {
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2"}

	task := NewSingleSelectTask(title, makeTestItems(options))

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyRight})
	selectedTask, ok := updated.(*SingleSelectTask)
	assert.True(t, ok, "Обновленная задача должна быть типа *SingleSelectTask")
	assert.True(t, selectedTask.IsDone(), "Задача должна завершиться после нажатия →")
	assert.Equal(t, options[0], selectedTask.GetSelected(), "Должна быть выбрана текущая опция")
}

// TestSingleSelectTaskBoundaries проверяет обработку граничных случаев
func TestSingleSelectTaskBoundaries(t *testing.T) {
	// Создаем задачу SingleSelectTask
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}
	selectTask := NewSingleSelectTask(title, makeTestItems(options))

	// Движемся вниз до последнего элемента
	for i := 0; i < len(options)-1; i++ {
		updatedTask, _ := selectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
		selectTask, _ = updatedTask.(*SingleSelectTask)
	}
	assert.Equal(t, len(options)-1, selectTask.cursor, "Курсор должен находиться на последнем элементе")

	// Ещё одно нажатие вниз возвращает курсор к первому элементу
	updatedTask, _ := selectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	selectTask, _ = updatedTask.(*SingleSelectTask)
	assert.Equal(t, 0, selectTask.cursor, "После достижения конца список должен зациклиться к началу")

	// Стрелка вверх с первого элемента переносит на последний
	updatedTask, _ = selectTask.Update(tea.KeyMsg{Type: tea.KeyUp})
	selectTask, _ = updatedTask.(*SingleSelectTask)
	assert.Equal(t, len(options)-1, selectTask.cursor, "Стрелка вверх на первом элементе должна переносить на последний")

	// Подтверждаем выбор последней опции
	updatedTask, _ = selectTask.Update(tea.KeyMsg{Type: tea.KeyEnter})
	selectTaskDone, _ := updatedTask.(*SingleSelectTask)
	finalView := selectTaskDone.FinalView(80)
	assert.Contains(t, finalView, options[len(options)-1], "Значение задачи должно содержать последнюю опцию")

	// Сбрасываем задачу и проверяем обратное направление
	selectTask = NewSingleSelectTask(title, makeTestItems(options))
	updatedTask, _ = selectTask.Update(tea.KeyMsg{Type: tea.KeyUp})
	selectTask, _ = updatedTask.(*SingleSelectTask)
	assert.Equal(t, len(options)-1, selectTask.cursor, "Первое нажатие вверх должно переносить на последний элемент")

	// После возврата в начало стрелкой вниз подтверждаем первую опцию
	updatedTask, _ = selectTask.Update(tea.KeyMsg{Type: tea.KeyDown})
	selectTask, _ = updatedTask.(*SingleSelectTask)
	assert.Equal(t, 0, selectTask.cursor, "После переноса назад курсор должен вернуться к началу")
	updatedTask, _ = selectTask.Update(tea.KeyMsg{Type: tea.KeyEnter})
	selectTaskDone, _ = updatedTask.(*SingleSelectTask)
	finalView = selectTaskDone.FinalView(80)
	assert.Contains(t, finalView, options[0], "Значение задачи должно содержать первую опцию")
}

func TestSingleSelectTaskDisabledItems(t *testing.T) {
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2", "Опция 3"}

	task := NewSingleSelectTask(title, makeTestItems(options))
	task = task.WithItemsDisabled([]int{1})

	assert.Equal(t, 0, task.GetSelectedIndex(), "Курсор должен оставаться на первом доступном элементе")

	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyDown})
	task, _ = updated.(*SingleSelectTask)
	assert.Equal(t, 2, task.GetSelectedIndex(), "Курсор должен перепрыгивать через отключённый элемент")

	task.WithDefaultItem(1)
	assert.Equal(t, 2, task.GetSelectedIndex(), "Значение по умолчанию не должно указывать на выключенный элемент")

	task = task.WithItemsDisabled(nil)
	updated, _ = task.Update(tea.KeyMsg{Type: tea.KeyUp})
	task, _ = updated.(*SingleSelectTask)
	assert.Equal(t, 1, task.GetSelectedIndex(), "После включения элемента курсор должен уметь на него переходить")
}

func TestSingleSelectTaskViewportIndicators(t *testing.T) {
	title := "Выберите опцию"
	options := []string{"Опция 1", "Опция 2", "Опция 3", "Опция 4"}

	task := NewSingleSelectTask(title, makeTestItems(options)).WithViewport(2)
	for i := 0; i < 3; i++ {
		updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyDown})
		task, _ = updated.(*SingleSelectTask)
	}
	viewWithCounters := task.View(80)
	assert.Contains(t, viewWithCounters, "▲", "Индикатор должен содержать символ стрелки")
	assert.Contains(t, viewWithCounters, "выше", "Индикатор должен указывать на элементы выше")

	task = NewSingleSelectTask(title, makeTestItems(options)).WithViewport(2, false)
	for i := 0; i < 3; i++ {
		updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyDown})
		task, _ = updated.(*SingleSelectTask)
	}
	viewWithoutCounters := task.View(80)
	assert.Contains(t, viewWithoutCounters, "▲", "Индикатор должен содержать символ стрелки")
	assert.NotContains(t, viewWithoutCounters, "above", "При отключении счётчиков текст не должен отображаться")
	assert.NotContains(t, viewWithoutCounters, "выше", "При отключении счётчиков текст не должен отображаться")
}

func TestSingleSelectTaskHelpTagRendering(t *testing.T) {
	items := []Item{
		{Key: "Опция 1", Name: "Опция 1", Description: "подсказка 1"},
		{Key: "Опция 2", Name: "Опция 2"},
	}
	task := NewSingleSelectTask("Выбор", items)
	view := task.View(80)
	assert.Contains(t, view, "Опция 1", "Отображается базовое название без подсказки")
	assert.Contains(t, view, "подсказка 1", "Подсказка должна отображаться под активным элементом")

	// Перемещаемся на элемент без подсказки
	updated, _ := task.Update(tea.KeyMsg{Type: tea.KeyDown})
	task, _ = updated.(*SingleSelectTask)
	view = task.View(80)
	assert.NotContains(t, view, "подсказка 1", "Пустая подсказка не должна добавлять строку")

	cleanView := stripANSI(view)
	helpIndent := strings.Repeat(" ", ui.MainLeftIndent)
	expectedHelp := indentLines(formatNavigationHelpText(defaults.SingleSelectHelp, 80), helpIndent)
	assert.Contains(t, cleanView, expectedHelp, "Подсказка должна отображаться в выводе")

	lines := strings.Split(cleanView, "\n")
	expectedLines := strings.Split(expectedHelp, "\n")
	firstHelpLine := expectedLines[0]
	hintLineIndex := -1
	for i, line := range lines {
		if line == firstHelpLine {
			hintLineIndex = i
			break
		}
	}
	assert.GreaterOrEqual(t, hintLineIndex, 0, "Первая строка подсказки должна отображаться в выводе")
	if hintLineIndex > 0 {
		prevLine := strings.TrimSpace(lines[hintLineIndex-1])
		assert.NotEqual(t, "", prevLine, "Строка подсказки не должна отделяться пустой строкой")
	}
}

func TestSingleSelectTaskDividerRecognition(t *testing.T) {
	divider := SingleSelectDividerItem
	items := []Item{
		{Key: "first", Name: "Первый пункт"},
		divider,
		{Key: "second", Name: "Второй пункт"},
	}

	task := NewSingleSelectTask("С разделителем", items)

	assert.True(t, task.isDisabled(1), "Элемент-разделитель должен быть недоступным")
	assert.Equal(t, 0, task.cursor, "Курсор должен указывать на первый доступный пункт")

	moved := task.moveCursorForward()
	assert.True(t, moved, "Курсор должен перейти к следующему доступному пункту")
	assert.Equal(t, 2, task.cursor, "Курсор должен пропустить разделитель")

	assert.Equal(t, -1, task.choiceIndex(divider.Name), "Разделитель не должен разрешать выбор по имени")
}

func TestSingleSelectTaskDividerViewRendering(t *testing.T) {
	divider := SingleSelectDividerItem
	items := []Item{
		{Key: "first", Name: "Первый пункт"},
		divider,
		{Key: "second", Name: "Второй пункт"},
	}

	task := NewSingleSelectTask("С разделителем", items)

	view := stripANSI(task.View(80))
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
		assert.NotContains(t, dividerText, "[", "Разделитель не должен отображаться как выбираемый пункт")
	}
}

func TestSingleSelectTaskDescriptionWrapping(t *testing.T) {
	description := "Очень длинное описание опции, которое должно автоматически переноситься на новую строку для сохранения читабельности списка выбора даже на узких экранах приложения."
	items := []Item{
		{Key: "opt1", Name: "Опция 1", Description: description},
		{Key: "opt2", Name: "Опция 2"},
	}

	task := NewSingleSelectTask("Выберите опцию", items)
	view := stripANSI(task.View(80))

	descIndex := strings.Index(view, "Очень длинное описание")
	require.Greater(t, descIndex, -1, "Описание должно присутствовать в выводе")

	helpIndex := strings.Index(view, defaults.SingleSelectHelp)
	require.Greater(t, helpIndex, -1, "Текст подсказки навигации должен присутствовать в выводе")
	require.Less(t, descIndex, helpIndex, "Описание должно выводиться перед подсказкой навигации")

	descSegment := view[descIndex:helpIndex]
	lines := strings.Split(descSegment, "\n")

	var nonEmpty []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}

	assert.Greater(t, len(nonEmpty), 1, "Описание должно переноситься на несколько строк")
	for _, line := range nonEmpty {
		assert.True(t, strings.HasPrefix(line, "  "), "Каждая строка описания должна содержать базовый отступ")
	}
}
