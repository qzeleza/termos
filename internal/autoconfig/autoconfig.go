// Package autoconfig — автоматическая детекция окружения и применение embedded-конфигурации
package autoconfig

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"github.com/qzeleza/ziva/internal/ui"
)

// init вызывается при импортировании пакета и запускает авто-конфигурацию
func init() { AutoConfigure() }

// AutoConfigure выполняет авто-детекцию embedded-окружения и применяет настройки по умолчанию
var once sync.Once

func AutoConfigure() {
	once.Do(func() {
		if isEmbeddedEnvironment() {
			applyEmbeddedDefaults()
		}
	})
}

// applyEmbeddedDefaults включает минимально инвазивные embedded-настройки ядра
// На текущем этапе это переключение упрощённой палитры и иконок для терминалов
// с ограниченной поддержкой цветов/UTF-8. По мере необходимости сюда можно
// добавить другие безопасные оптимизации по умолчанию.
func applyEmbeddedDefaults() {
	// 1) Упрощенная палитра/иконки
	ui.EnableEmbeddedMode()

	// 2) ASCII‑фолбэк для действительно ограниченных терминалов
	if isLimitedTerminal() {
		// Включаем ASCII-иконки и максимально совместимый рендер
		ui.EnableASCIIMode()
	}

	// 3) Консервативный лимит памяти рантайма
	// Терминология: предпочитаем явный лимит из ZIVA_MEMORY_LIMIT или GOMEMLIMIT.
	if limitStr := firstNonEmpty(os.Getenv("GOMEMLIMIT"), os.Getenv("ZIVA_MEMORY_LIMIT")); limitStr != "" {
		if bytes, err := parseMemoryLimit(limitStr); err == nil {
			debug.SetMemoryLimit(int64(bytes))
		}
	} else {
		// По умолчанию 256MB для embedded-профиля — мягкая рекомендация
		const defauiltLimit = 256 * 1024 * 1024
		debug.SetMemoryLimit(defauiltLimit)
	}
}

// isEmbeddedEnvironment автоматически определяет, является ли окружение embedded
func isEmbeddedEnvironment() bool {
	// Принудительная установка через переменную окружения
	if embedded := os.Getenv("ZIVA_EMBEDDED"); embedded != "" {
		v := strings.TrimSpace(strings.ToLower(embedded))
		return v == "1" || v == "true" || v == "yes" || v == "on"
	}

	// Проверяем, не является ли это современным десктопным терминалом
	if isModernDesktopTerminal() {
		return false
	}

	// Автоматическое определение по набору эвристик
	return isMemoryConstrained() ||
		isLimitedTerminal() ||
		isKnownEmbeddedEnvironment()
}

// isMemoryConstrained проверяет ограничения памяти
func isMemoryConstrained() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Принудительный лимит из переменной окружения (например, 64MB, 128KB, 1GB)
	if limitStr := os.Getenv("ZIVA_MEMORY_LIMIT"); limitStr != "" {
		if limit, err := parseMemoryLimit(limitStr); err == nil {
			return m.Sys < limit
		}
	}

	// Более консервативный порог для embedded устройств - только действительно ограниченные системы
	const memoryThreshold = 128 * 1024 * 1024 // 128MB (вместо 512MB)
	return m.Sys < memoryThreshold
}

// parseMemoryLimit парсит строку с лимитом памяти (например "64MB", "128KB", "1GB")
func parseMemoryLimit(limitStr string) (uint64, error) {
	s := strings.TrimSpace(limitStr)
	s = strings.ToUpper(s)

	var multiplier uint64 = 1
	var numStr string

	// Поддерживаем B, KB/MB/GB, KiB/MiB/GiB
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
	cleanNum, err := normalizeNumberString(numStr)
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseUint(cleanNum, 10, 64)
	if err != nil {
		return 0, err
	}
	return num * multiplier, nil
}

func normalizeNumberString(num string) (string, error) {
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

// firstNonEmpty возвращает первый непустой (после TrimSpace) элемент
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// isLimitedTerminal проверяет ограничения терминала и локали
func isLimitedTerminal() bool {
	term := strings.ToLower(os.Getenv("TERM"))

	// Уважать NO_COLOR спеку — считать терминал ограниченным
	if os.Getenv("NO_COLOR") != "" {
		return true
	}

	// Принудительный ASCII‑режим
	if v := strings.TrimSpace(strings.ToLower(os.Getenv("ZIVA_ASCII_ONLY"))); v == "1" || v == "true" || v == "yes" || v == "on" {
		return true
	}

	// Распространённые ограниченные терминалы
	if term == "linux" || term == "console" || term == "vt100" || term == "vt102" || term == "dumb" {
		return true
	}

	// Проверяем поддержку UTF‑8
	lang := strings.ToLower(os.Getenv("LANG"))
	lcAll := strings.ToLower(os.Getenv("LC_ALL"))
	lcCtype := strings.ToLower(os.Getenv("LC_CTYPE"))

	utf := strings.Contains(lang, "utf") || strings.Contains(lcAll, "utf") || strings.Contains(lcCtype, "utf")

	// COLORTERM часто присутствует на современных терминалах
	colorTerm := os.Getenv("COLORTERM") != ""

	return !utf || !colorTerm
}

// isModernDesktopTerminal проверяет, является ли терминал современным десктопным
func isModernDesktopTerminal() bool {
	term := strings.ToLower(os.Getenv("TERM"))
	colorTerm := os.Getenv("COLORTERM")

	// Современные терминалы с полной поддержкой цветов
	modernTerminals := []string{
		"xterm-256color", "screen-256color", "tmux-256color",
		"alacritty", "kitty", "iterm2", "vte", "gnome",
	}

	for _, modernTerm := range modernTerminals {
		if strings.Contains(term, modernTerm) {
			return true
		}
	}

	// COLORTERM=truecolor указывает на 24-bit цвета
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return true
	}

	return false
}

// isKnownEmbeddedEnvironment проверяет известные признаки embedded-окружений
func isKnownEmbeddedEnvironment() bool {
	// Характерные файлы/пути OpenWrt/Entware и др.
	embeddedPaths := []string{
		"/opt/etc/init.d",      // Entware
		"/etc/openwrt_release", // OpenWrt
	}
	for _, path := range embeddedPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Проверяем архитектуру CPU
	return isEmbeddedCPU()
}

// isEmbeddedCPU проверяет архитектуру CPU для embedded устройств
func isEmbeddedCPU() bool {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return false
	}
	cpuInfo := strings.ToLower(string(data))

	// Характерные признаки embedded процессоров
	embeddedCPUs := []string{
		"arm", "mips", "aarch64", "armv7", "cortex",
		"mediatek", "qualcomm", "broadcom", "rockchip",
	}
	for _, cpu := range embeddedCPUs {
		if strings.Contains(cpuInfo, cpu) {
			return true
		}
	}
	return false
}
