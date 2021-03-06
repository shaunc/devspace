---
title: Command - devspace remove deployment
sidebar_label: deployment
id: version-v4.0.1-devspace_remove_deployment
original_id: devspace_remove_deployment
---


Removes one or all deployments from devspace configuration

## Synopsis


```
devspace remove deployment [deployment-name] [flags]
```

```
#######################################################
############ devspace remove deployment ###############
#######################################################
Removes one or all deployments from the devspace
configuration (If you want to delete the deployed 
resources, run 'devspace purge -d deployment_name'):

devspace remove deployment devspace-default
devspace remove deployment --all
#######################################################
```
## Options

```
      --all    Remove all deployments
  -h, --help   help for deployment
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

* [devspace remove](../../cli/commands/devspace_remove)	 - Changes devspace configuration
