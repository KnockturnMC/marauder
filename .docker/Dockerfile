# Build image
# We are using alpine, AS alpine has the dependencies needed to do a CGO_ENABLED=0 build, allowing us to run from scratch.
FROM golang:1.24 AS builder

# The version of goreleaser to download
ARG GORELEASER_VERSION="v2.9.0"

# Install bash to run build scripts
# Install git to fetch version
# Install ca-certificats for root ssl cert which is needed in scratch image.
# Install wget to download goreleaser
RUN apt install -y curl bash git ca-certificates wget
RUN wget "https://github.com/goreleaser/goreleaser/releases/download/$GORELEASER_VERSION/goreleaser_Linux_x86_64.tar.gz" -O goreleaser.tar.gz \
    && tar -xzvf goreleaser.tar.gz goreleaser \
    && cp goreleaser /usr/bin/goreleaser

# Create the runner user in the build image, we copy the created files later.
# The user does not need a password or a shell, it is simply there to restrict permissions.
ENV USER=runner
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" "${USER}"

# Copy over project source files
WORKDIR /usr/src/marauder

COPY go.mod go.sum .

# Download dependencies without version dependant code.
RUN go get -u ./...

# Copy entire rest
COPY . .

# Build binaries
RUN goreleaser build --clean --snapshot --id "marauder-controller" --id "marauder-operator" --id "marauder-client-static"

# The running image, from scratch
FROM scratch AS base_runner

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy over OS provided certs.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# The controller image
FROM base_runner AS controller

# Copy binary
COPY --from=builder /usr/src/marauder/dist/marauder-controller_linux_amd64_v1/marauderctl /usr/bin/marauderctl

# Run the binary.
ENTRYPOINT ["/usr/bin/marauderctl"]

# The controller image
FROM base_runner AS operator

# Copy binary
COPY --from=builder /usr/src/marauder/dist/marauder-operator_linux_amd64_v1/marauderop /usr/bin/marauderop

# Run the binary.
ENTRYPOINT ["/usr/bin/marauderop"]

# The client runner image with marauder client installed.
FROM alpine:3.21 AS client

# Copy the statically compiled binary for alpine
COPY --from=builder /usr/src/marauder/dist/marauder-client-static_linux_amd64_v1/marauder-static /usr/bin/marauder
