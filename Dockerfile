FROM golang:1.19.5-alpine as base
ARG TARGETARCH
ARG VERSION
ARG COMMIT
ARG DATE
RUN apk add --no-cache git
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

## BUILDER STAGE ##
FROM base as builder
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o conftest -ldflags="-w -s -X github.com/open-policy-agent/conftest/internal/commands.version=${VERSION}" main.go

## TEST STAGE ##
FROM base as test
RUN go test -v ./...

## ACCEPTANCE STAGE ##
FROM base as acceptance
COPY --from=builder /app/conftest /app/conftest

RUN apk add --no-cache npm bash
RUN npm install -g bats

RUN bats acceptance.bats

## EXAMPLES STAGE ##
FROM base as examples
ENV TERRAFORM_VERSION=0.12.28 \
    KUSTOMIZE_VERSION=2.0.3

COPY --from=builder /app/conftest /usr/local/bin
COPY examples /examples

WORKDIR /tmp
RUN apk add --no-cache npm make git jq ca-certificates openssl unzip wget && \
    wget "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip" && \
    unzip "terraform_${TERRAFORM_VERSION}_linux_amd64.zip" -d /usr/local/bin

RUN wget -O /usr/local/bin/kustomize "https://github.com/kubernetes-sigs/kustomize/releases/download/v${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_${TARGETARCH}" && \
    chmod +x /usr/local/bin/kustomize

RUN go install cuelang.org/go/cmd/cue@latest

WORKDIR /examples

## RELEASE ##
FROM alpine:3.17.1

# Install git for protocols that depend on it when using conftest pull
RUN apk add --no-cache git

COPY --from=builder /app/conftest /
RUN ln -s /conftest /usr/local/bin/conftest
WORKDIR /project

ENTRYPOINT ["/conftest"]
CMD ["--help"]
