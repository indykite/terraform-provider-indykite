# checkov:skip=CKV_DOCKER_2:ensure that HEALTHCHECK instructions have been added
FROM golang:1.26.2-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS base

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

ENV APPUSER="appuser"
ENV APPGROUP="appgroup"
ENV APPUSER_HOME="/home/$APPUSER"
ENV CLOUDSDK_INSTALL_DIR="${APPUSER_HOME}"
ENV PATH="$PATH:${APPUSER_HOME}/google-cloud-sdk/bin"

# hadolint ignore=DL3018
RUN apk add --update --no-cache \
        bash \
        curl \
        jq \
        git \
        make \
        python3 \
        openssh-client \
    # Install OpenTofu (open source Terraform clone)
    && curl --proto '=https' --tlsv1.2 -fsSL https://get.opentofu.org/install-opentofu.sh -o install-opentofu.sh \
    && chmod +x install-opentofu.sh \
    && ./install-opentofu.sh --install-method apk \
    && rm -f install-opentofu.sh \
    && ln -s /usr/bin/tofu /usr/bin/terraform \
    # Add new user and not using root to run the tests for security reasons
    && addgroup -S "$APPGROUP" --gid 10001 \
    && adduser -S "$APPUSER" --uid 10001 \
        -G "$APPGROUP" \
        --disabled-password \
        --gecos "" \
        --home "$APPUSER_HOME"

# ─── Image without baked-in tests (for pipeline use) ────────────────────────
FROM base AS terraform-plugin-tests-ci

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

# Install gscloud for posting the results in the bucket
RUN curl -sSL https://sdk.cloud.google.com | bash \
    && apk info -v \
    && terraform -version

COPY --chown=10001:10001 docker/infra/ci/startscript.sh "${APPUSER_HOME}/startscript.sh"

# Switch to user
USER "$APPUSER"

# Checkout github repository
ENV CI=true

RUN mkdir "${APPUSER_HOME}/github"
WORKDIR "${APPUSER_HOME}/github"

# HEALTHCHECK is not here, as this just execute command and terminates
# checkov:skip=CKV_DOCKER_2:ensure that HEALTHCHECK instructions have been added
CMD ["../startscript.sh"]

# ─── Image with baked-in tests (for non-SaaS use) ──────────────────────
FROM base AS terraform-plugin-tests

RUN terraform -version

COPY --chown=10001:10001 docker/infra/full/startscript.sh "${APPUSER_HOME}/startscript.sh"
COPY --chown=10001:10001 tests/ ${APPUSER_HOME}/tests

# Switch to user
USER "$APPUSER"

# Checkout github repository
ENV CI=true

WORKDIR "${APPUSER_HOME}/tests"

# HEALTHCHECK is not here, as this just execute command and terminates
# checkov:skip=CKV_DOCKER_2:ensure that HEALTHCHECK instructions have been added
CMD ["../startscript.sh"]
