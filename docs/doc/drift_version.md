## drift version

Command to fetch the version of helm-drift installed

### Synopsis

This will help the user find what version of the helm-drift plugin he or she installed in her machine.

```
drift version [flags]
```

### Options

```
  -h, --help   help for version
```

### Options inherited from parent commands

```
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

### SEE ALSO

* [drift](drift.md)	 - A utility that helps in identifying drifts in infrastructure

###### Auto generated by spf13/cobra on 13-Oct-2024
