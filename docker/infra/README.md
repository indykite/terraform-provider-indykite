# Build infra image

The Dockerfile in this directory defines two images used for running the Terraform provider integration tests.

Each image has its own version file:

- [`ci/.version`](ci/.version) — version of the `terraform-tests-ci` (pipeline) image.
- [`full/.version`](full/.version) — version of the `terraform-tests` image.

The [`check-version.sh`](check-version.sh) pre-commit hook **auto-bumps these versions** based on which directory changed — see [Versioning](#versioning).

> **Note:** the per-image directories were renamed to match the image names:
> `pipeline/` → [`ci/`](ci/) (the `terraform-tests-ci` image) and
> `private_cloud/` → [`full/`](full/) (the `terraform-tests` image). If you
> have local references to the old paths (scripts, mounts, bookmarks), update
> them to `docker/infra/ci/` and `docker/infra/full/`.

## Images

The Dockerfile is multi-stage and produces two targets sharing a common `base` stage (Alpine + Go + OpenTofu + a non-root `appuser`).

### `terraform-tests-ci` (pipeline image)

- **Purpose:** the image used in our pipeline during deployment. It does **not** contain any tests baked in.
- **How it works:** at runtime the container clones the `terraform-provider-indykite` repository from GitHub, reads the test configuration from a GCP secret, runs `make upgrade_test_provider integration`, and uploads the report to a GCS bucket.
- **Extras:** includes the Google Cloud SDK (`gcloud`/`gsutil`) for secret access and result upload.
- **Entry point:** [`ci/startscript.sh`](ci/startscript.sh).
- **When it is (re)built:** whenever files in `docker/infra/` directory change.
- **Required env vars:** `GITHUB_USER`, `GITHUB_TOKEN`, `GITHUB`, `BUCKET_NAME` (and optionally `BRANCH`, `RUN_ENV`, `RELEASE_VERSION`, `SECRET_NAME`, `SLACK_WEBHOOK_URL`).

### `terraform-tests` (for non-SaaS users)

- **Purpose:** an image that contains the integration tests **baked in**. It is capable of running the tests without any GitHub or GCloud connectivity — the tests directory is copied into the image at build time.
- **How it works:** at runtime the container reads the test configuration from a local JSON file
  (mounted or provided at a path given by `TERRAFORM_CONFIG_FILE`), then runs `terraform apply`
  and `go test --tags=integration ./...` against the baked-in `tests/` directory, and finally
  destroys the created resources.
- **Entry point:** [`full/startscript.sh`](full/startscript.sh).
- **When it is (re)built:** whenever files in `docker/infra/` or `tests/` directories change.
- **Required env vars:** `TERRAFORM_CONFIG_FILE` (path to a JSON config file inside the container), optionally `RUN_ENV` (defaults to `cloud`).

### Differences at a glance

| | `terraform-tests-ci` (pipeline) | `terraform-tests` |
| --- | --- | --- |
| Tests baked in? | No — cloned at runtime | Yes — copied from `tests/` at build time |
| Needs GitHub access? | Yes (to clone the repo) | No |
| Needs GCloud access? | Yes (secrets + result upload) | No |
| Test configuration source | GCP Secret Manager | Local JSON file (`TERRAFORM_CONFIG_FILE`) |
| Includes `gcloud` SDK | Yes | No |
| Entry script | `ci/startscript.sh` | `full/startscript.sh` |

## Building the images manually

Both builds must be run **from the repository root** so that the Dockerfile can access the `tests/` directory (required by the baked-in tests image).

Build the ci image:

```bash
docker build \
    -f docker/infra/Dockerfile \
    --target terraform-plugin-tests-ci \
    -t terraform-plugin-tests-ci:local \
    .
```

Build the full image with the tests baked in:

```bash
docker build \
    -f docker/infra/Dockerfile \
    --target terraform-plugin-tests \
    -t terraform-plugin-tests:local \
    .
```

For a multi-platform build (as done in CI), use `docker buildx`:

```bash
docker buildx build \
    -f docker/infra/Dockerfile \
    --target terraform-plugin-tests-ci \
    --platform linux/amd64,linux/arm64,linux/arm/v7 \
    -t terraform-plugin-tests-ci:local \
    .
```

## Versioning

Version files hold a `vX.Y.Z` string (the `v` prefix is required). The [`check-version.sh`](check-version.sh) pre-commit hook auto-bumps the versions based on which directory changed:

| Change            | `ci/.version`          | `full/.version`        |
|-------------------|------------------------|------------------------|
| `docker/infra/**` | minor bump (`X.Y+1.0`) | minor bump (`X.Y+1.0`) |
| `tests/**`        | —                      | patch bump (`X.Y.Z+1`) |

If a `.version` file is already staged (i.e. you edited it yourself), the hook leaves that image alone — this is how you override the default to a major or hand-picked version.

When the hook bumps a file, it writes the new version and aborts the commit so you can review the change. The bump is **not** staged automatically — stage it yourself (the hook prints the exact `git add` command) and re-run `git commit` to complete the commit with the bumped version included.

## CI behaviour

The [`docker-build-tf-plugin-tests.yaml`](../../.github/workflows/docker-build-tf-plugin-tests.yaml) workflow is the single source of truth for when images are pushed:

- **On pull requests to `master`** — each affected image is built and pushed to the registry tagged with the version from its `.version` file.
- **On push to `master` (after merge)** — no rebuild; the existing versioned image is manifest-retagged as `:latest` via `docker buildx imagetools create`, preserving multi-arch manifests. Only images whose `.version` file changed in the merge get retagged.
- **On manual `workflow_dispatch`** — both images are rebuilt with their current versioned tag, but `:latest` is not touched.
