FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY gke-policy /gke-policy

USER nonroot:nonroot

ENTRYPOINT ["/gke-policy"]