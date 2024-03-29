package cmd

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "io"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDshDeleteCommand(
    out io.Writer, context *string, namespace *string, nodeName *string,
) *cobra.Command {
    dshDelete := &dshCmd{
        out: out,
    }

    cmd := &cobra.Command{
        Use:   "delete <daemonset> [<options>]",
        Short: "delete pods for <daemonset>",
        Long:
`Deletes matching pods from the daemonset and/or node. If only a daemonset is
specified, all pods for that daemonset will be deleted. If only a node is
specified, all pods belonging to daemonsets on that node will be deleted. If
both are specified, the pod in that daemonset on that node will be deleted.`,
        Args: cobra.MatchAll(cobra.ExactArgs(1)),
        RunE: func(cmd *cobra.Command, args []string) error {
            return dshDelete.deletePods(*context, *namespace, args[0], *nodeName)
        },
    }

    return cmd
}

func (sv *dshCmd) deletePods(
    ccontext string, namespace string, ds string, nodeName string,
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

    for _, pod := range pods {
        err := clientset.CoreV1().Pods(namespace).Delete(
            context.TODO(), pod.Name, metav1.DeleteOptions{},
        )
        if err != nil {
            fmt.Printf("Error deleting pod %s: %v\n", pod.Name, err)
        } else {
            fmt.Printf("pod \"%s\" deleted\n", pod.Name)
        }
    }
    return nil
}
