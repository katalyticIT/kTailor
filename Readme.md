
# kTailor

kTailor is a lightweight and blazing-fast Kubernetes Mutating Webhook that dynamically modifies Deployments on the fly. By utilizing simple, reusable YAML templates stored in ConfigMaps, it effortlessly injects sidecars, environment variables, or volumes without requiring changes to the original source manifests.

[img/kTailor_made_suit.png](kTailor tailors a custom-made suit for your container.)

## Introduction

In modern Kubernetes environments, developers often need to inject standard infrastructure components (like monitoring sidecars, proxy configurations, or specific environment variables) into their applications. Instead of cluttering every single Deployment manifest, **kTailor** centralizes these modifications.

kTailor follows the **KISS principle** (Keep It Simple, Stupid). It is designed to be:
* **Small & Fast:** Written in Go, it utilizes an efficient In-Memory Informer Cache to observe templates. It introduces near-zero latency to your deployment process.
* **Efficient:** It only targets Deployments carrying a specific trigger label and skips everything else.
* **Easy to Use:** Templates are written in plain Kubernetes-like YAML. No complex programming or policy languages are required.

**When NOT to use kTailor:**
If you need highly complex policy enforcement, conditional logic, or want to mutate/validate a wide variety of Kubernetes resources beyond standard Deployments, kTailor might be too simple for your use case. In those scenarios, we highly recommend looking into established policy engines like [Kyverno](https://kyverno.io/) or [OPA Gatekeeper](https://openpolicyagent.org/docs/latest/kubernetes-introduction/).

## Installation

Deploying kTailor is streamlined via the included `Makefile`.

**Prerequisites:**
* A running Kubernetes Cluster.
* `cert-manager` installed (required to automatically generate the TLS certificates for the webhook).
* `docker` and `go` installed on your local machine.

**Steps:**

1. **Build the local binary (optional):**
   ```bash
   make build
   ```
2. **Build and push the Docker image:**
   Ensure you adjust the `IMAGE_RGST` and `IMAGE_REPO` variables in the Makefile or pass them as environment variables.
   ```bash
   make docker-build
   make docker-push
   ```
3. **Deploy to Kubernetes:**
   This applies the RBAC roles, TLS certificates, base templates, and the webhook configuration.
   ```bash
   make deploy
   ```

To quickly rebuild the image and restart the pod during development, you can simply run: `make rollout`. Use `make help` to see all available commands.

## How it Works: Templates & Triggering

### Template Management via ConfigMaps
To maximize robustness and integrate seamlessly with GitOps workflows, kTailor templates are managed entirely as standard Kubernetes `ConfigMaps`. 

For the internal Informer Cache to discover a template, the ConfigMap **must** have the following label:
```yaml
labels:
  ktailor.io/template: "true"
```
The actual template YAML is simply placed inside the `data` section under the generic `template` key.

### Template Structure
A kTailor template consists of different modification segments:
* `modifyContainers`: Alters existing containers. Supports three operations:
  * `insertIfNotExists`: Adds the value only if the key is entirely missing.
  * `insertOrOverwrite`: Adds the value or brutally overwrites an existing one.
  * `setOrAppend`: Adds the value or appends it to an existing one (e.g., merging strings with a colon delimiter).
* `addInitContainers`: Injects completely new InitContainers.
* `addContainers`: Injects completely new sidecar containers.
* `addVolumes`: Attaches new volumes to the Pod spec.

### Triggering the Webhook (`local` vs. `central`)
To instruct kTailor to modify a Deployment, you add the `ktailor.io/fit` label to your Deployment manifest. The value format is `<scope>.<templateName>`.

* **`central.my-template`**: kTailor looks for the ConfigMap `my-template` in its **own namespace** (usually `ktailor`). These are globally managed by cluster administrators.
* **`local.my-template`**: kTailor looks for the ConfigMap in the **Deployment's namespace**. This allows application developers to write and manage their own templates.

**Security Note:** If security takes precedence over developer convenience (to prevent privilege escalation via local namespaces), cluster administrators can disable local templates by setting `allowCustomTemplates: false` in the `ktailor-config` ConfigMap.

## Examples

Here are two practical examples of what you can achieve with kTailor.

### 1. The `insert-env` Example
A basic central template to inject an environment variable into an existing container.

**The Template (deployed in the `ktailor` namespace):**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ktailor-test
  namespace: ktailor
  labels:
    ktailor.io/template: "true"
data:
  template: |
    kind: ktailor-template
    modifyContainers:
      insertOrOverwrite:
        env:
          - name: KTAILORTEST
            value: "Env set by central kTailor template."
```

**The Deployment (deployed anywhere):**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ktailor-insert-env
  labels:
    ktailor.io/fit: "central.ktailor-test"
spec:
  # ... standard deployment spec ...
```
When this deployment is applied, kTailor intercepts it and injects the `KTAILORTEST` environment variable before the Pod is created.

### 2. The `timetravel` Example (Advanced)
This is a classic infrastructure hack. It utilizes the `libfaketime` library to manipulate the system time for a specific container *without changing the actual node time*.

This template performs three actions at once:
1. It adds an `emptyDir` shared volume.
2. It injects an `InitContainer` that copies the `libfaketime.so` binary into the shared volume.
3. It modifies the main container to mount the shared volume and sets the `LD_PRELOAD` and `FAKETIME` environment variables to activate the time manipulation.

**The Template (`lft-plus222d`):**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: lft-plus222d
  namespace: ktailor
  labels:
    ktailor.io/template: "true"
data:
  template: |
    kind: ktailor-template
    modifyContainers:
      insertIfNotExists:
        volumeMounts:
          - name: shared-lft-volume
            mountPath: /lft_volume
      insertOrOverwrite:
        env:
          - name: FAKETIME
            value: "+222d"
      setOrAppend:
        env:
          - name: LD_PRELOAD
            value: "/lft_volume/libfaketime.so.1"
    
    addInitContainers:
      - name: inject-libfaketime
        image: katalytic/libfaketime_init:1.0
        env:
          - name: LFT_DESTPATH
            value: /lft_volume
        volumeMounts:
        - name: shared-lft-volume
          mountPath: /lft_volume
          
    addVolumes:
      volumes:
        - name: shared-lft-volume
          emptyDir: {}
```
By simply adding `ktailor.io/fit: "central.lft-plus222d"` to any Deployment, the application inside will instantly believe it is running 222 days in the future, completely abstracting the complex volume and init-container logic away from the developer.

## Acknowledgements
A quick note: AI tools were used to assist in the coding and documentation of this project.

