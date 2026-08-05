# checkov:skip=CKV_DOCKER_2:ensure that HEALTHCHECK instructions have been added
FROM golang:1.26.5-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2
LABEL version="v0.1.0"

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

# Switch to user (numeric uid:gid of $APPUSER:$APPGROUP, resolvable by the host system)
USER 10001:10001

# Prepare ~/.ssh with GitHub's public host key only. The SSH private key is NOT baked
# into the image; run_test.sh materializes it at container start from the
# SSH_PRIVATE_KEY environment variable (or use a mounted ~/.ssh/id_ed25519).
RUN mkdir ~/.ssh/ \
    && chmod 700 ~/.ssh/ \
    && ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

ENV GITHUB_BRANCH=master
ENV GITHUB_REPO="git@github.com:indykite/terraform-provider-indykite.git"

WORKDIR "${APPUSER_HOME}/github"

# trivy:ignore:AVD-DS-0026 - TODO: Add HEALTHCHECK instruction in your Dockerfile
# HEALTHCHECK
ENTRYPOINT ["${APPUSER_HOME}/run_test.sh"]
