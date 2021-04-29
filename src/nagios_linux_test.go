// +build linux

package main

import (
	"os"
	"sync"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/stretchr/testify/assert"
)

func Test_runCommand_InvalidCommandError(t *testing.T) {
	stdout, stderr, exit := runCommand("jdijfs")
	assert.Equal(t, -1, exit)
	assert.Equal(t, "", stdout)
	assert.NotEmpty(t, stderr)
}

func Test_runCommand_returns1(t *testing.T) {
	stdout, stderr, exit := runCommand("/bin/sh", "test/returns2.sh")
	assert.Equal(t, 2, exit)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
}

func Test_collectServiceCheck(t *testing.T) {
	i, _ := integration.New("test", "test")
	sc := serviceCheck{
		Name:        "testname",
		Command:     []string{"echo", "testout"},
		Labels:      map[string]string{"testkey": "testval"},
		ParseOutput: false,
	}
	serverName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	expectedMetrics := map[string]interface{}{
		"serviceCheck.name":    "testname",
		"serviceCheck.status":  float64(0),
		"serviceCheck.message": "testout\n",
		"serviceCheck.error":   "",
		"serviceCheck.command": "echo testout",
		"serverName":           serverName,
		"displayName":          "testname",
		"entityName":           serverName,
		"event_type":           "NagiosServiceCheckSample",
		"testkey":              "testval",
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go collectServiceCheck(sc, i, &wg, "NagiosServiceCheckSample")
	wg.Wait()

	id := integration.NewIDAttribute("executing_host", serverName)
	e, _ := i.Entity("testname", "serviceCheck", id)
	metrics := e.Metrics[0].Metrics

	assert.Equal(t, expectedMetrics, metrics)
}
