FROM golang:1.17-bullseye AS build

WORKDIR /app

COPY internal/ ./internal
COPY GNUmakefile *.go go.* ./
RUN go mod download
RUN make

FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /app/gke-policy /gke-policy

USER nonroot:nonroot

ENTRYPOINT ["/gke-policy"]