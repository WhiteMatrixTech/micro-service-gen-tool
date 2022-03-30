package pkg

import (
	"testing"
)

func TestGetLatestVersion(t *testing.T) {
	version := GetLatestVersionFromTagName()
	t.Log(version)
}
