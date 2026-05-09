module github.com/oaswrap/spec/adapter/echov5openapi/example

go 1.25.0

require (
	github.com/labstack/echo/v5 v5.1.0
	github.com/oaswrap/spec v0.4.2
	github.com/oaswrap/spec/adapter/echov5openapi v0.0.0
)

require (
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/oaswrap/spec-ui v0.2.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace github.com/oaswrap/spec => ../../..

replace github.com/oaswrap/spec/adapter/echov5openapi => ..
