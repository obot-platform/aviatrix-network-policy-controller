# aviatrix-network-policy-controller

A controller to create Aviatrix `FirewallPolicy` resources for Obot-deployed MCP servers.

## Helm values

The controller is intended to be installed by Obot when the Obot server is configured with a network policy provider chart. Obot supplies the runtime-specific values listed below and merges any YAML or JSON from `OBOT_SERVER_MCPNETWORK_POLICY_PROVIDER_VALUES` into the chart values.

| Value | Default | Description |
| --- | --- | --- |
| `image.repository` | `ghcr.io/obot-platform/aviatrix-network-policy-controller` | Controller image repository. |
| `image.tag` | `""` | Controller image tag. Defaults to the chart `appVersion`; if `appVersion` is a development version such as `0.0.0-dev`, the chart uses `main`. |
| `image.pullPolicy` | `Always` | Kubernetes image pull policy for the controller container. |
| `imagePullSecrets` | `[]` | Image pull secrets added to the controller pod. |
| `nameOverride` | `""` | Overrides the chart name used in generated resource names. |
| `fullnameOverride` | `""` | Overrides the full release name used in generated resource names. |
| `serviceAccount.create` | `true` | Reserved for service account configuration. The chart currently renders a service account for the controller. |
| `serviceAccount.name` | `""` | Existing or custom service account name for the controller. Defaults to the chart fullname. |
| `podSecurityContext` | See `chart/values.yaml` | Pod-level security context for the controller pod. |
| `securityContext` | See `chart/values.yaml` | Container-level security context for the controller container. |
| `resources` | See `chart/values.yaml` | CPU and memory requests and limits for the controller container. |
| `secretName` | `obot-network-policy-provider` | Secret containing the Obot network policy provider API key. Obot creates and rotates this secret when the provider is enabled. |
| `obotStorageURL` | `""` | Required. Internal HTTPS URL for Obot storage APIs. Obot sets this automatically when it installs the provider. |
| `obotStorageTokenFile` | `/var/run/secrets/obot-network-policy-provider/apiKey` | File path inside the controller container that contains the Obot storage API key. |
| `mcpRuntimeNamespace` | `obot-mcp` | Kubernetes namespace containing the MCP server runtime resources and Aviatrix `FirewallPolicy` objects. |
| `obot.serviceAccount.name` | `""` | Obot server service account name. Obot sets this automatically so the provider can bind back to the Obot runtime context. |
| `obot.serviceAccount.namespace` | `""` | Namespace containing the Obot server service account. Obot sets this automatically. |

## Obot configuration

Enable this provider from Obot by setting either `OBOT_SERVER_MCPNETWORK_POLICY_PROVIDER_CHART_NAME` with `OBOT_SERVER_MCPNETWORK_POLICY_PROVIDER_CHART_REPO`, or `OBOT_SERVER_MCPNETWORK_POLICY_PROVIDER_CHART_PATH` for a local chart. The MCP runtime backend must be `kubernetes`.

Use `OBOT_SERVER_MCPNETWORK_POLICY_PROVIDER_VALUES` to override chart values. For example:

```yaml
mcpRuntimeNamespace: custom-mcp-runtime
resources:
  requests:
    cpu: 100m
    memory: 128Mi
```

Obot always supplies `mcpRuntimeNamespace`, `obotStorageURL`, `secretName`, `obotStorageTokenFile`, and `obot.serviceAccount` defaults before applying this override blob.
