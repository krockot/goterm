// A simple shell server.
//
// Listens on tcp/2323 and serves login shells to clients.
// Connected clients get an 80x25 pty for a controlling terminal.
// 

package main

import(
    "term"
    "fmt"
    "net"
    "os"
    "syscall"
    "time"
)

func Serve(conn net.Conn) {
    defer conn.Close()
    pid,tty,pty,err := term.ForkPty(
        "/bin/login",
        []string{"/bin/login"},
        term.DefaultAttributes(),
        term.NewWindowSize(80,25))

    if err != nil {
        fmt.Fprintf(os.Stderr, "ForkExecPty failed: %v\n", err)
        return
    }

    pty.Close()

    defer os.Wait(pid, 0)
    defer syscall.Kill(pid, 9)
    defer pty.Close()

    running := true

    go func() {
        buffer := make([]byte, 64)
        for n,e := conn.Read(buffer); e == nil && running; n,e = conn.Read(buffer) {
            tty.Write(buffer[:n])
        }
        running = false
    }()

    go func() {
        buffer := make([]byte, 64)
        var n int
        var e os.Error
        for n,e = tty.Read(buffer); e == nil && running; n,e = tty.Read(buffer) {
            conn.Write(buffer[:n])
        }
        running = false
    }()

    tick := time.NewTicker(1e9)
    for running {
        select {
        case <-tick.C:
            msg, err := os.Wait(pid, os.WNOHANG)
            if err == nil && msg.Pid == pid {
                running = false
            }
        }
    }
}

func main() {
    l,err := net.Listen("tcp", ":23")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }

    i := 0
    for true {
        i++
        c,err := l.Accept()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            return
        }

        go Serve(c)
    }
}

