package cmd

import (
    "context"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/kubernetes"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getClientSet(context string) (*kubernetes.Clientset, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    configOverrides := &clientcmd.ConfigOverrides{
        CurrentContext: context,
    }
    kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
        loadingRules, configOverrides,
    )

    config, err := kubeConfig.ClientConfig()
    if err != nil {
        return nil, err
    }

    clientset, err := kubernetes.NewForConfig(config)
    return clientset, err
}

func getDaemonSetsForNode(
    clientset *kubernetes.Clientset, namespace string, nodeName string,
) ([]string, error) {
    _, daemonSets, err := getDaemonSetInfo(
        clientset, "", namespace, nodeName,
    )
    if err != nil {
        return nil, err
    }

    return daemonSets, err
}

func getPodsForDaemonSet(
    clientset *kubernetes.Clientset, daemonSetName, namespace string,
    nodeName string,
) ([]corev1.Pod, error) {
    pods, _, err := getDaemonSetInfo(
        clientset, daemonSetName, namespace, nodeName,
    )
    if err != nil {
        return nil, err
    }
    return pods, err
}

func getDaemonSetInfo(
    clientset *kubernetes.Clientset, daemonSetName, namespace string,
    nodeName string,
) ([]corev1.Pod, []string, error) {
    var pods []corev1.Pod
    ds_set := make(map[string]struct{})

    podList, err := clientset.CoreV1().Pods(namespace).List(
        context.TODO(), metav1.ListOptions{},
    )
    if err != nil {
        return nil, nil, err
    }

    for _, pod := range podList.Items {
        if nodeName != "" && pod.Spec.NodeName != nodeName {
            continue
        }
        for _, owner := range pod.OwnerReferences {
            if owner.Kind == "DaemonSet" && (
                    daemonSetName == "" || owner.Name == daemonSetName) {
                pods = append(pods, pod)
                ds_set[owner.Name] = struct{}{}
                break
            }
        }
    }

    var daemonSets []string
    for k := range ds_set {
        daemonSets = append(daemonSets, k)
    }

    return pods, daemonSets, nil
}

func countReadyContainers(
    containerStatuses []corev1.ContainerStatus,
) (int, int) {
    var readyCount, totalCount int
    for _, status := range containerStatuses {
        totalCount++
        if status.Ready {
            readyCount++
        }
    }
    return readyCount, totalCount
}
