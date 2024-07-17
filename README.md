# Helm Drift


[![Go Report Card](https://goreportcard.com/badge/github.com/nikhilsbhat/helm-drift)](https://goreportcard.com/report/github.com/nikhilsbhat/helm-drift) 
[![shields](https://img.shields.io/badge/license-MIT-blue)](https://github.com/nikhilsbhat/helm-drift/blob/master/LICENSE) 
[![shields](https://godoc.org/github.com/nikhilsbhat/helm-drift?status.svg)](https://godoc.org/github.com/nikhilsbhat/helm-drift)
[![shields](https://img.shields.io/github/v/tag/nikhilsbhat/helm-drift.svg)](https://github.com/nikhilsbhat/helm-drift/tags)
[![shields](https://img.shields.io/github/downloads/nikhilsbhat/helm-drift/total.svg)](https://github.com/nikhilsbhat/helm-drift/releases)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/helm-drift)](https://artifacthub.io/packages/search?repo=helm-drift)

The Helm plugin that comes in handy while identifying configuration drifts (mostly due to in-place edits) from the deployed Helm charts.

## Introduction

Deploying resources on Kubernetes through the Helm package manager offers simplicity, but maintaining them can be challenging.

While Helm ideally manages all resources, manual interventions are sometimes necessary, leading to configuration discrepancies from the deployed Helm charts. </br>These changes risk being overwritten by subsequent Helm releases, potentially resulting in lost configurations.

The Helm Drift plugin aims to address this issue by validating resources associated with a specific chart or release against Kubernetes. Leveraging kubectl [diff](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#diff) functionality, it identifies any discrepancies or drifts in configurations.

For further insights into the motivation behind creating this plugin, check out the accompanying blog post [blog](https://medium.com/@nikhilsbhat93/helm-plugin-to-identify-the-configuration-that-has-drifted-away-from-the-deployed-helm-release-72f05d26d8cf).

### Example
```shell
# rendering summary to table would render drifts in below format.
helm drift run prometheus-standalone example/chart/sample/ -f ~/path/to/example/chart/sample/override-config.yaml --skip-cleaning
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

# Invoking command without any output format would render detailed drifts as below.
helm drift run prometheus-standalone example/chart/sample/ -f ~/path/to/example/chart/sample/override-config.yaml --skip-cleaning
# executing above command would yield results something like below:
--------------------------------------------------------------------------------------------------
Release                                : sample
------------------------------------------------------------------------------------
Identified drifts in: 'StatefulSet' 'web'

-----------
diff -u -N /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/LIVE-1098999488/apps.v1.StatefulSet.sample.web /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/MERGED-836230905/apps.v1.StatefulSet.sample.web
--- /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/LIVE-1098999488/apps.v1.StatefulSet.sample.web	2024-07-07 18:39:33
+++ /var/folders/dm/40_kbx_56psgqt29q0wh2cxh0000gq/T/MERGED-836230905/apps.v1.StatefulSet.sample.web	2024-07-07 18:39:33
@@ -5,7 +5,7 @@
     meta.helm.sh/release-name: sample
     meta.helm.sh/release-namespace: sample
   creationTimestamp: "2024-07-07T11:32:21Z"
-  generation: 3
+  generation: 4
   labels:
     app.kubernetes.io/managed-by: Helm
   name: web
@@ -15,7 +15,7 @@
 spec:
   minReadySeconds: 10
   podManagementPolicy: OrderedReady
-  replicas: 2
+  replicas: 3
   revisionHistoryLimit: 10
   selector:
     matchLabels:
@@ -28,7 +28,7 @@
         app: nginx
     spec:
       containers:
-      - image: k8s.gcr.io/nginx-slim:0.9
+      - image: k8s.gcr.io/nginx-slim:0.8
         imagePullPolicy: IfNotPresent
         name: nginx
         ports:
-----------

------------------------------------------------------------------------------------
OOPS...! DRIFTS FOUND
------------------------------------------------------------------------------------
Total time spent on identifying drifts : 0.383524375
Total number of drifts found           : YES
Status                                 : FAILED
------------------------------------------------------------------------------------
```

## Suggestion

Try using the drift plugin with a custom diff tool instead for better results. **Ex**: diff tool, [dyff](https://github.com/homeport/dyff), This can be used by setting the flag `--custom-diff`

```shell
helm drift run prometheus-standalone -n monitoring --from-release --custom-diff "dyff between --omit-header --set-exit-code"
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
Make sure the appropriate command is used for the actions. To check the available commands and flags, use `helm drift --help`

```bash
Identifies configuration drifts (mostly due to in-place edits) in the Kubernetes workloads provisioned via Helm charts.

Usage:
  drift [command] [flags]

Available Commands:
  all         Identifies drifts from all releases from the cluster.
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  run         Identifies drifts from a selected chart or release.
  version     Command to fetch the version of helm-drift installed

Flags:
      --concurrency int          the value to be set for flag --concurrency of 'kubectl diff' (default 1)
  -h, --help                     help for drift
  -l, --log-level string         log level for the plugin helm drift (defaults to info) (default "info")
      --no-color                 enabling this would render output with no color
      --revision int             revision of your release from which the drifts to be detected
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --skip-crds                setting this would set '--skip-crds' for helm template command while generating templates
      --skip-tests               setting this would set '--skip-tests' for helm template command while generating templates
      --validate                 setting this would set '--validate' for helm template command while generating templates
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])
      --version string           specify a version constraint for the chart version to use, the value passed here would be used to set --version for helm template command while generating templates


Use "drift [command] --help" for more information about a command.
```

## Commands
### `run`

```shell
It lists all configuration drifts that are part of the specified chart or release, if one exists.

Usage:
  drift run [RELEASE] [CHART] [flags]

Examples:
helm drift run prometheus-standalone path/to/chart/prometheus-standalone -f ~/path/to/override-config.yaml
helm drift run prometheus-standalone --from-release

Flags:
      --consider-hooks                      when this is enabled, the flag 'ignore-hooks' holds no value
      --custom-diff KUBECTL_EXTERNAL_DIFF   custom diff command to use instead of default, the command passed here would be set under KUBECTL_EXTERNAL_DIFF.More information can be found here https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#diff
  -d, --disable-error-on-drift              enabling this would disable exiting with error if drifts were identified
      --from-release                        enable the flag to identify drifts from a release instead (disabled by default, works with command 'run' not with 'all')
  -h, --help                                help for run
      --ignore-hooks strings                list of hooks to ignore while identifying the drifts (default [hook-succeeded,hook-failed])
      --ignore-hpa-changes                  when enabled, the drifts caused on workload due to hpa scaling would be ignored
      --kind strings                        kubernetes resource names to limit the drift identification (--kind takes higher precedence over --name)
      --name string                         name of the kubernetes resource to limit the drift identification
  -o, --output string                       the format to which the output should be rendered to, it should be one of yaml|json|table, if nothing specified it sets to default
      --regex string                        regex used to split helm template rendered (default "---\\n# Source:\\s.*.")
      --skip strings                        kubernetes resource names to skip the drift identification (ex: --skip Deployments)
      --skip-cleaning                       enable the flag to skip cleaning the manifests rendered on to disk
      --skip-validation                     enable the flag if prerequisite validation needs to be skipped
      --temp-path string                    path on disk where the helm templates would be rendered on to (the same would be used be used by 'kubectl diff') (default "/Users/nikhil.bhat/.helm-drift/templates")

Global Flags:
      --concurrency int          the value to be set for flag --concurrency of 'kubectl diff' (default 1)
  -l, --log-level string         log level for the plugin helm drift (defaults to info) (default "info")
      --no-color                 enabling this would render output with no color
      --revision int             revision of your release from which the drifts to be detected
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --skip-crds                setting this would set '--skip-crds' for helm template command while generating templates
      --skip-tests               setting this would set '--skip-tests' for helm template command while generating templates
      --validate                 setting this would set '--validate' for helm template command while generating templates
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])
      --version string           specify a version constraint for the chart version to use, the value passed here would be used to set --version for helm template command while generating templates
```

### `all`

```shell
It lists all configuration drifts that are part of various releases present in the cluster.
Do note that this is expensive operation since multiple kubectl command would be executed in parallel.

Usage:
  drift all [flags]

Examples:
helm drift all --kube-context k3d-sample
helm drift all --kube-context k3d-sample -n sample

Flags:
      --consider-hooks                      when this is enabled, the flag 'ignore-hooks' holds no value
      --custom-diff KUBECTL_EXTERNAL_DIFF   custom diff command to use instead of default, the command passed here would be set under KUBECTL_EXTERNAL_DIFF.More information can be found here https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#diff
  -d, --disable-error-on-drift              enabling this would disable exiting with error if drifts were identified
  -h, --help                                help for all
      --ignore-hooks strings                list of hooks to ignore while identifying the drifts (default [hook-succeeded,hook-failed])
      --ignore-hpa-changes                  when enabled, the drifts caused on workload due to hpa scaling would be ignored
      --is-default-namespace                set this flag if drifts have to be checked specifically in 'default' namespace
      --kind strings                        kubernetes resource names to limit the drift identification (--kind takes higher precedence over --name)
      --name string                         name of the kubernetes resource to limit the drift identification
  -o, --output string                       the format to which the output should be rendered to, it should be one of yaml|json|table, if nothing specified it sets to default
      --regex string                        regex used to split helm template rendered (default "---\\n# Source:\\s.*.")
      --skip strings                        kubernetes resource names to skip the drift identification (ex: --skip Deployments)
      --skip-cleaning                       enable the flag to skip cleaning the manifests rendered on to disk
      --skip-release stringArray            list of helm releases to be skipped for identifying helm drifts, ex: ReleaseName=Namespace | ReleaseName=Namespace
      --skip-validation                     enable the flag if prerequisite validation needs to be skipped
      --temp-path string                    path on disk where the helm templates would be rendered on to (the same would be used be used by 'kubectl diff') (default "/Users/nikhil.bhat/.helm-drift/templates")

Global Flags:
      --concurrency int          the value to be set for flag --concurrency of 'kubectl diff' (default 1)
  -l, --log-level string         log level for the plugin helm drift (defaults to info) (default "info")
      --no-color                 enabling this would render output with no color
      --revision int             revision of your release from which the drifts to be detected
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --skip-crds                setting this would set '--skip-crds' for helm template command while generating templates
      --skip-tests               setting this would set '--skip-tests' for helm template command while generating templates
      --validate                 setting this would set '--validate' for helm template command while generating templates
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])
      --version string           specify a version constraint for the chart version to use, the value passed here would be used to set --version for helm template command while generating templates
```

## Documentation

Updated documentation on all available commands and flags can be found [here](https://github.com/nikhilsbhat/helm-drift/blob/master/docs/doc/drift.md).

## Caveats

Identifying drifts on `CRDs` would be tricky, and the plugin might not respond with the correct data.

If helm hooks are defined in the chart with `hook-succeeded` or `hook-failed`, one might always find drifts when identifying drifts from charts.</br>
Things would work perfectly when identifying drifts from the installed release.

Support for adding a `flag` to skip helm `hooks` if required, is under development.
