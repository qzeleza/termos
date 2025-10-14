package task

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/performance"
	"github.com/qzeleza/ziva/internal/ui"
)

// SelectionBitset представляет битовый набор для оптимальной работы с embedded устройствами
// Использует uint32 для лучшей производительности на 32-битных embedded системах
// Поддерживает до 32 элементов выбора, что покрывает большинство практических случаев
type SelectionBitset uint32

type dependencyAction struct {
	disable     []int
	enable      []int
	forceSelect []int
	forceClear  []int
}

type dependencyRule struct {
	onSelect   dependencyAction
	onDeselect dependencyAction
}

// MultiSelectDependencyActions описывает действия, выполняемые при смене состояния пункта.
type MultiSelectDependencyActions struct {
	Disable     []string
	Enable      []string
	ForceSelect []string
	ForceClear  []string
}

// MultiSelectDependencyRule задаёт набор действий при выборе и снятии выбора пункта меню.
type MultiSelectDependencyRule struct {
	OnSelect   MultiSelectDependencyActions
	OnDeselect MultiSelectDependencyActions
}

// Set устанавливает бит в позиции index
func (s *SelectionBitset) Set(index int) {
	if index >= 0 && index < 32 {
		*s |= 1 << index
	}
}

// Clear очищает бит в позиции index
func (s *SelectionBitset) Clear(index int) {
	if index >= 0 && index < 32 {
		*s &^= 1 << index
	}
}

// IsSet проверяет, установлен ли бит в позиции index
func (s SelectionBitset) IsSet(index int) bool {
	if index < 0 || index >= 32 {
		return false
	}
	return s&(1<<index) != 0
}

// Toggle переключает бит в позиции index
func (s *SelectionBitset) Toggle(index int) {
	if index >= 0 && index < 32 {
		*s ^= 1 << index
	}
}

// Count возвращает количество установленных битов (оптимизировано для 32-бит)
func (s SelectionBitset) Count() int {
	// Используем встроенную функцию подсчета битов для оптимизации
	// На современных процессорах это компилируется в инструкцию POPCNT
	return popcount32(uint32(s))
}

// popcount32 подсчитывает установленные биты в 32-битном числе
func popcount32(x uint32) int {
	// Алгоритм Брайана Кернигана - эффективен для разреженных битов
	count := 0
	for x != 0 {
		x &= x - 1 // Убираем самый младший установленный бит
		count++
	}
	return count
}

func clearIndexSet(set map[int]struct{}) {
	if set == nil {
		return
	}
	for idx := range set {
		delete(set, idx)
	}
}

// Clear очищает все биты
func (s *SelectionBitset) ClearAll() {
	*s = 0
}

// SetAll устанавливает биты для первых n позиций
func (s *SelectionBitset) SetAll(n int) {
	if n <= 0 {
		*s = 0
		return
	}
	if n >= 32 {
		*s = SelectionBitset(^uint32(0))
		return
	}
	*s = SelectionBitset((1 << n) - 1)
}

// MultiSelectTask позволяет выбрать несколько вариантов из списка.
type MultiSelectTask struct {
	BaseTask
	items                []choice               // Список вариантов выбора
	disabled             map[int]struct{}       // Итоговый набор отключённых пунктов
	staticDisabled       map[int]struct{}       // Статические блокировки из конфигурации
	dynamicDisabled      map[int]struct{}       // Динамические блокировки по зависимостям
	dependencies         map[int]dependencyRule // Правила зависимостей по индексам
	selected             SelectionBitset        // Битовый набор выбранных элементов (оптимизировано для embedded)
	fallbackMap          map[int]struct{}       // Резервная карта для списков > 64 элементов
	cursor               int                    // Текущая позиция курсора
	activeStyle          lipgloss.Style         // Стиль для активного элемента
	hasSelectAll         bool                   // Включена ли опция "Выбрать все"
	selectAllEnableText  string                 // Текст опции активации "Выбрать все"
	selectAllDisableText string                 // Текст опции отключения выбора
	selectAllStyle       lipgloss.Style         // Стиль отображения опции "Выбрать все"
	showHelpMessage      bool                   // Показывать ли сообщение-подсказку
	helpMessage          string                 // Текст сообщения-подсказки
	// Viewport (окно просмотра) для ограничения количества отображаемых элементов
	viewportSize     int  // Размер viewport (количество видимых элементов), 0 = показать все
	viewportStart    int  // Начальная позиция viewport в списке элементов
	showCounters     bool // Показывать ли счетчики для выбранных элементов
	requireSelection bool // Требовать выбор хотя бы одного элемента перед завершением задачи
}

// NewMultiSelectTask создает новую задачу множественного выбора.
//
// @param title Заголовок задачи
// @param items Список вариантов выбора
// @return Указатель на новую задачу множественного выбора
func NewMultiSelectTask(title string, items []Item) *MultiSelectTask {
	normalized := normalizeItems(items)

	task := &MultiSelectTask{
		BaseTask:             NewBaseTask(title),
		items:                normalized,
		disabled:             make(map[int]struct{}),
		staticDisabled:       make(map[int]struct{}),
		dynamicDisabled:      make(map[int]struct{}),
		dependencies:         make(map[int]dependencyRule),
		cursor:               0, // Начинаем с первого элемента списка
		activeStyle:          ui.ActiveStyle,
		hasSelectAll:         false,
		selectAllEnableText:  defaults.SelectAllDefaultText,
		selectAllDisableText: defaults.SelectAllDisableDefaultText,
		selectAllStyle:       ui.MenuActionDefaultStyle(),
		showHelpMessage:      false,
		helpMessage:          "",
		// Viewport по умолчанию отключен (показываем все элементы)
		viewportSize:     0,
		viewportStart:    0,
		showCounters:     true,
		requireSelection: false,
	}

	// Для embedded устройств: используем битсет для списков <= 32 элементов,
	// иначе fallback на карту
	if len(normalized) > 32 {
		task.fallbackMap = make(map[int]struct{})
	}

	task.ensureCursorSelectable()
	return task
}

// isDisabled проверяет, помечен ли элемент как недоступный
func (t *MultiSelectTask) isDisabled(index int) bool {
	if index < 0 || index >= len(t.items) {
		return true
	}
	if t.items[index].isDivider() {
		return true
	}
	_, exists := t.disabled[index]
	return exists
}

// ensureCursorSelectable пытается разместить курсор на ближайшем доступном элементе
func (t *MultiSelectTask) ensureCursorSelectable() bool {
	if len(t.items) == 0 {
		if t.hasSelectAll {
			t.cursor = -1
		} else {
			t.cursor = -1
		}
		return false
	}

	if t.hasSelectAll && t.cursor == -1 {
		return true
	}

	if t.cursor >= 0 && t.cursor < len(t.items) && !t.isDisabled(t.cursor) {
		return true
	}

	start := t.cursor
	if start < 0 {
		start = 0
	}

	if idx, ok := t.findEnabledForward(start); ok {
		t.cursor = idx
		return true
	}

	if idx, ok := t.findEnabledBackward(start - 1); ok {
		t.cursor = idx
		return true
	}

	if t.hasSelectAll {
		t.cursor = -1
		return true
	}

	t.cursor = -1
	return false
}

// findEnabledForward возвращает индекс первого доступного элемента начиная с from (включительно)
func (t *MultiSelectTask) findEnabledForward(from int) (int, bool) {
	if from < 0 {
		from = 0
	}
	for i := from; i < len(t.items); i++ {
		if !t.isDisabled(i) {
			return i, true
		}
	}
	return -1, false
}

// findEnabledBackward возвращает индекс предыдущего доступного элемента
func (t *MultiSelectTask) findEnabledBackward(from int) (int, bool) {
	if from >= len(t.items) {
		from = len(t.items) - 1
	}
	for i := from; i >= 0; i-- {
		if !t.isDisabled(i) {
			return i, true
		}
	}
	return -1, false
}

// moveCursorForward перемещает курсор на следующий доступный элемент
func (t *MultiSelectTask) moveCursorForward() bool {
	original := t.cursor
	if original == -1 {
		if idx, ok := t.findEnabledForward(0); ok && idx != original {
			t.cursor = idx
			return true
		}
		return false
	}

	if idx, ok := t.findEnabledForward(original + 1); ok && idx != original {
		t.cursor = idx
		return true
	}

	if t.hasSelectAll && original != -1 {
		t.cursor = -1
		return true
	}

	if idx, ok := t.findEnabledForward(0); ok && idx != original {
		t.cursor = idx
		return true
	}

	return false
}

// moveCursorBackward перемещает курсор на предыдущий доступный элемент
func (t *MultiSelectTask) moveCursorBackward() bool {
	original := t.cursor
	if original == -1 {
		if idx, ok := t.findEnabledBackward(len(t.items) - 1); ok && idx != original {
			t.cursor = idx
			return true
		}
		return false
	}

	if idx, ok := t.findEnabledBackward(original - 1); ok && idx != original {
		t.cursor = idx
		return true
	}

	if t.hasSelectAll {
		if original != -1 {
			t.cursor = -1
			return true
		}
		return false
	}

	if idx, ok := t.findEnabledBackward(len(t.items) - 1); ok && idx != original {
		t.cursor = idx
		return true
	}

	return false
}

// resolveDisabledIndices конвертирует произвольный ввод в индексы элементов списка
func (t *MultiSelectTask) resolveDisabledIndices(input interface{}) []int {
	var result []int
	if input == nil {
		return result
	}

	addIndex := func(idx int) {
		if idx < 0 || idx >= len(t.items) {
			return
		}
		for _, existing := range result {
			if existing == idx {
				return
			}
		}
		result = append(result, idx)
	}

	switch v := input.(type) {
	case int:
		addIndex(v)
	case []int:
		for _, idx := range v {
			addIndex(idx)
		}
	case string:
		if idx := t.choiceIndex(v); idx != -1 {
			addIndex(idx)
		}
	case []string:
		for _, val := range v {
			if idx := t.choiceIndex(val); idx != -1 {
				addIndex(idx)
			}
		}
	case []interface{}:
		for _, item := range v {
			for _, idx := range t.resolveDisabledIndices(item) {
				addIndex(idx)
			}
		}
	}

	return result
}

// choiceIndex возвращает индекс элемента по значению или -1, если элемент не найден
func (t *MultiSelectTask) choiceIndex(value string) int {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return -1
	}
	for i, item := range t.items {
		if item.isDivider() {
			continue
		}
		if strings.EqualFold(item.key, normalized) || strings.EqualFold(item.name, normalized) {
			return i
		}
	}
	return -1
}

// clearDisabledSelections снимает выбор с отключённых элементов
func (t *MultiSelectTask) clearDisabledSelections() {
	if len(t.items) > 32 {
		if t.fallbackMap == nil {
			return
		}
		for idx := range t.disabled {
			delete(t.fallbackMap, idx)
		}
		return
	}

	for idx := range t.disabled {
		t.selected.Clear(idx)
	}
}

// clearAllSelections снимает выбор со всех элементов
func (t *MultiSelectTask) clearAllSelections() {
	if len(t.items) > 32 {
		if t.fallbackMap == nil {
			return
		}
		for idx := range t.fallbackMap {
			delete(t.fallbackMap, idx)
		}
		return
	}

	t.selected.ClearAll()
}

func (t *MultiSelectTask) rebuildDisabled() {
	if t.disabled == nil {
		t.disabled = make(map[int]struct{})
	}
	clearIndexSet(t.disabled)
	for idx := range t.staticDisabled {
		t.disabled[idx] = struct{}{}
	}
	for idx := range t.dynamicDisabled {
		t.disabled[idx] = struct{}{}
	}
}

func (t *MultiSelectTask) resetDynamicDisabled() {
	if t.dynamicDisabled == nil {
		t.dynamicDisabled = make(map[int]struct{})
		return
	}
	clearIndexSet(t.dynamicDisabled)
}

func (t *MultiSelectTask) resolveDependencyTargets(keys []string) []int {
	if len(keys) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(keys))
	result := make([]int, 0, len(keys))
	for _, key := range keys {
		idx := t.choiceIndex(key)
		if idx == -1 {
			continue
		}
		if _, exists := seen[idx]; exists {
			continue
		}
		seen[idx] = struct{}{}
		result = append(result, idx)
	}
	return result
}

func (t *MultiSelectTask) resolveDependencyActionConfig(cfg MultiSelectDependencyActions) dependencyAction {
	return dependencyAction{
		disable:     t.resolveDependencyTargets(cfg.Disable),
		enable:      t.resolveDependencyTargets(cfg.Enable),
		forceSelect: t.resolveDependencyTargets(cfg.ForceSelect),
		forceClear:  t.resolveDependencyTargets(cfg.ForceClear),
	}
}

func (t *MultiSelectTask) clearDependencies() {
	if t.dependencies == nil {
		t.dependencies = make(map[int]dependencyRule)
		return
	}
	for idx := range t.dependencies {
		delete(t.dependencies, idx)
	}
}

// selectAllEnabled отмечает все доступные элементы выбранными
func (t *MultiSelectTask) selectAllEnabled() {
	if len(t.items) > 32 {
		if t.fallbackMap == nil {
			t.fallbackMap = make(map[int]struct{})
		}
		for i := range t.items {
			if t.isDisabled(i) {
				delete(t.fallbackMap, i)
				continue
			}
			t.fallbackMap[i] = struct{}{}
		}
		return
	}

	t.selected.ClearAll()
	for i := range t.items {
		if t.isDisabled(i) {
			continue
		}
		t.selected.Set(i)
	}
}

func (t *MultiSelectTask) isSelectedRaw(index int) bool {
	if index < 0 || index >= len(t.items) {
		return false
	}
	if t.items[index].isDivider() {
		return false
	}
	if len(t.items) > 32 {
		if t.fallbackMap == nil {
			return false
		}
		_, exists := t.fallbackMap[index]
		return exists
	}
	return t.selected.IsSet(index)
}

func (t *MultiSelectTask) setSelectedState(index int, active bool) bool {
	if index < 0 || index >= len(t.items) {
		return false
	}
	if t.items[index].isDivider() {
		return false
	}
	current := t.isSelectedRaw(index)
	if active == current {
		return false
	}
	if len(t.items) > 32 {
		if t.fallbackMap == nil {
			t.fallbackMap = make(map[int]struct{})
		}
		if active {
			t.fallbackMap[index] = struct{}{}
		} else {
			delete(t.fallbackMap, index)
		}
	} else {
		if active {
			t.selected.Set(index)
		} else {
			t.selected.Clear(index)
		}
	}
	return true
}

// WithViewport устанавливает размер viewport (окна просмотра) для ограничения количества отображаемых элементов.
// Это полезно для длинных списков, когда нужно показывать только часть элементов.
//
// @param size Количество элементов для отображения одновременно (0 = показать все)
// @return Указатель на задачу для цепочки вызовов
func (t *MultiSelectTask) WithViewport(size int, showCounters ...bool) *MultiSelectTask {
	if size < 0 {
		size = 0
	}
	t.viewportSize = size
	t.showCounters = true
	if len(showCounters) > 0 {
		t.showCounters = showCounters[0]
	}
	return t
}

// WithItemsDisabled помечает элементы меню как недоступные для выбора.
// Поддерживаются типы: int, []int, string, []string. Nil очищает список отключённых элементов.
func (t *MultiSelectTask) WithItemsDisabled(disabled interface{}) *MultiSelectTask {
	if t.staticDisabled == nil {
		t.staticDisabled = make(map[int]struct{})
	}
	clearIndexSet(t.staticDisabled)

	indices := t.resolveDisabledIndices(disabled)
	for _, idx := range indices {
		if idx >= 0 && idx < len(t.items) {
			t.staticDisabled[idx] = struct{}{}
		}
	}

	t.applyDependencies()
	return t
}

// WithDependencies задаёт правила динамической блокировки пунктов меню.
func (t *MultiSelectTask) WithDependencies(rules map[string]MultiSelectDependencyRule) *MultiSelectTask {
	t.clearDependencies()
	if len(rules) == 0 {
		// Даже при очистке необходимо пересчитать состояния, чтобы снять предыдущие блокировки
		t.applyDependencies()
		return t
	}

	for key, cfg := range rules {
		idx := t.choiceIndex(key)
		if idx == -1 {
			continue
		}
		t.dependencies[idx] = dependencyRule{
			onSelect:   t.resolveDependencyActionConfig(cfg.OnSelect),
			onDeselect: t.resolveDependencyActionConfig(cfg.OnDeselect),
		}
	}

	t.applyDependencies()
	return t
}

// updateViewport обновляет позицию viewport на основе текущего положения курсора
func (t *MultiSelectTask) updateViewport() {
	// Если viewport отключен, ничего не делаем
	if t.viewportSize <= 0 {
		return
	}

	// Получаем эффективную позицию курсора (с учетом опции "Выбрать все")
	effectiveCursor := t.cursor
	if t.hasSelectAll {
		effectiveCursor = t.cursor + 1 // +1 потому что опция "Выбрать все" занимает позицию -1
	}

	// Если курсор выше viewport, сдвигаем viewport вверх
	if effectiveCursor < t.viewportStart {
		t.viewportStart = effectiveCursor
	}

	// Если курсор ниже viewport, сдвигаем viewport вниз
	if effectiveCursor >= t.viewportStart+t.viewportSize {
		t.viewportStart = effectiveCursor - t.viewportSize + 1
	}

	// Убеждаемся, что viewport не выходит за границы списка
	if t.viewportStart < 0 {
		t.viewportStart = 0
	}

	maxStart := len(t.items) - t.viewportSize
	if t.hasSelectAll {
		maxStart = len(t.items) + 1 - t.viewportSize // +1 для опции "Выбрать все"
	}
	if maxStart < 0 {
		maxStart = 0
	}
	if t.viewportStart > maxStart {
		t.viewportStart = maxStart
	}
}

// getVisibleRange возвращает диапазон видимых элементов с учетом viewport
// Возвращает: startIdx, endIdx, showSelectAll
func (t *MultiSelectTask) getVisibleRange() (int, int, bool) {
	// Если viewport отключен, показываем все элементы
	if t.viewportSize <= 0 {
		return 0, len(t.items), t.hasSelectAll
	}

	// Определяем, показывать ли опцию "Выбрать все"
	showSelectAll := t.hasSelectAll && t.viewportStart == 0

	// Вычисляем диапазон элементов списка
	startIdx := t.viewportStart
	if t.hasSelectAll && startIdx > 0 {
		startIdx-- // Компенсируем опцию "Выбрать все"
	}

	if startIdx < 0 {
		startIdx = 0
	}

	endIdx := startIdx + t.viewportSize
	if showSelectAll {
		endIdx-- // Уменьшаем на 1, так как одно место занимает опция "Выбрать все"
	}

	if endIdx > len(t.items) {
		endIdx = len(t.items)
	}

	return startIdx, endIdx, showSelectAll
}

// WithSelectAll добавляет опцию "Выбрать все" в начало списка.
// При выборе этой опции все остальные пункты автоматически помечаются/снимаются.
//
// @param options Дополнительные параметры: сначала текст включения, затем текст отключения, далее стиль (lipgloss.Style)
// @return Указатель на задачу для цепочки вызовов
func (t *MultiSelectTask) WithSelectAll(options ...interface{}) *MultiSelectTask {
	t.hasSelectAll = true
	t.cursor = -1 // Начинаем с опции "Выбрать все"
	enableText := defaults.SelectAllDefaultText
	disableText := defaults.SelectAllDisableDefaultText
	style := ui.MenuActionDefaultStyle()
	stringIndex := 0

	for _, option := range options {
		switch v := option.(type) {
		case string:
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				continue
			}
			switch stringIndex {
			case 0:
				enableText = trimmed
			case 1:
				disableText = trimmed
			}
			stringIndex++
		case lipgloss.Style:
			style = v
		case *lipgloss.Style:
			if v != nil {
				style = *v
			}
		}
	}

	t.selectAllEnableText = enableText
	t.selectAllDisableText = disableText
	t.selectAllStyle = style
	t.ensureCursorSelectable()
	t.updateViewport()
	return t
}

// WithDefaultItems позволяет заранее отметить элементы списка выбранными при открытии задачи.
// Поддерживает выбор одного индекса/строки или списков значений ([]int, []string).
func (t *MultiSelectTask) WithDefaultItems(defauiltSelection interface{}) *MultiSelectTask {
	if defauiltSelection == nil || len(t.items) == 0 {
		return t
	}

	// Сбрасываем текущий выбор
	if len(t.items) > 32 {
		if t.fallbackMap == nil {
			t.fallbackMap = make(map[int]struct{})
		} else {
			for k := range t.fallbackMap {
				delete(t.fallbackMap, k)
			}
		}
		t.selected.ClearAll()
	} else {
		t.selected.ClearAll()
	}

	setSelected := func(index int) bool {
		if index < 0 || index >= len(t.items) {
			return false
		}
		if t.isDisabled(index) {
			return false
		}
		if len(t.items) > 32 && t.fallbackMap != nil {
			t.fallbackMap[index] = struct{}{}
		} else {
			t.selected.Set(index)
		}
		return true
	}

	anyApplied := false

	switch v := defauiltSelection.(type) {
	case int:
		anyApplied = setSelected(v) || anyApplied
	case string:
		if idx := t.choiceIndex(v); idx != -1 {
			if setSelected(idx) {
				anyApplied = true
			}
		}
	case []int:
		for _, idx := range v {
			if setSelected(idx) {
				anyApplied = true
			}
		}
	case []string:
		for _, val := range v {
			if idx := t.choiceIndex(val); idx != -1 {
				if setSelected(idx) {
					anyApplied = true
				}
			}
		}
	}

	if anyApplied {
		// Перемещаем курсор на первый выбранный элемент
		cursorSet := false
		for i := range t.items {
			if t.isSelected(i) {
				t.cursor = i
				cursorSet = true
				break
			}
		}
		if !cursorSet && t.hasSelectAll {
			t.cursor = -1
		}
	}

	t.applyDependencies()
	return t
}

// WithRequireSelection управляет необходимостью выбрать хотя бы один пункт перед завершением задачи.
// При required=true поведение соответствует прежнему: Enter без выбора покажет подсказку.
// По умолчанию (false) пользователь может подтвердить пустой выбор.
func (t *MultiSelectTask) WithRequireSelection(required bool) *MultiSelectTask {
	t.requireSelection = required
	return t
}

// toggleSelectAll переключает состояние всех элементов списка.
// Если все элементы выбраны - снимает выбор со всех.
// Если хотя бы один элемент не выбран - выбирает все.
func (t *MultiSelectTask) toggleSelectAll() {
	if len(t.items) == 0 {
		return
	}

	if t.isAllSelected() {
		t.clearAllSelections()
	} else {
		t.selectAllEnabled()
	}

	t.applyDependencies()
}

// isAllSelected проверяет, выбраны ли все элементы списка
func (t *MultiSelectTask) isAllSelected() bool {
	enabledCount := 0
	selectedCount := 0
	for i := range t.items {
		if t.isDisabled(i) {
			continue
		}
		enabledCount++
		if t.isSelected(i) {
			selectedCount++
		}
	}

	if enabledCount == 0 {
		return false
	}

	return enabledCount == selectedCount
}

// toggleSelection переключает состояние выбора элемента по индексу (оптимизировано для embedded)
func (t *MultiSelectTask) toggleSelection(index int) {
	if t.isDisabled(index) {
		return
	}

	if len(t.items) > 32 && t.fallbackMap != nil {
		// Используем fallback карту для больших списков
		if _, exists := t.fallbackMap[index]; exists {
			delete(t.fallbackMap, index)
		} else {
			t.fallbackMap[index] = struct{}{}
		}
	} else {
		// Используем битсет для оптимизации
		t.selected.Toggle(index)
	}

	t.applyDependencies()
}

func (t *MultiSelectTask) applyDependencies() {
	if len(t.dependencies) == 0 {
		t.resetDynamicDisabled()
		t.rebuildDisabled()
		t.clearDisabledSelections()
		t.ensureCursorSelectable()
		t.updateViewport()
		return
	}

	t.resetDynamicDisabled()
	maxIterations := len(t.items)*4 + 10
	if maxIterations < 16 {
		maxIterations = 16
	}
	for iteration := 0; iteration < maxIterations; iteration++ {
		stateChanged := false
		for idx, rule := range t.dependencies {
			var actions dependencyAction
			if t.isSelectedRaw(idx) {
				actions = rule.onSelect
			} else {
				actions = rule.onDeselect
			}

			for _, target := range actions.disable {
				if target < 0 || target >= len(t.items) {
					continue
				}
				if _, exists := t.dynamicDisabled[target]; !exists {
					t.dynamicDisabled[target] = struct{}{}
					stateChanged = true
				}
				if t.setSelectedState(target, false) {
					stateChanged = true
				}
			}

			for _, target := range actions.enable {
				if _, exists := t.dynamicDisabled[target]; exists {
					delete(t.dynamicDisabled, target)
					stateChanged = true
				}
			}

			for _, target := range actions.forceSelect {
				if t.setSelectedState(target, true) {
					stateChanged = true
				}
			}

			for _, target := range actions.forceClear {
				if t.setSelectedState(target, false) {
					stateChanged = true
				}
			}
		}

		if !stateChanged {
			break
		}
	}

	t.rebuildDisabled()
	t.clearDisabledSelections()
	t.ensureCursorSelectable()
	t.updateViewport()
}

// isSelected проверяет, выбран ли элемент по индексу (оптимизировано для embedded)
func (t *MultiSelectTask) isSelected(index int) bool {
	if t.isDisabled(index) {
		return false
	}
	if len(t.items) > 32 && t.fallbackMap != nil {
		_, exists := t.fallbackMap[index]
		return exists
	}
	return t.selected.IsSet(index)
}

// stopTimeout останавливает таймер
func (t *MultiSelectTask) stopTimeout() {
	// Если таймер активен, останавливаем его
	if t.timeoutEnabled && t.timeoutManager != nil && t.timeoutManager.IsActive() {
		t.timeoutManager.StopTimeout()
		t.showTimeout = false
	}
}

// Update handles key presses for navigation and selection.
func (t *MultiSelectTask) Update(msg tea.Msg) (Task, tea.Cmd) {
	if t.done {
		return t, nil
	}

	switch msg := msg.(type) {
	// Обработка сообщения о тайм-ауте
	case TimeoutMsg:
		// Когда истекает тайм-аут, применяем значение по умолчанию (если есть)
		t.applyDefaultValue()
		return t, nil
	// Обработка периодического обновления для счетчика времени
	case TickMsg:
		// Если таймер активен, продолжаем обновления
		if t.timeoutEnabled && t.timeoutManager != nil && t.timeoutManager.IsActive() {
			return t, t.timeoutManager.StartTicker()
		}
		return t, nil
	case tea.KeyMsg:
		// Сбрасываем сообщение-подсказку при любом нажатии клавиш (кроме Enter)
		if msg.String() != "enter" {
			t.showHelpMessage = false
			t.helpMessage = ""
		}

		switch msg.String() {
		case "up", "k":
			t.stopTimeout()
			if t.moveCursorBackward() {
				// Обновляем viewport после изменения позиции курсора
				t.updateViewport()
			}
			return t, nil
		case "down", "j":
			t.stopTimeout()
			if t.moveCursorForward() {
				// Обновляем viewport после изменения позиции курсора
				t.updateViewport()
			}
			return t, nil
		case " ", "right", "Right":
			// При выборе останавливаем таймер
			t.stopTimeout()

			// В любом случае выполняем выбор/переключение
			if t.hasSelectAll && t.cursor == -1 {
				// Нажатие пробела на опции "Выбрать все"
				t.toggleSelectAll()
			} else if t.cursor >= 0 && !t.isDisabled(t.cursor) {
				// Обычная логика выбора для элементов списка
				t.toggleSelection(t.cursor)
			}
		case "q", "Q", "esc", "Esc", "ctrl+c", "Ctrl+C", "left", "Left":
			t.stopTimeout()
			return t.confirmSelection()

		case "enter":
			t.stopTimeout()
			return t.confirmSelection()
		}
		// После обработки клавиш возвращаем команду для продолжения тикера
		if t.timeoutEnabled && t.timeoutManager != nil && t.timeoutManager.IsActive() {
			return t, t.timeoutManager.StartTicker()
		}
	}

	// Запускаем таймер при первом обновлении, если он включен и еще не активен
	if t.timeoutEnabled && t.timeoutManager != nil && !t.timeoutManager.IsActive() {
		return t, t.timeoutManager.StartTickerAndTimeout()
	}

	return t, nil
}

// confirmSelection завершает задачу, имитируя поведение клавиши Enter.
func (t *MultiSelectTask) confirmSelection() (Task, tea.Cmd) {
	keys, names := t.collectSelectionSnapshot()
	if len(keys) == 0 {
		if t.requireSelection {
			// Если требуется выбор, показываем подсказку и остаемся активными
			t.showHelpMessage = true
			t.helpMessage = defaults.NeedSelectAtLeastOne
			return t, nil
		}
		// Пустой выбор разрешен — завершаем без ошибок
		t.done = true
		t.icon = ui.IconDone
		t.finalValue = defaults.DefaultSuccessLabel
		t.SetError(nil)
		t.showHelpMessage = false
		t.helpMessage = ""
		return t, nil
	}

	// Есть выбранные элементы — завершаем с их перечислением
	t.done = true
	t.icon = ui.IconDone
	t.finalValue = strings.Join(names, defaults.DefaultSeparator)
	t.SetError(nil)
	t.showHelpMessage = false
	t.helpMessage = ""
	return t, nil
}

// Run запускает задачу выбора
func (t *MultiSelectTask) Run() tea.Cmd {
	// Запускаем таймер и тикер, если они включены
	if t.timeoutEnabled && t.timeoutManager != nil {
		return t.timeoutManager.StartTickerAndTimeout()
	}
	return nil
}

// applyDefaultValue применяет значение по умолчанию при истечении таймера
func (t *MultiSelectTask) applyDefaultValue() {
	// Если есть значение по умолчанию
	if t.defaultValue != nil {
		switch val := t.defaultValue.(type) {
		case []int:
			// Если это список индексов для выбора
			for _, index := range val {
				// Выбираем только корректные индексы
				if index >= 0 && index < len(t.items) && !t.isDisabled(index) {
					// Устанавливаем выбор для этого элемента
					if len(t.items) > 32 && t.fallbackMap != nil {
						t.fallbackMap[index] = struct{}{}
					} else {
						t.selected.Set(index)
					}
				}
			}
		case []string:
			// Если это список строк для выбора
			for _, strVal := range val {
				if idx := t.choiceIndex(strVal); idx != -1 && !t.isDisabled(idx) {
					if len(t.items) > 32 && t.fallbackMap != nil {
						t.fallbackMap[idx] = struct{}{}
					} else {
						t.selected.Set(idx)
					}
				}
			}
		}

		t.applyDependencies()

		// Проверяем, есть ли хотя бы один выбранный элемент
		hasSelection := false
		if len(t.items) > 32 && t.fallbackMap != nil {
			hasSelection = len(t.fallbackMap) > 0
		} else {
			hasSelection = t.selected.Count() > 0
		}

		// Если есть выбранные элементы, завершаем задачу
		if hasSelection {
			// Собираем выбранные элементы
			_, names := t.collectSelectionSnapshot()
			// Завершаем задачу
			t.done = true
			t.icon = ui.IconDone
			t.finalValue = strings.Join(names, defaults.DefaultSeparator)
			t.SetError(nil)
		}
	}
}

// View отрисовывает список вариантов выбора для пользователя с выделением активного элемента.
//
// @param width Ширина макета для отображения
// @return Строка с отформатированным представлением задачи
func (t *MultiSelectTask) View(width int) string {
	// Если задача завершена, возвращаем FinalView
	if t.done {
		return t.FinalView(width)
	}

	var sb strings.Builder
	// Заголовок задачи с префиксом активной задачи (поддерживает кастомные префиксы)
	titlePrefix := t.InProgressPrefix()

	// Получаем отформатированный таймер (если он активен)
	timerStr := t.RenderTimer()

	header := renderTitleWithWrap(titlePrefix, t.title, ui.ActiveTaskStyle, timerStr, width)
	sb.WriteString(header)
	sb.WriteString("\n")

	// Добавляем разделительную линию
	sb.WriteString(renderSelectionSeparator(width, t.showSelectionSeparator, titlePrefix))

	// Получаем диапазон видимых элементов с учетом viewport
	startIdx, endIdx, showSelectAll := t.getVisibleRange()

	// Отображаем опцию "Выбрать все" если она включена и видима в viewport
	if showSelectAll {
		checked := " "
		var itemPrefix string
		displayText := t.selectAllEnableText

		// Проверяем, выбраны ли все элементы
		if t.isAllSelected() {
			checked = ui.IconSelected
			displayText = t.selectAllDisableText
		}

		// Определяем префикс для опции "Выбрать все"
		if t.cursor == -1 {
			// Опция "Выбрать все" активна
			itemPrefix = ui.GetSelectItemPrefix("active")
		} else {
			// Опция "Выбрать все" не активна - всегда показываем префикс "above"
			// потому что курсор находится ниже неё (на элементах списка)
			itemPrefix = ui.GetSelectItemPrefix("above")
		}

		// Применяем стиль отображения
		styleToApply := t.selectAllStyle
		displayText = styleToApply.Render(displayText)

		// Формируем строку для отображения опции "Выбрать все"
		sb.WriteString(fmt.Sprintf("%s[%s] %s\n", itemPrefix, checked, displayText))
	}

	// Добавляем индикатор прокрутки вверх, если есть скрытые элементы выше
	// При наличии пункта "Выбрать все" индикатор должен показываться даже если startIdx == 0
	if t.viewportSize > 0 && (startIdx > 0 || (t.hasSelectAll && t.viewportStart > 0)) {
		// Используем точно такой же префикс как у элементов "above"
		indentPrefix := ui.GetSelectItemPrefix("above")
		// Определяем количество элементов выше
		itemsAbove := startIdx
		if t.hasSelectAll && t.viewportStart > 0 {
			// Если есть пункт "Выбрать все" и он скрыт, добавляем +1 к счетчику
			itemsAbove = t.viewportStart
		}
		var indicator string
		if t.showCounters {
			arrow := ui.UpArrowSymbol + " "
			indicator = fmt.Sprintf(defaults.ScrollAboveFormat, indentPrefix, arrow, itemsAbove)
		} else {
			indicator = fmt.Sprintf("%s %s", indentPrefix, ui.UpArrowSymbol)
		}
		// Не добавляем перенос строки в самой строке, чтобы не нарушать форматирование
		appendIndicatorWithPlainPipe(&sb, indicator)
		// Добавляем перенос строки отдельно
		sb.WriteString("\n")
	}

	activeHelp := ""

	// Отображаем только видимые элементы списка
	for i := startIdx; i < endIdx; i++ {
		if i >= len(t.items) {
			break
		}

		item := t.items[i]
		label := item.displayName()
		description := item.helpText()
		if item.isDivider() {
			var itemPrefix string
			switch {
			case t.cursor == i:
				itemPrefix = ui.GetSelectItemPrefix("active")
			case t.hasSelectAll && t.cursor == -1:
				itemPrefix = ui.GetSelectItemPrefix("below")
			case i < t.cursor:
				itemPrefix = ui.GetSelectItemPrefix("above")
			default:
				itemPrefix = ui.GetSelectItemPrefix("below")
			}
			styledLabel := ui.DisabledStyle.Render(label)
			sb.WriteString(fmt.Sprintf("%s%s\n", itemPrefix, styledLabel))
			if t.cursor == i && strings.TrimSpace(description) != "" {
				activeHelp = formatItemDescriptionText(description, width)
			}
			continue
		}
		checked := " "
		var itemPrefix string
		itemDisabled := t.isDisabled(i)
		isExit := isExitChoice(item)
		isBack := !isExit && isBackChoice(item)

		if t.isSelected(i) {
			checked = ui.IconSelected
		}

		if itemDisabled {
			label = ui.DisabledStyle.Render(label)
			checked = ui.DisabledStyle.Render(checked)
		}
		if !itemDisabled && t.cursor != i {
			switch {
			case isExit:
				label = ui.MenuExitItemStyle.Render(label)
			case isBack:
				label = ui.MenuBackItemStyle.Render(label)
			}
		}

		if t.cursor == i {
			itemPrefix = ui.GetSelectItemPrefix("active")
			label = t.activeStyle.Render(label)
		} else if t.hasSelectAll && t.cursor == -1 {
			itemPrefix = ui.GetSelectItemPrefix("below")
		} else if i < t.cursor {
			itemPrefix = ui.GetSelectItemPrefix("above")
		} else {
			itemPrefix = ui.GetSelectItemPrefix("below")
		}

		openBracket := "["
		closeBracket := "]"
		if itemDisabled {
			openBracket = ui.DisabledStyle.Render(openBracket)
			closeBracket = ui.DisabledStyle.Render(closeBracket)
		}
		sb.WriteString(fmt.Sprintf("%s%s%s%s %s\n", itemPrefix, openBracket, checked, closeBracket, label))

		if t.cursor == i && strings.TrimSpace(description) != "" {
			activeHelp = formatItemDescriptionText(description, width)
		}
	}

	// Добавляем индикатор прокрутки вниз, если есть скрытые элементы ниже
	if t.viewportSize > 0 && endIdx < len(t.items) {
		// Используем точно такой же префикс как у элементов "below"
		indentPrefix := ui.GetSelectItemPrefix("below")
		// Не добавляем перенос строки в конце, чтобы не нарушать форматирование
		remaining := len(t.items) - endIdx
		var indicator string
		if t.showCounters {
			arrow := ui.DownArrowSymbol + " "
			indicator = fmt.Sprintf(defaults.ScrollBelowFormat, indentPrefix, arrow, remaining)
		} else {
			indicator = fmt.Sprintf("%s %s", indentPrefix, ui.DownArrowSymbol)
		}
		appendIndicatorWithPlainPipe(&sb, indicator)
		// Добавляем перенос строки отдельно
		sb.WriteString("\n")
	}

	// Формируем отступ для подсказки
	helpIndent := performance.RepeatEfficient(" ", ui.MainLeftIndent)

	// Новая строка для подсказки
	var helpLine string
	if activeHelp != "" {
		helpLine = "\n"
	}

	// Добавляем сообщение-подсказку если нужно
	var warning string
	if t.showHelpMessage && t.helpMessage != "" {
		activeHelp = ""
		helpLine = ""
		warning = ui.GetErrorMessageStyle().Render(fmt.Sprintf("%s%s", helpIndent, t.helpMessage))
	}

	// Если есть опция "Выбрать все", добавляем её в подсказку
	helpText := defaults.MultiSelectHelp
	if t.hasSelectAll {
		helpText = defaults.MultiSelectHelpSelectAll
	}
	// Добавляем разделительную линию
	sb.WriteString("\n" + ui.DrawLine(width))
	// Добавляем сообщение-подсказку если нужно
	if warning != "" {
		sb.WriteString(warning + "\n")
	}
	// Если есть активный элемент, добавляем его подсказку
	if activeHelp != "" {
		sb.WriteString(ui.HelpTextStyle.Render(indentLines(activeHelp, helpIndent)))
	}
	// Добавляем подсказку
	if helpLine != "" {
		sb.WriteString(helpLine)
	}
	formattedHelp := indentLines(formatNavigationHelpText(helpText, width), helpIndent)
	sb.WriteString(ui.SubtleStyle.Render(formattedHelp))

	return sb.String()
}

func (t *MultiSelectTask) FinalView(width int) string {
	// Получаем базовое финальное представление
	result := t.BaseTask.FinalView(width)

	// Если задача завершилась успешно и есть дополнительные строки для вывода
	if t.icon == ui.IconDone && len(t.items) > 0 {
		_, names := t.collectSelectionSnapshot()
		if len(names) > 0 {
			result += "\n"
			for _, value := range names {
				result += ui.DrawSummaryLine(value, width)
			}
		}
	}

	return result
}

// WithDefaultOptions устанавливает варианты по умолчанию, которые будут выбраны при тайм-ауте.
// @param defauiltOptions Варианты по умолчанию (список индексов или строк)
// @param timeout Длительность тайм-аута
// @return Указатель на задачу для цепочки вызовов
func (t *MultiSelectTask) WithDefaultOptions(defauiltOptions interface{}, timeout time.Duration) *MultiSelectTask {
	t.WithTimeout(timeout, defauiltOptions)
	return t
}

// WithTimeout устанавливает тайм-аут для задачи множественного выбора
// @param duration Длительность тайм-аута
// @param defaultValue Значение по умолчанию (список индексов или строк)
// @return Указатель на задачу для цепочки вызовов
func (t *MultiSelectTask) WithTimeout(duration time.Duration, defaultValue interface{}) *MultiSelectTask {
	t.BaseTask.WithTimeout(duration, defaultValue)
	return t
}

// GetSelected возвращает список выбранных элементов
//
// @return []string Список выбранных пользователем вариантов
func (t *MultiSelectTask) GetSelected() []string {
	keys, _ := t.collectSelectionSnapshot()
	return keys
}

func (t *MultiSelectTask) collectSelectionSnapshot() ([]string, []string) {
	var keys []string
	var names []string
	for i := range t.items {
		if t.isSelected(i) {
			item := t.items[i]
			keys = append(keys, item.valueKey())
			names = append(names, item.displayName())
		}
	}
	return keys, names
}
