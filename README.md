# Helm Drift


[![Go Report Card](https://goreportcard.com/badge/github.com/nikhilsbhat/helm-drift)](https://goreportcard.com/report/github.com/nikhilsbhat/helm-drift) 
[![shields](https://img.shields.io/badge/license-MIT-blue)](https://github.com/nikhilsbhat/helm-drift/blob/master/LICENSE) 
[![shields](https://godoc.org/github.com/nikhilsbhat/helm-drift?status.svg)](https://godoc.org/github.com/nikhilsbhat/helm-drift)
[![shields](https://img.shields.io/github/v/tag/nikhilsbhat/helm-drift.svg)](https://github.com/nikhilsbhat/helm-drift/tags)
[![shields](https://img.shields.io/github/downloads/nikhilsbhat/helm-drift/total.svg)](https://github.com/nikhilsbhat/helm-drift/releases)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/helm-drift)](https://artifacthub.io/packages/search?repo=helm-drift)

The helm plugin that helps in identifying deviations(mostly due to in-place edits) in the configurations that are deployed via helm chart.

## Introduction

Kubernetes' resources can be deployed via package manager helm, it is easier to deploy but to manage the same require more effort.

If helm is used, strictly all resources should be managed by helm itself, but there are places where manual interventions are needed.</br>
This results in configuration drift from helm charts deployed.
These changes can be overridden by next helm release, what if the required changes are lost before adding it back to helm chart?

This helm plugin is intended to solve the same problem by validating the resources that are part of appropriate chart/release against kubernetes.

This leverages kubectl [diff](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#diff) to identify the drifts.

### Example
```shell
# By enabling --summary would render drifts as quick summary in table format.
helm drift run prometheus-standalone example/chart/sample/ -f ~/path/to/example/chart/sample/override-config.yaml --skip-cleaning --summary
       KIND      |         NAME          | DRIFT
-----------------|-----------------------|---------
  ServiceAccount | sample                | NO
  Service        | sample                | NO
  DaemonSet      | fluentd-elasticsearch | NO
  Pod            | nginx                 | NO
  Pod            | nginx-2               | NO
  ReplicaSet     | frontend              | NO
  Deployment     | sample                | NO
  StatefulSet    | web                   | YES
  Job            | pi                    | NO
  CronJob        | hello                 | NO
-----------------|-----------------------|---------
                          STATUS         | FAILED
                 ------------------------|---------
Namespace: 'sample' Release: 'sample'

# Invoking command with out flag --summary would render detailed drifts.
helm drift run prometheus-standalone example/chart/sample/ -f ~/path/to/example/chart/sample/override-config.yaml --skip-cleaning
# executing above command would yield results something like below:
------------------------------------------------------------------------------------
Identified drifts in: 'StatefulSet' 'web'

-----------
diff -u -N /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/LIVE-2873647491/apps.v1.StatefulSet.sample.web /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/MERGED-4261927724/apps.v1.StatefulSet.sample.web
--- /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/LIVE-2873647491/apps.v1.StatefulSet.sample.web	2023-03-25 23:33:06.000000000 +0530
+++ /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/MERGED-4261927724/apps.v1.StatefulSet.sample.web	2023-03-25 23:33:06.000000000 +0530
@@ -5,7 +5,7 @@
     meta.helm.sh/release-name: sample
     meta.helm.sh/release-namespace: sample
   creationTimestamp: "2023-03-24T06:15:02Z"
-  generation: 2
+  generation: 3
   labels:
     app.kubernetes.io/managed-by: Helm
   managedFields:
@@ -84,7 +84,6 @@
           f:spec:
             f:containers:
               k:{"name":"nginx"}:
-                f:image: {}
                 f:ports:
                   k:{"containerPort":8080,"protocol":"TCP"}:
                     .: {}
@@ -94,6 +93,24 @@
     manager: kubectl-edit
     operation: Update
     time: "2023-03-24T06:19:50Z"
+  - apiVersion: apps/v1
+    fieldsType: FieldsV1
+    fieldsV1:
+      f:spec:
+        f:template:
+          f:spec:
+            f:containers:
+              k:{"name":"nginx"}:
+                f:image: {}
+                f:ports:
+                  k:{"containerPort":80,"protocol":"TCP"}:
+                    .: {}
+                    f:containerPort: {}
+                    f:name: {}
+                    f:protocol: {}
+    manager: kubectl-client-side-apply
+    operation: Update
+    time: "2023-03-25T18:03:05Z"
   name: web
   namespace: sample
   resourceVersion: "14246"
@@ -114,10 +131,13 @@
         app: nginx
     spec:
       containers:
-      - image: k8s.gcr.io/nginx-slim:0.9
+      - image: k8s.gcr.io/nginx-slim:0.8
         imagePullPolicy: IfNotPresent
         name: nginx
         ports:
+        - containerPort: 80
+          name: web
+          protocol: TCP
         - containerPort: 8080
           name: web
           protocol: TCP
-----------
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
      --summary            if enabled, prints a quick summary in table format without printing actual drifts
      --temp-path string   path on disk where the helm templates would be rendered on to (the same would be used be used by 'kubectl diff') (default "/Users/nikhil.bhat/.helm-drift/templates")

Global Flags:
  -l, --log-level string         log level for the plugin helm drift (defaults to info) (default "info")
      --no-color                 enabling this would render summary with no color
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --skip-tests               setting this would set '--skip-tests' for helm template command while generating templates
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])
```

## Documentation

Updated documentation on all available commands and flags can be found [here](https://github.com/nikhilsbhat/helm-drift/blob/master/docs/doc/drift.md).