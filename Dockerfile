FROM golang:1.10 as builder
COPY . /go/src/github.com/newrelic/nri-nagios/
RUN cd /go/src/github.com/newrelic/nri-nagios && \
    make && \
    strip ./bin/nri-nagios

FROM newrelic/infrastructure:latest
ENV NRIA_IS_FORWARD_ONLY true
ENV NRIA_K8S_INTEGRATION true
COPY --from=builder /go/src/github.com/newrelic/nri-nagios/bin/nri-nagios /nri-sidecar/newrelic-infra/newrelic-integrations/bin/nri-nagios
COPY --from=builder /go/src/github.com/newrelic/nri-nagios/nagios-definition.yml /nri-sidecar/newrelic-infra/newrelic-integrations/definition.yml
USER 1000
