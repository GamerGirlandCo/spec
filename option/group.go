package option

// GroupConfig stores effective group-level settings.
type GroupConfig struct {
	Hide       bool
	Deprecated bool
	Tags       []string
	Security   []OperationSecurityConfig
}

// GroupOption mutates route-group behavior.
type GroupOption func(*GroupConfig)

// GroupHidden skips emitting all operations within the group scope.
func GroupHidden(hide ...bool) GroupOption {
	return func(cfg *GroupConfig) { cfg.Hide = optional(true, hide...) }
}

// GroupDeprecated marks all operations in the group deprecated.
func GroupDeprecated(deprecated ...bool) GroupOption {
	return func(cfg *GroupConfig) { cfg.Deprecated = optional(true, deprecated...) }
}

// GroupTags appends tags to all operations in the group.
func GroupTags(tags ...string) GroupOption {
	return func(cfg *GroupConfig) { cfg.Tags = append(cfg.Tags, tags...) }
}

// GroupSecurity appends one security requirement to all operations in the group.
func GroupSecurity(name string, scopes ...string) GroupOption {
	return func(cfg *GroupConfig) {
		cfg.Security = append(cfg.Security, OperationSecurityConfig{Name: name, Scopes: scopes})
	}
}
