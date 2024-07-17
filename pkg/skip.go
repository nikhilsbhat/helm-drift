package pkg

import (
	"github.com/thoas/go-funk"
	"helm.sh/helm/v3/pkg/release"
)

type resourcesToSkip []resourcesInfo

type Skip interface {
	skipResources(resource, namespace string) bool
}

func (skips resourcesToSkip) filterRelease(releases []*release.Release) []*release.Release {
	return funk.Filter(releases, func(release *release.Release) bool {
		for _, skip := range skips {
			if skip.name == release.Name && skip.namespace == release.Namespace {
				return false
			}
		}

		return true
	}).([]*release.Release)
}
