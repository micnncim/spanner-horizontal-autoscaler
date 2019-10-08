# Spanner Horizontal Autoscaler

*NOTE*: This is just experimental and PoC. Not for production usecase.

## Example

```yaml
apiVersion: spannerhorizontalautoscaler.k8s.io/v1alpha1
kind: SpannerInstance
metadata:
  name: spannerinstance-sample
spec:
  instanceId: instance-sample
  minNodes: 1
  maxNodes: 5
  cpuUtilizationThreshold: 70
```
