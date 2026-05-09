package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type GetUserRequest struct {
	ID string `path:"id" required:"true" description:"User identifier"`
}

type User struct {
	ID   string `json:"id" required:"true"`
	Name string `json:"name"`
}

func assertValidationContains(t *testing.T, err error, messages ...string) {
	t.Helper()
	require.Error(t, err)
	for _, message := range messages {
		assert.Contains(t, err.Error(), message)
	}
}
