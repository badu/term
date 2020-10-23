// +build darwin

package core

// The Darwin system is *almost* a real BSD system, but it suffers from a brain damaged TTY driver.
// This TTY driver does not actually wake up in poll() or similar calls, which means that we cannot reliably shut down the terminal without resorting to obscene custom C code and a dedicated poller thread.
//
// So instead, we do a best effort, and simply try to do the close in the background.
// Probably this will cause a leak of two goroutines and maybe also the file descriptor, meaning that applications on Darwin can't reinitialize the screen, but that's probably a very rare behavior, and accepting that is the best of some very poor alternative options.
//
// Maybe someday Apple will fix there tty driver, but its been broken for a long time (probably forever) so holding one's breath is contraindicated.

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

type termiosPrivate syscall.Termios

func (c *core) internalStart() error {
	var (
		e       error
		newtios termiosPrivate
		fd      uintptr
		tios    uintptr
		ioc     uintptr
	)
	c.termIOSPrv = &termiosPrivate{}

	if c.in, e = os.OpenFile("/dev/tty", os.O_RDONLY, 0); e != nil {
		goto failed
	}
	if c.out, e = os.OpenFile("/dev/tty", os.O_WRONLY, 0); e != nil {
		goto failed
	}

	tios = uintptr(unsafe.Pointer(c.termIOSPrv))
	ioc = uintptr(syscall.TIOCGETA)
	fd = uintptr(c.out.Fd())
	if _, _, e1 := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioc, tios, 0, 0, 0); e1 != 0 {
		e = e1
		goto failed
	}

	newtios = *c.termIOSPrv
	newtios.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	newtios.Oflag &^= syscall.OPOST
	newtios.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	newtios.Cflag &^= syscall.CSIZE | syscall.PARENB
	newtios.Cflag |= syscall.CS8

	tios = uintptr(unsafe.Pointer(&newtios))

	ioc = uintptr(syscall.TIOCSETA)
	if _, _, e1 := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioc, tios, 0, 0, 0); e1 != 0 {
		e = e1
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
	if c.out != nil {
		fd := uintptr(c.out.Fd())
		ioc := uintptr(syscall.TIOCSETAF)
		tios := uintptr(unsafe.Pointer(c.termIOSPrv))
		syscall.Syscall6(syscall.SYS_IOCTL, fd, ioc, tios, 0, 0, 0)
		c.out.Close()
	}

	// See above -- we background this call which might help, but really the tty is probably open.
	go func() {
		if c.in != nil {
			c.in.Close()
		}
	}()
	return nil
}

func (c *core) readWinSize() (int, int, error) {
	fd := uintptr(c.out.Fd())
	dim := [4]uint16{}
	dimp := uintptr(unsafe.Pointer(&dim))
	ioc := uintptr(syscall.TIOCGWINSZ)
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioc, dimp, 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(dim[1]), int(dim[0]), nil
}

func (c *core) Beep() error {
	if _, err := c.out.Write([]byte{byte(7)}); err != nil {
		if Debug {
			log.Printf("error writing to io : " + err.Error())
		}
	}
	return nil
}
