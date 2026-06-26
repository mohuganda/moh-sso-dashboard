# Releasing The Backend

## Automated Release

The `Backend Release` GitHub Actions workflow runs in two modes:

1. Manual dispatch: provide a version such as `1.2.3`. Do not include `backend/v` in the input.
2. Tag push: pushing `backend/v<version>` runs the same verification and publishes the GitHub release automatically.

If you prefer a release-PR flow, create a branch named `release/backend-v<version>` and merge it into `main`. The `Backend Release From PR` workflow will verify the merge commit, create the `backend/v<version>` tag, and publish the GitHub release automatically.

The workflow performs these operations before publishing the release:

1. Validate SemVer and confirm the tag is unused.
2. Verify formatting, vet, tests, and architecture boundaries.
3. Verify all version package/build paths are consistent.
4. Build Linux and macOS artifacts for amd64 and arm64.
5. Execute the Linux binary and verify version, commit, and clean state.
6. Generate SHA-256 checksums, an SPDX SBOM, and provenance attestations.
7. Create and push `backend/v<version>` only for manual dispatch.
8. Publish the GitHub release artifacts for manual dispatch, tag-push runs, and merged release PRs.

Pushing the component tag triggers the backend Docker image build. The image build embeds the same metadata, adds OCI labels, publishes provenance/SBOM data, and runs the image with `--version` as a post-build check.

The release workflow emits a metadata artifact with both:

- `release_tag`: `backend/v<version>` for the Git tag and GitHub Release
- `image_tag`: `<version>` for the Docker image tag

## Local Release Build

```bash
cd backend
GOCACHE=/private/tmp/moh-sso-go-build make build-release \
  VERSION=1.2.3 \
  COMMIT="$(git rev-parse HEAD)" \
  BUILD_TIME="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
  DIRTY=false
```

Release builds reject `VERSION=dev` and `DIRTY=true`.

## Hotfixes

Branch from the deployed release commit, apply only the required compatible fix, run full verification, and release the next PATCH version. Do not overwrite or move an existing release tag.

## Signing

OCI provenance and SBOM generation are enabled. Cosign signing can be added through GitHub OIDC without storing private signing keys in the repository. Production policy may require signature verification before deployment.
