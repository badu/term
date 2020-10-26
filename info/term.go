package info

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/badu/term/color"
)

var (
	// NotFound indicates that a suitable terminal entry could not be found.
	// This can result from either not having TERM set, or from the TERM failing to support certain minimal functionality, in particular absolute cursor addressability (the cup capability) is required.
	// For example, legacy "adm3" lacks this capability, whereas the slightly newer "adm3a" supports it.
	// This failure occurs most often with "dumb".
	NotFound = errors.New("terminal entry not found")
)

const (
	None  = 0
	XTerm = 1
)

type stackElem struct {
	s     string
	i     int
	isStr bool
	isInt bool
}

// headTailLinkedList is a head-tail linked list data structure implementation.
// It is based on a doubly linked list container, so that every operations time complexity is O(1).
// every operations over a headTailLinkedList are synchronized and safe for concurrent usage.
type headTailLinkedList struct {
	sync.RWMutex
	container *list.List
	capacity  int
}

// newDeque creates a headTailLinkedList.
func newDeque() *headTailLinkedList {
	return &headTailLinkedList{
		container: list.New(),
		capacity:  -1,
	}
}

// prepend inserts element at the front in a O(1) time complexity, returning true if successful or false if the headTailLinkedList is at capacity.
func (s *headTailLinkedList) prepend(item stackElem) bool {
	s.Lock()
	defer s.Unlock()

	if s.capacity < 0 || s.container.Len() < s.capacity {
		s.container.PushFront(item)
		return true
	}

	return false
}

// Shift removes the first element of the headTailLinkedList in a O(1) time complexity
func (s *headTailLinkedList) shift() stackElem {
	s.Lock()
	defer s.Unlock()

	var item interface{} = nil
	var firstContainerItem *list.Element = nil

	firstContainerItem = s.container.Front()
	if firstContainerItem != nil {
		item = s.container.Remove(firstContainerItem)
	}

	return item.(stackElem)
}

// first returns the first value stored in the headTailLinkedList in a O(1) time complexity
func (s *headTailLinkedList) first() stackElem {
	s.RLock()
	defer s.RUnlock()

	item := s.container.Front()
	if item != nil {
		return item.Value.(stackElem)
	} else {
		return stackElem{}
	}
}

type stack struct {
	*headTailLinkedList
}

func newStack() *stack {
	return &stack{
		headTailLinkedList: newDeque(),
	}
}

// push adds on an item on the top of the stack
func (s *stack) push(item string) {
	e := stackElem{
		s:     item,
		isStr: true,
	}
	s.prepend(e)
}

// pushInt adds on an item on the top of the stack
func (s *stack) pushInt(item int) {
	e := stackElem{
		i:     item,
		isInt: true,
	}
	s.prepend(e)
}

// pushBool adds on an item on the top of the stack
func (s *stack) pushBool(item bool) {
	if item {
		s.pushInt(1)
	} else {
		s.pushInt(0)
	}
}

// pop removes and returns the item on the top of the stack
func (s *stack) pop() string {
	item := s.shift()
	if item.isStr {
		return item.s
	}
	return strconv.Itoa(item.i)
}

// popInt removes and returns the item on the top of the stack
func (s *stack) popInt() int {
	item := s.shift()
	if item.isInt {
		return item.i
	}
	i, _ := strconv.Atoi(item.s)
	return i
}

// popBool removes and returns the item on the top of the stack
func (s *stack) popBool() bool {
	item := s.shift()
	if item.isStr {
		return item.s == "1"
	}
	return item.i == 1
}

// Term represents a info entry.
// Note that we use friendly names in Go, but when we write out JSON, we use the same names as info.
// The name, aliases and smous, rmous fields do not come from info directly.
type Term struct {
	Columns      int // cols
	Width        int
	Lines        int // lines
	Height       int
	Colors       int // colors
	Modifiers    int
	Name         string
	Bell         string // bell
	Clear        string // clear
	EnterCA      string // smcup
	ExitCA       string // rmcup
	ShowCursor   string // cnorm
	HideCursor   string // civis
	AttrOff      string // sgr0
	Underline    string // smul
	Bold         string // bold
	Blink        string // blink
	Reverse      string // rev
	Dim          string // dim
	Italic       string // sitm
	EnterKeypad  string // smkx
	ExitKeypad   string // rmkx
	SetFg        string // setaf
	SetBg        string // setab
	ResetFgBg    string // op
	SetCursor    string // cup
	CursorBack1  string // cub1
	CursorUp1    string // cuu1
	PadChar      string // pad
	KeyBackspace string // kbs
	KeyF1        string // kf1
	KeyF2        string // kf2
	KeyF3        string // kf3
	KeyF4        string // kf4
	KeyF5        string // kf5
	KeyF6        string // kf6
	KeyF7        string // kf7
	KeyF8        string // kf8
	KeyF9        string // kf9
	KeyF10       string // kf10
	KeyF11       string // kf11
	KeyF12       string // kf12
	KeyInsert    string // kich1
	KeyDelete    string // kdch1
	KeyHome      string // khome
	KeyEnd       string // kend
	KeyHelp      string // khlp
	KeyPgUp      string // kpp
	KeyPgDn      string // knp
	KeyUp        string // kcuu1
	KeyDown      string // kcud1
	KeyLeft      string // kcub1
	KeyRight     string // kcuf1
	KeyBacktab   string // kcbt
	KeyExit      string // kext
	KeyClear     string // kclr
	KeyPrint     string // kprt
	KeyCancel    string // kcan
	Mouse        string // kmous
	MouseMode    string // XM
	AltChars     string // acsc
	EnterAcs     string // smacs
	ExitAcs      string // rmacs
	EnableAcs    string // enacs
	KeyShfRight  string // kRIT
	KeyShfLeft   string // kLFT
	KeyShfHome   string // kHOM
	KeyShfEnd    string // kEND
	KeyShfInsert string // kIC
	KeyShfDelete string // kDC

	// emulations, so don't depend too much on them in your application.
	// Terminal support for these are going to vary amongst XTerm that shifted variants of left and right exist, but not up and down. true color support, and some additional keys.
	// These are non-standard extensions to info.

	StrikeThrough   string // smxx
	SetFgBg         string // setfgbg
	SetFgBgRGB      string // setfgbgrgb
	SetFgRGB        string // setfrgb
	SetBgRGB        string // setbrgb
	KeyShfUp        string // shift-up
	KeyShfDown      string // shift-down
	KeyShfPgUp      string // shift-kpp
	KeyShfPgDn      string // shift-knp
	KeyCtrlUp       string // ctrl-up
	KeyCtrlDown     string // ctrl-left
	KeyCtrlRight    string // ctrl-right
	KeyCtrlLeft     string // ctrl-left
	KeyMetaUp       string // meta-up
	KeyMetaDown     string // meta-left
	KeyMetaRight    string // meta-right
	KeyMetaLeft     string // meta-left
	KeyAltUp        string // alt-up
	KeyAltDown      string // alt-left
	KeyAltRight     string // alt-right
	KeyAltLeft      string // alt-left
	KeyCtrlHome     string
	KeyCtrlEnd      string
	KeyMetaHome     string
	KeyMetaEnd      string
	KeyAltHome      string
	KeyAltEnd       string
	KeyAltShfUp     string
	KeyAltShfDown   string
	KeyAltShfLeft   string
	KeyAltShfRight  string
	KeyMetaShfUp    string
	KeyMetaShfDown  string
	KeyMetaShfLeft  string
	KeyMetaShfRight string
	KeyCtrlShfUp    string
	KeyCtrlShfDown  string
	KeyCtrlShfLeft  string
	KeyCtrlShfRight string
	KeyCtrlShfHome  string
	KeyCtrlShfEnd   string
	KeyAltShfHome   string
	KeyAltShfEnd    string
	KeyMetaShfHome  string
	KeyMetaShfEnd   string
	Aliases         []string
	TrueColor       bool // true if the terminal supports direct color
}

type Commander struct {
	*paramsBuffer
	bGotos         *gotoCache
	bColors        *colorCache
	Colors         int // colors
	Columns        int // cols
	Lines          int // lines
	svars          [26]string
	bEnterCA       []byte
	bHideCursor    []byte
	bShowCursor    []byte
	bEnableAcs     []byte
	bClear         []byte
	bAttrOff       []byte
	bExitCA        []byte
	bExitKeypad    []byte
	bBold          []byte
	bUnderline     []byte
	bReverse       []byte
	bBlink         []byte
	bDim           []byte
	bItalic        []byte
	bStrikeThrough []byte
	bResetFgBg     []byte
	bEnableMouse   []byte
	bDisableMouse  []byte
	PadChar        string
	SetFg          string // setaf
	SetBg          string // setab
	SetFgBg        string // setfgbg
	SetFgBgRGB     string // setfgbgrgb
	SetFgRGB       string // setfrgb
	SetBgRGB       string // setbrgb
	SetCursor      string // cup
	EnterAcs       string // smacs
	ExitAcs        string // rmacs
	AltChars       string // acsc
	Clear          string // clear
	HasMouse       bool
	HasHideCursor  bool
}

type colorCache struct {
	mapb map[string][]byte
}

type gotoCache struct {
	mapb map[int][]byte
}

// paramsBuffer handles some persistent state for TParam.
// Technically we could probably dispense with this, but caching buffer arrays gives us a nice little performance boost.
// Furthermore, we know that TParam is rarely (never?) called reentrantly, so we can just reuse the same buffers, making it thread-safe by stashing a lock.
type paramsBuffer struct {
	sync.Mutex
	out bytes.Buffer
	buf bytes.Buffer
}

// start initializes the params buffer with the initial string data.
// It also locks the paramsBuffer.
// The caller must call End() when finished.
func (pb *paramsBuffer) start(s string) {
	pb.Lock()
	pb.out.Reset()
	pb.buf.Reset()
	pb.buf.WriteString(s)
}

// End returns the final output from TParam, but it also releases the lock.
func (pb *paramsBuffer) end() string {
	s := pb.out.String()
	pb.Unlock()
	return s
}

// next returns the next input character to the expander.
func (pb *paramsBuffer) next() (byte, error) {
	return pb.buf.ReadByte()
}

// put "emits" (rather schedules for output) a single byte character.
func (pb *paramsBuffer) put(ch byte) {
	pb.out.WriteByte(ch)
}

// putString schedules a string for output.
func (pb *paramsBuffer) putString(s string) {
	pb.out.WriteString(s)
}

// TParam takes a info parameterized string, such as setaf or cup, and evaluates the string, and returns the result with the parameter applied.
func (t *Commander) TParam(s string, ints ...int) string {
	var (
		dvars  [26]string
		a, b   string
		ai, bi int
		ab     bool
		params [9]int
	)
	stk := newStack()

	if t.paramsBuffer == nil {
		t.paramsBuffer = &paramsBuffer{}
	}
	t.paramsBuffer.start(s)

	// make sure we always have 9 parameters -- makes it easier later to skip checks
	for i := 0; i < len(params) && i < len(ints); i++ {
		params[i] = ints[i]
	}

	nest := 0

	for {
		ch, err := t.paramsBuffer.next()
		if err != nil {
			break
		}

		if ch != '%' {
			t.paramsBuffer.put(ch)
			continue
		}

		ch, err = t.paramsBuffer.next()
		if err != nil {
			// TODO Error
			break
		}

		switch ch {
		case '%': // quoted %
			t.paramsBuffer.put(ch)

		case 'i': // increment both parameters (ANSI cup support)
			params[0]++
			params[1]++

		case 'c', 's':
			// NB: these, and 'd' below are special cased for efficiency.
			// They could be handled by the richer format support below, less efficiently.
			a = stk.pop()
			t.paramsBuffer.putString(a)

		case 'd':
			ai = stk.popInt()
			t.paramsBuffer.putString(strconv.Itoa(ai))

		case '0', '1', '2', '3', '4', 'x', 'X', 'o', ':':
			// This is pretty suboptimal, but this is rarely used.
			// None of the mainstream terminals use any of this, and it would surprise me if this code is ever executed outside of test cases.
			f := "%"
			if ch == ':' {
				ch, _ = t.paramsBuffer.next()
			}
			f += string(ch)
			for ch == '+' || ch == '-' || ch == '#' || ch == ' ' {
				ch, _ = t.paramsBuffer.next()
				f += string(ch)
			}
			for (ch >= '0' && ch <= '9') || ch == '.' {
				ch, _ = t.paramsBuffer.next()
				f += string(ch)
			}
			switch ch {
			case 'd', 'x', 'X', 'o':
				ai = stk.popInt()
				t.paramsBuffer.putString(fmt.Sprintf(f, ai))
			case 'c', 's':
				a = stk.pop()
				t.paramsBuffer.putString(fmt.Sprintf(f, a))
			}

		case 'p': // push parameter
			ch, _ = t.paramsBuffer.next()
			ai = int(ch - '1')
			if ai >= 0 && ai < len(params) {
				stk.pushInt(params[ai])
			} else {
				stk.pushInt(0)
			}

		case 'P': // pop & store variable
			ch, _ = t.paramsBuffer.next()
			if ch >= 'A' && ch <= 'Z' {
				t.svars[int(ch-'A')] = stk.pop()
			} else if ch >= 'a' && ch <= 'z' {
				dvars[int(ch-'a')] = stk.pop()
			}

		case 'g': // recall & push variable
			ch, _ = t.paramsBuffer.next()
			if ch >= 'A' && ch <= 'Z' {
				stk.push(t.svars[int(ch-'A')])
			} else if ch >= 'a' && ch <= 'z' {
				stk.push(dvars[int(ch-'a')])
			}

		case '\'': // push(char)
			ch, _ = t.paramsBuffer.next()
			_, _ = t.paramsBuffer.next() // must be ' but we don't check
			stk.push(string(ch))

		case '{': // push(int)
			ai = 0
			ch, _ = t.paramsBuffer.next()
			for ch >= '0' && ch <= '9' {
				ai *= 10
				ai += int(ch - '0')
				ch, _ = t.paramsBuffer.next()
			}
			// ch must be '}' but no verification
			stk.pushInt(ai)

		case 'l': // push(strlen(pop))
			a = stk.pop()
			stk.pushInt(len(a))

		case '+':
			bi = stk.popInt()
			ai = stk.popInt()
			stk.pushInt(ai + bi)
		case '-':
			bi = stk.popInt()
			ai = stk.popInt()
			stk.pushInt(ai - bi)

		case '*':
			bi = stk.popInt()
			ai = stk.popInt()
			stk.pushInt(ai * bi)

		case '/':
			bi = stk.popInt()
			ai = stk.popInt()
			if bi != 0 {
				stk.pushInt(ai / bi)
			} else {
				stk.pushInt(0)
			}

		case 'm': // push(pop mod pop)
			bi = stk.popInt()
			ai = stk.popInt()
			if bi != 0 {
				stk.pushInt(ai % bi)
			} else {
				stk.pushInt(0)
			}

		case '&': // AND
			bi = stk.popInt()
			ai = stk.popInt()
			stk.pushInt(ai & bi)

		case '|': // OR
			bi = stk.popInt()
			ai = stk.popInt()
			stk.pushInt(ai | bi)

		case '^': // XOR
			bi = stk.popInt()
			ai = stk.popInt()
			stk.pushInt(ai ^ bi)

		case '~': // bit complement
			ai = stk.popInt()
			stk.pushInt(ai ^ -1)

		case '!': // logical NOT
			ai = stk.popInt()
			stk.pushBool(ai != 0)

		case '=': // numeric compare or string compare
			b = stk.pop()
			a = stk.pop()
			stk.pushBool(a == b)

		case '>': // greater than, numeric
			bi = stk.popInt()
			ai = stk.popInt()
			stk.pushBool(ai > bi)

		case '<': // less than, numeric
			bi = stk.popInt()
			ai = stk.popInt()
			stk.pushBool(ai < bi)

		case '?': // start conditional

		case 't':
			ab = stk.popBool()
			if ab {
				// just keep going
				break
			}
			nest = 0
		ifloop:
			// this loop consumes everything until we hit our else, or the end of the conditional
			for {
				ch, err = t.paramsBuffer.next()
				if err != nil {
					break
				}
				if ch != '%' {
					continue
				}
				ch, _ = t.paramsBuffer.next()
				switch ch {
				case ';':
					if nest == 0 {
						break ifloop
					}
					nest--
				case '?':
					nest++
				case 'e':
					if nest == 0 {
						break ifloop
					}
				}
			}

		case 'e':
			// if we got here, it means we didn't use the else in the 't' case above, and we should skip until the end of the conditional
			nest = 0
		elloop:
			for {
				ch, err = t.paramsBuffer.next()
				if err != nil {
					break
				}
				if ch != '%' {
					continue
				}
				ch, _ = t.paramsBuffer.next()
				switch ch {
				case ';':
					if nest == 0 {
						break elloop
					}
					nest--
				case '?':
					nest++
				}
			}

		case ';': // endif

		}
	}

	return t.paramsBuffer.end()
}

// WriteString emits the string to the writer, but expands inline padding indications (of the form $<[delay]> where [delay] is msec) to a suitable time (unless the info string indicates this isn't needed by specifying npc - no padding).
// All Term based strings should be emitted using this function.
func (t *Commander) WriteString(w io.Writer, s string) error {
	for {
		beg := strings.Index(s, "$<")
		if beg < 0 {
			// Most strings don't need padding!
			if _, err := io.WriteString(w, s); err != nil {
				return fmt.Errorf("could not write string : %v", err)
			}
			return nil
		}
		if _, err := io.WriteString(w, s[:beg]); err != nil {
			return fmt.Errorf("could not write string : %v", err)
		}
		s = s[beg+2:]
		end := strings.Index(s, ">")
		if end < 0 {
			// unterminated.. just emit bytes unadulterated
			if _, err := io.WriteString(w, "$<"+s); err != nil {
				return fmt.Errorf("could not write string : %v", err)
			}
			return nil
		}
		val := s[:end]
		s = s[end+1:]
		padDur := 0
		unit := time.Millisecond
		dot := false
	loop:
		for i := range val {
			switch val[i] {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				padDur *= 10
				padDur += int(val[i] - '0')
				if dot {
					unit /= 10
				}
			case '.':
				if !dot {
					dot = true
				} else {
					break loop
				}
			default:
				break loop
			}
		}

		// Curses historically uses padding to achieve "fine grained" delays.
		// We have much better clocks these days, and so we do not rely on padding but simply sleep a bit.
		if len(t.PadChar) > 0 {
			time.Sleep(unit * time.Duration(padDur))
		}
	}
}

// WriteString emits the string to the writer, but expands inline padding indications (of the form $<[delay]> where [delay] is msec) to a suitable time (unless the info string indicates this isn't needed by specifying npc - no padding).
// All Term based strings should be emitted using this function.
func (t *Commander) WriteBytes(w io.Writer, s []byte) error {
	for {
		beg := bytes.Index(s, []byte("$<"))
		if beg < 0 {
			// Most strings don't need padding!
			if _, err := w.Write(s); err != nil {
				return fmt.Errorf("could not write string : %v", err)
			}
			return nil
		}
		if _, err := w.Write(s[:beg]); err != nil {
			return fmt.Errorf("could not write string : %v", err)
		}
		s = s[beg+2:]
		end := bytes.Index(s, []byte(">"))
		if end < 0 {
			// unterminated.. just emit bytes unadulterated
			ns := []byte("$>")
			ns = append(ns, s...)
			if _, err := w.Write(ns); err != nil {
				return fmt.Errorf("could not write string : %v", err)
			}
			return nil
		}
		val := s[:end]
		s = s[end+1:]
		padDur := 0
		unit := time.Millisecond
		dot := false
	loop:
		for i := range val {
			switch val[i] {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				padDur *= 10
				padDur += int(val[i] - '0')
				if dot {
					unit /= 10
				}
			case '.':
				if !dot {
					dot = true
				} else {
					break loop
				}
			default:
				break loop
			}
		}

		// Curses historically uses padding to achieve "fine grained" delays.
		// We have much better clocks these days, and so we do not rely on padding but simply sleep a bit.
		if len(t.PadChar) > 0 {
			time.Sleep(unit * time.Duration(padDur))
		}
	}
}

// TColor returns a string corresponding to the given foreground and background colors.
// Either fg or bg can be set to -1 to elide.
func (t *Commander) TColor(fi, bi int) string {
	rv := ""
	// As a special case, we map bright colors to lower versions if the color table only holds 8.
	// For the remaining 240 colors, the user is out of luck.
	// Someday we could create a mapping table, but its not worth it.
	if t.Colors == 8 {
		if fi > 7 && fi < 16 {
			fi -= 8
		}
		if bi > 7 && bi < 16 {
			bi -= 8
		}
	}
	if t.Colors > fi && fi >= 0 {
		rv += t.TParam(t.SetFg, fi)
	}
	if t.Colors > bi && bi >= 0 {
		rv += t.TParam(t.SetBg, bi)
	}
	return rv
}

func (t *Commander) PutEnterCA(w io.Writer) {
	if _, err := w.Write(t.bEnterCA); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutHideCursor(w io.Writer) {
	if _, err := w.Write(t.bHideCursor); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutShowCursor(w io.Writer) {
	if _, err := w.Write(t.bShowCursor); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutEnableAcs(w io.Writer) {
	if _, err := w.Write(t.bEnableAcs); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutClear(w io.Writer) {
	if _, err := w.Write(t.bClear); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutAttrOff(w io.Writer) {
	if _, err := w.Write(t.bAttrOff); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutExitCA(w io.Writer) {
	if _, err := w.Write(t.bExitCA); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutExitKeypad(w io.Writer) {
	if _, err := w.Write(t.bExitKeypad); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutBold(w io.Writer) {
	if _, err := w.Write(t.bBold); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutUnderline(w io.Writer) {
	if _, err := w.Write(t.bUnderline); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutReverse(w io.Writer) {
	if _, err := w.Write(t.bReverse); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutBlink(w io.Writer) {
	if _, err := w.Write(t.bBlink); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutDim(w io.Writer) {
	if _, err := w.Write(t.bDim); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutItalic(w io.Writer) {
	if _, err := w.Write(t.bItalic); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutStrikeThrough(w io.Writer) {
	if _, err := w.Write(t.bStrikeThrough); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutResetFgBg(w io.Writer) {
	if _, err := w.Write(t.bResetFgBg); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutEnableMouse(w io.Writer) {
	if !t.HasMouse {
		return
	}
	if _, err := w.Write(t.bEnableMouse); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func (t *Commander) PutDisableMouse(w io.Writer) {
	if !t.HasMouse {
		return
	}
	if _, err := w.Write(t.bDisableMouse); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

// GoTo for addressing the cursor at the given row and column.
// The origin 0, 0 is in the upper left corner of the screen.
func (t *Commander) GoTo(w io.Writer, column, row int) {
	v, ok := t.bGotos.mapb[hash(row, column)]
	if !ok { // not found : storing
		gotoStr := t.TParam(t.SetCursor, row, column)
		v = []byte(gotoStr)
		t.bGotos.mapb[hash(row, column)] = v
	}
	if _, err := w.Write(v); err != nil {
		if Debug {
			log.Printf("error writing to out : %v", err)
		}
	}
}

func hash(row, column int) int {
	return ((row & 0xFFFF) << 16) | (column & 0xFFFF) // X is row, Y is column
}

func (t *Commander) WriteBothColors(w io.Writer, fg, bg color.Color, isDelighted bool) {
	fgAndBgNames := ""
	if isDelighted {
		fgAndBgNames = fmt.Sprintf("0xFF_%06X_%06X", fg.Hex(), bg.Hex())
	} else {
		fgAndBgNames = fmt.Sprintf("%06X_%06X", fg.Hex(), bg.Hex())
	}
	bgFg, has := t.bColors.mapb[fgAndBgNames]
	if !has {
		bgFgStr := ""
		if isDelighted {
			bgFgStr = t.TParam(t.SetFgBg, int(fg&0xff), int(bg&0xff))
		} else {
			r1, g1, b1 := fg.RGB()
			r2, g2, b2 := bg.RGB()
			bgFgStr = t.TParam(t.SetFgBgRGB, r1, g1, b1, r2, g2, b2)
		}
		bgFg = []byte(bgFgStr)
		t.bColors.mapb[fgAndBgNames] = bgFg
	}
	if err := t.WriteBytes(w, bgFg); err != nil {
		if Debug {
			log.Printf("[core-sendFgBg] error writing string : %v", err)
		}
	}
}

func (t *Commander) WriteColor(w io.Writer, c color.Color, isForeground, isDelighted bool) {
	colorName := ""
	if isDelighted {
		colorName = fmt.Sprintf("0xFF_%06X", c.Hex())
	} else {
		colorName = fmt.Sprintf("%06X", c.Hex())
	}
	cb, has := t.bColors.mapb[colorName]
	if !has {
		cs := ""
		if isDelighted {
			if isForeground {
				cs = t.TParam(t.SetFg, int(c&0xFF))
			} else {
				cs = t.TParam(t.SetBg, int(c&0xFF))
			}
		} else {
			r, g, b := c.RGB()
			if isForeground {
				cs = t.TParam(t.SetFgRGB, r, g, b)
			} else {
				cs = t.TParam(t.SetBgRGB, r, g, b)
			}
		}
		cb = []byte(cs)
		t.bColors.mapb[colorName] = cb
	}
	if err := t.WriteBytes(w, cb); err != nil {
		if Debug {
			log.Printf("[core-sendFgBg] error writing string : %v", err)
		}
	}
}

func NewCommander(ti *Term) *Commander {
	res := Commander{}
	// goto optimization : cache the goto instructions for each cell
	res.bGotos = &gotoCache{mapb: make(map[int][]byte)}
	res.bColors = &colorCache{mapb: make(map[string][]byte)}
	// optimisation : some of the commonly used strings are prepared as []byte
	buf := bytes.NewBuffer(nil)
	if err := res.WriteString(buf, ti.EnterCA); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bEnterCA = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.HideCursor); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bHideCursor = buf.Bytes()
	res.HasHideCursor = len(ti.HideCursor) != 0
	buf.Reset()
	if err := res.WriteString(buf, ti.ShowCursor); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bShowCursor = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.EnableAcs); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bEnableAcs = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.Clear); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bClear = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.AttrOff); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bAttrOff = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.ExitCA); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bExitCA = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.ExitKeypad); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bExitKeypad = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.Bold); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bBold = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.Underline); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bUnderline = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.Reverse); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bReverse = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.Blink); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bBlink = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.Dim); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bDim = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.Italic); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bItalic = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.StrikeThrough); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bStrikeThrough = buf.Bytes()
	buf.Reset()
	if err := res.WriteString(buf, ti.ResetFgBg); err != nil {
		if Debug {
			log.Printf("error making bytes : %v", err)
		}
	}
	res.bResetFgBg = buf.Bytes()

	res.HasMouse = len(ti.Mouse) != 0
	if res.HasMouse {
		enableMouse := res.TParam(ti.MouseMode, 1)
		res.bEnableMouse = []byte(enableMouse)
		disableMouse := res.TParam(ti.MouseMode, 0)
		res.bDisableMouse = []byte(disableMouse)
	}
	res.PadChar = ti.PadChar
	res.Colors = ti.Colors
	res.SetFg = ti.SetFg
	res.SetBg = ti.SetBg
	res.SetFgBg = ti.SetFgBg
	res.SetFgBgRGB = ti.SetFgBgRGB
	res.SetCursor = ti.SetCursor
	res.Clear = ti.Clear
	res.Lines = ti.Lines
	res.Columns = ti.Columns
	res.AltChars = ti.AltChars
	res.EnterAcs = ti.EnterAcs
	res.ExitAcs = ti.ExitAcs

	if Debug {
		log.Printf("EnterCA = %#v", res.bEnterCA)
		log.Printf("HideCursor = %#v", res.bHideCursor)
		log.Printf("ShowCursor = %#v", res.bShowCursor)
		log.Printf("EnableAcs = %#v", res.bEnableAcs)
		log.Printf("Clear = %#v", res.bClear)
		log.Printf("AttrOff = %#v", res.bAttrOff)
		log.Printf("ExitCA = %#v", res.bExitCA)
		log.Printf("ExitKeypad = %#v", res.bExitKeypad)
		log.Printf("Bold = %#v", res.bBold)
		log.Printf("Underline = %#v", res.bUnderline)
		log.Printf("Reverse = %#v", res.bReverse)
		log.Printf("Blink = %#v", res.bBlink)
		log.Printf("Dim = %#v", res.bDim)
		log.Printf("Italic = %#v", res.bItalic)
		log.Printf("StrikeThrough = %#v", res.bStrikeThrough)
		log.Printf("ResetFgBg = %#v", res.bResetFgBg)
		log.Printf("EnableMouse = %#v", res.bEnableMouse)
		log.Printf("DisableMouse = %#v", res.bDisableMouse)
	}
	return &res
}

var (
	mu    sync.Mutex
	infos = make(map[string]*Term)
)

// AddTerminfo can be called to register a new Term entry.
func AddTerminfo(t *Term) {
	mu.Lock()
	infos[t.Name] = t
	for _, x := range t.Aliases {
		infos[x] = t
	}
	mu.Unlock()
}

// RemoveAllInfos clears up some RAM after we've got what we needed (our Commander)
func RemoveAllInfos() {
	mu.Lock()
	infos = nil
	mu.Unlock()
}

// LookupTerminfo attempts to find a definition for the named $TERM.
func LookupTerminfo(name string) (*Term, error) {
	if name == "" {
		// else on windows: index out of bounds
		// on the name[0] reference below
		return nil, NotFound
	}

	addTrueColor := false
	switch os.Getenv("COLORTERM") {
	case "truecolor", "24bit", "24-bit":
		addTrueColor = true
	}
	mu.Lock()
	t := infos[name]
	mu.Unlock()

	// If the name ends in -truecolor, then fabricate an entry from the corresponding -256color, -color, or bare terminal.
	if t.TrueColor {
		addTrueColor = true
	} else if t == nil && strings.HasSuffix(name, "-truecolor") {
		suffixes := []string{
			"-256color",
			"-88color",
			"-color",
			"",
		}
		base := name[:len(name)-len("-truecolor")]
		for _, s := range suffixes {
			if t, _ = LookupTerminfo(base + s); t != nil {
				addTrueColor = true
				break
			}
		}
	}

	if t == nil {
		return nil, NotFound
	}

	switch os.Getenv("TERM_TRUECOLOR") {
	case "":
	case "disable":
		addTrueColor = false
	default:
		addTrueColor = true
	}

	// If the user has requested 24-bit color with $COLORTERM, then amend the value (unless already present).
	// This means we don't need to have a value present.
	if addTrueColor && t.SetFgBgRGB == "" && t.SetFgRGB == "" && len(t.SetBgRGB) == 0 {
		// Supply vanilla ISO 8613-6:1994 24-bit color sequences.
		t.SetFgRGB = "\x1b[38;2;%p1%d;%p2%d;%p3%dm"
		t.SetBgRGB = "\x1b[48;2;%p1%d;%p2%d;%p3%dm"
		t.SetFgBgRGB = "\x1b[38;2;%p1%d;%p2%d;%p3%d;48;2;%p4%d;%p5%d;%p6%dm"
	}

	return t, nil
}
