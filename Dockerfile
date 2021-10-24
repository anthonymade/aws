ARG GO_VERSION

FROM golang:${GO_VERSION}

ARG USER_ID
ARG GROUP_ID

RUN apt-get update && apt-get install -y xz-utils

RUN groupadd --gid $GROUP_ID gouser \
      && useradd --uid $USER_ID --gid $GROUP_ID --create-home gouser --home-dir /home/gouser

USER gouser

WORKDIR /usr/src/myap

RUN go install golang.org/x/tools/cmd/goimports@latest

COPY go* ./

RUN go mod download
