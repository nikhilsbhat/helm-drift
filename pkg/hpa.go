package pkg

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func (drift *Drift) IsManagedByHPA(name, kind, nameSpace string) (bool, error) {
	config, err := buildConfigWithContextFromFlags(drift.kubeContext, drift.kubeConfig)
	if err != nil {
		return false, &errors.DriftError{Message: fmt.Sprintf("building config with context errored with '%v'", err)}
	}

	// Create a Kubernetes clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return false, &errors.DriftError{Message: fmt.Sprintf("creating kubernetes clientsets errored with '%v'", err)}
	}

	response, err := clientSet.AutoscalingV2().HorizontalPodAutoscalers(nameSpace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	var isManagedByHPA bool

	for _, item := range response.Items {
		if item.Spec.ScaleTargetRef.Name == name && item.Spec.ScaleTargetRef.Kind == kind {
			drift.log.Debugf("the '%s' '%s' is managed by hpa hence the drifts for this would be suppressed if enabled", kind, name)

			isManagedByHPA = true

			break
		}
	}

	return isManagedByHPA, nil
}

func buildConfigWithContextFromFlags(context string, kubeConfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
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
