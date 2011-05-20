// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is ripped directly from Go sources src/pkg/syscall/exec_unix.go.
// Modified fork to support controlling terminal acquisition by the child.
// If there is a hell, I am surely going there for this.

package term

import (
	"syscall"
	"os"
	"unsafe"
)

type procAttr struct {
	Setsid	 bool		// Create session.
	Ctty	   int		 // Controlling terminal fd if Setsid (-1 for none)
	Ptrace	 bool		// Enable tracing.
	Dir		string	  // Current working directory.
	Env		[]string	// Environment.
	Files	  []int	   // File descriptors.
	Chroot	 string	  // Chroot.
	Credential *syscall.Credential // Credential.
}

// StartProcessOnTtty works like the POSIX os.StartProcess except that it
// initiates a new session in the child process and takes a tty fd to be
// handed over as the session's controlling terminal
func StartProcessOnTty(name string, argv []string, attr *os.ProcAttr, tty int) (pid int, err os.Error) {
	sysattr := &procAttr{
		Setsid: true,
		Ctty: tty,
		Dir: attr.Dir,
		Env: attr.Env,
	}

	if sysattr.Env == nil {
		sysattr.Env = os.Environ()
	}

	intfd := make([]int, len(attr.Files))
	for i,f := range attr.Files {
		if f == nil {
			intfd[i] = -1
		} else {
			intfd[i] = f.Fd()
		}
	}
	sysattr.Files = intfd

	pid, e := forkExec(name, argv, sysattr)
	if e != 0 {
		return -1,&os.PathError{"fork/exec", name, os.Errno(e)}
	}
	return pid,nil
}

func fcntl(fd, cmd, arg int) (int,int) {
	r,_,e := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), uintptr(cmd), uintptr(arg))
	return int(r),int(e)
}

func read(fd int, buf *byte, nbuf int) (int,int) {
	r,_,e := syscall.Syscall(syscall.SYS_READ, uintptr(fd), uintptr(unsafe.Pointer(buf)), uintptr(nbuf))
	return int(r),int(e)
}

func forkAndExecInChild(argv0 *byte, argv, envv []*byte, chroot, dir *byte, attr *procAttr, pipe int) (pid int, err int) {
	// Declare all variables at top in case any
	// declarations require heap allocation (e.g., err1).
	var r1, r2, err1 uintptr
	var nextfd int
	var i int

	// guard against side effects of shuffling fds below.
	fd := append([]int(nil), attr.Files...)

	darwin := syscall.OS == "darwin"

	// About to call fork.
	// No more allocation or calls of non-assembly functions.
	r1, r2, err1 = syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if err1 != 0 {
		return 0, int(err1)
	}

	// On Darwin:
	//	r1 = child pid in both parent and child.
	//	r2 = 0 in parent, 1 in child.
	// Convert to normal Unix r1 = 0 in child.
	if darwin && r2 == 1 {
		r1 = 0
	}

	if r1 != 0 {
		// parent; return PID
		return int(r1), 0
	}

	// Fork succeeded, now in child.

	// Enable tracing if requested.
	if attr.Ptrace {
		_, _, err1 = syscall.RawSyscall(syscall.SYS_PTRACE, uintptr(syscall.PTRACE_TRACEME), 0, 0)
		if err1 != 0 {
			goto childerror
		}
	}

	// Session ID
	if attr.Setsid {
		_, _, err1 = syscall.RawSyscall(syscall.SYS_SETSID, 0, 0, 0)
		if err1 != 0 {
			goto childerror
		}
		if attr.Ctty >= 0 {
			_, _, err1 = syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(attr.Ctty), uintptr(TIOCSCTTY), 0)
			if err1 != 0 {
				goto childerror
			}
		}
	}

	// Chroot
	if chroot != nil {
		_, _, err1 = syscall.RawSyscall(syscall.SYS_CHROOT, uintptr(unsafe.Pointer(chroot)), 0, 0)
		if err1 != 0 {
			goto childerror
		}
	}

	// User and groups
	if attr.Credential != nil {
		ngroups := uintptr(len(attr.Credential.Groups))
		groups := uintptr(0)
		if ngroups > 0 {
			groups = uintptr(unsafe.Pointer(&attr.Credential.Groups[0]))
		}
		_, _, err1 = syscall.RawSyscall(syscall.SYS_SETGROUPS, ngroups, groups, 0)
		if err1 != 0 {
			goto childerror
		}
		_, _, err1 = syscall.RawSyscall(syscall.SYS_SETGID, uintptr(attr.Credential.Gid), 0, 0)
		if err1 != 0 {
			goto childerror
		}
		_, _, err1 = syscall.RawSyscall(syscall.SYS_SETUID, uintptr(attr.Credential.Uid), 0, 0)
		if err1 != 0 {
			goto childerror
		}
	}

	// Chdir
	if dir != nil {
		_, _, err1 = syscall.RawSyscall(syscall.SYS_CHDIR, uintptr(unsafe.Pointer(dir)), 0, 0)
		if err1 != 0 {
			goto childerror
		}
	}

	// Pass 1: look for fd[i] < i and move those up above len(fd)
	// so that pass 2 won't stomp on an fd it needs later.
	nextfd = int(len(fd))
	if pipe < nextfd {
		_, _, err1 = syscall.RawSyscall(syscall.SYS_DUP2, uintptr(pipe), uintptr(nextfd), 0)
		if err1 != 0 {
			goto childerror
		}
		syscall.RawSyscall(syscall.SYS_FCNTL, uintptr(nextfd), syscall.F_SETFD, syscall.FD_CLOEXEC)
		pipe = nextfd
		nextfd++
	}
	for i = 0; i < len(fd); i++ {
		if fd[i] >= 0 && fd[i] < int(i) {
			_, _, err1 = syscall.RawSyscall(syscall.SYS_DUP2, uintptr(fd[i]), uintptr(nextfd), 0)
			if err1 != 0 {
				goto childerror
			}
			syscall.RawSyscall(syscall.SYS_FCNTL, uintptr(nextfd), syscall.F_SETFD, syscall.FD_CLOEXEC)
			fd[i] = nextfd
			nextfd++
			if nextfd == pipe { // don't stomp on pipe
				nextfd++
			}
		}
	}

	// Pass 2: dup fd[i] down onto i.
	for i = 0; i < len(fd); i++ {
		if fd[i] == -1 {
			syscall.RawSyscall(syscall.SYS_CLOSE, uintptr(i), 0, 0)
			continue
		}
		if fd[i] == int(i) {
			// dup2(i, i) won't clear close-on-exec flag on Linux,
			// probably not elsewhere either.
			_, _, err1 = syscall.RawSyscall(syscall.SYS_FCNTL, uintptr(fd[i]), syscall.F_SETFD, 0)
			if err1 != 0 {
				goto childerror
			}
			continue
		}
		// The new fd is created NOT close-on-exec,
		// which is exactly what we want.
		_, _, err1 = syscall.RawSyscall(syscall.SYS_DUP2, uintptr(fd[i]), uintptr(i), 0)
		if err1 != 0 {
			goto childerror
		}
	}

	// By convention, we don't close-on-exec the fds we are
	// started with, so if len(fd) < 3, close 0, 1, 2 as needed.
	// Programs that know they inherit fds >= 3 will need
	// to set them close-on-exec.
	for i = len(fd); i < 3; i++ {
		syscall.RawSyscall(syscall.SYS_CLOSE, uintptr(i), 0, 0)
	}

	// Time to exec.
	_, _, err1 = syscall.RawSyscall(syscall.SYS_EXECVE,
		uintptr(unsafe.Pointer(argv0)),
		uintptr(unsafe.Pointer(&argv[0])),
		uintptr(unsafe.Pointer(&envv[0])))

childerror:
	// send error code on pipe
	syscall.RawSyscall(syscall.SYS_WRITE, uintptr(pipe), uintptr(unsafe.Pointer(&err1)), uintptr(unsafe.Sizeof(err1)))
	for {
		syscall.RawSyscall(syscall.SYS_EXIT, 253, 0, 0)
	}

	// Calling panic is not actually safe,
	// but the for loop above won't break
	// and this shuts up the compiler.
	panic("unreached")
}

var zeroAttributes procAttr

func forkExec(argv0 string, argv []string, attr *procAttr) (pid int, err int) {
	var p [2]int
	var n int
	var err1 uintptr
	var wstatus syscall.WaitStatus

	if attr == nil {
		attr = &zeroAttributes
	}

	p[0] = -1
	p[1] = -1

	// Convert args to C form.
	argv0p := syscall.StringBytePtr(argv0)
	argvp := syscall.StringArrayPtr(argv)
	envvp := syscall.StringArrayPtr(attr.Env)

	if syscall.OS == "freebsd" && len(argv[0]) > len(argv0) {
		argvp[0] = argv0p
	}

	var chroot *byte
	if attr.Chroot != "" {
		chroot = syscall.StringBytePtr(attr.Chroot)
	}
	var dir *byte
	if attr.Dir != "" {
		dir = syscall.StringBytePtr(attr.Dir)
	}

	// Acquire the fork lock so that no other threads
	// create new fds that are not yet close-on-exec
	// before we fork.
	syscall.ForkLock.Lock()

	// Allocate child status pipe close on exec.
	if err = syscall.Pipe(p[0:]); err != 0 {
		goto error
	}
	if _, err = fcntl(p[0], syscall.F_SETFD, syscall.FD_CLOEXEC); err != 0 {
		goto error
	}
	if _, err = fcntl(p[1], syscall.F_SETFD, syscall.FD_CLOEXEC); err != 0 {
		goto error
	}

	// Kick off child.
	pid, err = forkAndExecInChild(argv0p, argvp, envvp, chroot, dir, attr, p[1])
	if err != 0 {
	error:
		if p[0] >= 0 {
			syscall.Close(p[0])
			syscall.Close(p[1])
		}
		syscall.ForkLock.Unlock()
		return 0, err
	}
	syscall.ForkLock.Unlock()

	// Read child error status from pipe.
	syscall.Close(p[1])
	n, err = read(p[0], (*byte)(unsafe.Pointer(&err1)), unsafe.Sizeof(err1))
	syscall.Close(p[0])
	if err != 0 || n != 0 {
		if n == unsafe.Sizeof(err1) {
			err = int(err1)
		}
		if err == 0 {
			err = syscall.EPIPE
		}

		// Child failed; wait for it to exit, to make sure
		// the zombies don't accumulate.
		_, err1 := syscall.Wait4(pid, &wstatus, 0, nil)
		for err1 == syscall.EINTR {
			_, err1 = syscall.Wait4(pid, &wstatus, 0, nil)
		}
		return 0, err
	}

	// Read got EOF, so pipe closed on exec, so exec succeeded.
	return pid, 0
}

