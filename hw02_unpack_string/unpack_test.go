package hw02unpackstring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		// uncomment if task with asterisk completed
		{input: `qwe\4\5`, expected: `qwe45`},
		{input: `qwe\45`, expected: `qwe44444`},
		{input: `qwe\\5`, expected: `qwe\\\\\`},
		{input: `qwe\\\3`, expected: `qwe\3`},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{"3abc", "45", "aaa10b"}
	for _, tc := range invalidStrings {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}

func TestUnpuckUtf8Strung(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"пробельные символы", "a\tb\nc", "a\tb\nc"},
		{"кириллица", "Привет", "Привет"},
		{"комбинирующие символы (буква + диакритика)", "a\u0301", "a\u0301"},
		{"комбинирующие символы (буква + диакритика) с повтором", "a\u03012", "a\u0301\u0301"},
		{"простой смайлик (одна руна)", "😊", "😊"},
		{"простой смайлик (одна руна) с повтором", "😊7", "😊😊😊😊😊😊😊"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := Unpack(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
