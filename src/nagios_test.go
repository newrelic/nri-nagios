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
				false,
			},
		},
	}

	os.Chmod("./test/testconfig.yaml", 0600)
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
		"entityName":           "serviceCheck:testname",
		"event_type":           "NagiosServiceCheckSample",
		"testkey":              "testval",
	}

	collectServiceCheck(sc, i, "NagiosServiceCheckSample")

	id := integration.NewIDAttribute("executing_host", serverName)
	e, _ := i.Entity("testname", "serviceCheck", id)
	metrics := e.Metrics[0].Metrics

	assert.Equal(t, expectedMetrics, metrics)
}

func Test_collectServiceCheck_InvalidNameError(t *testing.T) {
	i, _ := integration.New("test", "test")
	sc := serviceCheck{
		Name:        "",
		Command:     []string{"echo", "testout"},
		Labels:      map[string]string{"testkey": "testval"},
		ParseOutput: false,
	}

	collectServiceCheck(sc, i, "NagiosServiceCheckSample")

	e, _ := i.Entity("testname", "serviceCheck")

	assert.Equal(t, 0, len(e.Metrics))
}

func Test_collectServiceCheck_NoNameError(t *testing.T) {
	i, _ := integration.New("test", "test")
	sc := serviceCheck{
		Name:        "",
		Command:     []string{"echo", "testout"},
		Labels:      map[string]string{"testkey": "testval"},
		ParseOutput: false,
	}

	collectServiceCheck(sc, i, "NagiosServiceCheckSample")

	e, _ := i.Entity("testname", "serviceCheck")

	assert.Equal(t, 0, len(e.Metrics))
}

func Test_collectServiceCheck_InvalidCommandError(t *testing.T) {
	i, _ := integration.New("test", "test")
	sc := serviceCheck{
		Name:        "test",
		Command:     []string{},
		Labels:      map[string]string{"testkey": "testval"},
		ParseOutput: false,
	}

	collectServiceCheck(sc, i, "NagiosServiceCheckSample")

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

func Test_parseOutput1(t *testing.T) {
	case1 := `DISK OK - free space: / 3326 MB (56%);`

	expectedServiceOutput := case1
	expectedLongServiceOutput := ""
	expectedServicePerfData := map[string]float64{}

	serviceOutput, longServiceOutput, servicePerfData := parseOutput(case1)

	assert.Equal(t, expectedServiceOutput, serviceOutput)
	assert.Equal(t, expectedLongServiceOutput, longServiceOutput)
	assert.Equal(t, expectedServicePerfData, servicePerfData)

}

func Test_parseOutput2(t *testing.T) {
	case2 := `DISK OK - free space: /root 3326 MB (56%); | /root=2643MB;5948;5958;0;5968`

	expectedServiceOutput := "DISK OK - free space: /root 3326 MB (56%); "
	expectedLongServiceOutput := ""
	expectedServicePerfData := map[string]float64{
		"/root": 2643.0,
	}

	serviceOutput, longServiceOutput, servicePerfData := parseOutput(case2)

	assert.Equal(t, expectedServiceOutput, serviceOutput)
	assert.Equal(t, expectedLongServiceOutput, longServiceOutput)
	assert.Equal(t, expectedServicePerfData, servicePerfData)

}

func Test_parseOutput3(t *testing.T) {
	case3 := "DISK OK - free space: /root 3326 MB (56%); | /=2643MB;5948;5958;0;5968\n/ 15272 MB (77%);\n/boot 68 MB (69%); | /boot=68MB;88;93;0;98\n/home=69357MB;253404;253409;0;253414"

	expectedServiceOutput := "DISK OK - free space: /root 3326 MB (56%); "
	expectedLongServiceOutput := "/ 15272 MB (77%);\n/boot 68 MB (69%); "
	expectedServicePerfData := map[string]float64{
		"/":     2643.0,
		"/boot": 68.0,
		"/home": 69357.0,
	}

	serviceOutput, longServiceOutput, servicePerfData := parseOutput(case3)

	assert.Equal(t, expectedServiceOutput, serviceOutput)
	assert.Equal(t, expectedLongServiceOutput, longServiceOutput)
	assert.Equal(t, expectedServicePerfData, servicePerfData)

}

func Test_parseOutput4(t *testing.T) {
	case4 := `DISK OK - free space: /root 3326 MB (56%); | /root=2643MB test2=3452.0`

	expectedServiceOutput := "DISK OK - free space: /root 3326 MB (56%); "
	expectedLongServiceOutput := ""
	expectedServicePerfData := map[string]float64{
		"/root": 2643.0,
		"test2": 3452.0,
	}

	serviceOutput, longServiceOutput, servicePerfData := parseOutput(case4)

	assert.Equal(t, expectedServiceOutput, serviceOutput)
	assert.Equal(t, expectedLongServiceOutput, longServiceOutput)
	assert.Equal(t, expectedServicePerfData, servicePerfData)

}
