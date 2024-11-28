package pkg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/acarl005/stripansi"
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

func (drift *Drift) HasOnlyChangesScaledByHpa(diffOutput string) (bool, error) {
	customDiff := ""
	if drift.CustomDiff != "" {
		customDiff = drift.CustomDiff
	} else if os.Getenv("KUBECTL_EXTERNAL_DIFF") != "" {
		customDiff = os.Getenv("KUBECTL_EXTERNAL_DIFF")
	}

	diffToolUsed := ""
	if customDiff != "" {
		diffToolUsed = strings.Split(customDiff, " ")[0]
	} else {
		diffToolUsed = "diff"
	}

	drift.log.Infof("custom diff: %s", diffToolUsed)

	if diffToolUsed != "diff" && diffToolUsed != "dyff" {
		drift.log.Warnf("--ignore-hpa-changes currently only supports diff and dyff, not '%s'", diffToolUsed)
		return false, nil
	}

	hasOnlyChangesScaledByHpa := true

	diffOutput = stripansi.Strip(diffOutput)

	stringReader := strings.NewReader(diffOutput)
	scanner := bufio.NewScanner(stringReader)
	for scanner.Scan() {
		line := scanner.Text()

		if diffToolUsed == "diff" && !diffLineHasChangesNonHpaRelated(line) {
			continue
		} else if diffToolUsed == "dyff" && !dyffLineHasChangesNonHpaRelated(line) {
			continue
		}

		hasOnlyChangesScaledByHpa = false
		break
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return hasOnlyChangesScaledByHpa, nil
}

func diffLineHasChangesNonHpaRelated(line string) bool {
	// skip diff output lines starting with +++ or ---
	if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
		return false
	}

	// skip lines that have no changes
	if !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") {
		return false
	}

	// check if the line changed is related to replicas or generation, then continue, since we are looking for other fields changed besides replicas and generation
	if strings.Contains(line, "+  replicas:") || strings.Contains(line, "-  replicas:") ||
		strings.Contains(line, "+  generation:") || strings.Contains(line, "-  generation:") {
		return false
	}

	return true
}

func dyffLineHasChangesNonHpaRelated(line string) bool {
	// skip empty lines and lines starting with space
	if line == "" || strings.HasPrefix(line, " ") {
		return false
	}

	// check if the line changed is related to replicas or generation, then continue, since we are looking for other fields changed besides replicas and generation
	if strings.Contains(line, "spec.replicas") || strings.Contains(line, "metadata.generation") {
		return false
	}

	return true
}
