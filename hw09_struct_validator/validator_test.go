package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
		name        string
	}{
		{
			name: "valid user",
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    30,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{"12345678901", "10987654321"},
			},
			expectedErr: nil,
		},
		{
			name: "valid app",
			in: App{
				Version: "1.2.3",
			},
			expectedErr: nil,
		},
		{
			name: "token without tags - should be valid",
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			expectedErr: nil,
		},
		{
			name: "valid response",
			in: Response{
				Code: 200,
				Body: "OK",
			},
			expectedErr: nil,
		},
		{
			name: "int min and max combination",
			in: struct {
				Value int `validate:"min:10|max:20"`
			}{Value: 15},
			expectedErr: nil,
		},
		{
			name: "string with multiple rules via comma",
			in: struct {
				Name string `validate:"len:5,in:qwert,qwer"`
			}{Name: "qwert"},
			expectedErr: nil,
		},
		{
			name: "struct with unexported field only",
			in: struct {
				noexported string `validate:"len:5"`
			}{noexported: "12345"},
			expectedErr: nil,
		},
		{
			name: "tag with -",
			in: struct {
				Field string `validate:"-"`
			}{Field: "any"},
			expectedErr: nil,
		},
		{
			name: "empty tag",
			in: struct {
				Field string `validate:""`
			}{Field: "any"},
			expectedErr: nil,
		},
		{
			name: "programmatic error - invalid in parameter for int",
			in: struct {
				Field int `validate:"in:1,two,3"`
			}{Field: 1},
			expectedErr: errors.New("некорректное число в списке in"),
		},
		{
			name: "unknown validator for string",
			in: struct {
				Field string `validate:"unknown:xxx"`
			}{Field: "test"},
			expectedErr: errors.New("неизвестный валидатор для строки"),
		},
		{
			name: "unknown validator for int",
			in: struct {
				Field int `validate:"unknown:10"`
			}{Field: 5},
			expectedErr: errors.New("неизвестный валидатор для числа"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)

			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var valErrs ValidationErrors
			if errors.As(err, &valErrs) {
				t.Fatalf("expected programmatic error, got ValidationErrors: %v", valErrs)
			}
			if tt.expectedErr != nil && !strings.Contains(err.Error(), tt.expectedErr.Error()) {
				t.Errorf("expected error containing %q, got %q", tt.expectedErr.Error(), err.Error())
			}
		})
	}
}
