# New Relic Infrastructure Integration for Nagios

Reports the output of service checks 

## Requirements

All desired service checks must be pre-installed and be executable by the root user.

## Installation

* Download an archive file for the `Nagios` Integration
* Extract `nagios-definition.yml` and the `bin` directory into `/var/db/newrelic-infra/newrelic-integrations`
* Add execute permissions for the binary file `nr-nagios` (if required)
* Extract `nagios-config.yml.sample` and `service_check_config.yml.sample` into `/etc/newrelic-infra/integrations.d`

## Usage

To run the Nagios integration, you must have the agent installed (see [agent installation](https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation/install-infrastructure-linux)).

To use the Nagios integration, first rename `nagios-config.yml.sample` to `nagios-config.yml`, then configure the integration
by editing the fields in the file. 

You can view your data in Insights by creating your own NRQL queries. To do so, use the **NagiosServiceCheckSample** event type.

## Compatibility

* Supported OS: No restrictions

## Integration Development usage

Assuming you have the source code, you can build and run the Nagios integration locally

* Go to the directory of the Nagios Integration and build it
```
$ make
```

* The command above will execute tests for the Nagios integration and build an executable file called `nr-nagios` in the `bin` directory
```
$ ./bin/nr-nagios --help
```

For managing external dependencies, the [govendor tool](https://github.com/kardianos/govendor) is used. It is required to lock all external dependencies to a specific version (if possible) in the vendor directory.
