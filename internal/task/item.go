package task

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// / Item описывает элемент списка для задач выбора.
type Item struct {
	Key         string
	Name        string
	Description string
}

const (
	singleSelectDividerKey   = "__ziva_single_select_divider__"
	singleSelectDividerLabel = "├───────────────"
)

var (
	// SingleSelectDividerItem предоставляет стандартный элемент-разделитель
	// для списков выбора (single/multi). Он автоматически распознаётся задачами
	// и отображается как недоступный разделитель.
	SingleSelectDividerItem = Item{
		Key:         singleSelectDividerKey,
		Name:        singleSelectDividerLabel,
		Description: "",
	}
)

type choice struct {
	key         string
	name        string
	description string
	divider     bool
}

func (c choice) displayName() string {
	return c.name
}

func (c choice) helpText() string {
	return c.description
}

func (c choice) valueKey() string {
	return c.key
}

func (c choice) isDivider() bool {
	return c.divider
}

// / normalizeItems подготавливает элементы к использованию внутри задач.
func normalizeItems(source []Item) []choice {
	normalized := make([]choice, len(source))
	for i, it := range source {
		key := strings.TrimSpace(it.Key)
		name := strings.TrimSpace(it.Name)
		if key == "" && name != "" {
			key = name
		}
		if name == "" && key != "" {
			name = key
		}
		if key == "" && name == "" {
			key = fmt.Sprintf("item_%d", i+1)
			name = key
		}
		desc := strings.TrimSpace(it.Description)
		isDivider := isDividerChoice(key, name)
		normalized[i] = choice{
			key:         key,
			name:        name,
			description: desc,
			divider:     isDivider,
		}
	}
	applyDividerLabel(normalized)
	return normalized
}

func isDividerChoice(key, name string) bool {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey != "" && strings.EqualFold(trimmedKey, singleSelectDividerKey) {
		return true
	}
	trimmedName := strings.TrimSpace(name)
	return trimmedName == singleSelectDividerLabel
}

func dynamicDividerLabel(items []choice) string {
	maxLen := 0
	for _, item := range items {
		if item.isDivider() {
			continue
		}
		trimmed := strings.TrimSpace(item.displayName())
		if trimmed == "" {
			continue
		}
		length := utf8.RuneCountInString(trimmed)
		if length > maxLen {
			maxLen = length
		}
	}

	if maxLen == 0 {
		defaultLen := utf8.RuneCountInString(singleSelectDividerLabel)
		if defaultLen > 0 {
			maxLen = defaultLen - 1
		}
		if maxLen < 1 {
			maxLen = 3
		}
	}

	return strings.Repeat("─", maxLen+5)
}

func applyDividerLabel(choices []choice) {
	label := dynamicDividerLabel(choices)
	for i := range choices {
		if choices[i].isDivider() {
			choices[i].name = label
		}
	}
}
