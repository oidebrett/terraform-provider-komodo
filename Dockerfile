FROM --platform=linux/amd64 alpine:3.22

WORKDIR /workspace

RUN apk add go curl unzip bash sudo nodejs npm vim

ENV GOPATH=/root/go
ENV PATH=$PATH:$GOPATH/bin

# install terraform:
RUN curl -O https://releases.hashicorp.com/terraform/1.12.2/terraform_1.12.2_linux_amd64.zip && \
    unzip terraform_1.12.2_linux_amd64.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform_1.12.2_linux_amd64.zip