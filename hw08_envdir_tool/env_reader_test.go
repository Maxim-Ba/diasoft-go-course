package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckFileName(t *testing.T) {
	t.Run("name contains =", func(t *testing.T) {
		err := checkFileName("dsfsdfsdf=sdfsdfsdfsdf")
		require.ErrorIs(t, err, WrongFileName)
	})
	t.Run("name not contains =", func(t *testing.T) {
		err := checkFileName("dsfsdfsdf+sdfsdfsdfsdf")
		require.NoError(t, err)
	})
}

func TestTrimRight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{
			name:  "empty",
			input: "",
			want:  []byte(""),
		},
		{
			name:  "one word without spaces",
			input: "hello",
			want:  []byte("hello"),
		},
		{
			name:  "only spaces",
			input: "hello   ",
			want:  []byte("hello"),
		},
		{
			name:  "tabs at right",
			input: "hello\t\t\t",
			want:  []byte("hello"),
		},
		{
			name:  "tabs and spaces",
			input: "hello  \t \t ",
			want:  []byte("hello"),
		},
		{
			name:  "space at left",
			input: "  hello",
			want:  []byte("  hello"),
		},
		{
			name:  "inner spaces",
			input: "hel lo  ",
			want:  []byte("hel lo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimRight(tt.input)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("trimRight(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestReplaceNull(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "empty",
			input: []byte{},
			want:  "",
		},
		{
			name:  "without nulls",
			input: []byte("hello"),
			want:  "hello",
		},
		{
			name:  "null at center",
			input: []byte("he\x00llo"),
			want:  "he\nllo",
		},
		{
			name:  "two nulls",
			input: []byte("a\x00b\x00c"),
			want:  "a\nb\nc",
		},
		{
			name:  "only nulls 1",
			input: []byte{0, 0, 0},
			want:  "\n\n\n",
		},
		{
			name:  "only nulls 2",
			input: []byte("\x00\x00\x00"),
			want:  "\n\n\n",
		},
		{
			name:  "mix",
			input: []byte("go\x00lang\x00"),
			want:  "go\nlang\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceNull(tt.input)
			if got != tt.want {
				t.Errorf("replaceNull(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	tmpDir := t.TempDir()

	createFile := func(name, content string) string {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	tests := []struct {
		name        string
		content     string
		expected    string
		expectError bool
	}{
		{
			name:     "simple with newline",
			content:  "hello\n",
			expected: "hello",
		},
		{
			name:     "simple without newline",
			content:  "hello",
			expected: "hello",
		},
		{
			name:     "few strings",
			content:  "first line\nsecond line\n",
			expected: "first line",
		},
		{
			name:     "empty",
			content:  "",
			expected: "",
		},
		{
			name:     "only newline",
			content:  "\n",
			expected: "",
		},
		{
			name:     "with tabs and space (with newline)",
			content:  "value  \t\t\n",
			expected: "value",
		},
		{
			name:     "with tabs and space (without newline)",
			content:  "value  \t\t",
			expected: "value",
		},
		{
			name:     "spaces around",
			content:  "  value  \n",
			expected: "  value",
		},
		{
			name:     "with nulls",
			content:  "abc\x00def\n",
			expected: "abc\ndef",
		},
		{
			name:     "only tabs ans spaces with newline",
			content:  " \t\n",
			expected: "",
		},
		{
			name:     "only tabs ans spaces without newline",
			content:  " \t",
			expected: "",
		},
		{
			name:        "empty file name",
			content:     "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.name == "empty file name" {
				path = filepath.Join(tmpDir, "not-exist")
			} else {
				path = createFile("testfile", tt.content)
			}

			got, err := getValue(path)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("getValue() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestReadDir(t *testing.T) {
	originDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(originDir); chErr != nil {
			t.Error(chErr)
		}
	}()

	files := map[string]struct {
		content string
		want    EnvValue
	}{
		"EMPTY": {
			content: "",
			want:    EnvValue{Value: "", NeedRemove: true},
		},
		"WITH_TABS": {
			content: "value  \t\t\n",
			want:    EnvValue{Value: "value", NeedRemove: false},
		},
		"WITH_NULL": {
			content: "a\x00b\n",
			want:    EnvValue{Value: "a\nb", NeedRemove: false},
		},
	}

	for name, data := range files {
		err := os.WriteFile(name, []byte(data.content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Создаём поддиректорию – она должна игнорироваться
	if err := os.Mkdir("SUBDIR", 0o755); err != nil {
		t.Fatal(err)
	}

	env, err := ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}

	if len(env) != len(files) {
		t.Errorf("got %d entries, want %d", len(env), len(files))
	}

	for name, file := range files {
		t.Run(name, func(t *testing.T) {
			got, ok := env[name]
			if !ok {
				t.Errorf("missing key %q", name)
			}
			if got != file.want {
				t.Errorf("key %q: got %+v, want %+v", name, got, file.want)
			}
		})
	}
}
