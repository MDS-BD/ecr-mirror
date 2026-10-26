# ecr-mirror

Copy a container image between two private ECR registries, each reachable through its own AWS profile. Pure Go: no
`aws` CLI, no `docker`, no `docker login`, no `crane`. Multi-arch indexes are copied whole; blobs already in the target
are skipped.

```sh
go build -o bin/ecr-mirror ./cmd
ecr-mirror --src-profile profile1 --dst-profile profile2 \
  111111111111.dkr.ecr.eu-west-1.amazonaws.com/my-app:tag-1 \
  222222222222.dkr.ecr.eu-west-1.amazonaws.com/my-app-new:latest

ecr-mirror --platform linux/arm64 ...   # copy one platform instead of the whole index
ecr-mirror --no-clobber ...             # fail if the target tag already exists
ecr-mirror --no-create-repo ...         # do not create the target repository when missing
```

Without `--src-profile` / `--dst-profile` (or `ECR_MIRROR_SRC_PROFILE` / `ECR_MIRROR_DST_PROFILE`) the standard AWS
credential chain is used. Each side's region comes from its registry host. Run `ecr-mirror --help` for every flag.

## Install

```sh
export HOMEBREW_GITHUB_API_TOKEN=$(gh auth token)   # any token able to read MDS-BD repos
brew tap MDS-BD/tools git@github.com:MDS-BD/homebrew-tools.git
brew install --cask mds-bd/tools/ecr-mirror
```

## Permissions

| Side | Actions |
| ---- | ------- |
| source | `ecr:GetAuthorizationToken`, `ecr:BatchGetImage`, `ecr:GetDownloadUrlForLayer`, `ecr:BatchCheckLayerAvailability` |
| target | `ecr:GetAuthorizationToken`, `ecr:BatchCheckLayerAvailability`, `ecr:InitiateLayerUpload`, `ecr:UploadLayerPart`, `ecr:CompleteLayerUpload`, `ecr:PutImage`, `ecr:BatchGetImage`, plus `ecr:CreateRepository` unless `--no-create-repo` |

## How it works

1. Each image reference is parsed with [go-containerregistry](https://github.com/google/go-containerregistry);
   the registry ID and region come from the host via
   [amazon-ecr-credential-helper](https://github.com/awslabs/amazon-ecr-credential-helper) (`api.ExtractRegistry`).
2. A keychain maps each registry host to its profile's AWS config and asks the credential helper for an ECR token
   (`ecr:GetAuthorizationToken`, cached in `~/.ecr` like the docker helper).
3. The target repository is created if missing (ECR rejects pushes to unknown repositories).
4. The source manifest is fetched and pushed to the target with `remote.Pusher`, streaming blobs from source to target
   without touching disk. Layers already present are not uploaded again.

## Test

```sh
go test ./... -cover
```

## Release

Same as [aws-tunnel](https://github.com/MDS-BD/aws-tunnel#release): push a `v*` tag; GoReleaser builds, signs,
publishes and pushes the cask to the tap. The same repository secrets are needed.

## License and release compliance

The project is licensed under the MIT License; third-party licenses and notices are retained in `LICENSES/third-party/`.
The inventory and source provenance are recorded in `LICENSES/README.txt`; it includes third-party license texts and
notices for the audited runtime dependencies and Go toolchain snapshot. For the copy/paste procedure to rerun the
license audit, record the actual compiler, and refresh licenses before release, see the tap's
[license audit instructions](https://github.com/MDS-BD/homebrew-tools/blob/main/PUBLISHING.md#rerunning-the-license-audit).
