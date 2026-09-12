# This dockerfile is supposed to be run by goreleaser.
# Using plain `docker build` will not work as running `make` will put binaries in different
# paths than goreleaser. To build images locally, use goreleaser from within ./src directoryr:
# goreleaser release --snapshot --clean --skip=archive,sign,publish,announce,sbom,before,homebrew

FROM --platform=$BUILDPLATFORM alpine:3.24.1 AS terraform-downloader
ARG TARGETARCH

# renovate: datasource=github-releases depName=hashicorp/terraform
ARG TERRAFORM_VERSION=1.14.7

RUN apk add --no-cache curl unzip
RUN curl -fsSL -o terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip \
    && unzip terraform.zip \
    && mv terraform /usr/local/bin/terraform

FROM alpine:3.24.1
ARG TARGETPLATFORM

COPY $TARGETPLATFORM/kubara /usr/local/bin/kubara
COPY --from=terraform-downloader /usr/local/bin/terraform /usr/local/bin/terraform

RUN apk add --no-cache helm kubectl yq \
    && addgroup -S kubara && adduser -S kubara -G kubara

USER kubara
WORKDIR /home/kubara

ENV KUBARA_UPDATE_CHECK=0

ENTRYPOINT ["/usr/local/bin/kubara", "--kubeconfig", "/kubeconfig"]
