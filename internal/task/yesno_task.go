package task

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/ui"
)

// YesNoOption представляет варианты выбора для YesNoTask
type YesNoOption int

const (
	YesOption YesNoOption = iota
	NoOption
)

// YesNoTask представляет задачу выбора из двух опций: Да, Нет
// Это обертка над SingleSelectTask для консистентности UI
type YesNoTask struct {
	*SingleSelectTask
	question        string
	yesLabel        string
	noLabel         string
	selectedOption  YesNoOption
	showResultLine  bool
	noCountsAsError bool
}

func (t *YesNoTask) syncSelectedOption() {
	idx := t.GetSelectedIndex()
	switch idx {
	case 0:
		t.selectedOption = YesOption
		t.SetError(nil)
	case 1:
		t.selectedOption = NoOption
		if t.noCountsAsError {
			msg := fmt.Errorf("%s \"%s\"", defaults.DefaultSelectedLabel, defaults.DefaultNo)
			t.SetError(msg)
			t.SetStopOnError(false)
		} else {
			t.SetError(nil)
		}
	default:
		return
	}
}

// NewYesNoTask создает новую задачу выбора с двумя опциями
func NewYesNoTask(title, question string) *YesNoTask {
	// Создаем варианты выбора
	options := []Item{
		{Key: defaults.DefaultYes, Name: defaults.DefaultYes},
		{Key: defaults.DefaultNo, Name: defaults.DefaultNo},
	}

	// Создаем базовую задачу выбора
	selectTask := NewSingleSelectTask(title, options)

	return &YesNoTask{
		SingleSelectTask: selectTask,
		question:         question,
		yesLabel:         defaults.DefaultYes,
		noLabel:          defaults.DefaultNo,
		selectedOption:   YesOption,
		showResultLine:   true,
		noCountsAsError:  true,
	}
}

// stopTimeout останавливает таймер
func (t *YesNoTask) stopTimeout() {
	// Если таймер активен, останавливаем его
	if t.timeoutEnabled && t.timeoutManager != nil && t.timeoutManager.IsActive() {
		t.timeoutManager.StopTimeout()
		t.showTimeout = false
	}
}

// Update обновляет состояние задачи, делегируя логику базовому SingleSelectTask
func (t *YesNoTask) Update(msg tea.Msg) (Task, tea.Cmd) {
	if t.IsDone() {
		return t, nil
	}

	// Рассматриваем специальные случаи для YesNoTask
	switch msg := msg.(type) {
	case TimeoutMsg:
		// Когда истекает тайм-аут, применяем значение по умолчанию
		// Обрабатываем локально для надежности (как в SingleSelectTask)
		// Значение по умолчанию уже нормализовано в WithTimeout
		t.applyDefaultValue()
		return t, nil
	case TickMsg:
		// Если таймер активен, продолжаем обновления
		if t.timeoutEnabled && t.timeoutManager != nil && t.timeoutManager.IsActive() {
			return t, t.timeoutManager.StartTicker()
		}
		return t, nil
	case tea.KeyMsg:
		// Обрабатываем нажатие упрвляющих клавиш для отключения таймера
		switch msg.String() {
		case " ", "up", "down", "j", "k", "enter":
			t.stopTimeout()
		}
	}

	// Делегируем обработку базовому SingleSelectTask
	updatedTask, cmd := t.SingleSelectTask.Update(msg)
	t.SingleSelectTask = updatedTask.(*SingleSelectTask)

	// Если задача завершена, определяем выбранную опцию
	if t.IsDone() {
		t.syncSelectedOption()
	}

	return t, cmd
}

// Run запускает задачу выбора
func (t *YesNoTask) Run() tea.Cmd {
	// Запускаем таймер и тикер, если они включены
	if t.timeoutEnabled && t.timeoutManager != nil {
		return t.timeoutManager.StartTickerAndTimeout()
	}
	return nil
}

// View отображает текущее состояние задачи, делегируя логику базовому SingleSelectTask
// applyDefaultValue применяет значение по умолчанию при истечении таймера
func (t *YesNoTask) applyDefaultValue() {
	t.SingleSelectTask.applyDefaultValue()
	if t.IsDone() {
		t.syncSelectedOption()
	}
}

func (t *YesNoTask) View(width int) string {
	if t.IsDone() {
		return t.FinalView(width)
	}

	// Делегируем отображение базовому SingleSelectTask
	return t.SingleSelectTask.View(width)
}

// FinalView переопределяет отображение завершённой задачи.
// Если пользователь выбрал "Нет", с правой стороны выводится слово "ОТКАЗ"
// ярко-жёлтым цветом. Для ответа "Да" используется стандартное зелёное
// оформление выбранной опции.
func (t *YesNoTask) FinalView(width int) string {
	if t.icon == ui.IconError {
		return t.SingleSelectTask.FinalView(width)
	}

	// Для отмененных задач используем базовое представление, чтобы показать статус "ОТМЕНА"
	if t.icon == ui.IconCancelled {
		return t.BaseTask.FinalView(width)
	}

	// Используем новый префикс завершённой задачи
	// Успешной считается выбор без ошибок и ("Да" или "Нет", если она не считается ошибкой)
	success := !t.HasError() && (t.selectedOption == YesOption || !t.noCountsAsError)
	prefix := t.CompletedPrefix()
	if prefix == "" {
		prefix = ui.GetCompletedTaskPrefix(success)
	}

	titleStyle := lipgloss.NewStyle()
	if !success {
		titleStyle = ui.GetErrorStatusStyle()
	}

	// Сформируем правую часть строки
	var right string
	switch t.selectedOption {
	case YesOption:
		right = ui.TaskStatusSuccessStyle.Render(defaults.DefaultYesLabel)
	case NoOption:
		if t.noCountsAsError {
			// Для "Нет" выводим слово ОТКАЗ стилем ошибки
			right = ui.GetErrorStatusStyle().Render(defaults.DefaultNoLabel)
		} else {
			right = ui.TaskStatusSuccessStyle.Render(defaults.DefaultNoLabel)
		}
	}

	header := renderTitleWithWrap(prefix, t.title, titleStyle, right, width)

	var builder strings.Builder
	builder.WriteString(header)
	builder.WriteString("\n")
	// Если задача завершилась успешно и есть дополнительные строки для вывода
	if t.showResultLine && t.icon == ui.IconDone && t.cursor >= 0 && t.cursor < len(t.items) {
		builder.WriteString(ui.DrawSummaryLine(t.items[t.cursor].displayName(), width))
	}

	return builder.String()
}

// WithResultLine управляет отображением итоговой строки результата (совместимость)
func (t *YesNoTask) WithResultLine(show bool) *YesNoTask {
	t.showResultLine = show
	return t
}

/**
 * WithNoAsError включает интерпретацию ответа "Нет" как успешного результата.
 *
 * @return Указатель на задачу для цепочки вызовов.
 */
func (t *YesNoTask) WithNoAsError() *YesNoTask {
	t.noCountsAsError = false
	if t.selectedOption == NoOption && t.HasError() {
		t.SetError(nil)
	}
	return t
}

/**
 * ResultLineVisible сообщает, отображается ли итоговая строка результата.
 *
 * @return true, если строка результата включена.
 */
func (t *YesNoTask) ResultLineVisible() bool {
	return t.showResultLine
}

// WithCustomLabels позволяет изменить текст опций (перегрузка для 2 параметров)
// Возвращает *YesNoTask для возможности цепочки вызовов
func (t *YesNoTask) WithCustomLabels(yesLabel, noLabel string) *YesNoTask {
	if trimmed := strings.TrimSpace(yesLabel); trimmed != "" {
		t.yesLabel = trimmed
		if len(t.items) > 0 {
			t.items[0].name = trimmed
		}
	}
	if trimmed := strings.TrimSpace(noLabel); trimmed != "" {
		t.noLabel = trimmed
		if len(t.items) > 1 {
			t.items[1].name = trimmed
		}
	}
	return t
}

// WithDefaultItem позволяет задать опцию, которая будет выбрана по умолчанию при открытии задачи.
// Поддерживает YesNoOption, bool, индекс (int) и строку (string).
func (t *YesNoTask) WithDefaultItem(option interface{}) *YesNoTask {
	if option == nil {
		return t
	}

	selectByIndex := func(index int) {
		t.SingleSelectTask.WithDefaultItem(index)
		switch index {
		case 0:
			t.selectedOption = YesOption
		case 1:
			t.selectedOption = NoOption
		}
	}

	switch v := option.(type) {
	case YesNoOption:
		switch v {
		case YesOption:
			selectByIndex(0)
		case NoOption:
			selectByIndex(1)
		}
	case bool:
		if v {
			selectByIndex(0)
		} else {
			selectByIndex(1)
		}
	case int:
		selectByIndex(v)
	case string:
		if idx := t.SingleSelectTask.choiceIndex(v); idx != -1 {
			selectByIndex(idx)
		}
	}

	return t
}

// GetValue возвращает ответ пользователя (true для "Да", false для "Нет")
// @deprecated Рекомендуется использовать GetSelectedOption() для более ясной семантики
func (t *YesNoTask) GetValue() bool {
	switch t.selectedOption {
	case YesOption:
		return true
	case NoOption:
		return false
	default:
		return false
	}
}

// GetSelectedOption возвращает выбранную опцию
func (t *YesNoTask) GetSelectedOption() YesNoOption {
	return t.selectedOption
}

// IsYes возвращает true если выбрано "Да"
func (t *YesNoTask) IsYes() bool {
	return t.selectedOption == YesOption
}

// IsNo возвращает true если выбрано "Нет"
func (t *YesNoTask) IsNo() bool {
	return t.selectedOption == NoOption
}

// SetError устанавливает ошибку для задачи
func (t *YesNoTask) SetError(err error) {
	t.SingleSelectTask.SetError(err)
}

// HasError возвращает true, если при выполнении задачи произошла ошибка
func (t *YesNoTask) HasError() bool {
	return t.SingleSelectTask.HasError()
}

// Error возвращает ошибку, если она есть
func (t *YesNoTask) Error() error {
	return t.SingleSelectTask.Error()
}

// WithTimeout устанавливает тайм-аут для задачи Да/Нет
// @param duration Длительность тайм-аута
// @param defaultValue Значение по умолчанию ("Да", "Нет" или индекс 0/1)
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithTimeout(duration time.Duration, defaultValue interface{}) *YesNoTask {
	// Нормализуем тип значения по умолчанию до ожидаемых BaseTask/SingleSelectTask типов (int или string)
	var normalized interface{} = nil
	if defaultValue != nil {
		switch v := defaultValue.(type) {
		case YesNoOption:
			if v == YesOption {
				normalized = 0 // индекс "Да"
			} else {
				normalized = 1 // индекс "Нет"
			}
		case bool:
			if v {
				normalized = 0 // Да
			} else {
				normalized = 1 // Нет
			}
		case int:
			normalized = v
		case string:
			// Нормализуем строковое значение к индексу для языко-независимого сравнения
			normalized = t.normalizeStringToIndex(v)
		default:
			// Неизвестный тип — не задаем значение по умолчанию
			normalized = nil
		}
	}

	// Проксируем в SingleSelectTask с нормализованным значением
	t.SingleSelectTask.WithTimeout(duration, normalized)
	return t
}

// normalizeStringToIndex конвертирует строковое значение по умолчанию в соответствующий индекс
// Поддерживает языко-независимое сравнение для "Да"/"Нет" на разных языках
func (t *YesNoTask) normalizeStringToIndex(value string) interface{} {
	// Список известных вариантов для "Да" на разных языках
	yesVariants := []string{"да", "yes", "evet", "так", "так"}
	// Список известных вариантов для "Нет" на разных языках
	noVariants := []string{"нет", "no", "hayır", "не", "ні"}

	valueLower := strings.ToLower(strings.TrimSpace(value))

	// Проверяем среди вариантов "Да"
	for _, variant := range yesVariants {
		if strings.EqualFold(valueLower, variant) {
			return 0 // индекс "Да"
		}
	}

	// Проверяем среди вариантов "Нет"
	for _, variant := range noVariants {
		if strings.EqualFold(valueLower, variant) {
			return 1 // индекс "Нет"
		}
	}

	if idx := t.SingleSelectTask.choiceIndex(value); idx != -1 {
		return idx
	}

	return value
}

// WithTimeoutYes устанавливает тайм-аут с выбором "Да" по умолчанию (языко-независимый)
// @param duration Длительность тайм-аута
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithTimeoutYes(duration time.Duration) *YesNoTask {
	t.WithTimeout(duration, YesOption)
	return t
}

// WithTimeoutNo устанавливает тайм-аут с выбором "Нет" по умолчанию (языко-независимый)
// @param duration Длительность тайм-аута
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithTimeoutNo(duration time.Duration) *YesNoTask {
	t.WithTimeout(duration, NoOption)
	return t
}

// WithDefaultYes устанавливает "Да" как вариант по умолчанию (языко-независимый)
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithDefaultYes() *YesNoTask {
	t.WithDefaultItem(YesOption)
	return t
}

// WithDefaultNo устанавливает "Нет" как вариант по умолчанию (языко-независимый)
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithDefaultNo() *YesNoTask {
	t.WithDefaultItem(NoOption)
	return t
}

// WithDefaultOption устанавливает вариант по умолчанию и тайм-аут для задачи
// @param defauiltOption Может быть индексом (0 - Да, 1 - Нет) или строкой ("Да", "Нет")
// @param timeout Время ожидания в секундах до автовыбора
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithDefaultOption(defauiltOption interface{}, timeout time.Duration) *YesNoTask {
	// Используем метод WithTimeout для единообразия
	t.WithTimeout(timeout, defauiltOption)
	return t
}
