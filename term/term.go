// Package term provides a set of utilities for manipulating terminal and
// pseudo-terminal devices on BSD-like systems.

package term

import (
	//#include "goterm.h"
	//#cgo LDFLAGS: -lutil
	"C"
	"os"
	"syscall"
)

type Terminal struct {
	file *os.File
}

// Returns a terminal controller for the current process's tty and an Error,
// if any.  Equivalent to Open("/dev/tty").
func Mine() (*Terminal, os.Error) {
	file, err := os.OpenFile("/dev/tty", os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return &Terminal{file}, nil
}

// Returns a terminal controller for a given terminal device in the filesystem
// and an Error if any.
func Open(filename string) (*Terminal, os.Error) {
	file, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return &Terminal{file}, nil
}

// Returns the name of the terminal's tty device and an error if
// the terminal's fd is not a tty.
func (t *Terminal) Name() (string, os.Error) {
	cs := C.ttyname(C.int(t.Fd()))
	if cs == nil {
		return "", os.NewError("Not a tty")
	}
	return C.GoString(cs), nil
}

// Implements io.Closer
func (t *Terminal) Close() os.Error {
	if t == nil || t.file == nil {
		return os.EINVAL
	}
	return t.file.Close()
}

// Returns the integer Unix file descriptor referencing the terminal
func (t *Terminal) Fd() int {
	return t.file.Fd()
}

func (t *Terminal) File() *os.File {
	return t.file
}

// Implements io.Reader
func (t *Terminal) Read(buf []byte) (int, os.Error) {
	if t == nil || t.file == nil {
		return 0, os.EINVAL
	}
	return t.file.Read(buf)
}

// Implements io.Writer
func (t *Terminal) Write(buf []byte) (int, os.Error) {
	if t == nil || t.file == nil {
		return 0, os.EINVAL
	}
	return t.file.Write(buf)
}

// Attributes holds terminal control flags.  This structure is a direct analog
// of the standard termios struct defined in termios.h.
type Attributes struct {
	Input        uint32
	Output       uint32
	Control      uint32
	Local        uint32
	ControlChars [C.NCCS]byte
}

func DefaultAttributes() *Attributes {
	return &Attributes{
		IXON | ICRNL,
		OPOST | ONLCR,
		CS8 | CREAD | B38400,
		IEXTEN | CLOCAL | PARENB | ECHOK | ECHOE | ECHO | ICANON | ISIG,
		[C.NCCS]byte{
			3, 28, 127, 21, 4, 0, 1, 0,
			17, 19, 26, 0, 18, 15, 23, 22,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
	}
}

// WindowSize is an analog of the winsize struct defined in sys/ioctl.h.  It
// represents the dimensions (in characters or pixels) of a terminal window.
type WindowSize struct {
	Rows, Cols              uint16
	PixelWidth, PixelHeight uint16
}

func NewWindowSize(cols, rows uint16) *WindowSize {
	return &WindowSize{rows, cols, 0, 0}
}

func NewPixelWindowSize(width, height uint16) *WindowSize {
	return &WindowSize{0, 0, width, height}
}

func makeCAttributes(attr *Attributes) C.struct_termios {
	var cattr C.struct_termios
	cattr.c_iflag = C.tcflag_t(attr.Input)
	cattr.c_oflag = C.tcflag_t(attr.Output)
	cattr.c_cflag = C.tcflag_t(attr.Control)
	cattr.c_lflag = C.tcflag_t(attr.Local)
	for i := 0; i < C.NCCS; i++ {
		cattr.c_cc[i] = C.cc_t(attr.ControlChars[i])
	}
	return cattr
}

func makeGoAttributes(cattr *C.struct_termios) *Attributes {
	attr := &Attributes{}
	attr.Input = uint32(cattr.c_iflag)
	attr.Output = uint32(cattr.c_oflag)
	attr.Control = uint32(cattr.c_cflag)
	attr.Local = uint32(cattr.c_lflag)
	for i := 0; i < C.NCCS; i++ {
		attr.ControlChars[i] = byte(cattr.c_cc[i])
	}
	return attr
}

func makeCWindowSize(sz *WindowSize) C.struct_winsize {
	var winsize C.struct_winsize
	winsize.ws_row = C.ushort(sz.Rows)
	winsize.ws_col = C.ushort(sz.Cols)
	winsize.ws_xpixel = C.ushort(sz.PixelWidth)
	winsize.ws_ypixel = C.ushort(sz.PixelHeight)
	return winsize
}

func makeGoWindowSize(winsize *C.struct_winsize) *WindowSize {
	return &WindowSize{
		uint16(winsize.ws_row),
		uint16(winsize.ws_col),
		uint16(winsize.ws_xpixel),
		uint16(winsize.ws_ypixel)}
}

// Option flags for SetAttributes
const (
	NOW   = int(C.TCSANOW)
	DRAIN = int(C.TCSADRAIN)
	FLUSH = int(C.TCSAFLUSH)
)

// Gets a terminal's attributes.  Akin to tcgetattr().
func (t *Terminal) GetAttributes() (*Attributes, os.Error) {
	var cattr C.struct_termios
	result := int(C.tcgetattr(C.int(t.Fd()), &cattr))
	if result < 0 {
		return nil, os.NewError("Unable to get terminal attributes.")
	}
	return makeGoAttributes(&cattr), nil
}

// Sets a terminal's attributes.  Akin to tcsetattr().
func (t *Terminal) SetAttributes(attr *Attributes, options int) os.Error {
	cattr := makeCAttributes(attr)
	result := C.tcsetattr(C.int(t.Fd()), C.int(options), &cattr)
	if result < 0 {
		return os.NewError("Unable to set terminal attributes.")
	}

	return nil
}

// Convenience function for enabling specific attributes.  The flags set in
// the Attributes argument are ORed with the terminal's current attribute set.
// Control character settings are unchanged.
func (t *Terminal) EnableAttributes(attr *Attributes, options int) os.Error {
	oattr, err := t.GetAttributes()
	if err != nil {
		return err
	}
	oattr.Input |= attr.Input
	oattr.Output |= attr.Output
	oattr.Control |= attr.Control
	oattr.Local |= attr.Local
	return t.SetAttributes(oattr, options)
}

// Convenience function for disabling specific attributes.  The flags set in
// the Attributes argument are cleared from terminal's current attribute set.
// Control character settings are unchanged.
func (t *Terminal) DisableAttributes(attr *Attributes, options int) os.Error {
	oattr, err := t.GetAttributes()
	if err != nil {
		return err
	}
	oattr.Input &^= attr.Input
	oattr.Output &^= attr.Output
	oattr.Control &^= attr.Control
	oattr.Local &^= attr.Local
	return t.SetAttributes(oattr, options)
}

// Gets the window size of a terminal
func (t *Terminal) GetWindowSize() (*WindowSize, os.Error) {
	var winsize C.struct_winsize
	if C.goterm_get_window_size(C.int(t.Fd()), &winsize) < 0 {
		return nil, os.NewError("Unable to get window size.")
	}
	return makeGoWindowSize(&winsize), nil
}

// Sets the window size of a terminal
func (t *Terminal) SetWindowSize(sz *WindowSize) os.Error {
	winsize := makeCWindowSize(sz)
	if C.goterm_set_window_size(C.int(t.Fd()), &winsize) < 0 {
		return os.NewError("Unable to set window size.")
	}
	return nil
}

// Opens an available pseudo-terminal and returns Terminal controllers for the master and
// slave.  Also returns an Error, if any.
func OpenPty(attr *Attributes, size *WindowSize) (*Terminal, *Terminal, os.Error) {
	cattr := makeCAttributes(attr)
	csize := makeCWindowSize(size)

	// see Go sources @src/pkg/syscall/exec.go for ForkLock info
	syscall.ForkLock.RLock()
	result := C.goterm_openpty(&cattr, &csize)
	if result.result < 0 {
		syscall.ForkLock.RUnlock()
		return nil, nil, os.NewError("Unable to open pty.")
	}
	syscall.CloseOnExec(int(result.master))
	syscall.CloseOnExec(int(result.slave))
	syscall.ForkLock.RUnlock()

	master := &Terminal{os.NewFile(int(result.master), C.GoString(C.ttyname(result.master)))}
	slave := &Terminal{os.NewFile(int(result.slave), C.GoString(C.ttyname(result.slave)))}
	return master, slave, nil
}

// Opens a pseudo-terminal, forks, sets up the pty slave as the new child process's controlling terminal and
// stdin/stdout/stderr, and execs the given command in the child process.  Returns the master
// terminal control, the child pid, and an Error if any.
func ForkPty(name string, argv []string, attr *Attributes, size *WindowSize) (*Terminal, int, os.Error) {
	master, slave, err := OpenPty(attr, size)
	if err != nil {
		return nil, -1, err
	}

	procattr := &os.ProcAttr{
		Dir: "",
		Env: nil,
		Files: []*os.File{
			os.NewFile(slave.Fd(), "/dev/stdin"),
			os.NewFile(slave.Fd(), "/dev/stdout"),
			os.NewFile(slave.Fd(), "/dev/stderr"),
		},
		Sys: &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
		},
	}

	proc, err := os.StartProcess(name, argv, procattr)
	if err != nil {
		return nil, -1, err
	}
	return master, proc.Pid, nil
}
