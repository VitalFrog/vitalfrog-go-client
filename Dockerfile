#
# Builder
#

FROM golang:1.18-alpine AS builder

# Create a workspace for the app
WORKDIR /app

# Copy over the files
COPY . ./

WORKDIR /app/cmd/cli

# Build
RUN go build

RUN ls
#
# Runner
#

FROM alpine AS runner

WORKDIR /

# Copy from builder the final binary
COPY --from=builder /app/cmd/cli/cli /cli

ENTRYPOINT ["/cli"]
