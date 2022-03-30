package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLatestVersion(t *testing.T) {
	_, err := ReadTokenFromS3()
	if err != nil {
		t.Log("Errored during testing, this is expected during testing in Github Action if AWS credential is not setup")
	} else {
		version := GetLatestVersionFromTagName()
		assert.NotEmpty(t, version)
		t.Log(version)
	}
}
