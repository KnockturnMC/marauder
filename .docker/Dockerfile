# Build image
# We are using alpine, as alpine has the dependencies needed to do a CGO_ENABLED=0 build, allowing us to run from scratch.
FROM golang as builder

# Install bash to run build scripts
# Install git to fetch version
# Install ca-certificats for root ssl cert which is needed in scratch image.
RUN apt install -y curl bash git ca-certificates
RUN echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | tee /etc/apt/sources.list.d/goreleaser.list && \
    apt update && \
    apt install -y goreleaser

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
COPY . .

# Build binaries
RUN go work sync && goreleaser build --clean --snapshot --id "marauder-controller" --id "marauder-operator"

# The running image, from scratch
FROM scratch as base_runner

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy over OS provided certs.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# The controller image
FROM base_runner as controller

# Copy binary
COPY --from=builder /usr/src/marauder/dist/marauder-controller_linux_amd64_v1/marauderctl /usr/bin/marauderctl

# Run the binary.
ENTRYPOINT ["/usr/bin/marauderctl"]

# The controller image
FROM base_runner as operator

# Copy binary
COPY --from=builder /usr/src/marauder/dist/marauder-operator_linux_amd64_v1/marauderop /usr/bin/marauderop

# Run the binary.
ENTRYPOINT ["/usr/bin/marauderop"]

