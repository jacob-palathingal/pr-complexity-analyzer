package runner

import (
	"strings"
	"testing"
	"time"
)

func TestConfigWithDefaults(t *testing.T) {
	cfg := Config{
		BaseRef: "main",
	}

	cfg = cfg.withDefaults()

	if cfg.HeadRef != "HEAD" {
		t.Fatalf(
			"head ref = %q, want HEAD",
			cfg.HeadRef,
		)
	}

	if cfg.Format != "text" {
		t.Fatalf(
			"format = %q, want text",
			cfg.Format,
		)
	}

	if cfg.Timeout != defaultTimeout {
		t.Fatalf(
			"timeout = %s, want %s",
			cfg.Timeout,
			defaultTimeout,
		)
	}
}

func TestConfigValidateAcceptsSupportedValues(t *testing.T) {
	testCases := []Config{
		{
			BaseRef:    "main",
			HeadRef:    "HEAD",
			Format:     "text",
			LangFilter: "",
		},
		{
			BaseRef:    "main",
			HeadRef:    "HEAD",
			Format:     "json",
			LangFilter: "python",
		},
		{
			BaseRef:    "main",
			HeadRef:    "HEAD",
			Format:     "markdown",
			LangFilter: "py",
		},
		{
			BaseRef:    "main",
			HeadRef:    "HEAD",
			Format:     "text",
			LangFilter: "go",
		},
		{
			BaseRef:    "main",
			HeadRef:    "HEAD",
			Format:     "text",
			LangFilter: "golang",
		},
		{
			BaseRef:    "main",
			HeadRef:    "HEAD",
			Format:     "text",
			LangFilter: "all",
		},
	}

	for _, cfg := range testCases {
		cfg = cfg.withDefaults()

		if err := cfg.Validate(); err != nil {
			t.Fatalf(
				"expected valid config %+v, got %v",
				cfg,
				err,
			)
		}
	}
}

func TestConfigValidateRejectsMissingBaseRef(t *testing.T) {
	cfg := Config{
		HeadRef: "HEAD",
		Format:  "text",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected missing base ref error")
	}

	if !strings.Contains(err.Error(), "base ref") {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestConfigValidateRejectsUnsupportedFormat(t *testing.T) {
	cfg := Config{
		BaseRef: "main",
		HeadRef: "HEAD",
		Format:  "yaml",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected unsupported format error")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestConfigValidateRejectsUnsupportedLanguage(t *testing.T) {
	cfg := Config{
		BaseRef:    "main",
		HeadRef:    "HEAD",
		Format:     "text",
		LangFilter: "javascript",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected unsupported language error")
	}

	if !strings.Contains(err.Error(), "unsupported language") {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestConfigValidateRejectsNegativeThreshold(t *testing.T) {
	cfg := Config{
		BaseRef:   "main",
		HeadRef:   "HEAD",
		Format:    "text",
		Threshold: -1,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected negative threshold error")
	}

	if !strings.Contains(err.Error(), "threshold") {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestConfigValidateRejectsNegativeMinimumDelta(t *testing.T) {
	cfg := Config{
		BaseRef:  "main",
		HeadRef:  "HEAD",
		Format:   "text",
		MinDelta: -1,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected negative minimum delta error")
	}

	if !strings.Contains(err.Error(), "min-delta") {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestConfigValidateRejectsNegativeTimeout(t *testing.T) {
	cfg := Config{
		BaseRef: "main",
		HeadRef: "HEAD",
		Format:  "text",
		Timeout: -1 * time.Second,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected negative timeout error")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestNormalizeLanguageFilterAliases(t *testing.T) {
	testCases := map[string]string{
		"":       "",
		"all":    "",
		"ALL":    "",
		"py":     "python",
		"Python": "python",
		"golang": "go",
		"Go":     "go",
	}

	for input, expected := range testCases {
		actual := normalizeLangFilter(input)

		if actual != expected {
			t.Fatalf(
				"normalizeLangFilter(%q) = %q, want %q",
				input,
				actual,
				expected,
			)
		}
	}
}
