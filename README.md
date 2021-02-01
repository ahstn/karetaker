# Karetaker

[![codecov](https://codecov.io/gh/ahstn/karetaker/branch/main/graph/badge.svg?token=Y4LCYHK8OQ)](https://codecov.io/gh/ahstn/karetaker)

**WORK IN PROGRESS. Only for dev and experimental use.**

## Overview
The main intention of `karetaker` is to oversee and clean-up objects on your Kubernetes cluster. The best use case for this is on development clusters, where deployments, configmaps, etc have a tendency to be left running with no purpose.


At its core, `karetaker` has four main clean-up operations: age-based, completed, duplicated and un-used.

Each of these operations have their own logic and specific Kubernetes resources they act against:

- Duplicate - attempts to detect multiple copies of similar deployments.
- Completed - specially to target completed jobs.
- Age-Based - Target resources older than a specific age. (i.e. deploys older than 7 days)
- Un-used - Attempts to find resources that are no longer used (i.e. configmaps not referenced by another resource)

## Commands

### `karetaker age`
Target resources older than a specific age. (i.e. deploys older than 7 days)

```
➜ karetaker age -h
Find resources older than a certain age

Usage:
    karetaker [type] {flags}

Arguments:
    type                          type of resource (default: deployment)

Flags:
    -a, --age                     age boundary to filter on (default: 48h)
    -A, --allow                   allow list (CSV) of name patterns to ignore (i.e. 'istio')
    -h, --help                    displays usage information of the application or a command (default: false)
    -n, --namespace               kubernetes namespace (default: default)
   
Example:
    karetaker age -n default -a 48h deployment
```
To ignore certain objects, see : [Allow List](#allow-list).

### `karetaker unused`
Attempts to find resources that are no longer used, a primary example of this would be an existing configmap that isn't being referenced by a running deployment or pod.

To ignore certain objects, see : [Allow List](#allow-list).

```
➜ karetaker unused -h
Find resources not in use by another object

Usage:
    karetaker [type] {flags}

Arguments:
    type                          type of resource (default: configmap)

Flags:
    -A, --allow                   allow list (CSV) of name patterns to ignore (i.e. 'istio')
    -d, --dry-run                 if true, only show the resources (default: false)
    -h, --help                    displays usage information of the application or a command (default: false)
    -n, --namespace               kubernetes namespace (default: default)
Example:
    karetaker unused -n default secrets,configmaps
```

### `karetaker duplicate`
This commands aims to find similar or duplicate Kubernetes deployments. It's intended for finding similar Helm releases, but can be used for any deployment that has an "app name" and "instance" labels (i.e. `kubernetes.io/name` and `kubernetes.io/instance`).

```
➜ karetaker duplicate -h

Usage:
   karetaker [target] {flags}

Arguments: 
   target               label to target similarities and duplicates (default: kubernetes.io/instance)

Flags: 
   -f, --filter         deployments label filter (i.e. app=auth) 
   -h, --help           displays usage information of the application or a command (default: false)
   -n, --namespace      kubernetes namespace (default: default)
```

The `kubernetes.io/name` label is used to filter deployments for the target application and `kubernetes.io/instance` is used to find similar label values. Examples of the `instance` label could be the name of your release, the ticket identifier for a new application feature or the username of the engineer working on the feature.

With this in mind, if we have duplicate instances `kubectl get deploy` might look like the following:
```
kubectl get deploy -l 'kubernetes.io/name=app'
NAME                READY   UP-TO-DATE   AVAILABLE   AGE
app-20AprRelease    1/1     1            1           5d
app-20AprReleasee   1/1     1            1           5d
app-adam            1/1     1            1           3h
app-adam2           1/1     1            1           2d
app-adam5           1/1     1            1           49d
app-auth-531        1/1     1            1           17d
app-auth-531-2      1/1     1            1           16d
```

So when our app name label equals 'app' we have 6 deployments which will all have instances labels (`kuberetes.io/instance=adam`, `kubernetes.io/instance=adam2`, etc). From looking at these we can see that the April Release deployment has a "typo" duplicate that wasn't cleaned up and the engineer Adam has a release for 49 days ago that they potentially forgot about.

`karetaker duplicate` is designed to make us aware of these similar deployments and delete them, if we deem them unnecessary.

## Allow List
To ignore certain objects (i.e. `default-token` or `istio-ca`), all commands will support an "allow-list" as `-A or --allow`.

To use this flag, simply pass it a comma separated list, which is added onto the defaults. The "allow-list", will always contain `default-token, istio-ca, sh.helm.release` by default.

Partial matches are used rather than exact, so `-A istio` would result in any configmap or secret containing `istio` to be ignored.

## Backlog Items

In a roughly prioritised order:

- [x] Authenticate with Kuberentes (Out-of-Cluster Usage)
- [x] List Deployments older than X time
- [x] Identify duplicate Helm releases
- [x] List un-referenced configmaps & secrets
- [x] Allow list of resources/objects to ignore 
- [ ] Config file for batch execution  
- [ ] Add Logging for batch execution (i.e. logrus)
- [ ] Duplicate should consider pod image and possibly environment variables 
- [ ] Use default namespace from kubeconfig
- [ ] Authenticate using Service Account (In-Cluster Usage)
- [ ] List Deployments without a desired running replica(s)
- [ ] List Deployments using 90% of resource limits
- [ ] Integration Tests using KinD
- [ ] Add progress bars for ANSI terminals (i.e. spinners & emojis)
- [ ] Interactive Clean-Up CLI
