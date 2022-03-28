package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestS3GetAccessToken(t *testing.T) {
	token, err := ReadTokenFromS3()
	if err != nil {
		t.Log("Errored during testing, this is expected during testing in Github Action if AWS credential is not setup")
	} else {
		assert.NotEmpty(t, token)
		t.Log(token)
	}
}
