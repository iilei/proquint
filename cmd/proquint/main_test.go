package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	fn()
	_ = w.Close()
	return <-outC
}

func TestMainEncode(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"proquint", "encode", "42"}

	output := captureStdout(t, func() {
		main()
	})

	if strings.TrimSpace(output) != "babop" {
		t.Fatalf("expected babop, got %q", output)
	}
}

func TestMainDecode(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"proquint", "decode", "babop"}

	output := captureStdout(t, func() {
		main()
	})

	if strings.TrimSpace(output) != "42" {
		t.Fatalf("expected 42, got %q", output)
	}
}
