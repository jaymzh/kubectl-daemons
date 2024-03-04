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
    out io.Writer, namespace *string, nodeName *string,
) *cobra.Command {
    var container string
    var tail int
    var follow bool

    dshLog := &dshCmd{
        out: out,
    }

    cmd := &cobra.Command{
        Use:   "log",
        Short: "Logs for <ds>",
        Args: cobra.MatchAll(cobra.ExactArgs(1)),
        RunE: func(cmd *cobra.Command, args []string) error {
            return dshLog.getLogs(
                *namespace, args[0], *nodeName, container, follow, &tail,
            )
        },
    }

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
    namespace string, ds string, nodeName string, container string, follow bool,
    lines *int,
) error {
    clientset, err := getClientSet()
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