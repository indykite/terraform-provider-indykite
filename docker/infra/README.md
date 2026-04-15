# Build infra image

The Dockerfile in this directory defines two images used for running the Terraform provider integration tests.

Each image has its own version file:

- [`pipeline/.version`](pipeline/.version) — version of the `terraform-tests` (pipeline) image.
- [`private_cloud/.version`](private_cloud/.version) — version of the `terraform-tests-for-private-cloud` image.

The [`check-version.sh`](check-version.sh) pre-commit hook **auto-bumps these files** based on what you changed (see the *Versioning* section below). You normally do not need to edit them by hand.

## Images

The Dockerfile is multi-stage and produces two targets sharing a common `base` stage (Alpine + Go + OpenTofu + a non-root `appuser`).

### `terraform-tests` (pipeline image)

- **Purpose:** the image used in our pipeline during deployment. It does **not** contain any tests baked in.
- **How it works:** at runtime the container clones the `terraform-provider-indykite` repository from GitHub, reads the test configuration from a GCP secret, runs `make upgrade_test_provider integration`, and uploads the report to a GCS bucket.
- **Extras:** includes the Google Cloud SDK (`gcloud`/`gsutil`) for secret access and result upload.
- **Entry point:** [`pipeline/startscript.sh`](pipeline/startscript.sh).
- **When it is (re)built:** whenever `docker/infra/Dockerfile` or `docker/infra/pipeline/startscript.sh` changes.
- **Required env vars:** `GITHUB_USER`, `GITHUB_TOKEN`, `GITHUB`, `BUCKET_NAME` (and optionally `BRANCH`, `RUN_ENV`, `RELEASE_VERSION`, `SECRET_NAME`, `SLACK_WEBHOOK_URL`).

### `terraform-tests-for-private-cloud` (private cloud image)

- **Purpose:** an image that contains the integration tests **baked in**. It is capable of running the tests without any GitHub or GCloud connectivity — the tests directory is copied into the image at build time.
- **How it works:** at runtime the container reads the test configuration from a local JSON file
  (mounted or provided at a path given by `TERRAFORM_CONFIG_FILE`), then runs `terraform apply`
  and `go test --tags=integration ./...` against the baked-in `tests/` directory, and finally
  destroys the created resources.
- **Entry point:** [`private_cloud/startscript.sh`](private_cloud/startscript.sh).
- **When it is (re)built:** whenever `docker/infra/Dockerfile`, `docker/infra/private_cloud/startscript.sh`, or any file under `tests/` changes.
- **Required env vars:** `TERRAFORM_CONFIG_FILE` (path to a JSON config file inside the container), optionally `RUN_ENV` (defaults to `cloud`).

### Differences at a glance

| | `terraform-tests` (pipeline) | `terraform-tests-for-private-cloud` |
| --- | --- | --- |
| Tests baked in? | No — cloned at runtime | Yes — copied from `tests/` at build time |
| Needs GitHub access? | Yes (to clone the repo) | No |
| Needs GCloud access? | Yes (secrets + result upload) | No |
| Test configuration source | GCP Secret Manager | Local JSON file (`TERRAFORM_CONFIG_FILE`) |
| Includes `gcloud` SDK | Yes | No |
| Entry script | `pipeline/startscript.sh` | `private_cloud/startscript.sh` |

## Building the images manually

Both builds must be run **from the repository root** so that the Dockerfile can access the `tests/` directory (required by the private cloud image).

Build the pipeline image:

```bash
docker build \
    -f docker/infra/Dockerfile \
    --target terraform-tests \
    -t terraform-plugin-tests:local \
    .
```

Build the private cloud image:

```bash
docker build \
    -f docker/infra/Dockerfile \
    --target terraform-tests-for-private-cloud \
    -t terraform-tests-for-private-cloud:local \
    .
```

For a multi-platform build (as done in CI), use `docker buildx`:

```bash
docker buildx build \
    -f docker/infra/Dockerfile \
    --target terraform-tests \
    --platform linux/amd64,linux/arm64,linux/arm/v7 \
    -t terraform-plugin-tests:local \
    .
```

## Versioning

Version files hold an `X.Y.Z` string (with an optional `v` prefix). The [`check-version.sh`](check-version.sh) pre-commit hook auto-bumps them based on the staged files — highest-impact bump wins per image:

| Change                                       | `pipeline/.version`     | `private_cloud/.version` |
|----------------------------------------------|-------------------------|--------------------------|
| `docker/infra/Dockerfile`                    | minor bump (`X.Y+1.0`)  | minor bump (`X.Y+1.0`)   |
| `docker/infra/pipeline/startscript.sh`       | patch bump (`X.Y.Z+1`)  | —                        |
| `docker/infra/private_cloud/startscript.sh`  | —                       | patch bump (`X.Y.Z+1`)   |
| `tests/**`                                   | —                       | patch bump (`X.Y.Z+1`)   |

If a `.version` file is already staged (i.e. you edited it yourself), the hook leaves that image alone — this is how you override the default to a major or hand-picked version.

When the hook bumps a file, it `git add`s it and aborts the commit so you can review the change; re-running `git commit` completes the commit with the bumped version included.

## CI behaviour

The [`docker-build-tf-plugin-tests.yaml`](../../.github/workflows/docker-build-tf-plugin-tests.yaml) workflow is the single source of truth for when images are pushed:

- **On pull requests to `master`** — each affected image is built and pushed to the registry tagged with the version from its `.version` file.
- **On push to `master` (after merge)** — no rebuild; the existing versioned image is manifest-retagged as `:latest` via `docker buildx imagetools create`, preserving multi-arch manifests. Only images whose `.version` file changed in the merge get retagged.
- **On manual `workflow_dispatch`** — both images are rebuilt with their current versioned tag, but `:latest` is not touched.
