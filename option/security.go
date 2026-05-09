package option

import "github.com/oaswrap/spec/openapi"

// SecurityOption mutates a reusable security scheme definition.
type SecurityOption func(*securityConfig)

type securityConfig struct {
	scheme *openapi.SecurityScheme
}

// SecurityDescription sets security scheme description.
func SecurityDescription(description string) SecurityOption {
	return func(cfg *securityConfig) {
		if cfg.scheme == nil {
			cfg.scheme = &openapi.SecurityScheme{}
		}
		cfg.scheme.Description = &description
	}
}

// SecurityOAuth2MetadataURL sets oauth2 metadata discovery URL.
// It is only valid for OpenAPI 3.2.0 and for `oauth2` security schemes.
func SecurityOAuth2MetadataURL(url string) SecurityOption {
	return func(cfg *securityConfig) {
		if cfg.scheme == nil {
			cfg.scheme = &openapi.SecurityScheme{}
		}
		cfg.scheme.OAuth2MetadataURL = url
	}
}

// SecurityDeprecated marks a security scheme deprecated.
// It is only valid for OpenAPI 3.2.0.
func SecurityDeprecated(deprecated ...bool) SecurityOption {
	return func(cfg *securityConfig) {
		if cfg.scheme == nil {
			cfg.scheme = &openapi.SecurityScheme{}
		}
		cfg.scheme.Deprecated = optional(true, deprecated...)
	}
}

// SecurityAPIKey configures an `apiKey` security scheme.
//
// Example:
//
//	option.WithSecurity(
//		"apiKeyAuth",
//		option.SecurityAPIKey("X-API-Key", openapi.SecuritySchemeAPIKeyInHeader),
//	)
func SecurityAPIKey(name string, in openapi.SecuritySchemeAPIKeyIn) SecurityOption {
	return func(cfg *securityConfig) {
		current := currentSecurityScheme(cfg)
		cfg.scheme = &openapi.SecurityScheme{Type: "apiKey", Name: name, In: in}
		applySecurityCommon(cfg.scheme, current)
	}
}

// SecurityHTTPBearer configures an `http` security scheme.
func SecurityHTTPBearer(scheme string, bearerFormat ...string) SecurityOption {
	return func(cfg *securityConfig) {
		current := currentSecurityScheme(cfg)
		cfg.scheme = &openapi.SecurityScheme{Type: "http", Scheme: scheme}
		applySecurityCommon(cfg.scheme, current)
		if len(bearerFormat) > 0 {
			cfg.scheme.BearerFormat = &bearerFormat[0]
		}
	}
}

// SecurityOAuth2 configures an `oauth2` security scheme.
func SecurityOAuth2(flows openapi.OAuthFlows) SecurityOption {
	return func(cfg *securityConfig) {
		current := currentSecurityScheme(cfg)
		cfg.scheme = &openapi.SecurityScheme{Type: "oauth2", Flows: &flows}
		applySecurityCommon(cfg.scheme, current)
	}
}

// SecurityOAuth2Implicit configures an OAuth2 implicit flow.
func SecurityOAuth2Implicit(authorizationURL string, scopes map[string]string, opts ...OAuthFlowOption) SecurityOption {
	return SecurityOAuth2(openapi.OAuthFlows{
		Implicit: newOAuthFlow(openapi.OAuthFlow{AuthorizationURL: authorizationURL, Scopes: scopes}, opts...),
	})
}

// SecurityOAuth2Password configures an OAuth2 password flow.
func SecurityOAuth2Password(tokenURL string, scopes map[string]string, opts ...OAuthFlowOption) SecurityOption {
	return SecurityOAuth2(openapi.OAuthFlows{
		Password: newOAuthFlow(openapi.OAuthFlow{TokenURL: tokenURL, Scopes: scopes}, opts...),
	})
}

// SecurityOAuth2ClientCredentials configures an OAuth2 client credentials flow.
func SecurityOAuth2ClientCredentials(
	tokenURL string,
	scopes map[string]string,
	opts ...OAuthFlowOption,
) SecurityOption {
	return SecurityOAuth2(openapi.OAuthFlows{
		ClientCredentials: newOAuthFlow(openapi.OAuthFlow{TokenURL: tokenURL, Scopes: scopes}, opts...),
	})
}

// SecurityOAuth2AuthorizationCode configures an OAuth2 authorization code flow.
//
// Example:
//
//	option.WithSecurity(
//		"oauth2",
//		option.SecurityOAuth2AuthorizationCode(
//			"https://auth.example.com/oauth/authorize",
//			"https://auth.example.com/oauth/token",
//			map[string]string{"read": "Read access"},
//		),
//	)
func SecurityOAuth2AuthorizationCode(
	authorizationURL string,
	tokenURL string,
	scopes map[string]string,
	opts ...OAuthFlowOption,
) SecurityOption {
	return SecurityOAuth2(openapi.OAuthFlows{
		AuthorizationCode: newOAuthFlow(openapi.OAuthFlow{
			AuthorizationURL: authorizationURL,
			TokenURL:         tokenURL,
			Scopes:           scopes,
		}, opts...),
	})
}

// SecurityOAuth2DeviceAuthorization configures an OAuth2 device authorization flow.
// It is only valid for OpenAPI 3.2.0.
func SecurityOAuth2DeviceAuthorization(
	deviceAuthorizationURL string,
	tokenURL string,
	scopes map[string]string,
	opts ...OAuthFlowOption,
) SecurityOption {
	return SecurityOAuth2(openapi.OAuthFlows{
		DeviceAuthorization: newOAuthFlow(openapi.OAuthFlow{
			DeviceAuthorizationURL: deviceAuthorizationURL,
			TokenURL:               tokenURL,
			Scopes:                 scopes,
		}, opts...),
	})
}

// OAuthFlowOption mutates an OAuth2 flow.
type OAuthFlowOption func(*openapi.OAuthFlow)

// OAuthRefreshURL sets OAuth2 flow refreshUrl.
func OAuthRefreshURL(url string) OAuthFlowOption {
	return func(flow *openapi.OAuthFlow) { flow.RefreshURL = &url }
}

// SecurityMutualTLS configures a `mutualTLS` security scheme.
func SecurityMutualTLS() SecurityOption {
	return func(cfg *securityConfig) {
		current := currentSecurityScheme(cfg)
		cfg.scheme = &openapi.SecurityScheme{Type: "mutualTLS"}
		applySecurityCommon(cfg.scheme, current)
	}
}

func newOAuthFlow(flow openapi.OAuthFlow, opts ...OAuthFlowOption) *openapi.OAuthFlow {
	if flow.Scopes == nil {
		flow.Scopes = map[string]string{}
	}
	for _, opt := range opts {
		opt(&flow)
	}
	return &flow
}

// SecurityOpenIDConnect configures an `openIdConnect` security scheme.
func SecurityOpenIDConnect(url string) SecurityOption {
	return func(cfg *securityConfig) {
		current := currentSecurityScheme(cfg)
		cfg.scheme = &openapi.SecurityScheme{Type: "openIdConnect", OpenIDConnectURL: url}
		applySecurityCommon(cfg.scheme, current)
	}
}

func currentSecurityScheme(cfg *securityConfig) *openapi.SecurityScheme {
	if cfg.scheme == nil {
		return nil
	}
	current := *cfg.scheme
	return &current
}

func applySecurityCommon(scheme, current *openapi.SecurityScheme) {
	if current == nil {
		return
	}
	scheme.Description = current.Description
	scheme.OAuth2MetadataURL = current.OAuth2MetadataURL
	scheme.Deprecated = current.Deprecated
}
