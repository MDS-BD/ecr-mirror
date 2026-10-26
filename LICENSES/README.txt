ecr-mirror third-party notices
================================

This bundle applies to the ecr-mirror release binary built from ./cmd for
Darwin/Linux amd64/arm64, CGO_ENABLED=0. The included Go 1.27.3 toolchain and
vendor license files are a snapshot from the prior audit, established
2026-10-26; they do not pin or promise the compiler used for future releases.
Before release, rerun the audit procedure in the tap's PUBLISHING.md and
refresh the toolchain/vendor inventory and notices for the actual release
compiler. Runtime module versions were inventoried from the release dependency
closure using `go list -mod=readonly -deps -json ./cmd` for all four target
pairs and normalized to module/version pairs. Exact-version source retrieval
references follow in this file. The included license and notice files are
verbatim copies, not a replacement license or assertion of third-party
ownership. Dependencies remain under their respective terms.

Exact module-version-to-license-file mapping, including retrieval source URLs,
is in `manifest.tsv`. File contents are copied verbatim from those tagged
sources. The 31 external runtime modules are:
- github.com/aws/aws-sdk-go-v2 v1.47.1
- github.com/aws/aws-sdk-go-v2/config v1.33.6
- github.com/aws/aws-sdk-go-v2/credentials v1.20.6
- github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.1
- github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.4
- github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.4
- github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.4
- github.com/aws/aws-sdk-go-v2/service/ecr v1.66.1
- github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.38.10
- github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.19
- github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.4
- github.com/aws/aws-sdk-go-v2/service/signin v1.10.1
- github.com/aws/aws-sdk-go-v2/service/sso v1.38.1
- github.com/aws/aws-sdk-go-v2/service/ssooidc v1.43.1
- github.com/aws/aws-sdk-go-v2/service/sts v1.51.1
- github.com/aws/smithy-go v1.28.1
- github.com/awslabs/amazon-ecr-credential-helper/ecr-login v0.12.0
- github.com/docker/cli v29.7.2+incompatible
- github.com/docker/docker-credential-helpers v0.9.5
- github.com/google/go-containerregistry v0.22.1
- github.com/klauspost/compress v1.19.2
- github.com/mattn/go-colorable v0.1.14
- github.com/mattn/go-isatty v0.0.20
- github.com/mitchellh/go-homedir v1.1.0
- github.com/opencontainers/go-digest v1.0.0
- github.com/opencontainers/image-spec v1.1.1
- github.com/rs/zerolog v1.35.1
- github.com/sirupsen/logrus v1.9.4
- golang.org/x/sync v0.22.0
- golang.org/x/sys v0.47.0

The bundle retains each applicable exact-version module LICENSE, AWS SDK
module licenses and NOTICE, Docker CLI LICENSE and NOTICE, smithy-go LICENSE
and NOTICE plus the scanner-identified included BSD license files, and the
klauspost/compress root LICENSE. The latter expressly includes its nested
snapref/S2 BSD terms and xxhash MIT terms. No extra license has been
invented; the full source license text is kept intact.

The included Go 1.27.3 distribution LICENSE covers the Go runtime and
standard-library code in that prior audit snapshot. The distribution's
vendored golang.org/x/crypto, x/net, x/sys, and x/text source licenses are
included separately because portions were linked into the audited program.
They are historical inventory evidence; refresh them for the actual compiler
before release.

Audit source and retrieval references for exceptional attribution:
- Go: https://go.dev/dl/ (go1.27.3 source/distribution LICENSE and src/vendor
  for the prior audit snapshot; refresh using the actual compiler)
- AWS SDK: https://github.com/aws/aws-sdk-go-v2/tree/v1.47.1 and version-tagged
  submodule releases listed above
- smithy-go: https://github.com/aws/smithy-go/tree/v1.28.1
- Docker CLI: https://github.com/docker/cli/tree/v29.7.2
- klauspost/compress: https://github.com/klauspost/compress/tree/v1.19.2

Re-run the four-target go-list inventory after any dependency, compiler, or
release-target change, inspect all changed runtime package license/notice
files (including nested source), and update this manifest and archive contents
before distributing binaries.
