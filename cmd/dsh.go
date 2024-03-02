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
    var namespace string
    var nodeName string

    dshCmd := &cobra.Command{
        Use: "d",
        Short: "Various helpers for daemonsets",
        SilenceUsage: true,
        RunE: func (c *cobra.Command, args []string) error {
            return nil
        },
    }

    dshCmd.PersistentFlags().StringVarP(
        &namespace, "namespace", "n", "default", "Namespace to look in",
    )
    dshCmd.PersistentFlags().StringVarP(
        &nodeName, "node", "N", "", "Limit to pods on node",
    )

    dshCmd.AddCommand(newVersionCmd(streams.Out))
    dshCmd.AddCommand(NewDshGetCommand(streams.Out, &namespace, &nodeName))
    dshCmd.AddCommand(NewDshDeleteCommand(streams.Out, &namespace, &nodeName))
    return dshCmd
}
