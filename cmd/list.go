package cmd

import (
    "errors"
    "fmt"
    "github.com/spf13/cobra"
    "io"
)


func newDshListCommand(
    out io.Writer, context *string, namespace *string, nodeName *string,
) *cobra.Command {
    var output string

    dshList := &dshCmd{
        out: out,
    }

    cmd := &cobra.Command{
        Use:   "list [<node>] [<options>]",
        Short: "list daemonsets on a node. You can pass in the node as the arg, or use -N",
        Args: cobra.MatchAll(cobra.MaximumNArgs(1)),
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) == 1 {
                *nodeName = args[0]
            }
            return dshList.getDaemonSets(*context, *namespace, *nodeName, output)
        },
    }

    return cmd
}

func (sv *dshCmd) getDaemonSets(
    context string, namespace string, nodeName string, output string,
) error {
    if nodeName == "" {
        return errors.New("You must specify a node")
    }

    clientset, err := getClientSet(context)
    if err != nil {
        return err
    }

    daemonSets, err := getDaemonSetsForNode(clientset, namespace, nodeName)
    if err != nil {
        return err
    }

    if len(daemonSets) == 0 {
        fmt.Printf("No daemonsets found\n")
        return nil
    }

    for _, item := range daemonSets {
        fmt.Println(item)
    }

    return nil
}
