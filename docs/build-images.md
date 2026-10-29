# DoughOps container image builds

The `Build container images` GitHub Actions workflow builds this fork's five
components using the existing Make targets, Containerfiles, license notices and
digest-pinned distroless base images. Go is selected from the repository's
`make print-go-version` output. Images are built for `linux/amd64` (Intel/AMD x86-64 Kubernetes nodes).

Following [DoughOps' GHCR guidance](https://github.com/DoughOps/.github/blob/main/profile/README.md),
images belong to the organization and use the Actions-provided `GITHUB_TOKEN`.
The workflow also follows [image-builder](https://github.com/DoughOps/image-builder)
by publishing version/short-SHA tag pairs and a build summary with package links.
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

Tags follow image-builder's two-tag convention with the `-tks-custom` suffix.
`release-1.20.3` builds publish `1.20.3-tks-custom` and
`1.20.3-tks-custom-<7-character-sha>`. Version-tag builds similarly publish
`v1.20.3-tks-custom` and `v1.20.3-tks-custom-<7-character-sha>`; an existing
`-tks-custom` suffix is not added twice. Unversioned builds, such as `master`,
publish `latest-tks-custom` and `<7-character-sha>-tks-custom`.

Each component is built and published by its own matrix job. Pull-request builds
do not log in to GHCR or push images. Branch/tag/manual builds publish the image
under the generated tags after that component builds successfully. The Actions
run summary lists each component's status, platform, tags and GHCR package link.
All versions are retained; the workflow does not delete old packages.

The five existing Containerfiles share `hack/containers/.ci.yml`, which is
required and explicitly sets `schedule: none` and `keep_versions: all`, following
the reference repository's documented policy fields. This release workflow
supports those policies and fails if the file/fields are absent or changed to
unsupported values. The metadata file stays outside the generated image build
contexts. There is no scheduled rebuild, Docker Hub publication or
package-deletion token in this workflow. Those optional image-builder features
are not enabled for this release.

## Deploying the fork

Use the same published tag for all components. For example, with the packaged
cert-manager Helm chart and the `1.20.3-tks-custom` image tag:

```yaml
image:
  repository: ghcr.io/doughops/cert-manager-controller
  tag: "1.20.3-tks-custom"
acmesolver:
  image:
    repository: ghcr.io/doughops/cert-manager-acmesolver
    tag: "1.20.3-tks-custom"
cainjector:
  image:
    repository: ghcr.io/doughops/cert-manager-cainjector
    tag: "1.20.3-tks-custom"
webhook:
  image:
    repository: ghcr.io/doughops/cert-manager-webhook
    tag: "1.20.3-tks-custom"
startupapicheck:
  image:
    repository: ghcr.io/doughops/cert-manager-startupapicheck
    tag: "1.20.3-tks-custom"
featureGates: "DefaultPrivateKeyRotationPolicyAlways=false"
```

The fixed `1.20.3-tks-custom` tag moves when this fork is rebuilt. For reproducible
deployments, select the version/short-SHA tag or an image digest. The
repository's `values.yaml` supports image digests.
If the organization restricts package creation or image visibility, an org owner
must permit this repository's Actions token to publish these packages and set
the intended package visibility. Published packages use GHCR's configured
visibility; this workflow does not change organization or package settings.
New GHCR packages are private by default; set the intended visibility in package
settings or configure pull access for the Kubernetes cluster.
