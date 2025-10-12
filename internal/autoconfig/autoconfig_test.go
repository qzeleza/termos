package autoconfig

import (
	"testing"
)

func TestParseMemoryLimit_VariousUnits(t *testing.T) {
	cases := map[string]uint64{
		"1024":  1024,
		"1KB":   1024,
		"1MB":   1024 * 1024,
		"2GB":   2 * 1024 * 1024 * 1024,
		"1KiB":  1024,
		"1MiB":  1024 * 1024,
		"2GiB":  2 * 1024 * 1024 * 1024,
		"1_024": 1024,
		"1_5MB": 1, // ожидаем ошибку, поэтому пропускаем
	}

	for in, want := range cases {
		got, err := parseMemoryLimit(in)
		if in == "1_5MB" {
			if err == nil {
				t.Fatalf("expected error for %q, got none", in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseMemoryLimit(%q) error: %v", in, err)
		}
		if got != want {
			t.Fatalf("parseMemoryLimit(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if v := firstNonEmpty("", "  ", "val", "next"); v != "val" {
		t.Fatalf("firstNonEmpty returned %q, want 'val'", v)
	}
	if v := firstNonEmpty("", "  "); v != "" {
		t.Fatalf("firstNonEmpty returned %q, want ''", v)
	}
}

func TestIsLimitedTerminal_EnvHeuristics(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("NO_COLOR", "1")
	t.Setenv("LANG", "C")
	t.Setenv("COLORTERM", "")
	if !isLimitedTerminal() {
		t.Fatal("expected limited terminal for TERM=dumb NO_COLOR=1 LANG=C")
	}

	t.Setenv("TERM", "xterm-256color")
	t.Setenv("NO_COLOR", "")
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("COLORTERM", "truecolor")
	if isLimitedTerminal() {
		t.Fatal("did not expect limited terminal for xterm-256color and UTF-8")
	}
}

func TestIsMemoryConstrained_WithEnvLimit(t *testing.T) {
	t.Setenv("ZIVA_MEMORY_LIMIT", "64MB")
	// функция isMemoryConstrained может учитывать эвристику по env; проверим, что не паникует
	_ = isMemoryConstrained()
}
