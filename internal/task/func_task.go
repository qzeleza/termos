package task

import (
	"fmt"
	"strings"
	"time"

	"github.com/qzeleza/ziva/internal/performance"
	"github.com/qzeleza/ziva/internal/ui"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/qzeleza/ziva/internal/defaults"
)

/**
 * @brief Задача, выполняющая функцию и завершающаяся успехом или ошибкой.
 * @details Данный тип задачи позволяет выполнить произвольную функцию и отразить результат выполнения (успех или ошибка).
 * Дополнительно поддерживает вывод информационных строк под заголовком при успешном завершении.
 */
type FuncTask struct {
	BaseTask
	spinner  spinner.Model
	function func() error
	// summaryFunc функция, возвращающая массив строк для отображения
	// под заголовком задачи при успешном завершении
	summaryFunc func() []string
	// summaryLines содержит результат выполнения summaryFunc
	// (заполняется при успешном завершении основной функции)
	summaryLines []string
	err          error
	// successLabel отображается справа от заголовка при успешном завершении.
	// По умолчанию значение равно "ГОТОВО", но может быть переопределено методом WithSuccessLabel.
	successLabel string
}

/**
 * @brief Параметры для создания FuncTask через функциональные опции.
 */
type FuncTaskOption func(*FuncTask)

/**
 * @brief Устанавливает функцию для получения дополнительной информации.
 * @param summaryFunc Функция, возвращающая массив строк для отображения под заголовком.
 * @return Функциональная опция для NewFuncTask.
 */
func WithSummaryFunction(summaryFunc func() []string) FuncTaskOption {
	return func(t *FuncTask) {
		t.summaryFunc = summaryFunc
	}
}

/**
 * @brief Устанавливает флаг остановки очереди при ошибке.
 * @param stop Флаг остановки очереди при ошибке.
 * @return Функциональная опция для NewFuncTask.
 */
func WithStopOnError(stop bool) FuncTaskOption {
	return func(t *FuncTask) {
		t.stopOnError = stop
	}
}

/**
 * @brief Устанавливает текст успешного завершения.
 * @param label Текст для отображения при успешном завершении.
 * @return Функциональная опция для NewFuncTask.
 */
func WithSuccessLabelOption(label string) FuncTaskOption {
	return func(t *FuncTask) {
		if strings.TrimSpace(label) != "" {
			t.successLabel = label
		}
	}
}

/**
 * @brief Создает новую задачу типа FuncTask с функциональными опциями.
 * @param title Заголовок задачи.
 * @param funcAction Функция, которую необходимо выполнить.
 * @param options Функциональные опции для настройки задачи.
 * @return Указатель на созданную задачу FuncTask.
 * @details Рекомендуемый способ создания FuncTask с более читаемым API:
 *
 * task := NewFuncTaskWithOptions("Загрузка данных",
 *     func() error { return loadData() },
 *     WithSummaryFunction(func() []string {
 *         return []string{"Загружено 100 записей", "Обработано за 2.3с"}
 *     }),
 *     WithStopOnError(false),
 *     WithSuccessLabelOption("ЗАВЕРШЕНО"),
 *     WithSummaryIndent(4),
 * )
 */
func NewFuncTask(title string, funcAction func() error, options ...FuncTaskOption) *FuncTask {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.SpinnerStyle

	// Создаем базовую задачу
	baseTask := NewBaseTask(title)

	// Устанавливаем флаг stopOnError в true по умолчанию
	baseTask.SetStopOnError(true)

	// Создаем задачу с базовыми значениями
	task := &FuncTask{
		BaseTask:     baseTask,
		spinner:      s,
		function:     funcAction,
		summaryFunc:  nil,
		summaryLines: nil,
		successLabel: defaults.DefaultSuccessLabel,
	}

	// Применяем функциональные опции
	for _, option := range options {
		option(task)
	}

	return task
}

/**
 * @brief Устанавливает функцию для получения дополнительной информации при успешном завершении.
 * @param summaryFunc Функция, возвращающая массив строк для отображения под заголовком.
 * @return Указатель на задачу для возможности цепочки вызовов.
 * @details Переданные строки будут выведены под заголовком задачи с отступом
 * только при успешном завершении основной функции.
 */
func (t *FuncTask) WithSummary(summaryFunc func() []string) *FuncTask {
	t.summaryFunc = summaryFunc
	return t
}

/**
 * @brief Позволяет переопределить текст, выводимый справа при успешном выполнении задачи.
 * @param label Новый текст для отображения.
 * @return Указатель на задачу для возможности цепочки вызовов.
 * @details Возвращает *FuncTask, чтобы можно было использовать чейнинг:
 * task.NewFuncTask(...).WithSuccessLabel("DONE")
 */
func (t *FuncTask) WithSuccessLabel(label string) *FuncTask {
	if strings.TrimSpace(label) != "" {
		t.successLabel = label
	}
	return t
}

/**
 * @brief Запускает выполнение функции, связанной с задачей.
 * @return Команда для tea.Cmd.
 */
// Определяем специальный тип сообщения для успешного завершения
type funcTaskCompleteMsg struct{}

func (t *FuncTask) Run() tea.Cmd {
	return tea.Batch(t.spinner.Tick, func() tea.Msg {
		// Выполняем функцию и проверяем на ошибку
		err := t.function()
		if err != nil {
			// Устанавливаем ошибку в базовой задаче
			t.BaseTask.err = err
			return err
		}

		// Делаем задержку перед завершением
		// для лучшей визуальной анимации (если она включена)
		if defaults.IsCompletionDelayEnabled() {
			time.Sleep(defaults.DefaultCompletionDelay)
		}

		// Возвращаем специальное сообщение об успешном завершении
		return funcTaskCompleteMsg{}
	})
}

/**
 * @brief Обрабатывает сообщения для задачи типа FuncTask.
 * @param msg Сообщение для обработки.
 * @return Обновленная задача и команда tea.Cmd.
 */
func (t *FuncTask) Update(msg tea.Msg) (Task, tea.Cmd) {
	switch msg := msg.(type) {
	case funcTaskCompleteMsg:
		// Получили сообщение об успешном завершении функции
		t.done = true
		t.icon = ui.IconDone

		// Если определена функция для получения дополнительной информации, вызываем её
		if t.summaryFunc != nil {
			t.summaryLines = t.summaryFunc()
		}

		// Устанавливаем финальное значение для выравнивания по правому краю
		t.finalValue = ui.SuccessLabelStyle.Render(t.successLabel)
		return t, nil
	case error:
		// Получили ошибку от функции, помечаем задачу как завершенную с ошибкой
		t.err = msg
		// Устанавливаем ошибку в базовый тип
		t.BaseTask.err = msg
		t.done = true
		// Добавляем крестик слева и устанавливаем иконку для отображения в FinalView
		t.icon = ui.IconError

		// Сохраняем текст ошибки с применением стиля ErrorMessageStyle
		// Форматирование с отступом и переносами строк будет выполнено в FinalView
		t.finalValue = ui.GetErrorMessageStyle().Render(t.err.Error())
		return t, nil
	case spinner.TickMsg:
		// Если задача завершена, не обновляем спиннер
		if t.done {
			return t, nil
		}
		// Обновляем спиннер
		var cmd tea.Cmd
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd
	case tea.KeyMsg:
		// Обработка нажатия клавиш для возможности выхода из задачи
		switch msg.String() {
		case "q", "Q", "ctrl+c", "Esc", "esc":
			// Помечаем задачу как выполненную с отменой
			t.done = true
			t.icon = ui.IconCancelled
			cancelErr := fmt.Errorf(defaults.ErrorMsgCanceled)
			t.SetError(cancelErr)
			t.finalValue = ui.ErrorMessageStyle.Render(cancelErr.Error())
			return t, nil
		}
	}
	// Продолжаем выполнение для всех остальных сообщений
	return t, nil
}

/**
 * @brief Отображает текущее состояние задачи типа FuncTask.
 * @param width Ширина области отображения.
 * @return Строка с визуализацией состояния задачи.
 */
func (t *FuncTask) View(width int) string {
	if t.IsDone() {
		return t.FinalView(width)
	}
	// Используем новый префикс для активной задачи
	prefix := performance.FastConcat(
		performance.RepeatEfficient(" ", ui.MainLeftIndent),
		ui.TaskInProgressSymbol,
		" ",
	)
	result := fmt.Sprintf("%s%s%s\n", prefix, t.spinner.View(), ui.ActiveTaskStyle.Render(t.title))
	// Добавляем подсказку о навигации с новым отступом
	helpIndent := performance.RepeatEfficient(" ", ui.MainLeftIndent)
	result += "\n" + ui.DrawLine(width) + ui.SubtleStyle.Render(fmt.Sprintf("%s%s", helpIndent, defaults.TaskExitHint))

	return result
}

/**
 * @brief Отображает финальное состояние задачи FuncTask с дополнительной информацией.
 * @param width Ширина области отображения.
 * @return Строка с визуализацией завершенной задачи.
 * @details Переопределяет базовый FinalView для поддержки вывода дополнительных строк
 * при успешном завершении. Дополнительные строки выводятся с настраиваемым отступом.
 */
func (t *FuncTask) FinalView(width int) string {
	// Получаем базовое финальное представление
	result := t.BaseTask.FinalView(width)

	// Если задача завершилась успешно и есть дополнительные строки для вывода
	if t.icon == ui.IconDone && len(t.summaryLines) > 0 {
		result += t.drawSummaryLines(width)
	} else {
		// Если задача завершилась с ошибкой
		// Добавляем перенос строки, если есть дополнительные строки под заголовком
		result += "\n"
	}

	return result
}

// drawSummaryLines рисует дополнительные строки под заголовком задачи
func (t *FuncTask) drawSummaryLines(width int) string {

	// Добавляем верхнюю разделительную линию
	result := "\n"

	// Добавляем каждую строку с настраиваемым отступом
	for _, text_line := range t.summaryLines {
		if strings.TrimSpace(text_line) != "" { // Пропускаем пустые строки
			result += ui.DrawSummaryLine(text_line, width) // Добавляем дополнительные строки с отступом
		}
	}

	// Добавляем нижнюю разделительную линию
	// result += performance.FastConcat(
	// 	performance.RepeatEfficient(" ", ui.MainLeftIndent),
	// 	ui.VerticalLineSymbol,
	// )

	return result
}
