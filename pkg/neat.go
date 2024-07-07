package pkg

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (drift *Drift) neat(deviation deviation.Deviation) ([]byte, error) {
	config, err := clientcmd.BuildConfigFromFlags("", drift.kubeConfig)
	if err != nil {
		drift.log.Error(err)

		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %s", err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		drift.log.Error(err)

		return nil, err
	}

	gvr, err := findGVR(clientSet, "", deviation.APIVersion, deviation.Kind)
	if err != nil {
		log.Fatalf("Error finding GVR: %s", err.Error())
	}

	resource, err := dynamicClient.Resource(gvr).Namespace(drift.namespace).Get(context.TODO(), deviation.Resource, metav1.GetOptions{})
	if err != nil {
		drift.log.Errorf("fetching kubernetes manifests for '%s' '%s' in namespace '%s' errored with:  %v", deviation.Kind, deviation.Resource, drift.namespace, err)

		return nil, err
	}

	cleanedResource := cleanResource(drift.dropStandardHelmLabels(resource))

	yamlData, err := yaml.MarshalWithOptions(cleanedResource.Object, yaml.IndentSequence(true), yaml.UseLiteralStyleIfMultiline(true))
	if err != nil {
		return nil, err
	}

	return yamlData, nil
}

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

func findGVR(clientSet *kubernetes.Clientset, group, version, resourceType string) (schema.GroupVersionResource, error) {
	resources, err := clientSet.Discovery().ServerPreferredResources()
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	for _, resourceList := range resources {
		groupVersion, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}

		if filepath.Join(groupVersion.Group, groupVersion.Version) == version {
			for _, resource := range resourceList.APIResources {
				if strings.Contains(resource.Name, strings.ToLower(resourceType)) {
					return schema.GroupVersionResource{
						Group:    groupVersion.Group,
						Version:  groupVersion.Version,
						Resource: resource.Name,
					}, nil
				}
			}
		}
	}

	return schema.GroupVersionResource{}, &errors.DriftError{Message: fmt.Sprintf("resource type '%s' of version '%s' not found", resourceType, version)}
}
