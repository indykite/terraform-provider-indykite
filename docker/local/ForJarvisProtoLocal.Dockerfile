# checkov:skip=CKV_DOCKER_2:ensure that HEALTHCHECK instructions have been added
FROM golang:1.24-alpine@sha256:8f8959f38530d159bf71d0b3eb0c547dc61e7959d8225d1599cf762477384923
LABEL version="v0.0.1"

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

ENV APPUSER="appuser"
ENV APPGROUP="appgroup"
ENV APPUSER_HOME="/home/$APPUSER"

#You can start with any base Docker Image that works for you
# hadolint ignore=DL3018
RUN apk add --update --no-cache \
        curl \
        jq \
        git \
        openssh-client \
    && apk upgrade \
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
        --home "$APPUSER_HOME" \
    && apk info -v \
    && terraform -version

COPY run_tests_on_local_be.sh "${APPUSER_HOME}/run_test.sh"
RUN chmod +x "${APPUSER_HOME}/run_test.sh" \
    && mkdir "${APPUSER_HOME}/github" \
    && chown "$APPUSER":"$APPGROUP" "${APPUSER_HOME}/run_test.sh" "${APPUSER_HOME}/github"

# Switch to user
USER "$APPUSER"

# Add ssh private key into container
# trivy:ignore:AVD-DS-0031 - TODO: find a better way to pass it
ARG SSH_PRIVATE_KEY
RUN mkdir ~/.ssh/ \
    && echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_ed25519 \
    && chmod 600 ~/.ssh/id_ed25519 \
    && ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

ENV GITHUB_BRANCH=master
ENV GITHUB_REPO="git@github.com:indykite/terraform-provider-indykite.git"

WORKDIR "${APPUSER_HOME}/github"

# trivy:ignore:AVD-DS-0026 - TODO: Add HEALTHCHECK instruction in your Dockerfile
# HEALTHCHECK
ENTRYPOINT ["${APPUSER_HOME}/run_test.sh"]
