# Helm Drift


[![Go Report Card](https://goreportcard.com/badge/github.com/nikhilsbhat/helm-drift)](https://goreportcard.com/report/github.com/nikhilsbhat/helm-drift) 
[![shields](https://img.shields.io/badge/license-MIT-blue)](https://github.com/nikhilsbhat/helm-drift/blob/master/LICENSE) 
[![shields](https://godoc.org/github.com/nikhilsbhat/helm-drift?status.svg)](https://godoc.org/github.com/nikhilsbhat/helm-drift)
[![shields](https://img.shields.io/github/v/tag/nikhilsbhat/helm-drift.svg)](https://github.com/nikhilsbhat/helm-drift/tags)
[![shields](https://img.shields.io/github/downloads/nikhilsbhat/helm-drift/total.svg)](https://github.com/nikhilsbhat/helm-drift/releases)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/drift)](https://artifacthub.io/packages/search?repo=drift)


The helm plugin that helps in identifying deviations(mostly due to in-place edits) in the configurations that are deployed via helm chart.

## Introduction

Kubernetes' resources can be deployed via package manager helm, it is easier to deploy but to manage the same require more effort.

If helm is used, strictly all resources should be managed by helm itself, but there are places where manual interventions are needed.</br>
This results in configuration drift from helm charts deployed.</br>
These changes can be overridden by next helm release, what if the required changes are lost before adding it back to helm chart?

This helm plugin is intended to solve the same problem by validating the resources that are part of appropriate chart/release against kubernetes.

This leverages kubectl [diff](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#diff) to identify the drifts.

```shell
helm drift run prometheus-standalone example/chart/sample/ -f ~/path/to/example/chart/sample/override-config.yaml --skip-cleaning
# executing above command would yield results something like below:
diff -u -N /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/LIVE-1282120558/batch.v1.Job.default.pi /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/MERGED-3024449788/batch.v1.Job.default.pi
--- /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/LIVE-1282120558/batch.v1.Job.default.pi	2023-03-23 12:01:36.000000000 +0530
+++ /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/MERGED-3024449788/batch.v1.Job.default.pi	2023-03-23 12:01:36.000000000 +0530
@@ -0,0 +1,77 @@
+apiVersion: batch/v1
+kind: Job
+metadata:
+  annotations:
+    batch.kubernetes.io/job-tracking: ""
+  creationTimestamp: "2023-03-23T06:31:36Z"
+  generation: 1
+  labels:
+    controller-uid: 8e325a1c-5e66-4abf-9f55-af4acf763b5c
+    job-name: pi
+  managedFields:
+  - apiVersion: batch/v1
+    fieldsType: FieldsV1
+    fieldsV1:
+      f:spec:
+        f:backoffLimit: {}
+        f:completionMode: {}
+        f:completions: {}
+        f:parallelism: {}
+        f:suspend: {}
+        f:template:
+          f:spec:
+            f:containers:
+              k:{"name":"pi"}:
+                .: {}
+                f:command: {}
+                f:image: {}
+                f:imagePullPolicy: {}
+                f:name: {}
+                f:resources: {}
+                f:terminationMessagePath: {}
+                f:terminationMessagePolicy: {}
+            f:dnsPolicy: {}
+            f:restartPolicy: {}
+            f:schedulerName: {}
+            f:securityContext: {}
+            f:terminationGracePeriodSeconds: {}
+    manager: kubectl-client-side-apply
+    operation: Update
+    time: "2023-03-23T06:31:36Z"
+  name: pi
+  namespace: default
+  uid: 8e325a1c-5e66-4abf-9f55-af4acf763b5c
+spec:
+  backoffLimit: 4
+  completionMode: NonIndexed
+  completions: 1
+  parallelism: 1
+  selector:
+    matchLabels:
+      controller-uid: 8e325a1c-5e66-4abf-9f55-af4acf763b5c
+  suspend: false
+  template:
+    metadata:
+      creationTimestamp: null
+      labels:
+        controller-uid: 8e325a1c-5e66-4abf-9f55-af4acf763b5c
+        job-name: pi
+    spec:
+      containers:
+      - command:
+        - perl
+        - -Mbignum=bpi
+        - -wle
+        - print bpi(2000)
+        image: perl:5.34.0
+        imagePullPolicy: IfNotPresent
+        name: pi
+        resources: {}
+        terminationMessagePath: /dev/termination-log
+        terminationMessagePolicy: File
+      dnsPolicy: ClusterFirst
+      restartPolicy: Never
+      schedulerName: default-scheduler
+      securityContext: {}
+      terminationGracePeriodSeconds: 30
+status: {}
```
## Installation

```shell
helm plugin install https://github.com/nikhilsbhat/helm-drift
```
Use the executable just like any other go-cli application.

## Usage

```bash
helm drift [command] [flags]
```
Make sure appropriate command is used for the actions, to check the available commands and flags use `helm drift --help`

```bash
Identifies drifts (mostly due to in place edits) in the kubernetes workloads provisioned via helm charts.

Usage:
  drift [command] [flags]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  run         Identifies drifts from a selected chart/release
  version     Command to fetch the version of helm-drift installed

Flags:
  -h, --help                     help for drift
  -l, --log-level string         log level for the plugin helm drift (defaults to info) (default "info")
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --skip-tests               setting this would set '--skip-tests' for helm template command while generating templates
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])


Use "drift [command] --help" for more information about a command.
```

## Commands
### `run`

```shell
Lists all configuration drifts that are part of specified chart/release if exists.

Usage:
  drift run [RELEASE] [CHART] [flags]

Examples:
  helm drift run prometheus-standalone path/to/chart/prometheus-standalone -f ~/path/to/override-config.yaml
  helm drift run prometheus-standalone --from-release

Flags:
      --from-release       enable the flag to identify drifts from a release instead (disabled by default)
  -h, --help               help for run
      --regex string       regex used to split helm template rendered (default "---\\n# Source:\\s.*.")
      --skip-cleaning      enable the flag to skip cleaning the manifests rendered on to disk
      --skip-validation    enable the flag if prerequisite validation needs to be skipped
      --temp-path string   path on disk where the helm templates would be rendered on to (the same would be used be used by 'kubectl diff') (default "/Users/nikhil.bhat/.helm-drift/templates")

Global Flags:
  -l, --log-level string         log level for the plugin helm drift (defaults to info) (default "info")
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --skip-tests               setting this would set '--skip-tests' for helm template command while generating templates
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])
```

## Documentation

Updated documentation on all available commands and flags can be found [here](https://github.com/nikhilsbhat/helm-drift/blob/master/docs/doc/drift.md).