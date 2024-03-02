package main

import (
    "github.com/jaymzh/kubectl-daemons/cmd"
    "os"
    "k8s.io/cli-runtime/pkg/genericclioptions"
)

var version = "undefined"

func main() {
    cmd.SetVersion(version)

    dshCmd := cmd.NewDshCommand(
        genericclioptions.IOStreams{
            In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr,
        },
    )
    if err := dshCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
