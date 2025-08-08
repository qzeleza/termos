# Термос - текстовый интерфейс для работы на малых маршрутизаторах и устройствах, написанный в виде библиотеки на Go

**Термос** - это мощный и гибкий фреймворк для создания терминальных пользовательских интерфейсов (TUI) на Go. Он построен поверх библиотеки [Bubble Tea](https://github.com/charmbracelet/bubbletea) и предоставляет высокоуровневые компоненты для создания интерактивных консольных приложений.

## ✨ Особенности

- 🎯 **Готовые компоненты** - множественный выбор, одиночный выбор, ввод текста, да/нет
- 🎨 **Гибкая стилизация** - поддержка цветов, стилей и тем оформления
- 📱 **Адаптивность** - оптимизация для embedded устройств с ограниченными ресурсами
- 🔄 **Очереди задач** - система управления последовательностью задач
- ✅ **Валидация** - встроенная система валидации пользовательского ввода
- 🎭 **Анимации** - поддержка анимированных переходов и эффектов
- 🧠 **Умная память** - оптимизация использования памяти для встроенных систем

## 🚀 Быстрый старт

### Установка

```bash
go get github.com/qzeleza/termos
```

### Простой пример (Bubble Tea)

```go
package main

import (
    "fmt"
    "log"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/qzeleza/termos/common"
    "github.com/qzeleza/termos/task"
)

type model struct{ t common.Task }

func (m model) Init() tea.Cmd { return m.t.Run() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.t, cmd = m.t.Update(msg)
    if m.t.IsDone() {
        return m, tea.Quit
    }
    return m, cmd
}

func (m model) View() string {
    // Используем разумную ширину для отрисовки
    w := common.CalculateLayoutWidth(80)
    return m.t.View(w)
}

func main() {
    yesNo := task.NewYesNoTask("Хотите продолжить?", "Подтвердите ваше действие")
    if _, err := tea.NewProgram(model{t: yesNo}).Run(); err != nil {
        log.Fatal(err)
    }
    if yesNo.GetChoice() {
        fmt.Println("Пользователь выбрал: Да")
    } else {
        fmt.Println("Пользователь выбрал: Нет")
    }
}
```

## 📦 Компоненты

### Задачи (Tasks)

- **YesNoTask** - выбор "да/нет"
- **SingleSelectTask** - выбор одного элемента из списка
- **MultiSelectTask** - выбор нескольких элементов из списка  
- **InputTaskNew** - ввод текста с валидацией

### Очереди (планируется)

- Оркестрация последовательных задач и сбор статистики в пакете `query/` (в разработке)

### UI компоненты

- **Styles** - система стилизации
- **Colors** - управление цветовой схемой
- **Layout** - компоненты разметки

## 🎨 Кастомизация

### Стили

```go
import "github.com/qzeleza/termos/ui"

// Настройка пользовательских стилей
styles := ui.GetDefaultStyles()
styles.Title = styles.Title.Foreground(lipgloss.Color("#ff6b6b"))
styles.Selected = styles.Selected.Background(lipgloss.Color("#4ecdc4"))
```

### Embedded оптимизации

```go
// 1) Ограничивайте ширину рендера
import "github.com/qzeleza/termos/common"
w := common.CalculateLayoutWidth(terminalWidth)

// 2) Переиспользуйте стили вместо пересоздания в каждом кадре
import "github.com/qzeleza/termos/ui"
ui.TitleStyle = ui.TitleStyle.Foreground(ui.ColorBrightGreen)

// 3) Используйте пулы для сборки строк
import "github.com/qzeleza/termos/performance"
b := performance.GetBuffer()
defer performance.PutBuffer(b)
b.WriteString("...быстрая сборка строки...")
```

## 📚 Документация

- [Руководство разработчика](docs/DEVDOC.md)
- [Примеры использования](docs/EXAMPLES.md)
- [Горячие клавиши](docs/HOTKEYS.md)
- [Настройка стилей](docs/STYLING.md)
- [Оптимизация для embedded](docs/EMBEDDED_OPTIMIZATION_REPORT.md)
- [Решение проблем](docs/TROUBLESHOOTING.md)

## 🏗️ Архитектура

```
termos/
├── task/           # Основные компоненты задач
├── query/          # Система очередей
├── ui/             # UI компоненты и стили  
├── validation/     # Система валидации
├── performance/    # Оптимизации производительности
├── common/         # Общие утилиты
├── errors/         # Обработка ошибок
└── docs/           # Документация
```

## 🎯 Применение

Termos идеально подходит для:

- **CLI утилиты** - интерактивные консольные приложения
- **Инсталляторы** - пошаговые установщики и настройщики
- **Системные утилиты** - управление конфигурациями и сервисами
- **Embedded устройства** - роутеры, IoT устройства с ограниченными ресурсами
- **DevOps инструменты** - развертывание и мониторинг

## 🤝 Совместимость

- **Go версия:** 1.22+
- **Платформы:** Linux, macOS, Windows
- **Терминалы:** все современные терминалы с поддержкой ANSI
- **Embedded:** OpenWrt, Entware, роутеры с ≥32MB RAM

## 📊 Производительность

- **Память:** оптимизировано для работы с ≥16MB RAM
- **CPU:** эффективная работа на ARM, MIPS, x86 архитектурах
- **Анимации:** адаптивная частота кадров в зависимости от мощности системы

## 🐛 Отчеты об ошибках

Если вы нашли ошибку или хотите предложить улучшение:

1. Проверьте [существующие issue](https://github.com/qzeleza/termos/issues)
2. Создайте новый issue с подробным описанием
3. Приложите минимальный пример воспроизведения

## 📄 Лицензия

MIT License - см. файл [LICENSE](LICENSE) для подробностей.

## 🙏 Благодарности

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - основа для TUI
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - стилизация
- Сообщество Go разработчиков

---

**Создано с ❤️ для сообщества Go разработчиков**