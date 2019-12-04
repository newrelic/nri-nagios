//go:generate goversioninfo
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
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
	integrationVersion = "2.4.1"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	ServiceChecksConfig string
}

type serviceCheckConfig struct {
	ServiceChecks []serviceCheck `yaml:"service_checks"`
}

type serviceCheck struct {
	Name        string            `yaml:"name"`
	Command     []string          `yaml:"command"`
	Labels      map[string]string `yaml:"labels"`
	ParseOutput bool              `yaml:"parse_output"`
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
			if fileInfo.Mode().Perm() > 0640 {
				return nil, fmt.Errorf("service checks file permissions are not restrictive enough. File permissions must be more strict than 0640. See documentation for details")
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

	// Set the hostname
	serverName, err := os.Hostname()
	if err != nil {
		log.Error("Failed to collect the hostname. Setting it to localhost to be set by the agent")
		serverName = "localhost"
	}

	// Create the entity
	hostIDAttr := integration.NewIDAttribute("executing_host", serverName)
	e, err := i.Entity(sc.Name, "serviceCheck", hostIDAttr)
	if err != nil {
		log.Error("Must provide a name for each service check: %s", err.Error())
		return
	}

	// Run the command for the service check
	stdout, stderr, exit := runCommand(sc.Command[0], sc.Command[1:]...)

	// Create a metric set
	ms := e.NewMetricSet("NagiosServiceCheckSample",
		metric.Attribute{Key: "serverName", Value: serverName},
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

	if sc.ParseOutput {
		serviceOutput, longServiceOutput, parsedMetrics := parseOutput(stdout)
		for key, value := range parsedMetrics {
			if err := ms.SetMetric(key, value, metric.GAUGE); err != nil {
				log.Error("Failed to set metric %s for %s: %s", key, value, err.Error())
			}
		}

		if err := ms.SetMetric("serviceCheck.serviceOutput", serviceOutput, metric.ATTRIBUTE); err != nil {
			log.Error("Failed to set metric %s for %s: %s", "serviceCheck.serviceOutput", sc.Name, err.Error())
		}

		if err := ms.SetMetric("serviceCheck.longServiceOutput", longServiceOutput, metric.ATTRIBUTE); err != nil {
			log.Error("Failed to set metric %s for %s: %s", "serviceCheck.longServiceOutput", sc.Name, err.Error())
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
			log.Error("Failed to execute script `%s`: %s", name, err.Error())
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

func parseOutput(output string) (string, string, map[string]float64) {
	re := regexp.MustCompile(`^(?P<serviceOutput>[^|]+)(?:\|(?P<metrics1>[^\n]*)\n?)?(?P<longServiceOutput>[^|]*)?(?:\|(?P<metrics2>[^|]*))?$`)
	match := re.FindStringSubmatch(output)
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	rawMetrics := result["metrics1"] + "\n" + result["metrics2"]
	parsedMetrics := parseMetrics(rawMetrics)

	return result["serviceOutput"], result["longServiceOutput"], parsedMetrics
}

func parseMetrics(rawMetrics string) map[string]float64 {
	re := regexp.MustCompile(`(?P<key>[^\s;,]+)=(?P<val>[\d\.]+)`)
	matches := re.FindAllStringSubmatch(rawMetrics, -1)
	results := map[string]float64{}
	for _, match := range matches {
		value, _ := strconv.ParseFloat(match[2], 64)
		key := match[1]
		results[key] = value
	}

	return results
}
