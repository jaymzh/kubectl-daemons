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

func getPodsForDaemonSet(
    clientset *kubernetes.Clientset, daemonSetName, namespace string,
    nodeName string,
) ([]corev1.Pod, error) {
	var pods []corev1.Pod

    podList, err := clientset.CoreV1().Pods(namespace).List(
        context.TODO(), metav1.ListOptions{},
    )
    if err != nil {
        return nil, err
    }

	for _, pod := range podList.Items {
        if nodeName != "" && pod.Spec.NodeName != nodeName {
            continue
        }
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "DaemonSet" && (
                    daemonSetName == "" || owner.Name == daemonSetName) {
				pods = append(pods, pod)
				break
			}
		}
	}

	return pods, nil
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
