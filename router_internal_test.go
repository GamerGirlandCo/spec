package spec

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec/option"
)

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		path   string
		want   string
	}{
		{name: "empty both", prefix: "", path: "", want: "/"},
		{name: "empty prefix", prefix: "", path: "users", want: "/users"},
		{name: "empty path", prefix: "/api", path: "", want: "/api"},
		{name: "slash path", prefix: "/api/", path: "/", want: "/api/"},
		{name: "normal join", prefix: "/api/", path: "/users", want: "/api/users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, joinPath(tt.prefix, tt.path))
		})
	}
}

func TestEnsureLeadingSlashAndOptional(t *testing.T) {
	assert.Equal(t, "/", ensureLeadingSlash(""))
	assert.Equal(t, "/x", ensureLeadingSlash("x"))
	assert.Equal(t, "/x", ensureLeadingSlash("/x"))

	assert.Equal(t, "fallback", optional("fallback"))
	assert.Equal(t, "value", optional("fallback", "value"))
}

func TestGroupOperationOptions(t *testing.T) {
	t.Run("hidden group", func(t *testing.T) {
		opts, hidden := groupOperationOptions([]option.GroupOption{option.GroupHidden()})
		assert.Nil(t, opts)
		assert.True(t, hidden)
	})

	t.Run("maps group options to op options", func(t *testing.T) {
		opts, hidden := groupOperationOptions([]option.GroupOption{
			option.GroupDeprecated(),
			option.GroupTags("g1", "g2"),
			option.GroupSecurity("apiKey", "read"),
		})
		assert.False(t, hidden)
		assert.Len(t, opts, 3)
	})
}

func TestRoutePathRespectsPrefixForNonWebhook(t *testing.T) {
	r := &route{prefix: "/v1", isWebhook: false, state: &sharedState{mu: sync.Mutex{}}}
	r.Path("users")
	assert.Equal(t, "/v1/users", r.path)

	webhook := &route{prefix: "/v1", isWebhook: true, state: &sharedState{mu: sync.Mutex{}}}
	webhook.Path("user.created")
	assert.Equal(t, "user.created", webhook.path)
}
