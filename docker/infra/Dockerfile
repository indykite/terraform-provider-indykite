# checkov:skip=CKV_DOCKER_2:ensure that HEALTHCHECK instructions have been added
FROM golang:1.24-alpine

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

ENV APPUSER="appuser"
ENV APPGROUP="appgroup"
ENV APPUSER_HOME="/home/$APPUSER"
ENV APPUSER_HOME=/home/appuser
ENV CLOUDSDK_INSTALL_DIR=${APPUSER_HOME}
ENV PATH="$PATH:${APPUSER_HOME}/google-cloud-sdk/bin"

# hadolint ignore=DL3018
RUN apk add --update --no-cache \
        bash \
        curl \
        jq \
        git \
        python3 \
        openssh-client \
    # Install OpenTofu (open source Terraform clone)
    && curl --proto '=https' --tlsv1.2 -fsSL https://get.opentofu.org/install-opentofu.sh -o install-opentofu.sh \
    && chmod +x install-opentofu.sh \
    && ./install-opentofu.sh --install-method apk \
    && rm -f install-opentofu.sh \
    # Add new user and not using root to run the tests for security reasons
    && addgroup -S "$APPGROUP" --gid 10001 \
    && adduser -S "$APPUSER" --uid 10001 \
        -G "$APPGROUP" \
        --disabled-password \
        --gecos "" \
        --home "$APPUSER_HOME" \
    # Install gscloud for posting the results in the bucket
    && curl -sSL https://sdk.cloud.google.com | bash


COPY startscript.sh "${APPUSER_HOME}/startscript.sh"
RUN chmod +x "${APPUSER_HOME}/startscript.sh"

# Switch to user
USER "$APPUSER"

# Checkout github repository
ENV CI=true


RUN mkdir "${APPUSER_HOME}/github"
WORKDIR "${APPUSER_HOME}/github"

# trivy:ignore:AVD-DS-0026 - TODO: Add HEALTHCHECK instruction in your Dockerfile
# HEALTHCHECK
CMD ["../startscript.sh"]
