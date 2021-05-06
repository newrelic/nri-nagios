ARG GOLANG_VERSION=1.16

FROM golang:${GOLANG_VERSION} as builder-nagios
WORKDIR /code
COPY go.mod .
RUN go mod download

COPY . ./
RUN go build -o ./bin/nri-nagios cmd/nri-nagios/main.go; strip ./bin/nri-nagios

FROM newrelic/infrastructure:latest
ENV NRIA_IS_FORWARD_ONLY true
ENV NRIA_K8S_INTEGRATION true
COPY --from=builder-nagios /code/bin/nri-nagios /nri-sidecar/newrelic-infra/newrelic-integrations/bin/nri-nagios
COPY --from=builder-nagios /code/nagios-definition.yml /nri-sidecar/newrelic-infra/newrelic-integrations/definition.yaml
USER 1000
