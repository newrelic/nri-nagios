FROM golang:1.10 as builder
RUN go get -d github.com/newrelic/nri-nagios/... && \
    cd /go/src/github.com/newrelic/nri-nagios && \
    make && \
    strip ./bin/nr-nagios

FROM newrelic/infrastructure:latest
ENV NRIA_IS_FORWARD_ONLY true
ENV NRIA_K8S_INTEGRATION true
COPY --from=builder /go/src/github.com/newrelic/nri-nagios/bin/nr-nagios /nri-sidecar/newrelic-infra/newrelic-integrations/bin/nr-nagios
COPY --from=builder /go/src/github.com/newrelic/nri-nagios/nagios-definition.yml /nri-sidecar/newrelic-infra/newrelic-integrations/definition.yml
USER 1000
