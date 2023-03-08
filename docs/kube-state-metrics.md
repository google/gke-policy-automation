# Kube state metrics for GKE Policy Automation

The [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) agent is needed
**only for cluster scalability limits check**.

---

* [Overview](#overview)
* [Installation](#installation)
  * [Existing installations](#existing-installations)
  * [New deployment](#new-deployment)
* [Configuring metrics collection](#metrics-collection)
* [Used metrics](#used-metrics)

---

## Overview

The kube-state-metrics is a simple service that listens to Kubernetes API server and generates metrics
about the state of the objects. The GKE Policy Automation ingests the metrics provided by kube-state-metrics
via standalone Prometheus server or
[Google Cloud Managed Service for Prometheus](https://cloud.google.com/stackdriver/docs/managed-prometheus).

## Installation

The kube-state-metrics is running inside a Kubernetes pod and uses service account token for
read-only access to the Kubernetes cluster.

There are many ways of installing the kube-state-metrics agent, like those described in the
[kube-state-metrics official usage documentation](https://github.com/kubernetes/kube-state-metrics#usage).
The agent is also part of many monitoring stacks, like [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus/).

### Existing installations

If you are already using kube-state-metrics as a part of the other monitoring stack, like [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus/),
you can continue using it. You need to customize its configuration though.

For GKE Policy Automation scalability check, modify the kube-state-metrics container arguments to allow
additional labels for metrics.
This can be done by adding the following command line arguments to kube-state-metric:

 ```yaml
containers:
- args:
  - --metric-labels-allowlist=nodes=[cloud.google.com/gke-nodepool,topology.kubernetes.io/zone]
```

The GKE Policy Automation requires following, additional labels for `kube_node_labels` metric:

* `cloud.google.com/gke-nodepool` label for node's node pool information
* `topology.kubernetes.io/zone` label for node's compute zone information

### New deployment

You can use the following configuration to install Kube State Metrics
(includes any required configuration customizations):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: exporter
    app.kubernetes.io/name: kube-state-metrics
    app.kubernetes.io/version: 2.8.1
  name: kube-state-metrics
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: kube-state-metrics
  template:
    metadata:
      labels:
        app.kubernetes.io/component: exporter
        app.kubernetes.io/name: kube-state-metrics
        app.kubernetes.io/version: 2.8.1
    spec:
      automountServiceAccountToken: true
      containers:
      - image: registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.8.1
        args:
        - --metric-labels-allowlist=nodes=[cloud.google.com/gke-nodepool,topology.kubernetes.io/zone]
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          timeoutSeconds: 5
        name: kube-state-metrics
        ports:
        - containerPort: 8080
          name: http-metrics
        - containerPort: 8081
          name: telemetry
        readinessProbe:
          httpGet:
            path: /
            port: 8081
          initialDelaySeconds: 5
          timeoutSeconds: 5
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
          runAsUser: 65534
      nodeSelector:
        kubernetes.io/os: linux
      serviceAccountName: kube-state-metrics
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: exporter
    app.kubernetes.io/name: kube-state-metrics
    app.kubernetes.io/version: 2.8.1
  name: kube-state-metrics
  namespace: kube-system
spec:
  clusterIP: None
  ports:
  - name: http-metrics
    port: 8080
    targetPort: http-metrics
  - name: telemetry
    port: 8081
    targetPort: telemetry
  selector:
    app.kubernetes.io/name: kube-state-metrics
---
apiVersion: v1
automountServiceAccountToken: false
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: exporter
    app.kubernetes.io/name: kube-state-metrics
    app.kubernetes.io/version: 2.8.1
  name: kube-state-metrics
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app.kubernetes.io/component: exporter
    app.kubernetes.io/name: kube-state-metrics
    app.kubernetes.io/version: 2.8.1
  name: kube-state-metrics
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-state-metrics
subjects:
- kind: ServiceAccount
  name: kube-state-metrics
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app.kubernetes.io/component: exporter
    app.kubernetes.io/name: kube-state-metrics
    app.kubernetes.io/version: 2.8.1
  name: kube-state-metrics
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - secrets
  - nodes
  - pods
  - services
  - serviceaccounts
  - resourcequotas
  - replicationcontrollers
  - limitranges
  - persistentvolumeclaims
  - persistentvolumes
  - namespaces
  - endpoints
  verbs:
  - list
  - watch
- apiGroups:
  - apps
  resources:
  - statefulsets
  - daemonsets
  - deployments
  - replicasets
  verbs:
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - cronjobs
  - jobs
  verbs:
  - list
  - watch
- apiGroups:
  - autoscaling
  resources:
  - horizontalpodautoscalers
  verbs:
  - list
  - watch
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
- apiGroups:
  - policy
  resources:
  - poddisruptionbudgets
  verbs:
  - list
  - watch
- apiGroups:
  - certificates.k8s.io
  resources:
  - certificatesigningrequests
  verbs:
  - list
  - watch
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - list
  - watch
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  - volumeattachments
  verbs:
  - list
  - watch
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  - validatingwebhookconfigurations
  verbs:
  - list
  - watch
- apiGroups:
  - networking.k8s.io
  resources:
  - networkpolicies
  - ingressclasses
  - ingresses
  verbs:
  - list
  - watch
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - list
  - watch
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterrolebindings
  - clusterroles
  - rolebindings
  - roles
  verbs:
  - list
  - watch
```

## Metrics collection

### Self managed Prometheus

If self managed Prometheus collection is used, be sure to configure Prometheus scrapping
for kube-state-metrics, i.e. with a `PodMonitor` or `ServiceMonitor` objects and by annotating
kube-state-metric accordingly, i.e. with a `prometheus.io/scrape` annotation.

### Google Cloud Managed Service for Prometheus

If Google Cloud Managed Service for Prometheus is used, create the `PodMonitoring` object for kube-state-metrics:

```yaml
apiVersion: monitoring.googleapis.com/v1
kind: PodMonitoring
metadata:
  name: kube-state-metrics
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: kube-state-metrics
  endpoints:
  - port: http-metrics
    path: /metrics
    interval: 30s
```

As an alternative to the `PodMonitoring` you can use [ClusterPodMonitoring](https://github.com/GoogleCloudPlatform/prometheus-engine/blob/v0.5.0/doc/api.md#clusterpodmonitoring)
and label `kube-state-metrics` deployment accordingly.

## Used metrics

GKE Policy Automation uses the following metrics from `kube-state-metrics` agent:

* `kube_pod_info`
* `kube_pod_container_info`
* `kube_node_info`
* `kube_node_labels`
* `kube_service_info`
* `kube_horizontalpodautoscaler_info`
* `kube_secret_info`
