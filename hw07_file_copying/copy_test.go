package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Run("unsuported file", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := Copy(tmpDir, "dummy_output.txt", 0, 0)
		require.ErrorIs(t, err, ErrUnsupportedFile)
	})

	t.Run("offset exceeds file size", func(t *testing.T) {
		srcFile, err := os.CreateTemp(t.TempDir(), "src*.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer srcFile.Close()

		srcFile.WriteString("small content")
		srcPath := srcFile.Name()

		err = Copy(srcPath, "dummy.out", 1000, 0)
		fmt.Printf("asdasd %v", err)
		require.ErrorIs(t, err, ErrOffsetExceedsFileSize)
	})

	t.Run("copy simple file", func(t *testing.T) {
		srcFile, err := os.CreateTemp(t.TempDir(), "src*.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer srcFile.Close()

		content := make([]byte, 100)
		for i := range content {
			content[i] = byte('a' + i%26)
		}

		if _, err := srcFile.Write(content); err != nil {
			t.Fatal(err)
		}
		srcPath := srcFile.Name()
		tests := []struct {
			name     string
			offset   int64
			limit    int64
			expected []byte
		}{
			{
				name:     "copy whole file",
				offset:   0,
				limit:    0,
				expected: content,
			},
			{
				name:     "copy with offset",
				offset:   10,
				limit:    0,
				expected: content[10:],
			},
			{
				name:     "copy with limit",
				offset:   0,
				limit:    30,
				expected: content[:30],
			},
			{
				name:     "copy with offset and limit",
				offset:   20,
				limit:    40,
				expected: content[20 : 20+40],
			},
			{
				name:     "limit larger than remaining file",
				offset:   80,
				limit:    50,
				expected: content[80:],
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				dstFile, err := os.CreateTemp(t.TempDir(), "dst*.txt")
				if err != nil {
					t.Fatal(err)
				}
				dstPath := dstFile.Name()
				dstFile.Close()

				err = Copy(srcPath, dstPath, tc.offset, tc.limit)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				result, err := os.ReadFile(dstPath)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(result, tc.expected) {
					t.Errorf("expected %q, got %q", tc.expected, result)
				}
			})
		}
	})
}
