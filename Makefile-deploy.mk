deploy: deps-only
	@GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME) ./src || echo "Failed to compile"
	@ssh cb 'systemctl stop newrelic-infra'
	@scp bin/nr-memcached cb:/var/db/newrelic-infra/custom-integrations/bin || echo "Failed to copy binary"
	@scp memcached-config.yml cb:/etc/newrelic-infra/integrations.d/ || echo "Failed to copy config"
	@scp memcached-definition.yml cb:/var/db/newrelic-infra/custom-integrations/ || echo "Failed to copy definition"
	@ssh cb 'systemctl start newrelic-infra'
.PHONY: deploy

