package grapheme_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	. "github.com/badu/term/grapheme"
	"golang.org/x/text/transform"
	"gotest.tools/v3/assert"
)

func TestTransformer(t *testing.T) {
	tests := []struct {
		in, out string
		tr      transform.Transformer
	}{
		{
			in: "test", out: "test",
			tr: TString("", "x"),
		},
		{
			in: "a", out: "b",
			tr: TString("a", "b"),
		},
		{
			in: "yes", out: "no",
			tr: TString("yes", "no"),
		},
		{
			in: "what what what", out: "wut wut wut",
			tr: TString("what", "wut"),
		},
		{
			in: "???????", out: "*******",
			tr: TString("?", "*"),
		},
		{
			in: "no matches", out: "no matches",
			tr: TString("x", "y"),
		},
		{
			in: "hello", out: "heLLo",
			tr: TString("l", "L"),
		},
		{
			in: "hello", out: "hello",
			tr: TString("x", "X"),
		},
		{
			in: "", out: "",
			tr: TString("x", "X"),
		},
		{
			in: "radar", out: "<r>ada<r>",
			tr: TString("r", "<r>"),
		},
		{
			in: "banana", out: "b<>n<>n<>",
			tr: TString("a", "<>"),
		},
		{
			in: "banana", out: "b<><>a",
			tr: TString("an", "<>"),
		},
		{
			in: "banana", out: "b<>na",
			tr: TString("ana", "<>"),
		},
		{
			in: "banana", out: "banana",
			tr: TString("a", "a"),
		},
		{
			in: "xxx", out: "",
			tr: TString("x", ""),
		},
		{
			in: strings.Repeat("foo_", 8<<10), out: strings.Repeat("bar_", 8<<10),
			tr: TString("foo", "bar"),
		},
		{
			in: "a", out: "b",
			tr: RegexpString(regexp.MustCompile("a"), "b"),
		},
		{
			in: "testing", out: "x",
			tr: RegexpString(regexp.MustCompile(".*"), "x"),
		},
		{
			in: strings.Repeat("--ax-- --bx--", 4<<10), out: strings.Repeat("--xx-- --xx--", 4<<10),
			tr: RegexpString(regexp.MustCompile(`--\wx--`), "--xx--"),
		},
		{
			in: strings.Repeat("--1x-- --2x-- --3x--", 8<<10), out: strings.Repeat("--0x-- --1x-- --2x--", 8<<10),
			tr: RegexpStringSubmatchFunc(regexp.MustCompile(`--(\d)x--`), func(match []string) string {
				x, _ := strconv.Atoi(match[1])
				return fmt.Sprintf("--%dx--", x-1)
			}),
		},
		{
			in: "1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 ", out: strings.Repeat("num ", 16),
			tr: RegexpStringFunc(regexp.MustCompile(`\d+`), func(_ string) string {
				return "num"
			}),
		},
		{
			in: "bazzzz buzz foo what biz", out: "  foo what ",
			tr: Regexp(regexp.MustCompile(`b\w+z\w*`), nil),
		},
		{
			in: "a", out: "replaced",
			tr: RegexpSubmatchFunc(regexp.MustCompile("a(123)?"), func(_ [][]byte) []byte {
				return []byte("replaced")
			}),
		},
		{
			in: strings.Repeat("x", 10<<10), out: "",
			tr: Regexp(regexp.MustCompile(".*"), nil),
		},
		{
			in: "this is a test", out: "(this) (is) (a) (test)",
			tr: RegexpString(regexp.MustCompile(`\w+`), "($0)"),
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, _, err := transform.String(tt.tr, tt.in)
			assert.NilError(t, err)
			assert.DeepEqual(t, result, tt.out)
		})
	}
}

func TestOverflowDst(t *testing.T) {
	var calls int
	s := strings.Repeat("x", 8<<10)
	tr := RegexpStringFunc(regexp.MustCompile("x"), func(_ string) string {
		calls++
		return s
	})

	result, _, err := transform.String(tr, "x")
	assert.NilError(t, err)
	assert.Equal(t, result, s)
	assert.Equal(t, calls, 1)
}