// +build linux

package core

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

type termiosPrivate struct {
	tio *unix.Termios
}

func (c *core) internalStart() error {
	const (
		devTTY = "/dev/tty"
	)
	var (
		err error
		raw *unix.Termios
		tio *unix.Termios
	)

	if c.in, err = os.OpenFile(devTTY, os.O_RDONLY, 0); err != nil {
		goto failed
	}

	if c.out, err = os.OpenFile(devTTY, os.O_WRONLY, 0); err != nil {
		goto failed
	}

	tio, err = unix.IoctlGetTermios(int(c.out.Fd()), unix.TCGETS)
	if err != nil {
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
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8

	// This is setup for blocking reads.
	// In the past we attempted to use non-blocking reads, but now a separate input loop and timer copes with the problems we had on some systems (BSD/Darwin) where close hung forever.
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	err = unix.IoctlSetTermios(int(c.out.Fd()), unix.TCSETS, raw)
	if err != nil {
		goto failed
	}

	// Window size change. This is generated on some systems (including GNU) when the terminal driver’s record of the number of rows and columns on the screen is changed. The default action is to ignore it.
	// If a program does full-screen display, it should handle SIGWINCH. When the signal arrives, it should fetch the new screen size and reformat its display accordingly.
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
	return err
}

func (c *core) internalShutdown() error {
	signal.Stop(c.winSizeCh)
	if c.out != nil && c.termIOSPrv != nil {
		if err := unix.IoctlSetTermios(int(c.out.Fd()), unix.TCSETSF, c.termIOSPrv.tio); err != nil {
			return err
		}
		if err := c.out.Close(); err != nil {
			return err
		}
	}

	if c.in != nil {
		if err := c.in.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *core) readWinSize() (int, int, error) {
	wsz, err := unix.IoctlGetWinsize(int(c.out.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return -1, -1, err
	}

	rows, cols := int(wsz.Row), int(wsz.Col)
	if cols == 0 {
		if colsEnv := os.Getenv("COLUMNS"); len(colsEnv) > 0 {
			if cols, err = strconv.Atoi(colsEnv); err != nil {
				return -1, -1, err
			}
		} else {
			cols = c.info.Columns
		}
	}
	if rows == 0 {
		if rowsEnv := os.Getenv("LINES"); len(rowsEnv) > 0 {
			if rows, err = strconv.Atoi(rowsEnv); err != nil {
				return -1, -1, err
			}
		} else {
			rows = c.info.Lines
		}
	}
	return cols, rows, nil
}

func (c *core) Beep() error {
	if _, err := c.out.Write([]byte{byte(7)}); err != nil {
		if Debug {
			log.Printf("error writing to io : " + err.Error())
		}
	}
	return nil
}
