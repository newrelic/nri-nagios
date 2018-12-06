package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"gopkg.in/yaml.v2"
)

const (
	integrationName    = "com.newrelic.nagios"
	integrationVersion = "0.1.0"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	ServiceChecksConfig string
}

type serviceCheckConfig struct {
	ServiceChecks []serviceCheck `yaml:"service_checks"`
}

type serviceCheck struct {
	Name    string            `yaml:"name"`
	Command []string          `yaml:"command"`
	Labels  map[string]string `yaml:"labels"`
}

func main() {
	var args argumentList

	// Create Integration
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error("Failed to create integration: %s", err.Error())
		os.Exit(1)
	}

	// Set logging verbosity
	log.SetupLogging(args.Verbose)

	// Read the service checks definitions file
	conf, err := parseConfigFile(args.ServiceChecksConfig)
	if err != nil {
		log.Error("Config parsing failed: %s", err.Error())
		os.Exit(1)
	}

	// Run the service checks and store their result
	if args.HasMetrics() {
		for _, sc := range conf.ServiceChecks {
			collectServiceCheck(sc, i)
		}
	}

	// Publish the results
	if err := i.Publish(); err != nil {
		log.Error("Failed to publish integration: %s", err.Error())
		os.Exit(1)
	}
}

func parseConfigFile(configFile string) (*serviceCheckConfig, error) {
	// If on linux or macos, check that the service file is appropriately permissioned
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if fileInfo, _ := os.Stat(configFile); fileInfo != nil {
			if fileInfo.Mode().Perm() > 0600 {
				return nil, fmt.Errorf("service checks file permissions are not restrictive enough. Required file permissions are 0600. See documentation for details")
			}
		}
	}

	// Read the file into a string
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err.Error())
	}

	// Parse the file into a serviceCheckConfig struct
	var conf serviceCheckConfig
	err = yaml.UnmarshalStrict(yamlFile, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %s", err.Error())
	}

	return &conf, nil
}

func collectServiceCheck(sc serviceCheck, i *integration.Integration) {
	if len(sc.Command) == 0 {
		log.Error("All service checks require a command")
		return
	}

	// Create the entity
	e, err := i.Entity(sc.Name, "serviceCheck")
	if err != nil {
		log.Error("Must provide a name for each service check: %s", err.Error())
		return
	}

	// Run the command for the service check
	stdout, stderr, exit := runCommand(sc.Command[0], sc.Command[1:]...)

	// Create a metric set
	ms := e.NewMetricSet("NagiosServiceCheckSample",
		metric.Attribute{Key: "displayName", Value: sc.Name},
		metric.Attribute{Key: "entityName", Value: "serviceCheck:" + sc.Name},
	)

	// Add user-defined labels to the metric set
	for key, value := range sc.Labels {
		err := ms.SetMetric(key, value, metric.ATTRIBUTE)
		if err != nil {
			log.Error("Failed to create label %s: %s", key, err.Error())
		}
	}

	// Add each metric to the metric set
	for _, metric := range []struct {
		MetricName  string
		MetricValue interface{}
		MetricType  metric.SourceType
	}{
		{
			"serviceCheck.name",
			sc.Name,
			metric.ATTRIBUTE,
		},
		{
			"serviceCheck.status",
			exit,
			metric.GAUGE,
		},
		{
			"serviceCheck.message",
			stdout,
			metric.ATTRIBUTE,
		},
		{
			"serviceCheck.error",
			stderr,
			metric.ATTRIBUTE,
		},
		{
			"serviceCheck.command",
			strings.Join(sc.Command, " "),
			metric.ATTRIBUTE,
		},
	} {
		if err := ms.SetMetric(metric.MetricName, metric.MetricValue, metric.MetricType); err != nil {
			log.Error("Failed to set metric %s for %s: %s", metric.MetricName, sc.Name, err.Error())
		}
	}
}

func runCommand(name string, args ...string) (stdout string, stderr string, exitCode int) {
	// Create the command and buffers to save the stdout and stderr
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	// Run the command
	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	// Retrieve the exit code
	if err != nil {
		// Try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			exitCode = -1
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exit code should be zero
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	return
}
