package pkg

import (
	"testing"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"test0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadTokenFromS3()
			if err != nil {
				t.Log("Errored during testing, this is expected during testing in Github Action if AWS credential is not setup")
			} else {
				Install()
			}
		})
	}
}
