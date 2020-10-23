// +build nacl plan9 windows

package core

// This stub file is for systems that have no termios.

type termiosPrivate struct{}

func (c *core) internalStart() error {
	return ErrNoScreen
}

func (c *core) internalShutdown() error {
	return nil
}

func (c *core) readWinSize() (int, int, error) {
	return 0, 0, ErrNoScreen
}

func (c *core) Beep() error {
	return ErrNoScreen
}
