package cmd

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "io"
    "time"

    "golang.org/x/text/cases"
    "golang.org/x/text/language"
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDshDescribeCommand(
    out io.Writer, context *string, namespace *string, nodeName *string,
) *cobra.Command {
    dshDescribe := &dshCmd{
        out: out,
    }

    cmd := &cobra.Command{
        Use: "describe",
        Short: "describe pods for <ds>",
        Args: cobra.MatchAll(cobra.MaximumNArgs(1)),
        RunE: func(cmd *cobra.Command, args []string) error {
            ds := ""
            if len(args) == 1 {
                ds = args[0]
            }
            return dshDescribe.describePods(*context, *namespace, ds, *nodeName)
        },
    }

    return cmd
}

func (sv *dshCmd) describePods(
    ccontext string, namespace string, ds string, nodeName string,
) error {
    clientset, err := getClientSet(ccontext)
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
        podinfo, err := clientset.CoreV1().Pods(namespace).Get(
            context.TODO(), pod.Name, metav1.GetOptions{},
        )
        if err != nil {
            return err
        }

        events, err := clientset.CoreV1().Events(namespace).List(
            context.TODO(),
            metav1.ListOptions{
                FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name),
            },
        )
        if err != nil {
            return err
        }

        dumpPod(podinfo, events)
    }

    return nil
}

/*
 * There has to be some function in the kube libraries
 * I can use to do this... but I have yet to find one, so
 * here we are.
 */
func dumpPod(pod *v1.Pod, events *v1.EventList) {
    fmt.Printf("Name:         %s\n", pod.ObjectMeta.Name)
    fmt.Printf("Namespace:    %s\n", pod.ObjectMeta.Namespace)
    fmt.Printf("Priority:     %d\n", *pod.Spec.Priority)
    fmt.Printf("Node:         %s/%s\n", pod.Spec.NodeName, pod.Status.HostIP)
    fmt.Printf("Start Time:   %s\n", pod.Status.StartTime.Format(time.RFC1123))

    fmt.Println("Labels:")
    for key, value := range pod.Labels {
        fmt.Printf("              %s: %s\n", key, value)
    }

    fmt.Println("Annotations:")
    for key, value := range pod.Annotations {
        fmt.Printf("              %s: %s\n", key, value)
    }

    fmt.Printf("Status:       %s\n", pod.Status.Phase)
    fmt.Printf("IP:           %s\n", pod.Status.PodIP)

    fmt.Println("IPs:")
    for _, podIP := range pod.Status.PodIPs {
        fmt.Printf("  IP:           %s\n", podIP.IP)
    }

    if len(pod.OwnerReferences) > 0 {
        fmt.Printf("Controlled By:  %s/%s\n",
            pod.OwnerReferences[0].Kind, pod.OwnerReferences[0].Name,
        )
    } else {
        fmt.Println("Controlled By:  <none>")
    }

    fmt.Printf("Containers:\n")
    for i, containerStatus := range pod.Status.ContainerStatuses {
        containerSpec := pod.Spec.Containers[i]

        containerID := containerStatus.ContainerID
        containerName := containerStatus.Name
        imageID := containerStatus.ImageID

        fmt.Printf("  %s\n", containerName)
        fmt.Printf("    Container ID:  %s\n", containerID)
        fmt.Printf("    Image:         %s\n", containerSpec.Image)
        fmt.Printf("    Image ID:      %s\n", imageID)

        for _, port := range containerSpec.Ports {
            fmt.Printf("      - Port:       %v\n", port.ContainerPort)
            fmt.Printf("        Host Port:  %d\n", port.HostPort)
        }

        fmt.Println("    Command:")
        for _, cmdbit := range containerSpec.Command {
            fmt.Printf("      %s\n", cmdbit)
        }

        state := containerStatus.State
        fmt.Printf("    State:          ")
        switch {
        case state.Waiting != nil:
            fmt.Printf("Waiting\n")
        case state.Running != nil:
            fmt.Printf("Running\n")
            fmt.Printf(
                "      Started:      %s\n",
                state.Running.StartedAt.Format(time.RFC1123),
            )
        case state.Terminated != nil:
            fmt.Printf("Terminated\n")
            fmt.Printf("      Exit Code:    %d\n", state.Terminated.ExitCode)
        default:
            fmt.Printf("Unknown\n")
        }


        // transform the bool (true/false) into a string,
        // so we can capitolize it (True/False) to match kubectl
        readiness := fmt.Sprintf("%t", containerStatus.Ready)
        caser := cases.Title(language.English)
        upperReady := caser.String(readiness)
        fmt.Printf("    Ready:          %s\n", upperReady)

        restartCount := containerStatus.RestartCount
        fmt.Printf("    Restart Count:  %d\n", restartCount)

        fmt.Printf("    Environment:\n")
        for _, envVar := range containerSpec.Env {
            fmt.Printf("      - %s=%s\n", envVar.Name, envVar.Value)
        }

        fmt.Printf("    Mounts:\n")
        for _, mount := range containerSpec.VolumeMounts {
            fmt.Printf("      %s from %s", mount.MountPath, mount.Name)

            if mount.SubPath != "" {
                fmt.Printf(" (subpath: %s)", mount.SubPath)
            }

            if mount.ReadOnly {
                fmt.Printf(" (ro)")
            }

            fmt.Println()
        }
    }

    fmt.Println("Conditions:")
    for _, condition := range pod.Status.Conditions {
        fmt.Printf("  %-20s %v\n", condition.Type, condition.Status)
    }

    fmt.Println("Volumes:")
    for _, volume := range pod.Spec.Volumes {
        fmt.Printf("  %s:\n", volume.Name)

        if volume.VolumeSource.Projected != nil {
            fmt.Printf("    Type:                    Projected\n")
            for _, source := range volume.VolumeSource.Projected.Sources {
                if source.ServiceAccountToken != nil {
                    projection := source.ServiceAccountToken
                    fmt.Printf(
                        "    TokenExpirationSeconds:  %d\n",
                        *projection.ExpirationSeconds,
                    )
                } else if source.ConfigMap != nil {
                    projection := source.ConfigMap
                    fmt.Printf(
                        "    ConfigMapName:           %s\n",
                        projection.LocalObjectReference.Name,
                    )
                    fmt.Printf(
                        "    ConfigMapOptional:       %v\n",
                        projection.Optional,
                    )
                } else if source.DownwardAPI != nil {
                    fmt.Printf(
                        "    DownwardAPI:             %t\n",
                        true,
                    )
                } else {
                    fmt.Println("    (Volume source not recognized)")
                }
            }
        } else {
            fmt.Println("    (Volume source not recognized)")
        }
    }

    fmt.Printf(
        "QoS Class:                   %s\n",
        string(pod.Status.QOSClass),
    )

    fmt.Println("Node-Selectors:")
    for key, val := range pod.Spec.NodeSelector {
        fmt.Printf("  %ss=%s\n", key, val)
    }


    fmt.Println("Tolerations:")
    for _, toleration := range pod.Spec.Tolerations {
        fmt.Printf(
            "  %s:%s op=%s\n",
            toleration.Key,
            toleration.Value,
            toleration.Operator,
        )
    }

    fmt.Println("Events:")
    fmt.Printf(
        "  %-7s %-12s %-5s %-18s %s\n",
        "Type",
        "Reason",
        "Age",
        "From",
        "Message",
    )
    fmt.Printf(
        "  %-7s %-12s %-5s %-18s %s\n",
        "----",
        "------",
        "---",
        "----",
        "-------",
    )
    for _, event := range events.Items {
        fmt.Printf("  %-7s %-12s %-5s %-18s %s\n",
            event.Type,
            event.Reason,
            time.Since(event.LastTimestamp.Time).Round(time.Second),
            event.Source.Component,
            event.Message,
        )
    }
}
