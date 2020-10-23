// +build solaris illumos

package core

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

type termiosPrivate struct {
	tio *unix.Termios
}

func (c *core) internalStart() error {
	var (
		e   error
		raw *unix.Termios
		tio *unix.Termios
	)

	if c.in, e = os.OpenFile("/dev/tty", os.O_RDONLY, 0); e != nil {
		goto failed
	}
	if c.out, e = os.OpenFile("/dev/tty", os.O_WRONLY, 0); e != nil {
		goto failed
	}

	tio, e = unix.IoctlGetTermios(int(c.out.Fd()), unix.TCGETS)
	if e != nil {
		goto failed
	}

	c.termIOSPrv = &termiosPrivate{tio: tio}

	// make a local copy, to make it raw
	raw = &unix.Termios{
		Cflag: tio.Cflag,
		Oflag: tio.Oflag,
		Iflag: tio.Iflag,
		Lflag: tio.Lflag,
		Cc:    tio.Cc,
	}
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8

	// This is setup for blocking reads.  In the past we attempted to
	// use non-blocking reads, but now a separate input loop and timer
	// copes with the problems we had on some systems (BSD/Darwin)
	// where close hung forever.
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	e = unix.IoctlSetTermios(int(c.out.Fd()), unix.TCSETS, raw)
	if e != nil {
		goto failed
	}

	signal.Notify(c.winSizeCh, syscall.SIGWINCH)

	if w, h, e := c.readWinSize(); e == nil && w != 0 && h != 0 {
		c.resize(w, h, false)
	}

	return nil

failed:
	if c.in != nil {
		c.in.Close()
	}
	if c.out != nil {
		c.out.Close()
	}
	return e
}

func (c *core) internalShutdown() error {
	signal.Stop(c.winSizeCh)
	if c.out != nil && c.termIOSPrv != nil {
		unix.IoctlSetTermios(int(c.out.Fd()), unix.TCSETSF, c.termIOSPrv.tio)
		c.out.Close()
	}
	if c.in != nil {
		c.in.Close()
	}
	return nil
}

func (c *core) readWinSize() (int, int, error) {
	wsz, err := unix.IoctlGetWinsize(int(c.out.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return -1, -1, err
	}
	return int(wsz.Col), int(wsz.Row), nil
}

func (c *core) Beep() error {
	if _, err := c.out.Write([]byte{byte(7)}); err != nil {
		if Debug {
			log.Printf("error writing to io : " + err.Error())
		}
	}
	return nil
}
