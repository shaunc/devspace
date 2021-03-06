---
title: "Command - devspace update config"
sidebar_label: config
---


Converts the active config to the current config version

## Synopsis


```
devspace update config [flags]
```

```
#######################################################
############### devspace update config ################
#######################################################
Updates the currently active config to the newest
config version

Note: This does not upgrade the overwrite configs
#######################################################
```
## Options

```
  -h, --help   help for config
```

### Options inherited from parent commands

```
      --debug                 Prints the stack trace if an error occurs
      --kube-context string   The kubernetes context to use
  -n, --namespace string      The kubernetes namespace to use
      --no-warn               If true does not show any warning when deploying into a different namespace or kube-context than before
  -p, --profile string        The devspace profile to use (if there is any)
      --silent                Run in silent mode and prevents any devspace log output except panics & fatals
  -s, --switch-context        Switches and uses the last kube context and namespace that was used to deploy the DevSpace project
      --var strings           Variables to override during execution (e.g. --var=MYVAR=MYVALUE)
```

## See Also

* [devspace update](../../cli/commands/devspace_update)	 - Updates the current config
