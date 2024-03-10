package cmd

import (
    "github.com/spf13/cobra"
    "io"
    
    "k8s.io/cli-runtime/pkg/genericclioptions"
)

type dshCmd struct {
    out io.Writer
}

func NewDshCommand(streams genericclioptions.IOStreams) *cobra.Command {
    var context string
    var namespace string
    var nodeName string

    dshCmd := &cobra.Command{
        Use: "d <subcommand>",
        Short: "Various helpers for daemonsets",
        SilenceUsage: true,
        RunE: func (c *cobra.Command, args []string) error {
            return nil
        },
    }

    dshCmd.PersistentFlags().StringVarP(
        &context, "context", "", "", "Context",
    )
    dshCmd.PersistentFlags().StringVarP(
        &namespace, "namespace", "n", "default", "Namespace to look in",
    )
    dshCmd.PersistentFlags().StringVarP(
        &nodeName, "node", "N", "", "Limit to pods on node",
    )

    dshCmd.AddCommand(newVersionCommand(streams.Out))
    dshCmd.AddCommand(newDshGetCommand(streams.Out, &context, &namespace, &nodeName))
    dshCmd.AddCommand(newDshDeleteCommand(streams.Out, &context, &namespace, &nodeName))
    dshCmd.AddCommand(newDshDescribeCommand(streams.Out, &context, &namespace, &nodeName))
    dshCmd.AddCommand(newDshLogCommand(streams.Out, &context, &namespace, &nodeName))
    dshCmd.AddCommand(newDshListCommand(streams.Out, &context, &namespace, &nodeName))
    return dshCmd
}
