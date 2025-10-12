package task

import (
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/defaults"
	terrors "github.com/qzeleza/ziva/internal/errors"
	"github.com/qzeleza/ziva/internal/performance"
	"github.com/qzeleza/ziva/internal/ui"
	"github.com/qzeleza/ziva/internal/validation"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputType определяет тип ввода для InputTask
type InputType int

// Константы для различных типов ввода
const (
	InputTypeText     InputType = iota // Обычный текстовый ввод
	InputTypePassword                  // Ввод пароля (скрытый текст)
	InputTypeEmail                     // Ввод email
	InputTypeNumber                    // Ввод числа
	InputTypeIP                        // Ввод IP-адреса
	InputTypeDomain                    // Ввод доменного имени
)

// InputTaskNew представляет улучшенную задачу ввода с разделенными обязанностями
type InputTaskNew struct {
	BaseTask

	// Компоненты с разделенными обязанностями
	textInput    textinput.Model       // UI компонент ввода
	validator    validation.Validator  // Валидатор данных
	renderer     *InputRenderer        // Рендерер для отображения
	errorHandler *terrors.ErrorHandler // Обработчик ошибок

	// Состояние задачи
	inputType             InputType // Тип ввода
	validationErr         error     // Ошибка валидации
	value                 string    // Введенное значение
	prompt                string    // Подсказка для ввода
	placeholder           string    // Текст-заполнитель
	placeholderCustomized bool

	// Настройки
	width                   int  // Полная ширина поля ввода
	widthCustomized         bool // Пользовательская настройка полной ширины
	visibleLength           int  // Видимая длина поля ввода
	visibleLengthCustomized bool // Пользовательская настройка видимой длины
	maskInput               bool // Маскировать ввод (для паролей)
	allowEmpty              bool // Разрешить пустое значение
}

// NewInputTaskNew создает новую улучшенную задачу ввода
func NewInputTaskNew(title, prompt string) *InputTaskNew {
	// Создаем базовую задачу
	baseTask := NewBaseTask(title)
	baseTask.SetStopOnError(true)

	// Инициализируем компонент ввода
	ti := textinput.New()
	ti.Placeholder = defaults.DefaultPlaceholder
	ti.Focus()
	ti.CharLimit = defaults.MaxInputLength
	width := clampWidth(defaults.DefaultInputWidth)
	visibleLength := defaultVisibleLength(width)
	ti.Width = visibleLength
	ti.Cursor.Style = lipgloss.NewStyle().Background(ui.ColorLightBlue).Bold(true)
	ti.Cursor.TextStyle = lipgloss.NewStyle().Foreground(ui.ColorLightBlue).Bold(true)

	return &InputTaskNew{
		BaseTask:              baseTask,
		textInput:             ti,
		renderer:              NewInputRenderer(),
		errorHandler:          terrors.DefaultErrorHandler,
		inputType:             InputTypeText,
		prompt:                prompt,
		placeholder:           defaults.DefaultPlaceholder,
		placeholderCustomized: false,
		width:                 width,
		visibleLength:         visibleLength,
		maskInput:             false,
		allowEmpty:            false,
	}
}

// WithValidator устанавливает валидатор для задачи
func (t *InputTaskNew) WithValidator(validator validation.Validator) *InputTaskNew {
	t.validator = validator
	return t
}

// WithInputType устанавливает тип ввода
func (t *InputTaskNew) WithInputType(inputType InputType) *InputTaskNew {
	t.inputType = inputType

	// Автоматически настраиваем параметры в зависимости от типа
	switch inputType {
	case InputTypePassword:
		t.maskInput = true
		t.textInput.EchoMode = textinput.EchoPassword
		t.textInput.EchoCharacter = defaults.PasswordMask

		// Если валидатор не установлен, используем стандартный валидатор паролей
		if t.validator == nil {
			t.validator = validation.StandardPassword()
		}

	case InputTypeEmail:
		if t.validator == nil {
			t.validator = validation.Email()
		}

	case InputTypeNumber:
		if t.validator == nil {
			t.validator = validation.NewNumberValidator(defaults.DefaultNumberMin, defaults.DefaultNumberMax)
		}

	case InputTypeIP:
		if t.validator == nil {
			t.validator = validation.IP()
		}

	case InputTypeDomain:
		if t.validator == nil {
			t.validator = validation.Domain()
		}
	}

	return t
}

// WithWidth устанавливает ширину поля ввода
func (t *InputTaskNew) WithWidth(width int) *InputTaskNew {
	t.width = clampWidth(width)
	t.widthCustomized = true
	if !t.visibleLengthCustomized {
		t.visibleLength = defaultVisibleLength(t.width)
	} else {
		t.visibleLength = clampVisibleLength(t.visibleLength, t.width)
	}
	t.applyVisibleLength()
	return t
}

// WithVisibleLength задает видимую длину вводимых данных.
// Значение автоматически ограничивается допустимыми границами,
// чтобы сохранить корректный рендеринг виджета текстового ввода.
func (t *InputTaskNew) WithVisibleLength(length int) *InputTaskNew {
	t.visibleLengthCustomized = true
	t.visibleLength = clampVisibleLength(length, t.width)
	t.applyVisibleLength()
	return t
}

// applyVisibleLength обновляет ширину текстового ввода с учетом текущих настроек.
func (t *InputTaskNew) applyVisibleLength() {
	t.textInput.Width = t.visibleLength
}

// refreshAutoDimensions обновляет ширину и видимую длину вводимых данных
// автоматически в зависимости от текущего размера экрана.
//
// @param screenWidth Ширина экрана
func (t *InputTaskNew) refreshAutoDimensions(screenWidth int) {
	autoWidth, autoVisible := calculateAutoDimensions(screenWidth)
	if !t.widthCustomized {
		t.width = autoWidth
	}
	if !t.visibleLengthCustomized {
		if t.widthCustomized {
			t.visibleLength = defaultVisibleLength(t.width)
		} else {
			t.visibleLength = clampVisibleLength(autoVisible, t.width)
		}
	} else {
		t.visibleLength = clampVisibleLength(t.visibleLength, t.width)
	}
	t.applyVisibleLength()
}

// defaultVisibleLength вычисляет видимую длину вводимых данных по умолчанию.
// Значение автоматически ограничивается допустимыми границами,
// чтобы сохранить корректный рендеринг виджета текстового ввода.
func defaultVisibleLength(width int) int {
	return clampVisibleLength(adjustVisibleLength(width), width)
}

// adjustVisibleLength корректирует видимую длину вводимых данных,
// уменьшая её на количество символов, на которое по умолчанию
// уменьшается видимая область относительно полной ширины поля.
func adjustVisibleLength(width int) int {
	length := width - defaults.InputVisiblePadding
	if length < defaults.MinInputWidth {
		return width
	}
	return length
}

// clampVisibleLength ограничивает видимую длину вводимых данных
// минимальной и текущей полной шириной поля, чтобы сохранить
// корректный рендеринг виджета текстового ввода.
func clampVisibleLength(length, width int) int {
	if length < defaults.MinInputWidth {
		length = defaults.MinInputWidth
	}
	if width > 0 && length > width {
		length = width
	}
	if length > defaults.MaxInputWidth {
		length = defaults.MaxInputWidth
	}
	return length - defaults.InputVisiblePadding + 1
}

// clampWidth ограничивает ширину поля ввода минимальной и максимальной шириной,
// чтобы сохранить корректный рендеринг виджета текстового ввода.
func clampWidth(width int) int {
	if width < defaults.MinInputWidth {
		return defaults.MinInputWidth
	}
	if width > defaults.MaxInputWidth {
		return defaults.MaxInputWidth
	}
	return width
}

func calculateAutoDimensions(screenWidth int) (int, int) {
	layoutWidth := clampWidth(common.CalculateLayoutWidth(screenWidth))
	available := layoutWidth - inputPromptWidth()
	if available < defaults.MinInputWidth {
		available = defaults.MinInputWidth
	}
	fieldWidth := clampWidth(available)
	visible := fieldWidth - defaults.InputVisiblePadding
	return fieldWidth, clampVisibleLength(visible, fieldWidth)
}

func inputPromptWidth() int {
	return ui.MainLeftIndent + utf8.RuneCountInString(ui.CornerDownSymbol) + utf8.RuneCountInString(ui.HorizontalLineSymbol)
}

// WithPlaceholder устанавливает текст-заполнитель
func (t *InputTaskNew) WithPlaceholder(placeholder string) *InputTaskNew {
	t.placeholder = placeholder
	t.textInput.Placeholder = placeholder
	t.placeholderCustomized = true
	return t
}

// WithAllowEmpty разрешает пустые значения
func (t *InputTaskNew) WithAllowEmpty(allow bool) *InputTaskNew {
	t.allowEmpty = allow
	return t
}

// WithRenderer устанавливает пользовательский рендерер
func (t *InputTaskNew) WithRenderer(renderer *InputRenderer) *InputTaskNew {
	t.renderer = renderer
	return t
}

// WithStyle устанавливает стиль для рендерера
func (t *InputTaskNew) WithStyle(style lipgloss.Style) *InputTaskNew {
	t.renderer.WithStyle(style)
	return t
}

// GetValue возвращает введенное значение
func (t *InputTaskNew) GetValue() string {
	return t.value
}

// Run запускает задачу ввода
func (t *InputTaskNew) Run() tea.Cmd {
	// Запускаем мигание курсора и таймер, если он включен
	var cmds []tea.Cmd
	cmds = append(cmds, textinput.Blink)

	// Запускаем таймер и тикер, если они включены
	if t.timeoutEnabled && t.timeoutManager != nil {
		cmds = append(cmds, t.timeoutManager.StartTickerAndTimeout())
	}

	return tea.Batch(cmds...)
}

// Update обрабатывает сообщения и обновляет состояние задачи
func (t *InputTaskNew) Update(msg tea.Msg) (Task, tea.Cmd) {
	if t.done {
		return t, nil
	}

	switch msg := msg.(type) {
	case TimeoutMsg:
		// Применяем значение по умолчанию при истечении таймера
		t.applyDefaultValue()
		return t, nil
	case TickMsg:
		// Обрабатываем тик для обновления счетчика времени
		if t.timeoutEnabled && t.timeoutManager != nil && t.timeoutManager.IsActive() {
			return t, t.timeoutManager.StartTicker()
		}
		return t, nil
	case tea.KeyMsg:
		// При любом нажатии клавиши (кроме служебных) скрываем таймер
		key := msg.String()
		// Останавливаем таймер при любом вводе, кроме управляющих клавиш
		if t.timeoutEnabled && t.timeoutManager != nil && t.timeoutManager.IsActive() {
			// Отключаем таймер при вводе текста
			if key != "ctrl+c" && key != "esc" && key != "enter" {
				t.DisableTimeout()
			}
		}

		switch msg.String() {
		case "ctrl+c", "esc", "Ctrl+C", "Esc":
			// Отмена ввода
			return t.handleCancel()
		case "right", "Right":
			if performance.TrimSpaceEfficient(t.textInput.Value()) == "" {
				if value, defaultErr, _, hasDefault := t.resolveDefaultValue(); hasDefault {
					if defaultErr != nil {
						t.validationErr = defaultErr
						t.SetError(defaultErr)
					} else {
						t.textInput.SetValue(value)
						t.textInput.SetCursor(utf8.RuneCountInString(value))
						t.validationErr = nil
						t.SetError(nil)
						t.validateInput()
					}
					return t, nil
				}
			}
			var cmd tea.Cmd
			t.textInput, cmd = t.textInput.Update(msg)
			return t, cmd
		case "left", "Left":
			// Обрабатываем выход по стрелке влево при пустом поле ввода
			if performance.TrimSpaceEfficient(t.textInput.Value()) == "" {
				return t.handleCancel()
			}
			var cmd tea.Cmd
			t.textInput, cmd = t.textInput.Update(msg)
			return t, cmd
		case "enter":
			// Подтверждение ввода
			return t.handleSubmit()
		default:
			// Обновляем поле ввода
			var cmd tea.Cmd
			t.textInput, cmd = t.textInput.Update(msg)
			return t, cmd
		}

	case error:
		// Обработка ошибки валидации с контекстом
		taskErr := t.errorHandler.Handle(t.title, msg)
		t.validationErr = taskErr
		t.SetError(taskErr)
		return t, nil
	}

	// Обновляем компонент ввода
	var cmd tea.Cmd
	t.textInput, cmd = t.textInput.Update(msg)

	// Валидируем введенное значение в реальном времени
	t.validateInput()

	// Запускаем таймер при первом обновлении, если он включен и еще не активен
	if t.timeoutEnabled && t.timeoutManager != nil && !t.timeoutManager.IsActive() {
		return t, t.timeoutManager.StartTickerAndTimeout()
	}

	// Продолжаем тикер, если таймер активен
	if t.timeoutEnabled && t.timeoutManager != nil && t.timeoutManager.IsActive() {
		return t, tea.Batch(cmd, t.timeoutManager.StartTicker())
	}

	return t, cmd
}

// validateInput выполняет валидацию текущего ввода
func (t *InputTaskNew) validateInput() {
	currentValue := t.textInput.Value()

	// Проверяем пустое значение
	if currentValue == "" && !t.allowEmpty {
		if t.validator != nil {
			// Даем валидатору шанс проверить пустое значение
			if err := t.validator.Validate(currentValue); err != nil {
				t.validationErr = err
				return
			}
		}
		t.validationErr = nil
		return
	}

	// Выполняем валидацию если есть валидатор
	if t.validator != nil {
		if err := t.validator.Validate(currentValue); err != nil {
			// Создаем TaskError с контекстом валидации
			t.validationErr = terrors.NewValidationError(t.title, err).
				WithContext("input_type", t.inputType).
				WithContext("value_length", len(currentValue))
		} else {
			// Если валидация успешна — очищаем ошибку
			t.validationErr = nil
			t.SetError(nil)
		}
	} else {
		t.validationErr = nil
		t.SetError(nil)
	}
}

// handleSubmit обрабатывает подтверждение ввода
func (t *InputTaskNew) handleSubmit() (Task, tea.Cmd) {
	currentValue := t.textInput.Value()
	trimmed := performance.TrimSpaceEfficient(currentValue)

	if trimmed == "" {
		if value, defaultErr, _, hasDefault := t.resolveDefaultValue(); hasDefault {
			if defaultErr != nil {
				t.validationErr = defaultErr
				t.SetError(defaultErr)
				t.textInput.Focus()
				return t, nil
			}
			t.textInput.SetValue(value)
			t.textInput.SetCursor(utf8.RuneCountInString(value))
			currentValue = value
			trimmed = performance.TrimSpaceEfficient(currentValue)
			t.validationErr = nil
			t.SetError(nil)
		}
	}

	// Финальная валидация
	if !t.allowEmpty && trimmed == "" {
		emptyErr := terrors.NewValidationError(t.title, errors.New(defaults.ErrFieldRequired)).
			WithContext("required", true)
		t.validationErr = emptyErr
		t.SetError(emptyErr)
		t.textInput.Focus()
		return t, nil
	}

	if t.validator != nil {
		if err := t.validator.Validate(currentValue); err != nil {
			validationErr := terrors.NewValidationError(t.title, err).
				WithContext("final_validation", true).
				WithContext("input_type", t.inputType)
			t.validationErr = validationErr
			t.SetError(validationErr)
			t.textInput.Focus()
			return t, nil
		}
	}

	// Ввод успешен — очищаем ошибки и фиксируем значение
	t.validationErr = nil
	t.SetError(nil)
	t.value = currentValue
	t.done = true
	t.icon = ui.IconDone
	t.finalValue = ui.SuccessLabelStyle.Render(t.getDisplayValue())

	return t, nil
}

// handleCancel обрабатывает отмену ввода
func (t *InputTaskNew) handleCancel() (Task, tea.Cmd) {
	cancelErr := terrors.NewCancelError(t.title).
		WithContext("input_type", t.inputType).
		WithContext("partial_value", t.textInput.Value())

	t.SetError(cancelErr)
	t.done = true
	t.icon = ui.IconCancelled
	t.finalValue = ui.ErrorMessageStyle.Render(defaults.CancelShort)

	return t, nil
}

// getDisplayValue возвращает значение для отображения (маскирует пароли)
func (t *InputTaskNew) getDisplayValue() string {
	if t.maskInput {
		return performance.RepeatEfficient(string(defaults.PasswordMask), len(t.value))
	}
	return t.value
}

func stringifyDefaultValue(value interface{}) (string, bool) {
	switch v := value.(type) {
	case nil:
		return "", false
	case string:
		return v, true
	case fmt.Stringer:
		return v.String(), true
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), true
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), true
	case float32, float64:
		return fmt.Sprintf("%v", v), true
	case bool:
		return fmt.Sprintf("%t", v), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// resolveDefaultValue конвертирует и валидирует значение по умолчанию.
// Возвращает строковое значение, ошибку валидации (если есть), сообщение для пользователя и флаг наличия значения.
func (t *InputTaskNew) resolveDefaultValue() (string, *terrors.TaskError, string, bool) {
	if t.defaultValue == nil {
		return "", nil, "", false
	}

	var valueToSet string
	switch val := t.defaultValue.(type) {
	case string:
		valueToSet = val
	case int:
		valueToSet = fmt.Sprintf("%d", val)
	case float64:
		valueToSet = fmt.Sprintf("%f", val)
	default:
		valueToSet = fmt.Sprintf("%v", val)
	}

	if t.validator != nil {
		if err := t.validator.Validate(valueToSet); err != nil {
			validationErr := terrors.NewValidationError(t.title, err).
				WithContext("defauilt_value", true).
				WithContext("input_type", t.inputType)
			return "", validationErr, defaults.ErrDefaultValueInvalid, true
		}
	}

	if !t.allowEmpty && performance.TrimSpaceEfficient(valueToSet) == "" {
		emptyErr := terrors.NewValidationError(t.title, errors.New(defaults.ErrFieldRequired)).
			WithContext("defauilt_value", true)
		return "", emptyErr, defaults.ErrDefaultValueEmpty, true
	}

	return valueToSet, nil, "", true
}

// applyDefaultValue применяет значение по умолчанию при истечении таймера
func (t *InputTaskNew) applyDefaultValue() {
	valueToSet, defaultErr, failMessage, hasDefault := t.resolveDefaultValue()
	if !hasDefault {
		return
	}

	if defaultErr != nil {
		t.validationErr = defaultErr
		t.SetError(defaultErr)
		t.done = true
		t.icon = ui.IconError
		if failMessage == "" {
			failMessage = defaults.ErrDefaultValueInvalid
		}
		t.finalValue = ui.GetErrorMessageStyle().Render(failMessage)
		return
	}

	t.textInput.SetValue(valueToSet)
	t.value = valueToSet
	t.done = true
	t.icon = ui.IconDone
	t.finalValue = ui.SuccessLabelStyle.Render(t.getDisplayValue())
	t.validationErr = nil
	t.SetError(nil)
}

// View отображает текущее состояние задачи
func (t *InputTaskNew) View(width int) string {
	if !t.widthCustomized || !t.visibleLengthCustomized {
		t.refreshAutoDimensions(width)
	} else {
		t.visibleLength = clampVisibleLength(t.visibleLength, t.width)
		t.applyVisibleLength()
	}

	if t.IsDone() {
		return t.FinalView(width)
	}

	// Получаем отформатированный таймер (если он активен)
	title := t.title
	timerStr := t.RenderTimer()

	// Убедимся, что текстовое поле активно
	if !t.textInput.Focused() {
		t.textInput.Focus()
	}

	return t.renderer.RenderInput(
		title,
		t.textInput,
		t.validator,
		t.validationErr,
		t.inputType,
		t.InProgressPrefix(),
		t.SelectionSeparatorEnabled(),
		width,
		timerStr,
	)
}

// FinalView отображает финальное состояние задачи
func (t *InputTaskNew) FinalView(width int) string {
	completedPrefix := t.CompletedPrefix()
	if t.HasError() {
		return t.renderer.RenderFinal(t.title, "", true, t.Error(), completedPrefix, width)
	}

	return t.renderer.RenderFinal(t.title, t.getDisplayValue(), false, nil, completedPrefix, width)
}

// InputTaskBuilder предоставляет fluent API для создания InputTask
type InputTaskBuilder struct {
	task *InputTaskNew
}

// NewInputTaskBuilder создает новый построитель задач ввода
func NewInputTaskBuilder(title, prompt string) *InputTaskBuilder {
	return &InputTaskBuilder{
		task: NewInputTaskNew(title, prompt),
	}
}

// Password настраивает задачу для ввода пароля
func (b *InputTaskBuilder) Password() *InputTaskBuilder {
	b.task.WithInputType(InputTypePassword)
	return b
}

// Email настраивает задачу для ввода email
func (b *InputTaskBuilder) Email() *InputTaskBuilder {
	b.task.WithInputType(InputTypeEmail)
	return b
}

// Number настраивает задачу для ввода числа
func (b *InputTaskBuilder) Number(min, max int) *InputTaskBuilder {
	b.task.WithInputType(InputTypeNumber).WithValidator(validation.NewNumberValidator(min, max))
	return b
}

// IP настраивает задачу для ввода IP адреса
func (b *InputTaskBuilder) IP() *InputTaskBuilder {
	b.task.WithInputType(InputTypeIP)
	return b
}

// Domain настраивает задачу для ввода домена
func (b *InputTaskBuilder) Domain() *InputTaskBuilder {
	b.task.WithInputType(InputTypeDomain)
	return b
}

// VisibleLength задает видимую длину вводимых данных для задачи.
func (b *InputTaskBuilder) VisibleLength(length int) *InputTaskBuilder {
	b.task.WithVisibleLength(length)
	return b
}

// Required делает поле обязательным
func (b *InputTaskBuilder) Required() *InputTaskBuilder {
	b.task.WithAllowEmpty(false)
	return b
}

// Optional делает поле опциональным
func (b *InputTaskBuilder) Optional() *InputTaskBuilder {
	b.task.WithAllowEmpty(true)
	return b
}

// Width устанавливает ширину поля
func (b *InputTaskBuilder) Width(width int) *InputTaskBuilder {
	b.task.WithWidth(width)
	return b
}

// Placeholder устанавливает текст-заполнитель
func (b *InputTaskBuilder) Placeholder(placeholder string) *InputTaskBuilder {
	b.task.WithPlaceholder(placeholder)
	return b
}

// Validator устанавливает пользовательский валидатор
func (b *InputTaskBuilder) Validator(validator validation.Validator) *InputTaskBuilder {
	b.task.WithValidator(validator)
	return b
}

// Build возвращает готовую задачу
func (b *InputTaskBuilder) Build() *InputTaskNew {
	return b.task
}

// WithTimeout устанавливает тайм-аут для задачи ввода
// @param duration Длительность тайм-аута
// @param defaultValue Значение по умолчанию (строка)
// @return Указатель на задачу для цепочки вызовов
func (t *InputTaskNew) WithTimeout(duration time.Duration, defaultValue interface{}) *InputTaskNew {
	t.BaseTask.WithTimeout(duration, defaultValue)
	if !t.placeholderCustomized {
		if placeholder, ok := stringifyDefaultValue(defaultValue); ok && performance.TrimSpaceEfficient(placeholder) != "" {
			t.placeholder = placeholder
			t.textInput.Placeholder = placeholder
		}
	}
	return t
}
