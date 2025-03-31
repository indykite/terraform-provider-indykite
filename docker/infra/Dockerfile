FROM golang:1.23
#You can start with any base Docker Image that works for you
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
        software-properties-common \
        && rm -rf /var/lib/apt/lists/*

# Install terraform
RUN wget -O- https://apt.releases.hashicorp.com/gpg | \
    gpg --dearmor | \
    tee /usr/share/keyrings/hashicorp-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
    https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
    tee /etc/apt/sources.list.d/hashicorp.list \
    && apt-get update && apt-get install -y --no-install-recommends \
    terraform

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

CMD ["../startscript.sh"]
