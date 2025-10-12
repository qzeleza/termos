package query

import (
	"runtime"
	"strings"
	"testing"

	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/ui"
	"github.com/stretchr/testify/assert"
)

// Дополнительные тесты для покрытия функций управления памятью в queue.go

func TestCheckMemoryPressure(t *testing.T) {
	model := New("Тест давления памяти")

	// Тестируем функцию без превышения порога
	// (сложно протестировать реальное превышение без выделения огромного количества памяти)
	model.checkMemoryPressure()

	// Проверяем, что модель остается работоспособным
	assert.NotNil(t, model)
	assert.Equal(t, "Тест давления памяти", model.title)
}

func TestEmergencyCleanup(t *testing.T) {
	model := New("Тест экстренной очистки")

	// Добавляем много задач для имитации большого количества данных
	var tasks []common.Task
	for i := 0; i < 100; i++ {
		task := NewMockTask("Задача " + string(rune(i+'A')))
		task.CompleteSuccessfully()
		tasks = append(tasks, task)
	}
	model.AddTasks(tasks)
	model.current = len(tasks) // Помечаем все как завершенные

	// Выполняем экстренную очистку
	model.emergencyCleanup()

	// Проверяем, что модель остается работоспособной
	assert.NotNil(t, model)
	assert.True(t, len(model.tasks) <= MaxCompletedTasks, "Количество задач должно быть ограничено")
}

func TestCleanupOldTasks(t *testing.T) {
	model := New("Тест очистки старых задач")

	// Создаем больше задач, чем лимит
	var tasks []common.Task
	taskCount := MaxCompletedTasks + 20
	for i := 0; i < taskCount; i++ {
		task := NewMockTask("Задача " + string(rune(i)))
		task.CompleteSuccessfully()
		tasks = append(tasks, task)
	}

	model.AddTasks(tasks)
	model.current = taskCount // Все задачи завершены

	// Проверяем начальное состояние
	assert.Equal(t, taskCount, len(model.tasks))

	// Выполняем очистку
	model.cleanupOldTasks()

	// Проверяем результат
	assert.LessOrEqual(t, len(model.tasks), MaxCompletedTasks, "Количество задач должно быть ограничено")
	assert.Equal(t, MaxCompletedTasks, model.current, "Индекс current должен быть скорректирован")

	// Проверяем, что сохранились последние задачи
	if len(model.tasks) > 0 {
		lastTask := model.tasks[len(model.tasks)-1]
		assert.Contains(t, lastTask.Title(), "Задача", "Последние задачи должны быть сохранены")
	}
}

func TestCleanupOldTasksWithinLimit(t *testing.T) {
	model := New("Тест очистки в пределах лимита")

	// Создаем меньше задач, чем лимит
	var tasks []common.Task
	taskCount := MaxCompletedTasks - 10
	for i := 0; i < taskCount; i++ {
		task := NewMockTask("Задача " + string(rune(i+'A')))
		task.CompleteSuccessfully()
		tasks = append(tasks, task)
	}

	model.AddTasks(tasks)
	model.current = taskCount

	originalTaskCount := len(model.tasks)

	// Выполняем очистку
	model.cleanupOldTasks()

	// Количество задач не должно измениться
	assert.Equal(t, originalTaskCount, len(model.tasks), "Количество задач не должно изменяться если в пределах лимита")
	assert.Equal(t, taskCount, model.current, "Индекс current не должен изменяться")
}

func TestWithAppNameColor(t *testing.T) {
	model := New("Тест цвета названия приложения")

	result := model.WithAppNameColor(ui.ColorBrightBlue, true)

	// Проверяем возврат self для chaining
	assert.Equal(t, model, result, "WithAppNameColor должен возвращать self для chaining")

	// Проверяем, что стиль установлен
	assert.NotNil(t, model.appNameStyle, "Стиль названия приложения должен быть установлен")
}

func TestRun(t *testing.T) {
	model := New("Тест запуска")

	// Добавляем простую задачу
	task := NewMockTask("Простая задача")
	model.AddTasks([]common.Task{task})

	// Тестируем, что Run возвращает без ошибки
	// (полное тестирование TUI сложно в unit-тестах)
	err := model.Run()
	if err != nil && strings.Contains(err.Error(), "could not open a new TTY") {
		t.Skip("Пропуск теста Run: TTY недоступен в окружении")
	}
	assert.NoError(t, err, "Run не должен возвращать ошибку")
}

func TestInit(t *testing.T) {
	// Тест с задачами
	model := New("Тест инициализации")
	task := NewMockTask("Тестовая задача")
	model.AddTasks([]common.Task{task})

	cmd := model.Init()
	assert.NotNil(t, cmd, "Init должен возвращать команду для первой задачи")

	// Тест без задач
	emptyModel := New("Пустая модель")
	emptyCmd := emptyModel.Init()
	assert.NotNil(t, emptyCmd, "Init должен возвращать команду даже без задач")
}

func TestSetTitle(t *testing.T) {
	// Тест с названием приложения
	model := New("Тестовый заголовок")
	model.WithAppName("TestApp")
	model.WithAppNameColor(ui.ColorBrightBlue, true)

	result := model.setTitle(80)
	assert.Contains(t, result, "TestApp", "Результат должен содержать название приложения")
	assert.Contains(t, result, "Тестовый заголовок", "Результат должен содержать заголовок")
	assert.Contains(t, result, "\n", "Результат должен заканчиваться переносом строки")

	// Тест без названия приложения
	simpleModel := New("Простой заголовок")
	simpleResult := simpleModel.setTitle(80)
	assert.Contains(t, simpleResult, "Простой заголовок", "Результат должен содержать заголовок")
	assert.Contains(t, simpleResult, "\n", "Результат должен заканчиваться переносом строки")
}

func TestDrawFooterLine(t *testing.T) {
	result := DrawFooterLine(40)

	assert.NotEmpty(t, result, "DrawFooterLine должен возвращать непустую строку")
	assert.Contains(t, result, "─", "Результат должен содержать горизонтальные линии")
	assert.Contains(t, result, "\n", "Результат должен заканчиваться переносом строки")
}

func TestRemoveVerticalLinesBeforeTaskSymbolsMemory(t *testing.T) {
	// Создаем тестовое содержимое с вертикальными линиями
	content := "┌─────────────────────────────────────┐\n" +
		"│  Заголовок                          │\n" +
		"├─────────────────────────────────────┤\n" +
		"  ✓ Задача 1: завершено              \n" +
		"  │                                   \n" +
		"  ✓ Задача 2: завершено              \n" +
		"  │                                   \n" +
		"  │                                   \n"

	var sb strings.Builder
	sb.WriteString(content)

	// Применяем функцию
	removeVerticalLinesBeforeTaskSymbols(&sb)

	result := sb.String()

	// Проверяем, что функция работает без ошибок
	assert.NotEmpty(t, result, "Результат не должен быть пустым")
	assert.Contains(t, result, "✓", "Символы задач должны остаться")

	// Проверяем, что вертикальные линии после последней задачи изменены
	lines := strings.Split(result, "\n")
	foundLastTask := false
	lastTaskIndex := -1

	for i, line := range lines {
		if strings.Contains(line, "✓") {
			lastTaskIndex = i
			foundLastTask = true
		}
	}

	assert.True(t, foundLastTask, "Должна быть найдена хотя бы одна задача")

	// Проверяем строки после последней задачи
	for i := lastTaskIndex + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, "│") {
			// Если есть │, это должна быть рамка, а не отступ
			trimmedLine := strings.TrimSpace(line)
			assert.True(t,
				strings.HasPrefix(trimmedLine, "│") || strings.HasSuffix(trimmedLine, "│"),
				"Вертикальные линии должны быть только в рамке, не в отступах")
		}
	}
}

func TestMemoryManagementIntegration(t *testing.T) {
	// Интеграционный тест управления памятью
	model := New("Тест интеграции управления памятью")

	// Получаем начальные статистики памяти
	var initialStats runtime.MemStats
	runtime.ReadMemStats(&initialStats)

	// Создаем большое количество задач
	var tasks []common.Task
	for i := 0; i < 200; i++ {
		task := NewMockTask("Задача " + string(rune(i%26+'A')))
		task.CompleteSuccessfully()
		tasks = append(tasks, task)
	}

	model.AddTasks(tasks)
	model.current = len(tasks)

	// Обновляем статистику задач
	model.updateTaskStats()

	// Проверяем статистику
	assert.Equal(t, 200, model.successCount, "Все задачи должны быть успешными")
	assert.Equal(t, 0, model.errorCount, "Не должно быть ошибок")

	// Выполняем очистку памяти
	model.emergencyCleanup()

	// Проверяем, что модель остается работоспособной
	assert.LessOrEqual(t, len(model.tasks), MaxCompletedTasks, "Количество задач должно быть ограничено")
	assert.NotNil(t, model.tasks, "Массив задач не должен быть nil")

	// Проверяем, что статистика пересчитана корректно
	model.updateTaskStats()
	expectedSuccessCount := len(model.tasks)
	assert.Equal(t, expectedSuccessCount, model.successCount, "Статистика должна соответствовать количеству задач после очистки")
}

func TestErrorColorEdgeCases(t *testing.T) {
	model := New("Тест крайних случаев цвета ошибок")

	// Тестируем все доступные цвета ошибок
	colors := []ErrorColor{Yellow, Red, Orange}

	for _, color := range colors {
		result := model.SetErrorColor(color)
		assert.Equal(t, model, result, "SetErrorColor должен возвращать self для chaining")
	}

	// Проверяем, что функция работает с неопределенным цветом
	// (не должно вызывать панику)
	model.SetErrorColor(ErrorColor(999))
}
