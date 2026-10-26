# This is a multi-stage Dockerfile. The first stage contains the whole Go
# toolchain and is only used to compile the application. The second stage
# starts from a small, empty image and only receives the finished binary.
# The final image is therefore much smaller and contains no source code
# or build tools.

# ---- Stage 1: builder ----
# The Go image provides everything needed to compile the application.
# "AS builder" names this stage so the next stage can copy files from it.
FROM golang:1.27 AS builder

# All following commands run inside /app.
WORKDIR /app

# Copy the complete project (source code, frontend, docs, scripts) into the image.
COPY . /app

# Download the dependencies and compile the binary to /app/ordersystem.
# The script builds with CGO_ENABLED=0, so the binary has no dependency on
# system C libraries and also runs on Alpine.
RUN sh /app/scripts/build-application.sh

# ---- Stage 2: run ----
# Alpine is a very small Linux image, which is enough to run a static binary.
FROM alpine

WORKDIR /app

# Only the compiled binary is taken over from the builder stage. The frontend
# and the OpenAPI docs are embedded in the binary, so nothing else is needed.
COPY --from=builder /app/ordersystem /app/ordersystem

# Documents that the server listens on port 3000.
EXPOSE 3000

# The command that is executed when the container starts.
CMD ["/app/ordersystem"]
