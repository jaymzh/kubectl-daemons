// +build windows

package cmd

import (
    "k8s.io/client-go/tools/remotecommand"
)

func monitorTerminalResize(sizeQueue chan remotecommand.TerminalSize) {
    // no idea what to do for windows, so we'll ignore it for now
}
