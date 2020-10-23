package encoding

// The names of these constants are chosen to match Terminfo names,
// modulo case, and changing the prefix from ACS_ to Rune.  These are
// the runes we provide extra special handling for, with ASCII fallbacks
// for terminals that lack them.
const (
	Sterling = '£'
	DArrow   = '↓'
	LArrow   = '←'
	RArrow   = '→'
	UArrow   = '↑'
	Bullet   = '·'
	Board    = '░'
	CkBoard  = '▒'
	Degree   = '°'
	Diamond  = '◆'
	GEqual   = '≥'
	Pi       = 'π'
	HLine    = '─'
	Lantern  = '§'
	Plus     = '┼'
	LEqual   = '≤'
	LLCorner = '└'
	LRCorner = '┘'
	NEqual   = '≠'
	PlMinus  = '±'
	S1       = '⎺'
	S3       = '⎻'
	S7       = '⎼'
	S9       = '⎽'
	Block    = '█'
	TTee     = '┬'
	RTee     = '┤'
	LTee     = '├'
	BTee     = '┴'
	ULCorner = '┌'
	URCorner = '┐'
	VLine    = '│'
	Space    = ' '
)
