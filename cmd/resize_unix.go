// +build !windows

package cmd

import (
    "os"
    "os/signal"
    "syscall"

    "golang.org/x/term"
    "k8s.io/client-go/tools/remotecommand"
)

func monitorTerminalResize(sizeQueue chan remotecommand.TerminalSize) {
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGWINCH)
    defer signal.Stop(ch)

    for range ch {
        if width, height, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
            sizeQueue <- remotecommand.TerminalSize{
                Width: uint16(width), Height: uint16(height),
            }
        }
    }
}
