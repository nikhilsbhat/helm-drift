package k8s_test

import (
	"testing"

	"github.com/nikhilsbhat/helm-drift/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type K8sTestSuite struct {
	suite.Suite
	resource     string
	resourceHook string
}

func (suite *K8sTestSuite) SetupTest() {
	suite.resource = `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "1"
    meta.helm.sh/release-name: sample
    meta.helm.sh/release-namespace: sample
  labels:
    app.kubernetes.io/instance: sample
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/managed-by: helm
    app.kubernetes.io/name: sample
    app.kubernetes.io/version: 1.16.0
    helm.sh/chart: sample-0.1.0
  name: sample
  namespace: sample
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app.kubernetes.io/instance: sample
      app.kubernetes.io/name: sample
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app.kubernetes.io/instance: sample
        app.kubernetes.io/name: sample
    spec:
      containers:
      - image: nginx:1.16.0
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /
            port: http
            scheme: HTTP
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        name: sample
        ports:
        - containerPort: 80
          name: http
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /
            port: http
            scheme: HTTP
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      serviceAccount: sample
      serviceAccountName: sample
      terminationGracePeriodSeconds: 30`

	suite.resourceHook = `apiVersion: batch/v1
kind: Job
metadata:
  name: "sample-hook-succeeded"
  labels:
    app.kubernetes.io/managed-by: "Helm"
    app.kubernetes.io/instance: "sample"
    app.kubernetes.io/version: 1.16.0
    helm.sh/chart: "sample-0.1.0"
  annotations:
    # This is what defines this resource as a hook. Without this line, the
    # job is considered part of the release.
    "helm.sh/hook": post-install
    "helm.sh/hook-weight": "-5"
    "helm.sh/hook-delete-policy": hook-failed,hook-succeeded
spec:
  template:
    metadata:
      name: "sample"
      labels:
        app.kubernetes.io/managed-by: "Helm"
        app.kubernetes.io/instance: "sample"
        helm.sh/chart: "sample-0.1.0"
    spec:
      restartPolicy: Never
      containers:
        - name: post-install-job
          image: "alpine:3.3"
          command: ["/bin/sleep","10"]`
}

func TestK8sTestSuite(t *testing.T) {
	suite.Run(t, new(K8sTestSuite))
}

func (suite *K8sTestSuite) TestResource_GetNameSpace() {
	name, err := k8s.NewResource().GetMetadata(suite.resource, "sample", logrus.New())
	suite.NoError(err)
	suite.Equal("sample", name)
}

func (suite *K8sTestSuite) TestResource_GetName() {
	name, err := k8s.NewResource().GetMetadata(suite.resource, "name", nil)
	suite.NoError(err)
	suite.Equal("sample", name)
}

func (suite *K8sTestSuite) TestResource_GetKind() {
	kind, err := k8s.NewResource().Get(suite.resource, "", nil)
	suite.NoError(err)
	suite.Equal("Deployment", kind)
}

func (suite *K8sTestSuite) TestResource_IsHelmHookTrue() {
	kind, err := k8s.NewResource().IsHelmHook(suite.resourceHook, []string{"hook-succeeded", "hook-failed"})
	suite.NoError(err)
	suite.True(kind)
}
