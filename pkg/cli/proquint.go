package cli

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
)

var (
	proquintConsonants = []string{"b", "d", "f", "g", "h", "j", "k", "l",
		"m", "n", "p", "r", "s", "t", "v", "z"}
	proquintVowels = []string{"a", "i", "o", "u"}

	proquintConsonantIndex = make(map[rune]uint16, len(proquintConsonants))
	proquintVowelIndex     = make(map[rune]uint16, len(proquintVowels))
)

func init() {
	for i, s := range proquintConsonants {
		proquintConsonantIndex[rune(s[0])] = uint16(i)
	}
	for i, s := range proquintVowels {
		proquintVowelIndex[rune(s[0])] = uint16(i)
	}
}

func runProquint(args []string) {
	if len(args) == 0 {
		printProquintUsage()
		return
	}

	switch args[0] {
	case "encode":
		runProquintEncode(args[1:])
	case "decode":
		runProquintDecode(args[1:])
	case "--help", "-h":
		printProquintUsage()
	default:
		runProquintEncode(args)
	}
}

func runProquintEncode(args []string) {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			printProquintEncodeUsage()
			return
		}
	}

	raw, padGroups, err := parseProquintArgs(args)
	if err != "" {
		fmt.Fprintln(os.Stderr, err)
		printProquintEncodeUsage()
		os.Exit(1)
	}

	value, ok := parseBigInt(raw)
	if !ok {
		fmt.Fprintf(os.Stderr, "invalid number: %s\n", raw)
		os.Exit(1)
	}

	if value.Sign() < 0 {
		fmt.Fprintln(os.Stderr, "number must be non-negative")
		os.Exit(1)
	}

	fmt.Println(formatProquint(value, padGroups))
}

func runProquintDecode(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printProquintDecodeUsage()
		return
	}

	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "unexpected argument")
		printProquintDecodeUsage()
		os.Exit(1)
	}

	value, err := decodeProquint(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(value.Text(10))
}

func printProquintUsage() {
	fmt.Println("Usage:")
	fmt.Println("  proquint encode [--pad-groups=N] <number>")
	fmt.Println("  proquint decode <proquint>")
	fmt.Println("Generate or decode a proquint.")
}

func printProquintEncodeUsage() {
	fmt.Println("Usage: proquint encode [--pad-groups=N] <number>")
	fmt.Println("Generate a proquint from a decimal or 0x-prefixed hex bigint.")
}

func printProquintDecodeUsage() {
	fmt.Println("Usage: proquint decode <proquint>")
	fmt.Println("Decode a proquint back to a bigint.")
}

func parseProquintArgs(args []string) (string, int, string) {
	var raw string
	padGroups := 0

	for i := 0; i < len(args); i++ {
		a := args[i]

		if after, ok := strings.CutPrefix(a, "--pad-groups="); ok {
			count, err := strconv.Atoi(after)
			if err != nil || count < 0 {
				return "", 0, "invalid pad-groups value"
			}
			padGroups = count
			continue
		}

		if a == "--pad-groups" {
			if i+1 >= len(args) {
				return "", 0, "missing value for --pad-groups"
			}
			count, err := strconv.Atoi(args[i+1])
			if err != nil || count < 0 {
				return "", 0, "invalid pad-groups value"
			}
			padGroups = count
			i++
			continue
		}

		if strings.HasPrefix(a, "-") {
			return "", 0, "unknown option: " + a
		}

		if raw != "" {
			return "", 0, "unexpected argument"
		}

		raw = a
	}

	if raw == "" {
		return "", 0, "missing number"
	}

	return raw, padGroups, ""
}

func parseBigInt(raw string) (*big.Int, bool) {
	num := new(big.Int)
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		return num.SetString(raw[2:], 16)
	}
	return num.SetString(raw, 10)
}

func formatProquint(value *big.Int, padGroups int) string {
	if value.Sign() == 0 {
		words := []string{chunkToWord(0)}
		for len(words) < padGroups {
			words = append([]string{chunkToWord(0)}, words...)
		}
		return strings.Join(words, "-")
	}

	n := new(big.Int).Set(value)
	mask := big.NewInt(0xFFFF)
	words := make([]string, 0, (n.BitLen()+15)/16)

	for n.Sign() > 0 {
		chunk := new(big.Int).And(n, mask)
		words = append(words, chunkToWord(uint16(chunk.Uint64())))
		n.Rsh(n, 16)
	}

	for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
		words[i], words[j] = words[j], words[i]
	}

	for len(words) < padGroups {
		words = append([]string{chunkToWord(0)}, words...)
	}

	return strings.Join(words, "-")
}

func decodeProquint(encoded string) (*big.Int, error) {
	parts := strings.Split(encoded, "-")
	if len(parts) == 0 {
		return nil, errors.New("empty proquint")
	}

	n := new(big.Int)
	for _, part := range parts {
		chunk, err := wordToChunk(part)
		if err != nil {
			return nil, err
		}
		n.Lsh(n, 16)
		n.Add(n, new(big.Int).SetUint64(uint64(chunk)))
	}

	return n, nil
}

func wordToChunk(word string) (uint16, error) {
	if len(word) != 5 {
		return 0, fmt.Errorf("invalid proquint word length: %q", word)
	}

	word = strings.ToLower(word)
	c1, ok := proquintConsonantIndex[rune(word[0])]
	if !ok {
		return 0, fmt.Errorf("invalid proquint consonant: %q", string(word[0]))
	}
	v1, ok := proquintVowelIndex[rune(word[1])]
	if !ok {
		return 0, fmt.Errorf("invalid proquint vowel: %q", string(word[1]))
	}
	c2, ok := proquintConsonantIndex[rune(word[2])]
	if !ok {
		return 0, fmt.Errorf("invalid proquint consonant: %q", string(word[2]))
	}
	v2, ok := proquintVowelIndex[rune(word[3])]
	if !ok {
		return 0, fmt.Errorf("invalid proquint vowel: %q", string(word[3]))
	}
	c3, ok := proquintConsonantIndex[rune(word[4])]
	if !ok {
		return 0, fmt.Errorf("invalid proquint consonant: %q", string(word[4]))
	}

	return uint16((c1 << 12) | (v1 << 10) | (c2 << 6) | (v2 << 4) | c3), nil
}

func chunkToWord(chunk uint16) string {
	return proquintConsonants[(chunk>>12)&0xF] +
		proquintVowels[(chunk>>10)&0x3] +
		proquintConsonants[(chunk>>6)&0xF] +
		proquintVowels[(chunk>>4)&0x3] +
		proquintConsonants[chunk&0xF]
}
