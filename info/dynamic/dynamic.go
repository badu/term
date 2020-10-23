package dynamic

import (
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/badu/term/info"
)

type termcap struct {
	name    string
	desc    string
	aliases []string
	bools   map[string]bool
	nums    map[string]int
	strs    map[string]string
}

func (tc *termcap) getnum(s string) int {
	return (tc.nums[s])
}

func (tc *termcap) getflag(s string) bool {
	return (tc.bools[s])
}

func (tc *termcap) getstr(s string) string {
	return (tc.strs[s])
}

const (
	none = iota
	control
	escaped
)

var errNotAddressable = errors.New("terminal not cursor addressable")

func unescape(s string) string {
	// Various escapes are in \x format.  Control codes are
	// encoded as ^M (carat followed by ASCII equivalent).
	// escapes are: \e, \E - escape
	//  \0 NULL, \n \l \r \t \b \f \s for equivalent C escape.
	buf := &bytes.Buffer{}
	esc := none

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch esc {
		case none:
			switch c {
			case '\\':
				esc = escaped
			case '^':
				esc = control
			default:
				buf.WriteByte(c)
			}
		case control:
			buf.WriteByte(c ^ 1<<6)
			esc = none
		case escaped:
			switch c {
			case 'E', 'e':
				buf.WriteByte(0x1b)
			case '0', '1', '2', '3', '4', '5', '6', '7':
				if i+2 < len(s) && s[i+1] >= '0' && s[i+1] <= '7' && s[i+2] >= '0' && s[i+2] <= '7' {
					buf.WriteByte(((c - '0') * 64) + ((s[i+1] - '0') * 8) + (s[i+2] - '0'))
					i = i + 2
				} else if c == '0' {
					buf.WriteByte(0)
				}
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case 's':
				buf.WriteByte(' ')
			default:
				buf.WriteByte(c)
			}
			esc = none
		}
	}
	return (buf.String())
}

func (tc *termcap) setupterm(name string) error {
	cmd := exec.Command("infocmp", "-1", name)
	output := &bytes.Buffer{}
	cmd.Stdout = output

	tc.strs = make(map[string]string)
	tc.bools = make(map[string]bool)
	tc.nums = make(map[string]int)

	if err := cmd.Run(); err != nil {
		return err
	}

	// Now parse the output.
	// We get comment lines (starting with "#"), followed by
	// a header line that looks like "<name>|<alias>|...|<desc>"
	// then capabilities, one per line, starting with a tab and ending
	// with a comma and newline.
	lines := strings.Split(output.String(), "\n")
	for len(lines) > 0 && strings.HasPrefix(lines[0], "#") {
		lines = lines[1:]
	}

	// Ditch trailing empty last line
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	header := lines[0]
	if strings.HasSuffix(header, ",") {
		header = header[:len(header)-1]
	}
	names := strings.Split(header, "|")
	tc.name = names[0]
	names = names[1:]
	if len(names) > 0 {
		tc.desc = names[len(names)-1]
		names = names[:len(names)-1]
	}
	tc.aliases = names
	for _, val := range lines[1:] {
		if (!strings.HasPrefix(val, "\t")) ||
			(!strings.HasSuffix(val, ",")) {
			return (errors.New("malformed infocmp: " + val))
		}

		val = val[1:]
		val = val[:len(val)-1]

		if k := strings.SplitN(val, "=", 2); len(k) == 2 {
			tc.strs[k[0]] = unescape(k[1])
		} else if k := strings.SplitN(val, "#", 2); len(k) == 2 {
			u, err := strconv.ParseUint(k[1], 0, 0)
			if err != nil {
				return (err)
			}
			tc.nums[k[0]] = int(u)
		} else {
			tc.bools[val] = true
		}
	}
	return nil
}

// LoadTerminfo creates a Terminfo by for named terminal by attempting to parse
// the output from infocmp.  This returns the terminfo entry, a description of
// the terminal, and either nil or an error.
func LoadTerminfo(name string) (*info.Term, string, error) {
	var tc termcap
	if err := tc.setupterm(name); err != nil {
		if err != nil {
			return nil, "", err
		}
	}
	t := &info.Term{}
	// If this is an alias record, then just emit the alias
	t.Name = tc.name
	if t.Name != name {
		return t, "", nil
	}
	t.Aliases = tc.aliases
	t.Colors = tc.getnum("colors")
	t.Columns = tc.getnum("cols")
	t.Lines = tc.getnum("lines")
	t.Bell = tc.getstr("bel")
	t.Clear = tc.getstr("clear")
	t.EnterCA = tc.getstr("smcup")
	t.ExitCA = tc.getstr("rmcup")
	t.ShowCursor = tc.getstr("cnorm")
	t.HideCursor = tc.getstr("civis")
	t.AttrOff = tc.getstr("sgr0")
	t.Underline = tc.getstr("smul")
	t.Bold = tc.getstr("bold")
	t.Blink = tc.getstr("blink")
	t.Dim = tc.getstr("dim")
	t.Italic = tc.getstr("sitm")
	t.Reverse = tc.getstr("rev")
	t.EnterKeypad = tc.getstr("smkx")
	t.ExitKeypad = tc.getstr("rmkx")
	t.SetFg = tc.getstr("setaf")
	t.SetBg = tc.getstr("setab")
	t.SetCursor = tc.getstr("cup")
	t.CursorBack1 = tc.getstr("cub1")
	t.CursorUp1 = tc.getstr("cuu1")
	t.KeyF1 = tc.getstr("kf1")
	t.KeyF2 = tc.getstr("kf2")
	t.KeyF3 = tc.getstr("kf3")
	t.KeyF4 = tc.getstr("kf4")
	t.KeyF5 = tc.getstr("kf5")
	t.KeyF6 = tc.getstr("kf6")
	t.KeyF7 = tc.getstr("kf7")
	t.KeyF8 = tc.getstr("kf8")
	t.KeyF9 = tc.getstr("kf9")
	t.KeyF10 = tc.getstr("kf10")
	t.KeyF11 = tc.getstr("kf11")
	t.KeyF12 = tc.getstr("kf12")
	t.KeyInsert = tc.getstr("kich1")
	t.KeyDelete = tc.getstr("kdch1")
	t.KeyBackspace = tc.getstr("kbs")
	t.KeyHome = tc.getstr("khome")
	t.KeyEnd = tc.getstr("kend")
	t.KeyUp = tc.getstr("kcuu1")
	t.KeyDown = tc.getstr("kcud1")
	t.KeyRight = tc.getstr("kcuf1")
	t.KeyLeft = tc.getstr("kcub1")
	t.KeyPgDn = tc.getstr("knp")
	t.KeyPgUp = tc.getstr("kpp")
	t.KeyBacktab = tc.getstr("kcbt")
	t.KeyExit = tc.getstr("kext")
	t.KeyCancel = tc.getstr("kcan")
	t.KeyPrint = tc.getstr("kprt")
	t.KeyHelp = tc.getstr("khlp")
	t.KeyClear = tc.getstr("kclr")
	t.AltChars = tc.getstr("acsc")
	t.EnterAcs = tc.getstr("smacs")
	t.ExitAcs = tc.getstr("rmacs")
	t.EnableAcs = tc.getstr("enacs")
	t.Mouse = tc.getstr("kmous")
	t.KeyShfRight = tc.getstr("kRIT")
	t.KeyShfLeft = tc.getstr("kLFT")
	t.KeyShfHome = tc.getstr("kHOM")
	t.KeyShfEnd = tc.getstr("kEND")

	// Terminfo lacks descriptions for a bunch of modified keys,
	// but modern XTerm and emulators often have them.  Let's add them,
	// if the shifted right and left arrows are defined.
	if t.KeyShfRight == "\x1b[1;2C" && t.KeyShfLeft == "\x1b[1;2D" {
		t.KeyShfUp = "\x1b[1;2A"
		t.KeyShfDown = "\x1b[1;2B"
		t.KeyMetaUp = "\x1b[1;9A"
		t.KeyMetaDown = "\x1b[1;9B"
		t.KeyMetaRight = "\x1b[1;9C"
		t.KeyMetaLeft = "\x1b[1;9D"
		t.KeyAltUp = "\x1b[1;3A"
		t.KeyAltDown = "\x1b[1;3B"
		t.KeyAltRight = "\x1b[1;3C"
		t.KeyAltLeft = "\x1b[1;3D"
		t.KeyCtrlUp = "\x1b[1;5A"
		t.KeyCtrlDown = "\x1b[1;5B"
		t.KeyCtrlRight = "\x1b[1;5C"
		t.KeyCtrlLeft = "\x1b[1;5D"
		t.KeyAltShfUp = "\x1b[1;4A"
		t.KeyAltShfDown = "\x1b[1;4B"
		t.KeyAltShfRight = "\x1b[1;4C"
		t.KeyAltShfLeft = "\x1b[1;4D"

		t.KeyMetaShfUp = "\x1b[1;10A"
		t.KeyMetaShfDown = "\x1b[1;10B"
		t.KeyMetaShfRight = "\x1b[1;10C"
		t.KeyMetaShfLeft = "\x1b[1;10D"

		t.KeyCtrlShfUp = "\x1b[1;6A"
		t.KeyCtrlShfDown = "\x1b[1;6B"
		t.KeyCtrlShfRight = "\x1b[1;6C"
		t.KeyCtrlShfLeft = "\x1b[1;6D"

		t.KeyShfPgUp = "\x1b[5;2~"
		t.KeyShfPgDn = "\x1b[6;2~"
	}
	// And also for Home and End
	if t.KeyShfHome == "\x1b[1;2H" && t.KeyShfEnd == "\x1b[1;2F" {
		t.KeyCtrlHome = "\x1b[1;5H"
		t.KeyCtrlEnd = "\x1b[1;5F"
		t.KeyAltHome = "\x1b[1;9H"
		t.KeyAltEnd = "\x1b[1;9F"
		t.KeyCtrlShfHome = "\x1b[1;6H"
		t.KeyCtrlShfEnd = "\x1b[1;6F"
		t.KeyAltShfHome = "\x1b[1;4H"
		t.KeyAltShfEnd = "\x1b[1;4F"
		t.KeyMetaShfHome = "\x1b[1;10H"
		t.KeyMetaShfEnd = "\x1b[1;10F"
	}

	// And the same thing for rxvt and workalikes (Eterm, aterm, etc.)
	// It seems that urxvt at least send escaped as ALT prefix for these,
	// although some places seem to indicate a separate ALT key sesquence.
	if t.KeyShfRight == "\x1b[c" && t.KeyShfLeft == "\x1b[d" {
		t.KeyShfUp = "\x1b[a"
		t.KeyShfDown = "\x1b[b"
		t.KeyCtrlUp = "\x1b[Oa"
		t.KeyCtrlDown = "\x1b[Ob"
		t.KeyCtrlRight = "\x1b[Oc"
		t.KeyCtrlLeft = "\x1b[Od"
	}
	if t.KeyShfHome == "\x1b[7$" && t.KeyShfEnd == "\x1b[8$" {
		t.KeyCtrlHome = "\x1b[7^"
		t.KeyCtrlEnd = "\x1b[8^"
	}

	// If the kmous entry is present, then we need to record the
	// the codes to enter and exit mouse mode.  Sadly, this is not
	// part of the terminfo databases anywhere that I've found, but
	// is an extension.  The escapedape codes are documented in the XTerm
	// manual, and all terminals that have kmous are expected to
	// use these same codes, unless explicitly configured otherwise
	// vi XM.  Note that in any event, we only known how to parse either
	// x11 or SGR mouse events -- if your terminal doesn't support one
	// of these two forms, you maybe out of luck.
	t.MouseMode = tc.getstr("XM")
	if t.Mouse != "" && t.MouseMode == "" {
		// we anticipate that all xterm mouse tracking compatible
		// terminals understand mouse tracking (1000), but we hope
		// that those that don't understand any-event tracking (1003)
		// will at least ignore it.  Likewise we hope that terminals
		// that don't understand SGR reporting (1006) just ignore it.
		t.MouseMode = "%?%p1%{1}%=%t%'h'%Pa%e%'l'%Pa%;" +
			"\x1b[?1000%ga%c\x1b[?1002%ga%c\x1b[?1003%ga%c\x1b[?1006%ga%c"
	}

	// We only support colors in ANSI 8 or 256 color mode.
	if t.Colors < 8 || t.SetFg == "" {
		t.Colors = 0
	}
	if t.SetCursor == "" {
		return nil, "", errNotAddressable
	}

	// For padding, we lookup the pad char.  If that isn't present,
	// and npc is *not* set, then we assume a null byte.
	t.PadChar = tc.getstr("pad")
	if t.PadChar == "" {
		if !tc.getflag("npc") {
			t.PadChar = "\u0000"
		}
	}

	// For terminals that use "standard" SGR sequences, lets combine the
	// foreground and background together.
	if strings.HasPrefix(t.SetFg, "\x1b[") &&
		strings.HasPrefix(t.SetBg, "\x1b[") &&
		strings.HasSuffix(t.SetFg, "m") &&
		strings.HasSuffix(t.SetBg, "m") {
		fg := t.SetFg[:len(t.SetFg)-1]
		r := regexp.MustCompile("%p1")
		bg := r.ReplaceAllString(t.SetBg[2:], "%p2")
		t.SetFgBg = fg + ";" + bg
	}

	return t, tc.desc, nil
}