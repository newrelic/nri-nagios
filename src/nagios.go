package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/newrelic/infra-integrations-sdk/data/metric"

	"gopkg.in/yaml.v2"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
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
	}

	log.SetupLogging(args.Verbose)

	conf, err := parseConfigFile(args.ServiceChecksConfig)
	if err != nil {
		log.Error("Config parsing failed: %s", err.Error())
		os.Exit(1)
	}

	if args.HasMetrics() {
		for _, sc := range conf.ServiceChecks {
			collectServiceCheck(sc, i)
		}
	}

	// Create Entity
	if err := i.Publish(); err != nil {
		log.Error("Failed to publish integration: %s", err.Error())
		os.Exit(1)
	}
}

func parseConfigFile(configFile string) (*serviceCheckConfig, error) {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err.Error())
	}

	var conf serviceCheckConfig
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %s", err.Error())
	}

	return &conf, nil
}

func collectServiceCheck(sc serviceCheck, i *integration.Integration) {
	e, err := i.Entity(sc.Name, "serviceCheck")
	if err != nil {
		log.Error("Failed to get entity for service check %s: %s", sc.Name, err.Error())
	}
	stdout, stderr, exit := runCommand(sc.Command[0], sc.Command[1:]...)
	if err != nil {
		log.Error("Failed to run command %s: %s", sc.Name, err.Error())
	}

	ms := e.NewMetricSet("NagiosServiceCheckSample",
		metric.Attribute{Key: "displayName", Value: sc.Name},
		metric.Attribute{Key: "entityName", Value: "serviceCheck:" + sc.Name},
	)

	for key, value := range sc.Labels {
		err := ms.SetMetric(key, value, metric.ATTRIBUTE)
		if err != nil {
			log.Error("Failed to create label %s: %s", key, err.Error())
		}
	}

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
		err := ms.SetMetric(metric.MetricName, metric.MetricValue, metric.MetricType)
		if err != nil {
			log.Error("Failed to set metric %s for %s: %s", metric.MetricName, sc.Name, err.Error())
		}
	}
}

func run_command(args []string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Start()
	if err != nil {
		return "", 0, err
	}

	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
	}

	return out.String(), exitCode, nil
}

func runCommand(name string, args ...string) (stdout string, stderr string, exitCode int) {
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be gotten, and stderr will likely
			// be an empty string, so we use the default fail code, and format err
			// to string and set to stderr
			exitCode = -1
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	return
}
