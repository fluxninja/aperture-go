# syntax=docker/dockerfile:1
FROM golang:1.19-buster AS builder

WORKDIR /src
COPY --link . .

# App build
ARG GIT_COMMIT_HASH
ENV GIT_COMMIT_HASH=${GIT_COMMIT_HASH}

ARG GIT_BRANCH
ENV GIT_BRANCH=${GIT_BRANCH}

ENV GOPRIVATE github.com/fluxninja/aperture

RUN --mount=type=cache,target=/go/pkg/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    --mount=type=secret,id=ssh-script,dst=/root/go_mod_ssh_config.sh \
    --mount=type=secret,id=aperture-key,dst=/root/.ssh/aperture.pub \
    --mount=type=ssh <<-EOF
    bash /root/go_mod_ssh_config.sh
    go mod download
    # When aperture.tech domain is properly set, this should be tweaked.
    APERTURE_PKG="github.com/fluxninja/aperture"
    APERTURE_VERSION="$(grep ${APERTURE_PKG}.v go.mod | awk '{print $NF}')"
    APERTURE_PATH="/go/pkg/mod/github.com/fluxninja/aperture@${APERTURE_VERSION}"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 TARGET="/aperture-go-example" PREFIX="aperture" SOURCE="./example" LDFLAGS="-s -w" sh "${APERTURE_PATH}/pkg/info/build.sh"
EOF

# Final image
FROM alpine:3.15.0
COPY --link --from=builder /aperture-go-example /aperture-go-example
ENTRYPOINT ["/aperture-go-example"]
