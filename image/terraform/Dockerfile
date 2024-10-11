FROM debian:bullseye-slim
LABEL authors="Brookatlas"

RUN apt-get update && apt-get install -y \
    wget \
    unzip \
    git \
    && rm -rf /var/lib/apt/lists/*

ENV TERRAFORM_VERSION=1.8.0
RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip \
    && unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin/ \
    && chmod +x /usr/local/bin/terraform \
    && rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip

WORKDIR /workspace
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

ENTRYPOINT ["/bin/bash", "-c", "./entrypoint.sh"]
