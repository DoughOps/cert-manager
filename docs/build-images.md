# DoughOps container image builds

The `Build container images` GitHub Actions workflow builds this fork's five
components using the existing Make targets, Containerfiles, license notices and
digest-pinned distroless base images. Go is selected from the repository's
`make print-go-version` output. Images are built for `linux/amd64` (Intel/AMD x86-64 Kubernetes nodes).

Following [DoughOps' GHCR guidance](https://github.com/DoughOps/.github/blob/main/profile/README.md),
images belong to the organization and use the Actions-provided `GITHUB_TOKEN`.
No additional registry password or token secret is required. The workflow grants
`packages: write` for publication and supplies the fork's source/revision labels.
License references point to this fork at the built commit. GHCR repository names
are lowercase.

## Triggers and image names

Pull requests targeting `master` or `release-*` build all images without registry
login or publication. Pushes to those branches, version tags starting with `v`,
and manual runs publish images. The workflow runs only in `DoughOps/cert-manager`.
Manual dispatch is available after GitHub has registered the workflow on the
repository's default branch.

| Component | Image |
| --- | --- |
| Controller | `ghcr.io/doughops/cert-manager-controller` |
| ACME solver | `ghcr.io/doughops/cert-manager-acmesolver` |
| CA injector | `ghcr.io/doughops/cert-manager-cainjector` |
| Webhook | `ghcr.io/doughops/cert-manager-webhook` |
| Startup API check | `ghcr.io/doughops/cert-manager-startupapicheck` |

Every published image has a `sha-<full-commit-sha>` tag. Branch builds also have
their branch tag, such as `release-1.20.3`; version-tag builds retain the exact
tag, such as `v1.20.3-doughops.1`. Only the default branch receives `latest`.

Each component is built and published by its own matrix job. Pull-request builds
do not log in to GHCR or push images. Branch/tag/manual builds publish the image
under the generated tags after that component builds successfully.

## Deploying the fork

Use the same published tag for all components. For example, with the packaged
cert-manager Helm chart and the `release-1.20.3` image tag:

```yaml
image:
  repository: ghcr.io/doughops/cert-manager-controller
  tag: release-1.20.3
acmesolver:
  image:
    repository: ghcr.io/doughops/cert-manager-acmesolver
    tag: release-1.20.3
cainjector:
  image:
    repository: ghcr.io/doughops/cert-manager-cainjector
    tag: release-1.20.3
webhook:
  image:
    repository: ghcr.io/doughops/cert-manager-webhook
    tag: release-1.20.3
startupapicheck:
  image:
    repository: ghcr.io/doughops/cert-manager-startupapicheck
    tag: release-1.20.3
featureGates: "DefaultPrivateKeyRotationPolicyAlways=false"
```

For reproducible deployments, select a commit or version tag instead of the
moving branch tag. The repository's `values.yaml` also supports image digests.
If the organization restricts package creation or image visibility, an org owner
must permit this repository's Actions token to publish these packages and set
the intended package visibility. Published packages use GHCR's configured
visibility; this workflow does not change organization or package settings.
