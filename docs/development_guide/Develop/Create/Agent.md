# Introduction

This document provides a guide on how to create a new agent microservice

# What is an agent?


For this guide, we can assume that we are going to add a new agent named **testAgent**

# Infrastructure 
Creating a new agent mainly requires adding configurations that will allow helm to deploy these agents. This section describes the steps needed to setup the configuration for an agent.

## Namespace
Each agent requires their own namespace in the cluster, add the following to the namespace yaml file in `charts/namespaces/templates/namespaces.yml`:

```yaml

---
apiVersion: v1
kind: Namespace
metadata:
  name: testAgent
  annotations:
    "helm.sh/resource-policy": keep
    "app.kubernetes.io/managed-by": "Helm"
    "config.linkerd.io/trace-collector": collector.linkerd-jaeger:55678 # or 14268?

---

apiVersion: v1
kind: Secret
metadata:
  name: rabbit
  namespace: testAgent
type: Opaque
data:
  password: {{ .Values.secret.password | b64enc | quote }}

```

## Agent Template Files (charts/agents/templates)
First, create a file with the name of your agent, so in this case:
`charts/agents/templates/testAgent.yaml`, with the following content:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testAgent
  namespace: testAgent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: testAgent
  template:
    metadata:
      labels:
        app: testAgent
      annotations:
        "linkerd.io/inject" : enabled
    spec:
      serviceAccountName: job-creator-testAgent
      containers:
        - name: testAgent
          image: {{ .Values.dockerArtifactAccount }}/agent:{{ .Values.branchNameTag }}
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
          env:
            - name: DATA_STEWARD_NAME
              value: TESTAGENT 
            - name: OC_AGENT_HOST
              value: {{ .Values.tracingEndpoint }}
          # resources:
          #   requests:
          #     memory: "128Mi"
          #   limits:
          #     memory: "256Mi"
        - name: sidecar
          image: {{ .Values.dockerArtifactAccount }}/sidecar:{{ .Values.branchNameTag }}
          imagePullPolicy: Always
          env:
            - name: AMQ_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: rabbit
                  key: password
            - name: AMQ_USER
              value: normal_user
            - name: OC_AGENT_HOST
              value: {{ .Values.tracingEndpoint }}

---

apiVersion: v1
kind: Service
metadata:
  name: testAgent
  namespace: testAgent
spec:
  selector:
    app: testAgent
  ports:
    - name: http-testAgent-api
      protocol: TCP
      port: 8080
      targetPort: 8080
  type: ClusterIP


---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: testAgent-ingress
  namespace: testAgent
  annotations:
    nginx.ingress.kubernetes.io/service-upstream: "true"
spec:
  ingressClassName: nginx
  rules:
    - host: testAgent.testAgent.svc.cluster.local
      http:
        paths:
          - pathType: Prefix
            path: "/agent/v1/sqlDataRequest/testAgent"
            backend:
              service:
                name: testAgent
                port:
                  number: 8080
```

### Cluster Role file (charts/agents/template/cluster_role.yaml)

Now you should add your agent to the cluster_role file, two additions are required:
First, before the `rbac` blocks:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: job-creator-testAgent
  namespace: testAgent
---
```

And then in the end:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: job-creator-testAgent
  namespace: testAgent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: job-creator
subjects:
  - kind: ServiceAccount
    name: job-creator-testAgent
    namespace: testAgent
---
```

# Code 

