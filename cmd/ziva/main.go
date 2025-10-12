package main

import (
	// Встроенные импорты не требуются

	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/qzeleza/ziva"
	"github.com/qzeleza/ziva/internal/common"
	"github.com/qzeleza/ziva/internal/defaults"
	"github.com/qzeleza/ziva/internal/task"
)

var (
	memoryLimitDemoFlag   = flag.String("memory-limit", "", "строка лимита памяти для демонстрации строгого парсинга (например, 512MB, 1_5GB)")
	summaryWidthDemoFlag  = flag.Int("demo-summary-width", 60, "ширина в символах для предпросмотра форматирования итоговой сводки")
	summaryTextDemoFlag   = flag.String("demo-summary-text", "", "кастомный текст сводки для демонстрации переноса строк")
	showSummaryPreviewFlg = flag.Bool("demo-summary-preview", true, "выводить предпросмотр форматирования итоговой сводки в лог перед запуском TUI")
)

// pingResult представляет результат проверки подключения
type pingResult struct {
	Ping   string
	Loss   string
	Status string
}

func main() {
	// Настраиваем язык интерфейса и предупреждаем о возможных ограничениях терминала
	activeLang := configureLanguage()
	warnTerminalCapabilities(activeLang)
	ziva.SetDefaultLanguage("ru")

	// Демонстрация строгого парсинга лимитов памяти
	if limit := strings.TrimSpace(*memoryLimitDemoFlag); limit != "" {
		if bytes, err := parseMemoryLimitForDemo(limit); err != nil {
			log.Printf("⚙️  Парсер лимитов памяти: значение %q отклонено: %v", limit, err)
		} else {
			log.Printf("⚙️  Парсер лимитов памяти: значение %q принято (%d байт)", limit, bytes)
		}
	}

	// Настраиваем текст сводки для демонстрации переноса по ширине
	defaultSummary := "Итог развертывания: проверены все подсистемы и подготовлены рекомендации для последующих шагов.\nПожалуйста, просмотрите сводку, чтобы увидеть форматирование длинного текста."
	if customText := strings.TrimSpace(*summaryTextDemoFlag); customText != "" {
		defaultSummary = customText
	}
	defaults.SummaryCompleted = defaultSummary

	if *showSummaryPreviewFlg {
		layoutWidth := ziva.CalculateLayoutWidth(*summaryWidthDemoFlag)
		log.Printf("🛈 Предпросмотр сводки (ширина %d):", layoutWidth)
		for _, line := range previewSummaryLines(defaultSummary, layoutWidth) {
			log.Printf("   %s", line)
		}
	}

	// Демонстрируем управление стилями для пунктов «Назад» и «Выход»
	ziva.SetExitMenuItemStyle()                            // подсветка без аргументов
	ziva.SetBackMenuItemStyle(ziva.MenuItemDefaultStyle()) // возврат оставляем белым
	log.Println("🛈 Стиль меню: «Выход» подсвечен ярким цианом, «Назад» оставлен в базовом стиле для сравнения.")

	// Заголовок и краткое описание для TUI
	header := "Демонстрация всех типов задач Жива™"

	// Формируем очередь задач
	const (
		componentCLI        = "cli"
		componentServer     = "server"
		componentAgent      = "agent"
		componentWeb        = "web"
		componentDocs       = "docs"
		componentBuild      = "build"
		componentArtifacts  = "artifacts"
		componentViewport   = "viewport"
		componentInput      = "input"
		componentMulti      = "multi"
		componentSingle     = "single"
		componentValidation = "validation"
		componentBack       = "back"
		componentExit       = "exit"
		diagnosticLogging   = "logging"
		diagnosticDebug     = "debug"
		diagnosticTracing   = "tracing"
		diagnosticMetrics   = "metrics"
	)

	msel := []ziva.Item{
		{Key: componentCLI, Name: "CLI", Description: "Командный интерфейс для администрирования"},
		{Key: componentServer, Name: "Сервер", Description: "Бэкенд сервисы"},
		{Key: componentAgent, Name: "Агент"},
		{Key: componentWeb, Name: "Web UI", Description: "Веб-интерфейс для пользователей"},
		{Key: componentDocs, Name: "Документация", Description: "Автоматическая генерация документации"},
		{Key: componentBuild, Name: "Компилировать", Description: "Сборка исполняемых файлов"},
		{Key: componentArtifacts, Name: "Выходные данные", Description: "Архивация результатов"},
		ziva.Divider,
		{Key: componentViewport, Name: "Область просмотра"},
		{Key: componentInput, Name: "Поле ввода", Description: "Пример текстовой задачи"},
		{Key: componentMulti, Name: "Мультивыбор", Description: "Дополнительные параметры"},
		ziva.Divider,
		{Key: componentSingle, Name: "Одиночный выбор", Description: "Переключатель режимов"},
		{Key: componentValidation, Name: "Проверка ввода", Description: "Встроенные валидаторы"},
		{Key: componentBack, Name: "Назад", Description: "Возврат к предыдущему шагу"},
		{Key: componentExit, Name: "Выйти", Description: "Завершить выбор"},
	}

	const (
		envDevelopment = "development"
		envStaging     = "staging"
		envProduction  = "production"
		envCustom      = "custom"
		envCancel      = "cancel"
		envBack        = "back"
		envExit        = "exit"
	)

	ssel := []ziva.Item{
		{Key: envDevelopment, Name: "development", Description: "Среда разработки"},
		{Key: envStaging, Name: "staging", Description: "Промежуточная среда"},
		{Key: envProduction, Name: "production", Description: "Боевая среда"},
		ziva.Divider,
		{Key: envCustom, Name: "другое", Description: "Пользовательское значение"},
		{Key: envCancel, Name: "отмена", Description: "Отмена выбора"},
		ziva.Divider,
		{Key: envBack, Name: "назад", Description: "Вернуться к предыдущему пункту"},
		{Key: envExit, Name: "выход", Description: "Выход из программы"},
	}

	// Создаем очередь и добавляем задачи
	queue := ziva.NewQueue(header)
	queue.WithAppName("Жива™", "версия 1.0.0")
	// queue.WithOutResultLine()
	// queue.WithTasksNumbered(false, "[%d]")

	// 1) Задачи мультивыбора (без и с пунктом "Выбрать все")
	//    Пример без "Выбрать все"
	ms1 := ziva.NewMultiSelectTask("Выберите компоненты установки Выберите компоненты установки Выберите компоненты установки Выберите компоненты установки Выберите компоненты установки Выберите компоненты установки Выберите компоненты установки Выберите компоненты установки", msel).
		WithViewport(15, false).
		WithTimeout(3*time.Second, []string{componentCLI, componentServer}).
		WithItemsDisabled([]string{componentAgent, componentWeb})

	diagnosticsItems := []ziva.Item{
		{Key: diagnosticLogging, Name: "Логирование", Description: "Писать события в файл"},
		{Key: diagnosticDebug, Name: "Отладка", Description: "Расширенный вывод отладки"},
		{Key: diagnosticTracing, Name: "Трассировка", Description: "Профилирование производительности"},
		{Key: diagnosticMetrics, Name: "Метрики", Description: "Экспортировать Prometheus показатели"},
	}

	diagnosticsTask := ziva.NewMultiSelectTask("Настройка диагностики", diagnosticsItems).
		WithViewport(15, true).
		WithDependencies(map[string]ziva.MultiSelectDependencyRule{
			diagnosticLogging: {
				OnSelect: ziva.MultiSelectDependencyActions{
					Enable: []string{diagnosticDebug},
				},
				OnDeselect: ziva.MultiSelectDependencyActions{
					Disable:    []string{diagnosticDebug},
					ForceClear: []string{diagnosticDebug, diagnosticTracing},
				},
			},
			diagnosticDebug: {
				OnSelect: ziva.MultiSelectDependencyActions{
					ForceSelect: []string{diagnosticLogging},
				},
			},
			diagnosticTracing: {
				OnSelect: ziva.MultiSelectDependencyActions{
					ForceSelect: []string{diagnosticLogging},
				},
			},
		}).
		WithSelectAll().
		WithDefaultItems([]string{diagnosticLogging, diagnosticMetrics})

	securityItems := []ziva.Item{
		{Key: "firewall", Name: "Firewall", Description: "Базовая сетевой фильтр"},
		{Key: "ids", Name: "IDS", Description: "Система обнаружения вторжений"},
		{Key: "waf", Name: "WAF", Description: "Веб-фильтр для приложений"},
	}

	securityTask := ziva.NewMultiSelectTask("Минимальные требования безопасности мумолчанию Оставьте пустым и нажмите → или Enter, чтобы принять значение по умолчанию.", securityItems).
		WithViewport(3, false).
		WithRequireSelection(true)

	// //    Пример с пунктом "Выбрать все"
	// ms2 := ziva.NewMultiSelectTask("Выберите модули для сборки", ssel).
	// 	WithViewport(3, false).
	// 	WithSelectAll("Выбрать все").
	// 	WithTimeout(10*time.Second, []string{envDevelopment, envStaging}).
	// 	WithDefaultItems([]string{envDevelopment, envStaging}).
	// 	WithItemsDisabled([]string{envProduction, envCustom})

	// 2) Одиночный выбор
	ss := ziva.NewSingleSelectTask(
		"Выберите среду развертывания Выберите компоненты установки Выберите компоненты установки Выберите компоненты установки Выберите компоненты установки Выберите компоненты установки",
		ssel,
	).WithViewport(9).
		// WithTimeout(3*time.Second, envStaging).
		WithDefaultItem(envProduction)

	// 3) Ввод с использованием всех стандартных валидаторов
	//    Валидация будет происходить в момент подтверждения (Enter)
	v := ziva.DefaultValidators

	// inPath := task.NewInputTaskNew("Путь к файлу/директории", "Введите путь:").
	// 	WithValidator(v.Path())

	// 4) Задача-выполнение функции (FuncTask)
	//    Выполняет полезную работу и выводит результат в финальном представлении задачи (без fmt.Print)
	errorTaskRun := true
	var fn *ziva.FuncTask
	if errorTaskRun {
		data := pingResult{}
		fn = ziva.NewFuncTask(
			"Проверка соединения умолчанию Оставьте пустым и нажмите → или Enter, чтобы принять значение по умолчанию.",
			func() error {
				// return checkConnection(&data)
				return errors.New("симуляция ошибки в середине выполнения очереди\nне ясная причина стимуляции проблемы\nдополнительная информация")
			},
			// Выводим краткую сводку под заголовком после успеха
			ziva.WithSummaryFunction(func() []string {
				return []string{
					"Пинг: " + data.Ping,
					"Потери пакетов: " + data.Loss,
				}
			}),
			// Не останавливать очередь при ошибке (для демонстрации поведения)
			ziva.WithStopOnError(false),
		)
		// queue.AddTasks(fn)
	}
	// 5) Подтверждение Да/Нет (например, для сохранения настроек)
	// Используем языко-независимый метод вместо строки "Да"
	ys := ziva.NewYesNoTask("Сохранение конфигурации мумолчанию Оставьте пустым и нажмите → или Enter, чтобы принять значение по умолчанию.", "Сохранить изменения? ").
		WithTimeoutYes(2 * time.Second)
	// ys.WithoutResultLine()
	ys.WithNoAsError()

	inRequired := task.NewInputTaskNew("Обязательное поле", "Нельзя оставлять пустым:").
		WithValidator(v.Required())

	inDefault := task.NewInputTaskNew(
		"Поле со значением по умолчанию значение по умолчанию Оставьте пустым и нажмите → или Enter, чтобы принять значение по умолчанию.",
		"Оставьте пустым и нажмите → или Enter, чтобы принять значение по умолчанию.",
	).WithValidator(v.Required()).
		WithTimeout(45*time.Second, "значение по умолчанию Оставьте пустым и нажмите → или Enter, чтобы принять значение по умолчанию.")
	inDefault.ShowTimeout(true)
	inDefault.WithPlaceholder("значение по умолчанию Оставьте пустым и нажмите → или Enter, чтобы принять значение по умолчанию.")

	queue.AddTasks(
		inDefault,
		fn,
		securityTask,
		ys,
		ms1,
		diagnosticsTask,
		inRequired,
	)

	// очередь-заглушка, сюда добавляем задачи, которые не ходим пока чтобы они выполнялись
	queue1 := ziva.NewQueue(header)
	queue1.AddTasks(
		ms1,
		diagnosticsTask,
		securityTask,
		inRequired,

		ys,
		fn,
		ss,
	)
	// inUsername := ziva.NewInputTask("Имя пользователя", "Введите username:").
	// 	WithValidator(v.Username()).
	// 	WithTimeout(10*time.Second, "Alex")

	// inEmail := task.NewInputTaskNew("Email", "Введите email:").
	// 	WithInputType(task.InputTypeEmail).WithValidator(v.Email()).
	// 	WithTimeout(10*time.Second, "defauilt@example.com")

	// inOptionalEmail := task.NewInputTaskNew("Доп. Email (опционально)", "Введите email или оставьте пустым:").
	// 	WithInputType(task.InputTypeEmail).WithValidator(v.OptionalEmail())

	// inURL := task.NewInputTaskNew("URL", "Введите URL (http/https):").
	// 	WithValidator(v.URL())

	// inPort := task.NewInputTaskNew("Порт", "Введите порт (1-65535):").
	// 	WithInputType(task.InputTypeNumber).WithValidator(v.Port())

	// inRange := task.NewInputTaskNew("Число в диапазоне", "Введите число [10..100]:").
	// 	WithInputType(task.InputTypeNumber).WithValidator(v.Range(10, 100))

	// inIPv4 := task.NewInputTaskNew("IPv4", "Введите IPv4 адрес:").
	// 	WithInputType(task.InputTypeIP).WithValidator(v.IPv4())

	// inIPv6 := task.NewInputTaskNew("IPv6", "Введите IPv6 адрес:").
	// 	WithInputType(task.InputTypeIP).WithValidator(v.IPv6())

	// inIPAny := task.NewInputTaskNew("IP (любой)", "Введите IP адрес:").
	// 	WithInputType(task.InputTypeIP).WithValidator(v.IP())

	// inDomain := task.NewInputTaskNew("Домен", "Введите доменное имя:").
	// 	WithInputType(task.InputTypeDomain).WithValidator(v.Domain())

	// inAlphaNum := task.NewInputTaskNew("Только буквы и цифры", "Введите значение:").
	// 	WithValidator(v.AlphaNumeric())

	// inMinLen := task.NewInputTaskNew("Мин. длина", "Минимум 5 символов:").
	// 	WithValidator(v.MinLength(5))

	// inMaxLen := task.NewInputTaskNew("Макс. длина", "Не более 10 символов:").
	// 	WithValidator(v.MaxLength(10))

	// inExactLen := task.NewInputTaskNew("Точная длина", "Ровно 8 символов:").
	// 	WithValidator(v.Length(8))

	// inStdPwd := task.NewInputTaskNew("Пароль (стандарт)", "Введите пароль (>=8):").
	// 	WithInputType(task.InputTypePassword).WithValidator(v.StandardPassword())

	// inStrongPwd := task.NewInputTaskNew("Пароль (сильный)", "Введите пароль (>=12):").
	// 	WithInputType(task.InputTypePassword).WithValidator(v.StrongPassword())

	// queue.AddTasks(
	// 	ss, ms1,
	// 	//  inEmail, inOptionalEmail,
	// 	// inPath, inURL, inPort, inRange,
	// 	// inIPv4, inIPv6, inIPAny, inDomain,
	// 	// inAlphaNum, inMinLen, inMaxLen, inExactLen,
	// 	// inStdPwd, inStrongPwd,
	// 	ys,
	// 	inRequired,
	// )

	// Запускаем TUI c очередью задач. Результаты отображаются внутри интерфейса;
	// дополнительный вывод через fmt.Print не используется.
	if err := queue.Run(); err != nil {
		// Обработка ошибки
		log.Fatalf("Ошибка при запуске очереди: %v", err)
	}
}

func parseMemoryLimitForDemo(limitStr string) (uint64, error) {
	s := strings.TrimSpace(limitStr)
	s = strings.ToUpper(s)

	var multiplier uint64 = 1
	var numStr string

	switch {
	case strings.HasSuffix(s, "B") && !strings.HasSuffix(s, "KB") && !strings.HasSuffix(s, "MB") && !strings.HasSuffix(s, "GB") && !strings.HasSuffix(s, "KIB") && !strings.HasSuffix(s, "MIB") && !strings.HasSuffix(s, "GIB"):
		multiplier = 1
		numStr = strings.TrimSuffix(s, "B")
	case strings.HasSuffix(s, "KIB"):
		multiplier = 1024
		numStr = strings.TrimSuffix(s, "KIB")
	case strings.HasSuffix(s, "MIB"):
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(s, "MIB")
	case strings.HasSuffix(s, "GIB"):
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "GIB")
	case strings.HasSuffix(s, "KB"):
		multiplier = 1024
		numStr = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "GB"):
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "GB")
	default:
		numStr = s
	}

	numStr = strings.TrimSpace(numStr)
	if numStr == "" {
		return 0, fmt.Errorf("empty number")
	}

	cleanNum, err := normalizeNumberForDemo(numStr)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseUint(cleanNum, 10, 64)
	if err != nil {
		return 0, err
	}
	return value * multiplier, nil
}

func normalizeNumberForDemo(num string) (string, error) {
	if !strings.Contains(num, "_") {
		return num, nil
	}

	parts := strings.Split(num, "_")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid numeric format")
	}

	for i, part := range parts {
		if part == "" {
			return "", fmt.Errorf("invalid numeric format")
		}
		if i == 0 {
			if len(part) == 0 || len(part) > 3 {
				return "", fmt.Errorf("invalid numeric format")
			}
		} else if len(part) != 3 {
			return "", fmt.Errorf("invalid numeric format")
		}
	}

	return strings.Join(parts, ""), nil
}

func previewSummaryLines(summary string, width int) []string {
	if width <= 0 {
		width = 60
	}

	effectiveWidth := width - common.LayoutWrapMargin
	if effectiveWidth < 1 {
		effectiveWidth = 1
	}

	text := strings.ReplaceAll(summary, "\n", " ")
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return []string{}
	}

	runes := []rune(text)
	var lines []string
	start := 0
	for start < len(runes) {
		end := start + effectiveWidth
		if end >= len(runes) {
			lines = append(lines, string(runes[start:]))
			break
		}

		cut := end
		for cut > start && runes[cut-1] != ' ' {
			cut--
		}
		if cut == start {
			cut = end
		}

		line := strings.TrimSpace(string(runes[start:cut]))
		if line != "" {
			lines = append(lines, line)
		}

		start = cut
		for start < len(runes) && runes[start] == ' ' {
			start++
		}
	}

	return lines
}

func checkConnection(result *pingResult) error {
	time.Sleep(2 * time.Second)
	cmd := exec.Command("ping", "-c", "1", "google.com")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Парсим результат команды ping
	outputStr := string(output)

	// Извлекаем время пинга
	if strings.Contains(outputStr, "time=") {
		parts := strings.Split(outputStr, "time=")
		if len(parts) > 1 {
			timePart := strings.Fields(parts[1])[0]
			result.Ping = timePart
		}
	}

	// Извлекаем потери пакетов
	if strings.Contains(outputStr, "packet loss") {
		parts := strings.Split(outputStr, " packet loss")
		if len(parts) > 0 {
			lossFields := strings.Fields(parts[0])
			if len(lossFields) > 0 {
				result.Loss = lossFields[len(lossFields)-1]
			}
		}
	}

	// Определяем статус на основе успешности команды
	if !strings.Contains(outputStr, "1 received") {
		result.Status = "FAILED"
	} else {
		result.Status = "OK"
	}

	return nil
}

// configureLanguage разбирает язык из CLI/конфига, проверяет доступность локали и применяет его.
func configureLanguage() string {
	if defLang := strings.TrimSpace(os.Getenv("ZIVA_DEFAULT_LANG")); defLang != "" {
		ziva.SetDefaultLanguage(defLang)
	}
	langFlag := flag.String("lang", "", "язык интерфейса Ziva (например, ru или en)")
	flag.Parse()

	lang := strings.TrimSpace(*langFlag)
	if lang == "" {
		lang = strings.TrimSpace(os.Getenv("ZIVA_LANG"))
	}
	if lang == "" {
		lang = "ru"
	}
	lang = strings.ToLower(lang)

	if strings.HasPrefix(lang, "ru") && !localeAvailable("ru_RU") {
		printRussianLocaleHint()
		lang = "en"
	}

	return ziva.SetLanguage(lang)
}

// localeAvailable проверяет присутствие заданной локали в системе.
func localeAvailable(locale string) bool {
	if strings.HasPrefix(locale, "en") {
		return true
	}
	cmd := exec.Command("locale", "-a")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	list := strings.Split(strings.ToLower(string(output)), "\n")
	targets := []string{"ru_ru.utf-8", "ru_ru.utf8", "ru_ru", "ru"}
	for _, line := range list {
		line = strings.TrimSpace(line)
		for _, target := range targets {
			if line == target {
				return true
			}
		}
	}
	return false
}

// printRussianLocaleHint выводит рекомендации по установке русской локали и шрифта.
func printRussianLocaleHint() {
	fmt.Println("⚠️ Не удалось найти локаль ru_RU.UTF-8. Переключаю интерфейс на английский.")
	fmt.Println("   Установите русскую локаль командой (Debian/Ubuntu): sudo locale-gen ru_RU.UTF-8 && sudo update-locale LANG=ru_RU.UTF-8")
	fmt.Println("   Для Entware/BusyBox: opkg install locale-full glibc-binary-locales && export LANG=ru_RU.UTF-8")
	fmt.Println("   При необходимости настройте шрифт: setterm -reset && setterm -store")
}

// warnTerminalCapabilities предупреждает о возможных ограничениях терминала.
func warnTerminalCapabilities(lang string) {
	langEnv := strings.ToLower(os.Getenv("LANG"))
	term := strings.ToLower(os.Getenv("TERM"))
	colorTerm := strings.TrimSpace(os.Getenv("COLORTERM"))

	if !strings.Contains(langEnv, "utf") {
		fmt.Println("⚠️ Текущая локаль не содержит UTF-8. Псевдографика может отображаться некорректно.")
		fmt.Println("   Совет: export LANG=ru_RU.UTF-8 && export LC_ALL=ru_RU.UTF-8")
	}
	if colorTerm == "" {
		fmt.Println("⚠️ Терминал не сообщает о поддержке цвета (COLORTERM пуст). Включите цветной режим или используйте современный эмулятор.")
	}
	if term == "linux" || strings.Contains(term, "vt100") || strings.Contains(term, "busybox") {
		fmt.Println("ℹ️ Для Entware/BusyBox установите шрифт UTF-8: setterm -store && setterm -font latarcyrheb-sun32")
	}
}
