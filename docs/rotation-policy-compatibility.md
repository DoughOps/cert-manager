# Private key rotation compatibility in v1.20.3

This fork restores the v1.19 behavior of the controller feature gate
`DefaultPrivateKeyRotationPolicyAlways`. Upstream v1.20.3 still accepts the
gate, but always defaults an omitted rotation policy to `Always`.

The gate remains enabled by default. To make certificates with an omitted
`spec.privateKey.rotationPolicy` use `Never`, start the controller built from
this fork with:

```text
--feature-gates=DefaultPrivateKeyRotationPolicyAlways=false
```

With the Helm chart, set the controller's value:

```yaml
featureGates: "DefaultPrivateKeyRotationPolicyAlways=false"
```

Use the controller image built from this fork when deploying the chart.
The upstream v1.20.3 image does not contain this compatibility change.
This flag is configured on the controller; the webhook needs no corresponding
flag.

| Controller gate | Omitted policy | Explicit `Never` | Explicit `Always` |
| --- | --- | --- | --- |
| `true` (default) | `Always` | `Never` | `Always` |
| `false` | `Never` | `Never` | `Always` |

`Never` reuses an existing matching private key in the target Secret. A key is
generated when the Secret or key is missing, or the key cannot be decoded.
An existing key with a mismatching algorithm or size requires user intervention.
The controller applies this default at runtime without writing it into the
Certificate spec. Existing Certificates
that explicitly store `rotationPolicy: Always` continue to use `Always` until
that field is changed or omitted.
