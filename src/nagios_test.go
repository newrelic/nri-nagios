package main

import (
	"os"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/stretchr/testify/assert"
)

func Test_parseConfigFile(t *testing.T) {
	expectedConfig := &serviceCheckConfig{
		[]serviceCheck{
			{
				"test",
				[]string{"test", "args"},
				map[string]string{"env": "testenv"},
			},
		},
	}

	config, err := parseConfigFile("./test/testconfig.yaml")

	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, config)
}

func Test_parseConfigFile_FileNotExistError(t *testing.T) {
	config, err := parseConfigFile("./test/nonexist.yaml")

	assert.Error(t, err)
	assert.Nil(t, config)
}

func Test_parseConfigFile_InvalidYamlError(t *testing.T) {
	config, err := parseConfigFile("./test/invalid.yaml")
	os.Chmod("./test/invalid.yaml", 0600)

	assert.Error(t, err)
	assert.Nil(t, config)
}

func Test_parseConfigFile_UnrestrictiveError(t *testing.T) {
	config, err := parseConfigFile("./test/invalid.yaml")
	os.Chmod("./test/invalid.yaml", 0777)

	assert.Error(t, err)
	assert.Nil(t, config)
}

func Test_collectServiceCheck(t *testing.T) {
	i, _ := integration.New("test", "test")
	sc := serviceCheck{
		Name:    "testname",
		Command: []string{"echo", "testout"},
		Labels:  map[string]string{"testkey": "testval"},
	}

	expectedMetrics := map[string]interface{}{
		"serviceCheck.name":    "testname",
		"serviceCheck.status":  float64(0),
		"serviceCheck.message": "testout\n",
		"serviceCheck.error":   "",
		"serviceCheck.command": "echo testout",
		"displayName":          "testname",
		"entityName":           "serviceCheck:testname",
		"event_type":           "NagiosServiceCheckSample",
		"testkey":              "testval",
	}

	collectServiceCheck(sc, i)

	e, _ := i.Entity("testname", "serviceCheck")
	metrics := e.Metrics[0].Metrics

	assert.Equal(t, expectedMetrics, metrics)
}

func Test_collectServiceCheck_InvalidNameError(t *testing.T) {
	i, _ := integration.New("test", "test")
	sc := serviceCheck{
		Name:    "",
		Command: []string{"echo", "testout"},
		Labels:  map[string]string{"testkey": "testval"},
	}

	collectServiceCheck(sc, i)

	e, _ := i.Entity("testname", "serviceCheck")

	assert.Equal(t, 0, len(e.Metrics))
}

func Test_collectServiceCheck_NoNameError(t *testing.T) {
	i, _ := integration.New("test", "test")
	sc := serviceCheck{
		Name:    "",
		Command: []string{"echo", "testout"},
		Labels:  map[string]string{"testkey": "testval"},
	}

	collectServiceCheck(sc, i)

	e, _ := i.Entity("testname", "serviceCheck")

	assert.Equal(t, 0, len(e.Metrics))
}

func Test_collectServiceCheck_InvalidCommandError(t *testing.T) {
	i, _ := integration.New("test", "test")
	sc := serviceCheck{
		Name:    "test",
		Command: []string{},
		Labels:  map[string]string{"testkey": "testval"},
	}

	collectServiceCheck(sc, i)

	e, _ := i.Entity("testname", "serviceCheck")

	assert.Equal(t, 0, len(e.Metrics))
}

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
