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

const (
	toolDiff = "diff"
	toolDyff = "dyff"
)

func (drift *Drift) IsManagedByHPA(name, kind, nameSpace string) (bool, error) {
	hpaTargets, err := drift.hpaTargets(nameSpace)
	if err != nil {
		return false, err
	}

	_, isManagedByHPA := hpaTargets[hpaTargetKey(name, kind)]
	if isManagedByHPA {
		drift.log.Debugf("the '%s' '%s' is managed by hpa hence the drifts for this would be suppressed if enabled", kind, name)
	}

	return isManagedByHPA, nil
}

func (drift *Drift) hpaTargets(nameSpace string) (map[string]struct{}, error) {
	drift.hpaCacheMu.RLock()

	if hpaTargets, ok := drift.hpaCache[nameSpace]; ok {
		drift.hpaCacheMu.RUnlock()

		return hpaTargets, nil
	}

	drift.hpaCacheMu.RUnlock()

	clientSet, err := drift.getKubeClient()
	if err != nil {
		return nil, err
	}

	response, err := clientSet.AutoscalingV2().HorizontalPodAutoscalers(nameSpace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	hpaTargets := make(map[string]struct{}, len(response.Items))
	for _, item := range response.Items {
		hpaTargets[hpaTargetKey(item.Spec.ScaleTargetRef.Name, item.Spec.ScaleTargetRef.Kind)] = struct{}{}
	}

	drift.hpaCacheMu.Lock()
	if drift.hpaCache == nil {
		drift.hpaCache = make(map[string]map[string]struct{})
	}

	drift.hpaCache[nameSpace] = hpaTargets
	drift.hpaCacheMu.Unlock()

	return hpaTargets, nil
}

func (drift *Drift) getKubeClient() (kubernetes.Interface, error) {
	drift.kubeClientOnce.Do(func() {
		config, err := buildConfigWithContextFromFlags(drift.kubeContext, drift.kubeConfig)
		if err != nil {
			drift.kubeClientErr = &errors.DriftError{Message: fmt.Sprintf("building config with context errored with '%v'", err)}

			return
		}

		drift.kubeClient, drift.kubeClientErr = kubernetes.NewForConfig(config)
		if drift.kubeClientErr != nil {
			drift.kubeClientErr = &errors.DriftError{Message: fmt.Sprintf("creating kubernetes clientsets errored with '%v'", drift.kubeClientErr)}
		}
	})

	return drift.kubeClient, drift.kubeClientErr
}

func buildConfigWithContextFromFlags(context string, kubeConfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func hpaTargetKey(name, kind string) string {
	return kind + "/" + name
}

func (drift *Drift) HasOnlyChangesScaledByHpa(diffOutput string) (bool, error) {
	customDiff := drift.CustomDiff
	if customDiff == "" {
		customDiff = os.Getenv("KUBECTL_EXTERNAL_DIFF")
	}

	diffToolUsed := toolDiff
	if customDiff != "" {
		diffToolUsed = strings.Split(customDiff, " ")[0]
	}

	drift.log.Infof("custom diff: %s", diffToolUsed)

	if diffToolUsed != toolDiff && diffToolUsed != toolDyff {
		drift.log.Warnf("--ignore-hpa-changes currently only supports %s and %s, not '%s'", toolDiff, toolDyff, diffToolUsed)

		return false, nil
	}

	diffOutput = stripansi.Strip(diffOutput)

	scanner := bufio.NewScanner(strings.NewReader(diffOutput))

	for scanner.Scan() {
		line := scanner.Text()

		isNonHpaChange := (diffToolUsed == toolDiff && diffLineHasChangesNonHpaRelated(line)) ||
			(diffToolUsed == toolDyff && dyffLineHasChangesNonHpaRelated(line))

		if !isNonHpaChange {
			continue
		}

		return false, nil
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return true, nil
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

	// check if the line changed is related to replicas or generation, then continue,
	// since we are looking for other fields changed besides replicas and generation
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

	// check if the line changed is related to replicas or generation, then continue,
	// since we are looking for other fields changed besides replicas and generation
	if strings.Contains(line, "spec.replicas") || strings.Contains(line, "metadata.generation") {
		return false
	}

	return true
}
