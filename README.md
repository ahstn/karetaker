# Karetaker

**WORK IN PROGRESS. Only for dev and experimental use.**

#### Todo 

- [ ] Authenticate with Kuberentes (Out-of-Cluster Usage)
- [ ] List Deployments older than X time
- [ ] Add Logging for batch execution (i.e. logrus) 
- [ ] Add progress bars for ANSI terminals (i.e. spinners & emojis)
- [ ] List Deployments without a desired running replica(s)
- [ ] List Deployments using 90% of resource limits
- [ ] Identify duplicate Helm releases
- [ ] Interactive Clean-Up CLI
- [ ] Read Config for Batch Clean-up
- [ ] Authenticate using Service Account (In-Cluster Usage)
- [ ] Integration Tests using KinD


## Commands

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
