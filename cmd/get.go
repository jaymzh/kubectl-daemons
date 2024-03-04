package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "io"
    "os"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/cli-runtime/pkg/printers"

)


func newDshGetCommand(
    out io.Writer, namespace *string, nodeName *string,
) *cobra.Command {
    dshGet := &dshCmd{
        out: out,
    }

    cmd := &cobra.Command{
        Use:   "get",
        Short: "get pods for <ds>",
        Args: cobra.MatchAll(cobra.MaximumNArgs(1)),
        RunE: func(cmd *cobra.Command, args []string) error {
            ds := ""
            if len(args) == 1 {
                ds = args[0]
            }
            return dshGet.getPods(*namespace, ds, *nodeName)
        },
    }

    return cmd
}

func (sv *dshCmd) getPods(
    namespace string, ds string, nodeName string,
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

    printer := printers.NewTablePrinter(printers.PrintOptions{})
    table := metav1.Table{
        ColumnDefinitions: []metav1.TableColumnDefinition{
            {Name: "NAME"},
            {Name: "READY"},
            {Name: "STATUS"},
            {Name: "RESTARTS"},
            {Name: "AGE"},
        },
    }

    for _, pod := range pods {
        readyCount, totalCount :=
            countReadyContainers(pod.Status.ContainerStatuses)
        row := metav1.TableRow{
            Cells: []interface{}{
                pod.Name,
                fmt.Sprintf("%d/%d", readyCount, totalCount),
                string(pod.Status.Phase),
                fmt.Sprintf("%d", pod.Status.ContainerStatuses[0].RestartCount),
                pod.ObjectMeta.CreationTimestamp.Time.String(),
            },
        }
        table.Rows = append(table.Rows, row)
    }

    err = printer.PrintObj(&table, os.Stdout)
    return err
}
