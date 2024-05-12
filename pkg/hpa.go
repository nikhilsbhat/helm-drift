package pkg

import (
	"bufio"
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (drift *Drift) IsManagedByHPA(name, kind, nameSpace string) (bool, error) {
	config, err := clientcmd.BuildConfigFromFlags("", drift.kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	// Create a Kubernetes clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	response, err := clientSet.AutoscalingV2().HorizontalPodAutoscalers(nameSpace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	var isManagedByHPA bool

	for _, item := range response.Items {
		if item.Spec.ScaleTargetRef.Name == name && item.Spec.ScaleTargetRef.Kind == kind {
			drift.log.Debugf("the '%s' '%s' is managed by hpa so enabling '--server-side=true'", kind, name)

			isManagedByHPA = true

			break
		}

		drift.log.Debugf("looks like the '%s' '%s' is not managed by hpa hence not setting additional '--field-manager'", kind, name)
	}

	return isManagedByHPA, nil
}

func (drift *Drift) WasScaledByHpa(diffOutput string) (bool, error) {
	stringReader := strings.NewReader(diffOutput)

	scanner := bufio.NewScanner(stringReader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "+  replicas:") {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}
