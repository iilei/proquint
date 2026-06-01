package cli

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
	"testing"
)

type exitCode int

func captureOutput(fn func()) (string, string) {
	oldStdout, oldStderr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr
	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
	}()

	outC := make(chan string)
	errC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	fn()

	_ = wOut.Close()
	_ = wErr.Close()

	return <-outC, <-errC
}

func captureOutputWithExit(fn func()) (string, string, int) {
	oldStdout, oldStderr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr

	outC := make(chan string, 1)
	errC := make(chan string, 1)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	code := -1
	func() {
		defer func() {
			_ = wOut.Close()
			_ = wErr.Close()
		}()
		defer func() {
			if r := recover(); r != nil {
				if ec, ok := r.(exitCode); ok {
					code = int(ec)
				} else {
					panic(r)
				}
			}
		}()
		fn()
	}()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout := <-outC
	stderr := <-errC

	return stdout, stderr, code
}

func TestParseProquintArgs(t *testing.T) {
	cases := []struct {
		name        string
		args        []string
		expectedRaw string
		expectedPad int
		expectedErr string
	}{
		{
			name:        "value only",
			args:        []string{"42"},
			expectedRaw: "42",
			expectedPad: 0,
		},
		{
			name:        "pad-groups long flag",
			args:        []string{"--pad-groups=2", "42"},
			expectedRaw: "42",
			expectedPad: 2,
		},
		{
			name:        "pad-groups separate flag",
			args:        []string{"--pad-groups", "2", "42"},
			expectedRaw: "42",
			expectedPad: 2,
		},
		{
			name:        "missing number",
			args:        []string{},
			expectedErr: "missing number",
		},
		{
			name:        "invalid pad-groups",
			args:        []string{"--pad-groups=-1", "42"},
			expectedErr: "invalid pad-groups value",
		},
		{
			name:        "missing pad-groups value",
			args:        []string{"--pad-groups"},
			expectedErr: "missing value for --pad-groups",
		},
		{
			name:        "unknown option",
			args:        []string{"-x", "42"},
			expectedErr: "unknown option: -x",
		},
		{
			name:        "unexpected argument",
			args:        []string{"42", "43"},
			expectedErr: "unexpected argument",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, pad, err := parseProquintArgs(tc.args)
			if raw != tc.expectedRaw || pad != tc.expectedPad || err != tc.expectedErr {
				t.Fatalf("expected (%q, %d, %q), got (%q, %d, %q)", tc.expectedRaw, tc.expectedPad, tc.expectedErr, raw, pad, err)
			}
		})
	}
}

func TestParseBigInt(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		expected string
	}{
		{"decimal", "42", "42"},
		{"hex lowercase", "0x2a", "42"},
		{"hex uppercase", "0X2A", "42"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, ok := parseBigInt(tc.raw)
			if !ok {
				t.Fatalf("expected parse success for %q", tc.raw)
			}
			if got := val.Text(10); got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestFormatProquint(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"decimal 42", "42", "babop"},
		{"hex max 256-bit", "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", strings.Repeat("zuzuz-", 15) + "zuzuz"},
		{"decimal max 256-bit", "115792089237316195423570985008687907853269984665640564039457584007913129639935", strings.Repeat("zuzuz-", 15) + "zuzuz"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, ok := parseBigInt(tc.input)
			if !ok {
				t.Fatal("failed to parse test input")
			}
			if got := formatProquint(val, 0); got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestFormatProquintPadGroups(t *testing.T) {
	val, ok := parseBigInt("42")
	if !ok {
		t.Fatal("failed to parse test input")
	}

	if got := formatProquint(val, 2); got != "babab-babop" {
		t.Fatalf("expected %q, got %q", "babab-babop", got)
	}
}

func TestFormatProquintZeroPad(t *testing.T) {
	val := big.NewInt(0)
	if got := formatProquint(val, 2); got != "babab-babab" {
		t.Fatalf("expected %q, got %q", "babab-babab", got)
	}
}

func TestChunkToWordRoundTrip(t *testing.T) {
	for _, chunk := range []uint16{0, 1, 0x1234, 0xFFFF} {
		t.Run(fmt.Sprintf("chunk-%04x", chunk), func(t *testing.T) {
			word := chunkToWord(chunk)
			parsed, err := wordToChunk(word)
			if err != nil {
				t.Fatalf("wordToChunk failed: %v", err)
			}
			if parsed != chunk {
				t.Fatalf("expected %d, got %d", chunk, parsed)
			}
		})
	}
}

func TestDecodeProquint(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"babop to decimal 42", "babop", "42"},
		{"babab-babop to decimal 42", "babab-babop", "42"},
		{"zuzuz groups to max decimal", strings.Repeat("zuzuz-", 15) + "zuzuz", "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := decodeProquint(tc.input)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if got := val.Text(10); got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestDecodeProquintErrors(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "invalid proquint word length: \"\""},
		{"invalid length", "baba", "invalid proquint word length: \"baba\""},
		{"invalid consonant c1", "aabab", "invalid proquint consonant: \"a\""},
		{"invalid consonant c2", "baixp", "invalid proquint consonant: \"i\""},
		{"invalid consonant c3", "babaq", "invalid proquint consonant: \"q\""},
		{"invalid vowel v1", "bxbop", "invalid proquint vowel: \"x\""},
		{"invalid vowel v2", "babxp", "invalid proquint vowel: \"x\""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeProquint(tc.input)
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, err.Error())
			}
		})
	}
}

func TestRunProquintHelp(t *testing.T) {
	stdout, _ := captureOutput(func() {
		runProquint([]string{"--help"})
	})

	if !strings.Contains(stdout, "Usage:") {
		t.Fatalf("expected help text, got %q", stdout)
	}
}

func TestRunProquintDefaultEncode(t *testing.T) {
	stdout, _ := captureOutput(func() {
		runProquint([]string{"42"})
	})

	if stdout != "babop\n" {
		t.Fatalf("expected babop, got %q", stdout)
	}
}

func TestRunProquintDecode(t *testing.T) {
	stdout, _ := captureOutput(func() {
		runProquint([]string{"decode", "babop"})
	})

	if stdout != "42\n" {
		t.Fatalf("expected 42, got %q", stdout)
	}
}

func TestRunProquintEmptyArgs(t *testing.T) {
	stdout, _, code := captureOutputWithExit(func() {
		runProquint([]string{})
	})

	if code != -1 {
		t.Fatalf("expected no exit, got code %d", code)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Fatalf("expected usage output, got %q", stdout)
	}
}

func TestRunProquintEncodeHelp(t *testing.T) {
	stdout, _ := captureOutput(func() {
		runProquint([]string{"encode", "--help"})
	})

	if !strings.Contains(stdout, "Usage: proquint encode") {
		t.Fatalf("expected encode usage output, got %q", stdout)
	}
}

func TestRunProquintEncodeShortHelp(t *testing.T) {
	stdout, _ := captureOutput(func() {
		runProquint([]string{"encode", "-h"})
	})

	if !strings.Contains(stdout, "Usage: proquint encode") {
		t.Fatalf("expected encode usage output, got %q", stdout)
	}
}

func TestRunProquintEncodePadGroups(t *testing.T) {
	stdout, _ := captureOutput(func() {
		runProquint([]string{"encode", "--pad-groups", "2", "42"})
	})

	if stdout != "babab-babop\n" {
		t.Fatalf("expected babab-babop, got %q", stdout)
	}
}

func TestRunProquintDecodeHelp(t *testing.T) {
	stdout, _ := captureOutput(func() {
		runProquint([]string{"decode", "--help"})
	})

	if !strings.Contains(stdout, "Usage: proquint decode") {
		t.Fatalf("expected decode usage output, got %q", stdout)
	}
}

func TestRunProquintDecodeShortHelp(t *testing.T) {
	stdout, _ := captureOutput(func() {
		runProquint([]string{"decode", "-h"})
	})

	if !strings.Contains(stdout, "Usage: proquint decode") {
		t.Fatalf("expected decode usage output, got %q", stdout)
	}
}

func TestRunProquintDecodeNoArgs(t *testing.T) {
	stdout, _ := captureOutput(func() {
		runProquint([]string{"decode"})
	})

	if !strings.Contains(stdout, "Usage: proquint decode") {
		t.Fatalf("expected decode usage output, got %q", stdout)
	}
}

func TestExecuteNoArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) { panic(exitCode(code)) }
	os.Args = []string{"proquint"}

	_, stderr, code := captureOutputWithExit(func() {
		Execute()
	})

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Fatalf("expected usage output, got %q", stderr)
	}
}

func TestExecuteHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) { panic(exitCode(code)) }
	os.Args = []string{"proquint", "--help"}

	stdout, stderr, code := captureOutputWithExit(func() {
		Execute()
	})

	if code != -1 {
		t.Fatalf("expected no exit, got code %d", code)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Fatalf("expected usage output, got %q", stderr)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout, got %q", stdout)
	}
}

func TestRunProquintEncodeInvalidNumber(t *testing.T) {
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) { panic(exitCode(code)) }

	_, stderr, code := captureOutputWithExit(func() {
		runProquint([]string{"encode", "abc"})
	})

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "invalid number") {
		t.Fatalf("expected invalid number message, got %q", stderr)
	}
}

func TestRunProquintEncodeInvalidPadGroups(t *testing.T) {
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) { panic(exitCode(code)) }

	_, stderr, code := captureOutputWithExit(func() {
		runProquint([]string{"encode", "--pad-groups=-1", "42"})
	})

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "invalid pad-groups value") {
		t.Fatalf("expected pad-groups error, got %q", stderr)
	}
}

func TestRunProquintEncodeUnknownOption(t *testing.T) {
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) { panic(exitCode(code)) }

	_, stderr, code := captureOutputWithExit(func() {
		runProquint([]string{"encode", "-x", "42"})
	})

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "unknown option: -x") {
		t.Fatalf("expected unknown option message, got %q", stderr)
	}
}

func TestRunProquintDecodeUnexpectedArg(t *testing.T) {
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) { panic(exitCode(code)) }

	_, stderr, code := captureOutputWithExit(func() {
		runProquint([]string{"decode", "babop", "extra"})
	})

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "unexpected argument") {
		t.Fatalf("expected unexpected argument message, got %q", stderr)
	}
}

func TestRunProquintDecodeInvalidProquint(t *testing.T) {
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) { panic(exitCode(code)) }

	_, stderr, code := captureOutputWithExit(func() {
		runProquint([]string{"decode", "bad"})
	})

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "invalid proquint word length") {
		t.Fatalf("expected invalid proquint message, got %q", stderr)
	}
}
