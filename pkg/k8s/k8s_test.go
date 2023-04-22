package k8s_test

import (
	"testing"

	"github.com/nikhilsbhat/helm-drift/pkg/k8s"
	"github.com/stretchr/testify/suite"
)

type K8sTestSuite struct {
	suite.Suite
	resource string
}

func (suite *K8sTestSuite) SetupTest() {
	suite.resource = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample
  labels:
    helm.sh/chart: sample-0.1.0
    app.kubernetes.io/name: sample
    app.kubernetes.io/instance: sample
    app.kubernetes.io/version: "1.16.0"
    app.kubernetes.io/managed-by: Helm
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: sample
      app.kubernetes.io/instance: sample
  template:
    metadata:
      labels:
        app.kubernetes.io/name: sample
        app.kubernetes.io/instance: sample
    spec:
      serviceAccountName: sample
      securityContext:
        { }
      containers:
        - name: sample
          securityContext:
            { }
          image: "nginx:1.16.0"
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 80
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /
              port: http
          readinessProbe:
            httpGet:
              path: /
              port: http
          resources:
            { }
`
}

func TestK8sTestSuite(t *testing.T) {
	suite.Run(t, new(K8sTestSuite))
}

func (suite *K8sTestSuite) TestResource_GetName() {
	name, err := k8s.NewResource().GetName(suite.resource)
	suite.NoError(err)
	suite.Equal("sample", name)
}

func (suite *K8sTestSuite) TestResource_GetKind() {
	kind, err := k8s.NewResource().GetKind(suite.resource)
	suite.NoError(err)
	suite.Equal("Deployment", kind)
}
