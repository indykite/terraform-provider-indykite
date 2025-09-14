# checkov:skip=CKV_DOCKER_2:ensure that HEALTHCHECK instructions have been added
FROM golang:1.24

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

#You can start with any base Docker Image that works for you
#        software-properties-common \
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
		build-essential \
        git \
        wget \
        curl \
        jq \
        python3 \
        openssh-client \
        ca-certificates \
        gnupg \
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

# Install gscloud for posting the results in the bucket
ENV CLOUDSDK_INSTALL_DIR=${APPUSER_HOME}
RUN curl -sSL https://sdk.cloud.google.com | bash
ENV PATH="$PATH:${APPUSER_HOME}/google-cloud-sdk/bin"


COPY startscript.sh ${APPUSER_HOME}/startscript.sh
RUN chmod +x ${APPUSER_HOME}/startscript.sh

# Switch to user
USER appuser

# Checkout github repository
ENV CI=true


RUN mkdir ${APPUSER_HOME}/github
WORKDIR ${APPUSER_HOME}/github

# trivy:ignore:AVD-DS-0026 - TODO: Add HEALTHCHECK instruction in your Dockerfile
# HEALTHCHECK
CMD ["../startscript.sh"]
