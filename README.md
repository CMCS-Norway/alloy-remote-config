# Alloy Remote Config Server

A lightweight server that implements the Grafana Alloy Remote Config API. This server allows Alloy agents to fetch their configurations remotely, supporting dynamic configuration updates without agent restarts.

## How It Works

This server provides a simple API that matches the expected endpoints for Grafana Alloy remote configuration:

- `GET /api/v1/configs` - Lists all available configurations
- `GET /api/v1/configs/{configID}` - Retrieves a specific configuration
- `HEAD /api/v1/configs/{configID}` - Checks if a configuration exists/has changed
- `POST /api/v1/configs/{configID}` - Creates or updates a configuration
- `GET /health` - Health check endpoint

## Storage Options

This server supports two Kubernetes-friendly storage approaches:

### Option 1: ConfigMap-based Storage (Recommended for GitOps)

Use Kubernetes ConfigMaps to store your Alloy configuration files. This approach works well with GitOps workflows:

1. Define your Alloy configs in your Git repository
2. Use kustomize's `configMapGenerator` to create ConfigMaps
3. Mount these ConfigMaps into the server container

Benefits:
- Fully GitOps compatible
- Version-controlled configurations
- Easy rollbacks
- No need for persistent volumes

```yaml
# Example kustomization.yml
configMapGenerator:
  - name: alloy-configs
    files:
      - configs/dnk.alloy
      - configs/other-config.alloy
```

### Option 2: emptyDir Volume

For more dynamic scenarios or when configs are managed through the API:

```yaml
volumes:
  - name: config-storage
    emptyDir: {}
```

Benefits:
- Supports direct API updates
- No persistent storage required
- Configs can be managed programmatically

Note: Configurations will be lost when pods are restarted with this approach.

## Deployment

### Prerequisites

- Kubernetes cluster
- kubectl access
- (Optional) ArgoCD for GitOps

### Kubernetes Deployment

1. Update the deployment.yml to use your preferred storage option:

```yaml
# For ConfigMap storage
volumes:
  - name: config-storage
    configMap:
      name: alloy-configs
```

2. Deploy using kubectl or ArgoCD:

```bash
kubectl apply -k /path/to/manifests
```

## Managing Configurations

### Using ConfigMaps (GitOps approach)

1. Store your configuration files in your Git repo (e.g., `configs/dnk.alloy`)
2. Use kustomize to generate ConfigMaps
3. When you update configurations, commit changes to Git
4. ArgoCD or your CI/CD pipeline will apply the changes

Example file structure:
```
├── kustomization.yml
├── manifests/
│   ├── deployment.yml
│   ├── service.yml
│   └── ingress.yml
└── configs/
    ├── dnk.alloy
    └── other-config.alloy
```

### Using the API

You can also manage configurations using the API:

```bash
# Upload/update a configuration
curl -X POST https://alloy-config.cmcs.dev/api/v1/configs/dnk \
  -H "Content-Type: application/river" \
  --data-binary @/path/to/dnk.alloy
```

## Alloy Client Configuration

Configure your Alloy agents to use this server:

```river
remotecfg {
  url = "https://alloy-config.cmcs.dev/api/v1/configs/dnk"
  oauth2 {
    client_id = "your-client-id"
    client_secret = "your-client-secret"
    scopes = ["your-scope"]
    token_url = "https://your-oauth-server/token"
  }
  poll_frequency = "5m"
}
```

## Building the Docker Image

```bash
docker build -t ghcr.io/your-org/alloy-remote-config-server:latest .
docker push ghcr.io/your-org/alloy-remote-config-server:latest
```

Or use the provided GitHub Actions workflow to automatically build and push the image when you push changes to the repository.
