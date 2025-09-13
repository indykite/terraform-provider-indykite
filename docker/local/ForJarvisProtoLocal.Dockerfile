# checkov:skip=CKV_DOCKER_2:ensure that HEALTHCHECK instructions have been added
FROM golang:1.24
LABEL version="v0.0.1"

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

#You can start with any base Docker Image that works for you
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    wget \
    curl \
    openssh-client \
    ca-certificates \
    gnupg \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

# Install terraform
# hadolint ignore=DL3008,DL3016
RUN curl -fsSL https://apt.releases.hashicorp.com/gpg | \
    gpg --dearmor | \
    tee /usr/share/keyrings/hashicorp-archive-keyring.gpg > /dev/null \
    && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
    https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
    tee /etc/apt/sources.list.d/hashicorp.list > /dev/null \
	&& rm -rf /var/lib/apt/lists/* \
    && apt-get update && apt-get install -y --no-install-recommends terraform

# Add new user and not using root to run the tests for security reasons
RUN useradd --create-home -u 10001 appuser

ENV APPUSER_HOME=/home/appuser

COPY run_tests_on_local_be.sh ${APPUSER_HOME}/run_test.sh
RUN chmod +x ${APPUSER_HOME}/run_test.sh \
    && chown appuser:appuser ${APPUSER_HOME}/run_test.sh \
    && mkdir ${APPUSER_HOME}/github \
    && chown appuser:appuser ${APPUSER_HOME}/github

# Switch to user
USER appuser

# Add ssh private key into container
# trivy:ignore:AVD-DS-0031 - TODO: find a better way to pass it
ARG SSH_PRIVATE_KEY
RUN mkdir ~/.ssh/ \
    && echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_ed25519 \
    && chmod 600 ~/.ssh/id_ed25519 \
    && ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

ENV GITHUB_BRANCH=master
ENV GITHUB_REPO="git@github.com:indykite/terraform-provider-indykite.git"

WORKDIR ${APPUSER_HOME}/github

# trivy:ignore:AVD-DS-0026 - TODO: Add HEALTHCHECK instruction in your Dockerfile
# HEALTHCHECK
ENTRYPOINT ["${APPUSER_HOME}/run_test.sh"]
