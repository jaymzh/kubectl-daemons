package cmd

import (
    "context"
    "errors"
    "fmt"
    "github.com/spf13/cobra"
    "io"
    "os"
    "os/signal"
    "syscall"

    "k8s.io/client-go/kubernetes/scheme"
    "k8s.io/client-go/tools/remotecommand"
    "golang.org/x/term"

    v1 "k8s.io/api/core/v1"
)


func newDshExecCommand(
    out io.Writer, context *string, namespace *string, nodeName *string,
) *cobra.Command {
    var container string
    var stdin bool
    var tty bool

    dshExec := &dshCmd{
        out: out,
    }

    cmd := &cobra.Command{
        Use:   "exec <daemonset> [<options>] -- <command> [args...]",
        Short: "execute arbitrary commands in pod for <daemonset>",
        Args: cobra.MatchAll(cobra.MinimumNArgs(1)),
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) > 1 && cmd.ArgsLenAtDash() != -1 {
                remoteCommand := args[cmd.ArgsLenAtDash():]
                return dshExec.execPod(
                    *context, *namespace, args[0], *nodeName, container, stdin,
                    tty, remoteCommand,
                )
            } else {
                return errors.New("At least some command is required")
            }
        },
    }

    cmd.Flags().StringVarP(
        &container, "container", "c", "", "The container to exec into",
    )
    cmd.Flags().BoolVarP(
        &stdin, "stdin", "i", false, "Pass stdin to the container",
    )
    cmd.Flags().BoolVarP(
        &tty, "tty", "t", false, "Stdin is a TTY",
    )
    return cmd
}

type terminalSizeQueue struct {
    sizeQueue chan remotecommand.TerminalSize
}

func (t *terminalSizeQueue) Next() *remotecommand.TerminalSize {
    size, ok := <- t.sizeQueue
    if !ok {
        return nil
    }
    return &size
}

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

func (sv *dshCmd) execPod(
    kcontext string, namespace string, ds string, nodeName string,
    container string, stdin bool, tty bool, cmd []string,
) error {
    clientset, config, err := getClientSet(kcontext)
    if err != nil {
        return err
    }

    pods, err := getPodsForDaemonSet(clientset, ds, namespace, nodeName)
    if err != nil {
        return err
    }

    if len(pods) == 0 {
        fmt.Printf("No pods found\n")
        return nil
    }

    if len(pods) > 1 {
        fmt.Printf("More than one pod found, wut?!")
        return nil
    }

    req := clientset.CoreV1().RESTClient().
        Post().
        Resource("pods").
        Name(pods[0].Name).
        Namespace(namespace).
        SubResource("exec").
        VersionedParams(&v1.PodExecOptions{
            Command:   cmd,
            Container: container,
            Stdin:     stdin,
            Stdout:    true,
            Stderr:    true,
            TTY:       tty,
        }, scheme.ParameterCodec)

    exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
    if err != nil {
        return err
    }

    var streamOptions remotecommand.StreamOptions
    if tty {
        initialState, err := term.MakeRaw(int(os.Stdin.Fd()))
        if err != nil {
            return err
        }
        defer func() {
            if err := term.Restore(int(os.Stdin.Fd()), initialState); err != nil {
                // Handle the error, e.g., log it or print it.
                fmt.Fprintf(os.Stderr, "Error restoring terminal: %v\n", err)
            }
        }()

        // This queue has to be made with size 1 so that sending the original
        // size before there's a listener won't cause a freeze
        sizeQueue := make(chan remotecommand.TerminalSize, 1)
        tQueue := &terminalSizeQueue{sizeQueue: sizeQueue}

        // Send the initial terminal size.
        if width, height, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
            sizeQueue <- remotecommand.TerminalSize{
                Width: uint16(width), Height: uint16(height),
            }
        }

        go monitorTerminalResize(sizeQueue)

        streamOptions = remotecommand.StreamOptions{
            Stdin:             os.Stdin,
            Stdout:            os.Stdout,
            Stderr:            os.Stdout,
            Tty:               tty,
            TerminalSizeQueue: tQueue,
        }
    } else {
        streamOptions = remotecommand.StreamOptions{
            Stdin:  os.Stdin,
            Stdout: os.Stdout,
            Stderr: os.Stdout,
            Tty:    tty,
        }
    }

	if !stdin {
		streamOptions.Stdin = nil
	}

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    err = exec.StreamWithContext(ctx, streamOptions)
    return err
}
