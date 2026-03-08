package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestRunCmd(t *testing.T) {
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Создаём pipe для захвата вывода команды
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Канал для получения вывода
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	env := Environment{
		"FOO":   EnvValue{Value: "bar", NeedRemove: false},
		"EMPTY": EnvValue{NeedRemove: true},
		"BAZ":   EnvValue{Value: "qux", NeedRemove: false},
	}

	os.Setenv("FOO", "original_foo")
	os.Setenv("EMPTY", "should_be_removed")
	os.Setenv("BAZ", "original_baz")

	cmd := []string{"sh", "-c", "echo FOO=$FOO; echo EMPTY=$EMPTY; echo BAZ=$BAZ"}
	code := RunCmd(cmd, env)

	w.Close()
	out := <-outC

	if code != 0 {
		t.Errorf("expected code 0, got %d", code)
	}

	expected := "FOO=bar\nEMPTY=\nBAZ=qux\n"
	if out != expected {
		t.Errorf("unexpected output:\n%q\nwant:\n%q", out, expected)
	}

	// Проверяем, что оригинальное окружение процесса не изменилось
	if os.Getenv("FOO") != "original_foo" {
		t.Error("original FOO was modified")
	}
	if os.Getenv("EMPTY") != "should_be_removed" {
		t.Error("original EMPTY was modified")
	}
	if os.Getenv("BAZ") != "original_baz" {
		t.Error("original BAZ was modified")
	}
}
