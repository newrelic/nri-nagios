package main

import (
	"bytes"
	"context"
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

var (
	args ArgumentList
)

type ArgumentList struct {
	sdkArgs.DefaultArgumentList
}

type ServiceCheckConfig struct {
	ServiceChecks []struct {
		Name    string   `yaml:"name"`
		Command []string `yaml:"command"`
	} `yaml:"service_checks"`
}

func main() {
	// Create Integration
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error("Failed to create integration: %s", err.Error())
	}

	log.SetupLogging(args.Verbose)

	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		log.Error("Failed to read yaml file")
		os.Exit(1)
	}

	var conf ServiceCheckConfig
	err = yaml.Unmarshal(yamlFile, &conf)

	for _, sc := range conf.ServiceChecks {
		e, err := i.Entity(sc.Name, "serviceCheck")
		if err != nil {
			log.Error("Failed to get entity for service check %s: %s", sc.Name, err.Error())
		}
		out, exit, err := run_command(sc.Command)
		if err != nil {
			log.Error("Failed to run command %s: %s", sc.Name, err.Error())
		}

		ms := e.NewMetricSet("NagiosServiceCheckSample",
			metric.Attribute{Key: "displayName", Value: sc.Name},
			metric.Attribute{Key: "entityName", Value: "serviceCheck:" + sc.Name},
		)

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
				out,
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

	// Create Entity
	if err := i.Publish(); err != nil {
		log.Error("Failed to publish integration: %s", err.Error())
		os.Exit(1)
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
