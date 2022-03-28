package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestS3GetAccessToken(t *testing.T) {
	token, _ := ReadTokenFromS3()
	assert.NotEmpty(t, token)
	t.Log(token)
}
