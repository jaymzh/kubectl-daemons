package cmd

import (
    "encoding/json"
    "gopkg.in/yaml.v3"
    "fmt"
    "github.com/spf13/cobra"
    "io"
    "os"
    "strings"
    "time"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/cli-runtime/pkg/printers"
    v1 "k8s.io/api/core/v1"
)


func newDshGetCommand(
    out io.Writer, namespace *string, nodeName *string,
) *cobra.Command {
    var output string

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
            return dshGet.getPods(*namespace, ds, *nodeName, output)
        },
    }

    cmd.Flags().StringVarP(
        &output, "output", "o", "", "Output format. One of wide, json, yaml.",
    )

    return cmd
}

func (sv *dshCmd) getPods(
    namespace string, ds string, nodeName string, output string,
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

    var table metav1.Table
    var printer printers.ResourcePrinter
    if output == "" || output == "wide" {
        printer = printers.NewTablePrinter(printers.PrintOptions{})
        if output == "" {
            table = metav1.Table{
                ColumnDefinitions: []metav1.TableColumnDefinition{
                    {Name: "NAME"},
                    {Name: "READY"},
                    {Name: "STATUS"},
                    {Name: "RESTARTS"},
                    {Name: "AGE"},
                },
            }
        } else if output == "wide" {
            // when I append(table.ColumnDefinitions, ...) I get
            // syntax errors. No idea why.
            table = metav1.Table{
                ColumnDefinitions: []metav1.TableColumnDefinition{
                    {Name: "NAME"},
                    {Name: "READY"},
                    {Name: "STATUS"},
                    {Name: "RESTARTS"},
                    {Name: "AGE"},
                    {Name: "IP"},
                    {Name: "NODE"},
                    {Name: "NOMINATED NODE"},
                    {Name: "READINESS GATES"},
                },
            }
        }
    }

    for _, pod := range pods {
        switch output {
        case "json":
            jsonData, err := json.MarshalIndent(pod, "", "    ")
            if err != nil {
                return err
            }
            fmt.Println(string(jsonData))
        case "yaml":
            yamlData, err := yaml.Marshal(pod)
            if err != nil {
                return err
            }
            fmt.Println(string(yamlData))
        case "", "wide":
            readyCount, totalCount :=
                countReadyContainers(pod.Status.ContainerStatuses)
            age := time.Since(pod.ObjectMeta.CreationTimestamp.Time).Round(
                time.Second,
            )
            row := metav1.TableRow{
                Cells: []interface{}{
                    pod.Name,
                    fmt.Sprintf("%d/%d", readyCount, totalCount),
                    string(pod.Status.Phase),
                    fmt.Sprintf(
                        "%d",
                        pod.Status.ContainerStatuses[0].RestartCount,
                    ),
                    age,
                },
            }
            if (output == "wide") {
                nominatedNode := "<none>"
                if pod.Status.NominatedNodeName != "" {
                    nominatedNode = pod.Status.NominatedNodeName
                }
                var readinessGates []string
                ready := false
                for _, condition := range pod.Status.Conditions {
                    if condition.Type == v1.PodReady &&
                            condition.Status == v1.ConditionTrue {
                        ready = true
                        break
                    }
                    readinessGates = append(
                        readinessGates,
                        string(condition.Type),
                    )
                }
                var readinessGatesStr string
                if ready || len(readinessGates) == 0 {
                    readinessGatesStr = "<none>"
                } else {
                    readinessGatesStr = strings.Join(readinessGates, ", ")
                }

                row.Cells = append(
                    row.Cells,
                    pod.Status.PodIP,
                    pod.Spec.NodeName,
                    nominatedNode,
                    readinessGatesStr,
                )
            }

            table.Rows = append(table.Rows, row)
        }
    }

    if output == "" || output == "wide" {
        err = printer.PrintObj(&table, os.Stdout)
        return err
    }
    return nil
}
