package cmd

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "io"
    "os"
    "errors"

    v1 "k8s.io/api/core/v1"
)

func newDshLogCommand(
    out io.Writer, context *string, namespace *string, nodeName *string,
) *cobra.Command {
    var container string
    var tail int
    var follow bool

    dshLog := &dshCmd{
        out: out,
    }

    cmd := &cobra.Command{
        Use:   "log <daemonset> [<options>]",
        Short: "get logs for <daemonset>",
        Long:
`Get logs for pods matching a given daemonset and node. Any combination is
allowed.  If only a node is specified logs from all pods owned by a daemonset
on that node will be shown. If only a daemonset is specified, all pods in that
daemonset will have their logs shown.`,
        Args: cobra.MatchAll(cobra.ExactArgs(1)),
        RunE: func(cmd *cobra.Command, args []string) error {
            return dshLog.getLogs(
                *context, *namespace, args[0], *nodeName, container, follow, &tail,
            )
        },
    }

    cmd.Aliases = []string{"logs"}

    cmd.Flags().IntVarP(
        &tail, "tail", "t", 0, "Number of lines",
    )
    cmd.Flags().StringVarP(
        &container, "container", "c", "", "Container to get logs for",
    )
    cmd.Flags().BoolVarP(
        &follow, "follow", "f", false, "Specify if the logs should be stream",
    )


    return cmd
}

func (sv *dshCmd) getLogs(
    ccontext string, namespace string, ds string, nodeName string, container string,
    follow bool, lines *int,
) error {
    clientset, _, err := getClientSet(ccontext)
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
        return errors.New("Matched more then one pod somehow")
    }

    var tlines *int64
    if lines != nil && *lines != 0 {
        tlines = new(int64)
        *tlines = int64(*lines)
    }
    logOptions := &v1.PodLogOptions{
        Container: container,
        TailLines: tlines,
        Follow: follow,
    }

    podLog, err := clientset.CoreV1().Pods(namespace).GetLogs(
        pods[0].Name, logOptions,
    ).Stream(context.TODO())
    if err != nil {
        fmt.Printf("Error retrieving logs: %v\n", err)
        os.Exit(1)
    }
    defer podLog.Close()

    buf := make([]byte, 4096)
    for {
        bytesRead, err := podLog.Read(buf)
        if err != nil {
            break
        }
        if bytesRead > 0 {
            fmt.Print(string(buf[:bytesRead]))
        }
    }
    return nil
}
