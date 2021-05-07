// +build integration

package tests

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/newrelic/nri-nagios/tests/jsonschema"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

const (
	containerName = "nri-nagios"

)

func executeDockerCompose(containerName string, envVars []string) (string, string, error) {
	cmdLine := []string{"run"}
	for i := range envVars {
		cmdLine = append(cmdLine, "-e")
		cmdLine = append(cmdLine, envVars[i])
	}
	cmdLine = append(cmdLine, containerName)
	fmt.Printf("executing: docker-compose %s\n", strings.Join(cmdLine, " "))
	cmd := exec.Command("docker-compose", cmdLine...)
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	stdout := outbuf.String()
	stderr := errbuf.String()
	return stdout, stderr, err
}

func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}

func TestSuccessConnection(t *testing.T) {
	envVars := []string{
		"SERVICE_CHECKS_CONFIG=/code/tests/testdata/testconfig.yaml",
	}
	stdout, _, err := executeDockerCompose(containerName, envVars)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	response := string(stdout)
	schemaURI := filepath.Join("testdata","nagios-schema.json")
	err=jsonschema.Validate(schemaURI, response)
	assert.Nil(t, err)
	assert.Equal(t, "com.newrelic.nagios", gjson.Get(response, "name").String())
	assert.Equal(t, "3", gjson.Get(response, "protocol_version").String())
}
