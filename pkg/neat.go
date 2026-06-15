package pkg

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func cleanResource(resource *unstructured.Unstructured) *unstructured.Unstructured {
	unstructured.RemoveNestedField(resource.Object, "status")
	unstructured.RemoveNestedField(resource.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(resource.Object, "metadata", "selfLink")
	unstructured.RemoveNestedField(resource.Object, "metadata", "uid")
	unstructured.RemoveNestedField(resource.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(resource.Object, "metadata", "generation")
	unstructured.RemoveNestedField(resource.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(resource.Object, "metadata", "deletionTimestamp")
	unstructured.RemoveNestedField(resource.Object, "metadata", "deletionGracePeriodSeconds")
	unstructured.RemoveNestedField(resource.Object, "metadata", "deletionGracePeriodSeconds")

	return resource
}

func (drift *Drift) dropStandardHelmLabels(resource *unstructured.Unstructured) *unstructured.Unstructured {
	helmLabels := []string{
		"app.kubernetes.io/name",
		"helm.sh/chart",
		"app.kubernetes.io/managed-by",
		"app.kubernetes.io/instance",
		"app.kubernetes.io/version",
		"app.kubernetes.io/component",
		"app.kubernetes.io/part-of",
	}
	helmAnnotations := []string{"meta.helm.sh/release-name", "meta.helm.sh/release-namespace"}

	for _, label := range helmLabels {
		labels, found, err := unstructured.NestedStringMap(resource.Object, "metadata", "labels")
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if found {
			delete(labels, label)

			if err := unstructured.SetNestedStringMap(resource.Object, labels, "metadata", "labels"); err != nil {
				log.Fatalf("error: %v", err)
			}
		}
	}

	for _, annotation := range helmAnnotations {
		annotations, found, err := unstructured.NestedStringMap(resource.Object, "metadata", "annotations")
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if found {
			delete(annotations, annotation)

			if err := unstructured.SetNestedStringMap(resource.Object, annotations, "metadata", "annotations"); err != nil {
				log.Fatalf("error: %v", err)
			}
		}

		unstructured.RemoveNestedField(resource.Object, "metadata.annotations", annotation)
	}

	return resource
}
