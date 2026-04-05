package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Тест проверяет, что при уровне INFO логируются сообщения уровней INFO, WARN и ERROR,
	// но не логируются сообщения уровня DEBUG.
	t.Run("Info level should log info, warn and error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewWithWriter("INFO", &buf, &buf)

		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")

		output := buf.String()

		// Проверяем, что DEBUG не логируется на уровне INFO.
		if strings.Contains(output, "debug message") {
			t.Error("Debug message should not be logged at INFO level")
		}
		if !strings.Contains(output, "info message") {
			t.Error("Info message should be logged at INFO level")
		}
		if !strings.Contains(output, "warn message") {
			t.Error("Warn message should be logged at INFO level")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Error message should be logged at INFO level")
		}
	})

	// Тест проверяет, что при уровне ERROR логируются только сообщения уровня ERROR.
	t.Run("Error level should log only errors", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewWithWriter("ERROR", &buf, &buf)

		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")

		output := buf.String()

		if strings.Contains(output, "debug message") {
			t.Error("Debug message should not be logged at ERROR level")
		}
		if strings.Contains(output, "info message") {
			t.Error("Info message should not be logged at ERROR level")
		}
		if strings.Contains(output, "warn message") {
			t.Error("Warn message should not be logged at ERROR level")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Error message should be logged at ERROR level")
		}
	})

	// Тест проверяет, что при уровне DEBUG логируются все сообщения.
	t.Run("Debug level should log everything", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewWithWriter("DEBUG", &buf, &buf)

		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")

		output := buf.String()

		if !strings.Contains(output, "debug message") {
			t.Error("Debug message should be logged  at DEBUG level")
		}
		if !strings.Contains(output, "info message") {
			t.Error("Info message should be logged at DEBUG level")
		}
		if !strings.Contains(output, "warn message") {
			t.Error("Warn message should be logged at DEBUG level")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Error message should be logged at DEBUG level")
		}
	})

	// Тест проверяет корректность форматированного вывода (методы *f).
	t.Run("Test formatted logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewWithWriter("INFO", &buf, &buf)

		logger.Infof("formatted %s %d", "message", 42)

		output := buf.String()
		if !strings.Contains(output, "formatted message 42") {
			t.Errorf("Expected formatted message, got: %s", output)
		}
	})

	// Тест проверяет корректность парсинга различных строковых представлений уровней логирования.
	t.Run("Test level parsing", func(t *testing.T) {
		tests := []struct {
			input    string
			expected Level
		}{
			{"DEBUG", LevelDebug},
			{"debug", LevelDebug},
			{"INFO", LevelInfo},
			{"info", LevelInfo},
			{"WARN", LevelWarn},
			{"warn", LevelWarn},
			{"WARNING", LevelWarn}, // альтернативное название
			{"ERROR", LevelError},
			{"error", LevelError},
			{"unknown", LevelInfo}, // для неизвестных уровней
		}

		for _, tt := range tests {
			result := parseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		}
	})
}
