package pkg_test

import (
	"github.com/nikhilsbhat/helm-drift/pkg"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHelmTemplate_DropStandardLabels(t *testing.T) {
	manifest := `apiVersion: v1
kind: Service
metadata:
  annotations:
    meta.helm.sh/release-name: elastic
    meta.helm.sh/release-namespace: elk
  labels:
    app: elasticsearch-master
    app.kubernetes.io/managed-by: Helm
    chart: elasticsearch
    heritage: Helm
    release: elastic
  name: elasticsearch-master
  namespace: elk
spec:
  clusterIP: 10.43.44.194
  clusterIPs:
  - 10.43.44.194
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - name: http
    port: 9200
  - name: transport
    port: 9300
  selector:
    app: elasticsearch-master
    chart: elasticsearch
    release: elastic`

	expected := `apiVersion: v1
kind: Service
metadata:
    annotations: {}
    labels:
        app: elasticsearch-master
        chart: elasticsearch
        heritage: Helm
        release: elastic
    name: elasticsearch-master
    namespace: elk
spec:
    clusterIP: 10.43.44.194
    clusterIPs:
        - 10.43.44.194
    ipFamilies:
        - IPv4
    ipFamilyPolicy: SingleStack
    ports:
        - name: http
          port: 9200
        - name: transport
          port: 9300
    selector:
        app: elasticsearch-master
        chart: elasticsearch
        release: elastic
`

	t.Run("", func(t *testing.T) {
		template, err := pkg.NewHelmTemplate(manifest).DropStandardLabels(logrus.New())
		assert.NoError(t, err)
		assert.Equal(t, expected, template)
	})
}
