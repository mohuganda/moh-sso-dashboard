# Releasing The Backend

## Automated Release

Run the `Backend Release` GitHub Actions workflow and provide a version such as `1.2.3`. Do not include `backend/v` in the input.

The workflow performs these operations before creating a tag:

1. Validate SemVer and confirm the tag is unused.
2. Verify formatting, vet, tests, and architecture boundaries.
3. Verify all version package/build paths are consistent.
4. Build Linux and macOS artifacts for amd64 and arm64.
5. Execute the Linux binary and verify version, commit, and clean state.
6. Generate SHA-256 checksums, an SPDX SBOM, and provenance attestations.
7. Create and push `backend/v<version>`.
8. Publish the GitHub release artifacts.

Pushing the component tag triggers the backend Docker image build. The image build embeds the same metadata, adds OCI labels, publishes provenance/SBOM data, and runs the image with `--version` as a post-build check.

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
