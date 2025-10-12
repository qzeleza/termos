// Package ziva предоставляет библиотеку для создания интерактивных TUI приложений.
// Эта библиотека позволяет создавать разнообразные задачи пользовательского интерфейса такие как:
// - Задачи выбора Да/Нет
// - Задачи одиночного выбора из списка
// - Задачи множественного выбора
// - Задачи ввода текста с валидацией
// - Задачи выполнения функций
// - Очереди задач с автоматическим выполнением
//
// Пример использования:
//
//	queue := ziva.NewQueue("Пример использования Ziva")
//
//	// Создаем задачу выбора Да/Нет
//	confirm := ziva.NewYesNoTask("Подтверждение", "Продолжить выполнение?")
//
//	// Создаем задачу выбора из списка
//	options := []string{"development", "staging", "production"}
//	env := ziva.NewSingleSelectTask("Выбор среды", options)
//
//	// Добавляем задачи в очередь
//	queue.AddTasks(confirm, env)
//
//	// Запускаем очередь
//	err := queue.Run()
//	if err != nil {
//		log.Fatal(err)
//	}
package ziva

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/qzeleza/ziva/internal/autoconfig"
	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/performance"
	"github.com/qzeleza/ziva/internal/query"
	"github.com/qzeleza/ziva/internal/task"
	"github.com/qzeleza/ziva/internal/ui"
	"github.com/qzeleza/ziva/internal/validation"
)

// Task представляет собой интерфейс для выполнения задач в очереди.
// Этот интерфейс используется как в пакете task, так и в пакете query.
type Task = common.Task

// YesNoOption представляет варианты выбора для YesNoTask
type YesNoOption = task.YesNoOption

// Item описывает элемент списка для задач выбора.
type Item = task.Item

var (
	// SingleSelectDivider предоставляет стандартный разделитель для списков одиночного выбора.
	SingleSelectDivider = task.SingleSelectDividerItem
	// Divider — псевдоним для стандартного разделителя, чтобы использовать короткое имя.
	Divider = SingleSelectDivider
)

// MultiSelectDependencyActions описывает действия, выполняемые при изменении состояния пунктов мультивыбора.
type MultiSelectDependencyActions = task.MultiSelectDependencyActions

// MultiSelectDependencyRule определяет, что должно произойти при выборе и снятии отметки с пункта мультивыбора.
type MultiSelectDependencyRule = task.MultiSelectDependencyRule

const (
	// YesOption - опция "Да"
	YesOption = task.YesOption
	// NoOption - опция "Нет"
	NoOption = task.NoOption
)

// Queue представляет очередь задач для выполнения
type Queue struct {
	model *query.Model
}

// DisableCompletionDelay полностью отключает задержку завершения задач (по умолчанию включена).
func DisableCompletionDelay() {
	defaults.SetCompletionDelayEnabled(false)
}

// SetExitMenuItemStyle обновляет стиль отображения пунктов меню выхода.
// Передайте ziva.MenuActionDefaultStyle() для стандартной подсветки
// или ziva.MenuItemDefaultStyle() для оформления как обычных пунктов.
func SetExitMenuItemStyle(style lipgloss.Style) {
	ui.SetMenuExitItemStyle(style)
}

// SetBackMenuItemStyle обновляет стиль отображения пунктов меню возврата.
// Передайте ziva.MenuActionDefaultStyle() для стандартной подсветки
// или ziva.MenuItemDefaultStyle() для оформления как обычных пунктов.
func SetBackMenuItemStyle(style lipgloss.Style) {
	ui.SetMenuBackItemStyle(style)
}

// ResetExitMenuItemStyle возвращает стиль пункта выхода к подсвеченному значению по умолчанию.
func ResetExitMenuItemStyle() {
	ui.ResetMenuExitItemStyle()
}

// ResetBackMenuItemStyle возвращает стиль пункта "Назад" к подсвеченному значению по умолчанию.
func ResetBackMenuItemStyle() {
	ui.ResetMenuBackItemStyle()
}

// MenuActionDefaultStyle возвращает стиль подсветки специальных пунктов меню (яркий циан, жирный).
func MenuActionDefaultStyle() lipgloss.Style {
	return ui.MenuActionDefaultStyle()
}

// MenuItemDefaultStyle возвращает стиль, совпадающий с обычными пунктами меню (белый текст).
func MenuItemDefaultStyle() lipgloss.Style {
	return ui.MenuItemPlainStyle()
}

// ----------------------------------------------------------------------------
// Queue
// ----------------------------------------------------------------------------

// NewQueue создает новую очередь задач с заданным заголовком
//
// @param title Заголовок очереди
// @return Указатель на новую очередь задач
func NewQueue(title string) *Queue {
	return &Queue{
		model: query.New(title),
	}
}

// AddTasks добавляет задачи в очередь
//
// @param tasks Задачи, которые будут добавлены в очередь
// @return Указатель на очередь задач
func (q *Queue) AddTasks(tasks ...Task) *Queue {
	q.model.AddTasks(tasks)
	return q
}

// WithAppName устанавливает название приложения в заголовке
//
// @param appName Название приложения
// @param version Необязательная строка версии приложения
// @return Указатель на очередь задач
func (q *Queue) WithAppName(appName string, version ...string) *Queue {
	q.model.WithAppName(appName, version...)
	return q
}

// WithSummary включает/выключает отображение сводки по завершению
//
// @param show Флаг, указывающий, нужно ли отображать сводку
// @param lineEnabled Флаг, указывающий, нужно ли отображать линию сводки, под которой выводится сводка
// @return Указатель на очередь задач
func (q *Queue) WithOutSummary() *Queue {
	q.model.WithSummary(false)

	return q
}

// WithResultFormatting включает форматирование результатов задач с разделительными линиями.
// Если enabled=true, то перед каждым результатом задачи будет добавляться разделительная линия
// из префикса и указанного количества символов "─".
// При enabled=false поведение остается как и раньше - результаты выводятся сразу после строки с задачей.
// @param enabled - включить/выключить форматирование
func (q *Queue) WithOutResultLine() *Queue {
	q.model.WithResultFormatting(false)
	return q
}

// WithTasksNumbered включает отображение номеров для задач и задаёт формат (например "[%02d]" или "(%d)").
//
// @param keepFirstSymbol Флаг, указывающий, нужно ли сохранять первый символ формата
// @param numberFormat Формат номеров задач
// @return Указатель на очередь задач
func (q *Queue) WithTasksNumbered(keepFirstSymbol bool, numberFormat string) *Queue {
	q.model.WithTasksNumbered(true, keepFirstSymbol, numberFormat)
	return q
}

// WithAppNameColor устанавливает цвет текста и стиль названия приложения.
//
// @param textColor Цвет текста
// @param bold Флаг, указывающий, нужно ли сделать текст жирным
// @return Указатель на очередь задач
func (q *Queue) WithAppNameColor(textColor lipgloss.TerminalColor, bold bool) *Queue {
	q.model.WithAppNameColor(textColor, bold)
	return q
}

// WithTitleColor устанавливает цвет заголовка.
//
// @param titleColor Цвет заголовка
// @param bold Флаг, указывающий, нужно ли сделать заголовок жирным
// @return Указатель на очередь задач
func (q *Queue) WithTitleColor(titleColor lipgloss.TerminalColor, bold bool) *Queue {
	q.model.WithTitleColor(titleColor, bold)
	return q
}

// WithClearScreen включает/выключает очистку экрана перед запуском очереди задач
//
// @param clear Флаг, указывающий, нужно ли очищать экран перед запуском очереди задач
// @param hardClear Флаг, указывающий, нужно ли использовать твердую очистку экрана (true) или нет (false)
// @return Указатель на очередь задач
func (q *Queue) WithClearScreen(clear bool, hardClear ...bool) *Queue {
	q.model.WithClearScreen(clear, hardClear...)
	return q
}

// SetErrorColor устанавливает цвет для отображения ошибок в очереди.
//
// @param color Цвет для отображения ошибок
// @return Указатель на очередь задач
func (q *Queue) SetErrorColor(color query.ErrorColor) *Queue {
	q.model.SetErrorColor(color)
	return q
}

// Run запускает выполнение очереди задач
//
// @return Ошибка, если она возникла
func (q *Queue) Run() error {
	return q.model.Run()
}

// ----------------------------------------------------------------------------
// YesNoTask
// ----------------------------------------------------------------------------

// NewYesNoTask создает новую задачу выбора Да/Нет
//
// @param title Заголовок задачи
// @param question Вопрос, который будет задан пользователю
// @return Указатель на новую задачу выбора Да/Нет
func NewYesNoTask(title, question string) *YesNoTask {
	return &YesNoTask{task.NewYesNoTask(title, question)}
}

// YesNoTask представляет задачу выбора из двух опций: Да, Нет
type YesNoTask struct {
	*task.YesNoTask
}

// WithTimeoutYes устанавливает тайм-аут с выбором "Да" по умолчанию (языко-независимый)
//
// @param duration Длительность тайм-аута
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithTimeoutYes(duration time.Duration) *YesNoTask {
	t.YesNoTask.WithTimeoutYes(duration)
	return t
}

// WithTimeoutNo устанавливает тайм-аут с выбором "Нет" по умолчанию (языко-независимый)
//
// @param duration Длительность тайм-аута
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithTimeoutNo(duration time.Duration) *YesNoTask {
	t.YesNoTask.WithTimeoutNo(duration)
	return t
}

// WithDefaultYes устанавливает "Да" как вариант по умолчанию (языко-независимый)
//
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithDefaultYes() *YesNoTask {
	t.YesNoTask.WithDefaultYes()
	return t
}

// WithDefaultNo устанавливает "Нет" как вариант по умолчанию (языко-независимый)
//
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithDefaultNo() *YesNoTask {
	t.YesNoTask.WithDefaultNo()
	return t
}

// WithDefaultYesAndTimeout комбинированный метод: устанавливает "Да" по умолчанию и тайм-аут
//
// @param duration Длительность тайм-аута
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithDefaultYesAndTimeout(duration time.Duration) *YesNoTask {
	t.YesNoTask.WithDefaultYes()
	t.YesNoTask.WithTimeoutYes(duration)
	return t
}

// WithDefaultNoAndTimeout комбинированный метод: устанавливает "Нет" по умолчанию и тайм-аут
//
// @param duration Длительность тайм-аута
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithDefaultNoAndTimeout(duration time.Duration) *YesNoTask {
	t.YesNoTask.WithDefaultNo()
	t.YesNoTask.WithTimeoutNo(duration)
	return t
}

// WithCustomLabels позволяет изменить текст опций
//
// @param yesLabel Текст для опции "Да"
// @param noLabel Текст для опции "Нет"
// @return Указатель на задачу для цепочки вызовов
func (t *YesNoTask) WithCustomLabels(yesLabel, noLabel string) *YesNoTask {
	t.YesNoTask.WithCustomLabels(yesLabel, noLabel)
	return t
}

// GetSelectedOption возвращает выбранную опцию
//
// @return Выбранная опция
func (t *YesNoTask) GetSelectedOption() YesNoOption {
	return t.YesNoTask.GetSelectedOption()
}

// IsYes возвращает true если выбрано "Да"
//
// @return true если выбрано "Да"
func (t *YesNoTask) IsYes() bool {
	return t.YesNoTask.IsYes()
}

// IsNo возвращает true если выбрано "Нет"
//
// @return true если выбрано "Нет"
func (t *YesNoTask) IsNo() bool {
	return t.YesNoTask.IsNo()
}

/**
 * WithResultLine управляет отображением строки итогового результата.
 *
 * @param show Флаг, указывающий, нужно ли показывать строку результата.
 * @return Указатель на задачу для цепочки вызовов.
 */
func (t *YesNoTask) WithoutResultLine() *YesNoTask {
	t.YesNoTask.WithResultLine(false)
	return t
}

/**
 * WithNoAsError включает интерпретацию ответа "Нет" как успешного результата.
 *
 * @return Указатель на задачу для цепочки вызовов.
 */
func (t *YesNoTask) WithNoAsError() *YesNoTask {
	t.YesNoTask.WithNoAsError()
	return t
}

// ----------------------------------------------------------------------------
// SingleSelectTask
// ----------------------------------------------------------------------------

// NewSingleSelectTask создает новую задачу выбора одного варианта из списка
//
// @param title Заголовок задачи
// @param items Список элементов списка
// @return Указатель на новую задачу выбора
func NewSingleSelectTask(title string, items []Item) *SingleSelectTask {
	return &SingleSelectTask{task.NewSingleSelectTask(title, items)}
}

// SingleSelectTask представляет задачу для выбора одного варианта из списка
type SingleSelectTask struct {
	*task.SingleSelectTask
}

// WithViewport устанавливает размер viewport (окна просмотра) для ограничения количества отображаемых элементов.
// Это полезно для длинных списков, когда нужно показывать только часть элементов.
//
// @param size Количество элементов для отображения одновременно (0 = показать все)
// @param showCounters Флаг, указывающий, нужно ли отображать счетчики элементов
// @return Указатель на задачу для цепочки вызовов
func (t *SingleSelectTask) WithViewport(size int, showCounters ...bool) *SingleSelectTask {
	t.SingleSelectTask.WithViewport(size, showCounters...)
	return t
}

// WithTimeout устанавливает тайм-аут для задачи с значением по умолчанию
//
// @param duration Тайм-аут в миллисекундах
// @param defaultValue Значение по умолчанию
// @return Указатель на задачу для цепочки вызовов
func (t *SingleSelectTask) WithTimeout(duration time.Duration, defaultValue interface{}) *SingleSelectTask {
	t.SingleSelectTask.WithTimeout(duration, defaultValue)
	return t
}

// WithItemsDisabled помечает элементы меню как недоступные для выбора.
// Поддерживаются типы: int, []int, string, []string. Nil очищает список отключённых элементов.
//
// @param disabled Список индексов или значений, которые нужно отключить
// @return Указатель на задачу для цепочки вызовов
func (t *SingleSelectTask) WithItemsDisabled(disabled interface{}) *SingleSelectTask {
	t.SingleSelectTask.WithItemsDisabled(disabled)
	return t
}

// WithDefaultItem задает элемент, который будет подсвечен при открытии задачи.
//
// @param selection Индекс или значение элемента, который будет подсвечен
// @return Указатель на задачу для цепочки вызовов
func (t *SingleSelectTask) WithDefaultItem(selection interface{}) *SingleSelectTask {
	t.SingleSelectTask.WithDefaultItem(selection)
	return t
}

// GetSelected возвращает выбранное значение
//
// @return Выбранное значение
func (t *SingleSelectTask) GetSelected() string {
	return t.SingleSelectTask.GetSelected()
}

// GetSelectedIndex возвращает индекс выбранного элемента
//
// @return Индекс выбранного элемента
func (t *SingleSelectTask) GetSelectedIndex() int {
	return t.SingleSelectTask.GetSelectedIndex()
}

// ----------------------------------------------------------------------------
// MultiSelectTask
// ----------------------------------------------------------------------------

// NewMultiSelectTask создает новую задачу множественного выбора
//
// @param title Заголовок задачи
// @param items Список элементов списка
// @return Указатель на новую задачу множественного выбора
func NewMultiSelectTask(title string, items []Item) *MultiSelectTask {
	return &MultiSelectTask{task.NewMultiSelectTask(title, items)}
}

// MultiSelectTask представляет задачу для выбора нескольких вариантов из списка
type MultiSelectTask struct {
	*task.MultiSelectTask
}

// WithViewport устанавливает размер viewport (окна просмотра) для ограничения количества отображаемых элементов.
// Это полезно для длинных списков, когда нужно показывать только часть элементов.
//
// @param size Количество элементов для отображения одновременно (0 = показать все)
// @param showCounters Флаг, указывающий, нужно ли отображать счетчики элементов
// @return Указатель на задачу для цепочки вызовов
func (t *MultiSelectTask) WithViewport(size int, showCounters ...bool) *MultiSelectTask {
	t.MultiSelectTask.WithViewport(size, showCounters...)
	return t
}

// WithTimeout устанавливает тайм-аут для задачи с значениями по умолчанию
//
// @param duration Тайм-аут в миллисекундах
// @param defaultValues Значения по умолчанию
// @return Указатель на задачу для цепочки вызовов
func (t *MultiSelectTask) WithTimeout(duration time.Duration, defaultValues interface{}) *MultiSelectTask {
	t.MultiSelectTask.WithTimeout(duration, defaultValues)
	return t
}

// WithItemsDisabled помечает элементы меню как недоступные для выбора.
// Поддерживаются типы: int, []int, string, []string. Nil очищает список отключённых элементов.
//
// @param disabled Список индексов или значений, которые нужно отключить
// @return Указатель на задачу для цепочки вызовов
func (t *MultiSelectTask) WithItemsDisabled(disabled interface{}) *MultiSelectTask {
	t.MultiSelectTask.WithItemsDisabled(disabled)
	return t
}

// WithSelectAll добавляет опцию "Выбрать все" в начало списка
//
// @param options Дополнительные параметры: текст включения, текст отключения, стиль (lipgloss.Style)
// @return Указатель на задачу для цепочки вызовов
func (t *MultiSelectTask) WithSelectAll(options ...interface{}) *MultiSelectTask {
	t.MultiSelectTask.WithSelectAll(options...)
	return t
}

// WithDefaultItems задает элементы, которые будут отмечены при открытии задачи.
//
// @param defauiltSelection Индекс или значение элемента, который будет подсвечен
// @return Указатель на задачу для цепочки вызовов
func (t *MultiSelectTask) WithDefaultItems(defauiltSelection interface{}) *MultiSelectTask {
	t.MultiSelectTask.WithDefaultItems(defauiltSelection)
	return t
}

// WithRequireSelection управляет требованием выбрать хотя бы один пункт перед завершением задачи.
func (t *MultiSelectTask) WithRequireSelection(required bool) *MultiSelectTask {
	t.MultiSelectTask.WithRequireSelection(required)
	return t
}

// WithDependencies задаёт правила динамического включения и отключения пунктов меню.
// Правила принимаются в виде карты ключа элемента к действиям при выборе и снятии отметки.
func (t *MultiSelectTask) WithDependencies(rules map[string]MultiSelectDependencyRule) *MultiSelectTask {
	t.MultiSelectTask.WithDependencies(rules)
	return t
}

// GetSelected возвращает список выбранных элементов
//
// @return Список выбранных элементов
func (t *MultiSelectTask) GetSelected() []string {
	return t.MultiSelectTask.GetSelected()
}

// ----------------------------------------------------------------------------
// InputTask
// ----------------------------------------------------------------------------

// NewInputTask создает новую задачу ввода текста
//
// @param title Заголовок задачи
// @param prompt Подсказка для ввода
// @return Указатель на новую задачу ввода текста
func NewInputTask(title, prompt string) *InputTask {
	return &InputTask{task.NewInputTaskNew(title, prompt)}
}

// InputTask представляет задачу ввода текста
type InputTask struct {
	*task.InputTaskNew
}

// WithTimeout устанавливает тайм-аут для задачи с значением по умолчанию
//
// @param duration Тайм-аут в миллисекундах
// @param defaultValue Значение по умолчанию
// @return Указатель на задачу для цепочки вызовов
func (t *InputTask) WithTimeout(duration time.Duration, defaultValue string) *InputTask {
	t.InputTaskNew.WithTimeout(duration, defaultValue)
	return t
}

// WithValidator устанавливает валидатор для ввода
//
// @param validator Валидатор для ввода
// @return Указатель на задачу для цепочки вызовов
func (t *InputTask) WithValidator(validator validation.Validator) *InputTask {
	t.InputTaskNew.WithValidator(validator)
	return t
}

// WithInputType устанавливает тип ввода (пароль, email и т.д.)
//
// @param inputType Тип ввода
// @return Указатель на задачу для цепочки вызовов
func (t *InputTask) WithInputType(inputType task.InputType) *InputTask {
	t.InputTaskNew.WithInputType(inputType)
	return t
}

// WithVisibleLength устанавливает видимую длину поля ввода.
// Длинные значения будут прокручиваться по мере перемещения курсора
// или ввода новых символов за пределами указанной длины.
// Значение автоматически ограничивается минимальной и текущей полной
// шириной поля, чтобы сохранить корректный рендеринг.
func (t *InputTask) WithVisibleLength(length int) *InputTask {
	t.InputTaskNew.WithVisibleLength(length)
	return t
}

// GetValue возвращает введенное значение
//
// @return Введенное значение
func (t *InputTask) GetValue() string {
	return t.InputTaskNew.GetValue()
}

// InputType представляет тип поля ввода
type InputType = task.InputType

const (
	InputTypeText     = task.InputTypeText
	InputTypePassword = task.InputTypePassword
	InputTypeEmail    = task.InputTypeEmail
	InputTypeNumber   = task.InputTypeNumber
	InputTypeIP       = task.InputTypeIP
	InputTypeDomain   = task.InputTypeDomain
)

// ----------------------------------------------------------------------------
// FuncTask
// ----------------------------------------------------------------------------

// NewFuncTask создает новую задачу выполнения функции
//
// @param title Заголовок задачи
// @param fn Функция, которая будет выполнена
// @param opts Опции для конфигурации задачи
// @return Указатель на новую задачу выполнения функции
func NewFuncTask(title string, fn func() error, opts ...task.FuncTaskOption) *FuncTask {
	return &FuncTask{task.NewFuncTask(title, fn, opts...)}
}

// FuncTask представляет задачу выполнения функции
type FuncTask struct {
	*task.FuncTask
}

// WithStopOnError устанавливает флаг остановки очереди при ошибке
//
// @param stop Флаг остановки очереди при ошибке
// @return Указатель на задачу для цепочки вызовов
func (t *FuncTask) WithStopOnError(stop bool) *FuncTask {
	t.SetStopOnError(stop)
	return t
}

// Валидаторы - экспорт фабрики валидаторов
var DefaultValidators = validation.DefaultFactory

// FuncTaskOption представляет опцию для конфигурации FuncTask
type FuncTaskOption = task.FuncTaskOption

// Опции для FuncTask
var (
	WithSummaryFunction = task.WithSummaryFunction
	WithStopOnError     = task.WithStopOnError
)

// Стили для текста
var (
	ErrorStatusStyle   = ui.ErrorStatusStyle
	ErrorMessageStyle  = ui.ErrorMessageStyle
	CancelStyle        = ui.CancelStyle
	SubtleStyle        = ui.SubtleStyle
	SelectionStyle     = ui.SelectionStyle
	SelectionNoStyle   = ui.SelectionNoStyle
	ActiveStyle        = ui.ActiveStyle
	InputStyle         = ui.InputStyle
	SpinnerStyle       = ui.SpinnerStyle
	ActiveTitleStyle   = ui.ActiveTitleStyle
	ActiveTaskStyle    = ui.ActiveTaskStyle
	SuccessLabelStyle  = ui.SuccessLabelStyle
	FinishedLabelStyle = ui.FinishedLabelStyle
)

// Коды цветов
var (
	GreenBright   = ui.ColorBrightGreen
	RedBright     = ui.ColorBrightRed
	RedDark       = ui.ColorDarkRed
	YellowBright  = ui.ColorBrightYellow
	YellowDark    = ui.ColorDarkYellow
	OrangeBright  = ui.ColorBrightOrange
	OrangeDark    = ui.ColorDarkOrange
	BlueBright    = ui.ColorBrightBlue
	BlueDark      = ui.ColorDarkBlue
	CyanBright    = ui.ColorBrightCyan
	CyanDark      = ui.ColorDarkCyan
	MagentaBright = ui.ColorBrightMagenta
	WhiteBright   = ui.ColorBrightWhite
	GrayBright    = ui.ColorBrightGray
	GrayDark      = ui.ColorDarkGray
	LightBlue     = ui.ColorLightBlue
	Black         = ui.ColorBlack
	DarkGreen     = ui.ColorDarkGreen
)

// ----------------------------------------------------------------------------
// Автоконфигурация для встроенных систем
// ----------------------------------------------------------------------------

// AutoConfigure автоматически настраивает Ziva для оптимальной работы
// на текущей системе. Анализирует доступную память, архитектуру процессора
// и характеристики терминала для выбора наилучших настроек производительности.
//
// Рекомендуется вызывать в начале программы для встроенных систем и
// устройств с ограниченными ресурсами.
func AutoConfigure() {
	autoconfig.AutoConfigure()
}

// Is64Bit возвращает true, если приложение скомпилировано для 64-битной архитектуры.
// Полезно для определения доступного адресного пространства и выбора алгоритмов
// оптимизации памяти.
//
// @return true для 64-битных систем, false для 32-битных
func Is64Bit() bool {
	return autoconfig.Is64Bit()
}

// ----------------------------------------------------------------------------
// Утилиты производительности
// ----------------------------------------------------------------------------

// RepeatEfficient создает строку путем повторения заданной строки указанное количество раз.
// Использует оптимизированный алгоритм для минимизации аллокаций памяти.
//
// @param s Строка для повторения
// @param count Количество повторений
// @return Результирующая строка
func RepeatEfficient(s string, count int) string {
	return performance.RepeatEfficient(s, count)
}

// JoinEfficient объединяет срез строк с указанным разделителем.
// Оптимизирован для минимизации аллокаций памяти при работе с большими объемами данных.
//
// @param parts Срез строк для объединения
// @param separator Разделитель между строками
// @return Объединенная строка
func JoinEfficient(parts []string, separator string) string {
	return performance.JoinEfficient(parts, separator)
}

// FastConcat выполняет быстрое объединение произвольного количества строк.
// Предварительно вычисляет общую длину для минимизации перевыделения памяти.
//
// @param parts Строки для объединения
// @return Объединенная строка
func FastConcat(parts ...string) string {
	return performance.FastConcat(parts...)
}

// CleanWhitespaceEfficient удаляет лишние пробельные символы из строки.
// Заменяет последовательности пробелов, табуляций и переносов строк одним пробелом.
//
// @param s Исходная строка
// @return Очищенная строка
func CleanWhitespaceEfficient(s string) string {
	return performance.CleanWhitespaceEfficient(s)
}

// ----------------------------------------------------------------------------
// Управление цветами и стилями
// ----------------------------------------------------------------------------

// SetErrorColor устанавливает пользовательские цвета для отображения ошибок.
// Позволяет настроить цветовую схему для сообщений об ошибках и статусов.
//
// @param errorsColor Цвет для текста сообщений об ошибках
// @param statusColor Цвет для статусных меток (например, "ОШИБКА", "ОТКАЗ")
func SetErrorColor(errorsColor lipgloss.TerminalColor, statusColor lipgloss.TerminalColor) {
	ui.SetErrorColor(errorsColor, statusColor)
}

// ResetErrorColors сбрасывает цвета ошибок к значениям по умолчанию.
// Полезно для возврата к стандартной цветовой схеме после экспериментов.
func ResetErrorColors() {
	ui.ResetErrorColors()
}

// EnableEmbeddedMode включает режим встроенной системы с упрощенной цветовой палитрой.
// Рекомендуется для маршрутизаторов, IoT устройств и систем с ограниченными
// возможностями терминала.
func EnableEmbeddedMode() {
	ui.EnableEmbeddedMode()
}

// EnableASCIIMode включает ASCII-режим для терминалов без поддержки Unicode.
// Заменяет специальные символы на ASCII-эквиваленты.
func EnableASCIIMode() {
	ui.EnableASCIIMode()
}

// IsEmbeddedColorMode возвращает true, если включен режим встроенной системы.
//
// @return true, если активен режим встроенной системы
func IsEmbeddedColorMode() bool {
	return ui.IsEmbeddedColorMode()
}

// ----------------------------------------------------------------------------
// Расширенные валидаторы
// ----------------------------------------------------------------------------

// NewPasswordValidator создает валидатор паролей с указанной минимальной длиной.
// Проверяет сложность пароля, наличие различных типов символов.
//
// @param minLength Минимальная длина пароля
// @return Настроенный валидатор паролей
func NewPasswordValidator(minLength int) validation.Validator {
	return validation.NewPasswordValidator(minLength)
}

// NewEmailValidator создает валидатор email адресов.
// Проверяет корректность формата email согласно RFC стандартам.
//
// @return Валидатор email адресов
func NewEmailValidator() validation.Validator {
	return validation.NewEmailValidator()
}

// NewIPValidator создает валидатор IP адресов с настраиваемой поддержкой IPv4/IPv6.
//
// @param allowIPv4 Разрешить IPv4 адреса
// @param allowIPv6 Разрешить IPv6 адреса
// @return Настроенный валидатор IP адресов
func NewIPValidator(allowIPv4, allowIPv6 bool) validation.Validator {
	return validation.NewIPValidator(allowIPv4, allowIPv6)
}

// NewIPv4Validator создает валидатор только для IPv4 адресов.
//
// @return Валидатор IPv4 адресов
func NewIPv4Validator() validation.Validator {
	return validation.NewIPv4Validator()
}

// NewIPv6Validator создает валидатор только для IPv6 адресов.
//
// @return Валидатор IPv6 адресов
func NewIPv6Validator() validation.Validator {
	return validation.NewIPv6Validator()
}

// NewDomainValidator создает валидатор доменных имен.
// Проверяет корректность формата доменного имени согласно RFC стандартам.
//
// @return Валидатор доменных имен
func NewDomainValidator() validation.Validator {
	return validation.NewDomainValidator()
}

// NewTextValidator создает валидатор текста с ограничениями по длине.
//
// @param minLen Минимальная длина текста (0 = без ограничений)
// @param maxLen Максимальная длина текста (0 = без ограничений)
// @return Настроенный валидатор текста
func NewTextValidator(minLen, maxLen int) validation.Validator {
	return validation.NewTextValidator(minLen, maxLen)
}

// ----------------------------------------------------------------------------
// Утилиты для работы со строками и интернированием
// ----------------------------------------------------------------------------

// InternString выполняет интернирование строки для экономии памяти.
// Повторные вызовы с одинаковыми строками возвращают ссылку на один объект в памяти.
// Особенно эффективно для часто повторяющихся строк в UI.
//
// @param s Строка для интернирования
// @return Интернированная строка
func InternString(s string) string {
	return ui.InternString(s)
}

// GetCacheStats возвращает статистику кэша интернированных строк.
// Полезно для мониторинга использования памяти и оптимизации производительности.
//
// @return size - текущий размер кэша, capacity - максимальная емкость
func GetCacheStats() (size int, capacity int) {
	return ui.GetCacheStats()
}

// ClearInternCache очищает кэш интернированных строк.
// Рекомендуется вызывать периодически в долгоработающих приложениях.
func ClearInternCache() {
	ui.ClearInternCache()
}

// CalculateLayoutWidth вычисляет оптимальную ширину макета на основе ширины экрана.
// Учитывает отступы и ограничения для обеспечения читаемости на разных устройствах.
//
// @param screenWidth Ширина экрана в символах
// @return Рекомендуемая ширина макета
func CalculateLayoutWidth(screenWidth int) int {
	return common.CalculateLayoutWidth(screenWidth)
}
