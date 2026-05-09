package echov5openapi

import (
	"github.com/labstack/echo/v5"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/option"
)

type route struct {
	echoRoute echo.RouteInfo
	specRoute spec.Route
}

var _ Route = (*route)(nil)

func (r *route) Method() string {
	return r.echoRoute.Method
}

func (r *route) Path() string {
	return r.echoRoute.Path
}

func (r *route) Name() string {
	return r.echoRoute.Name
}

func (r *route) With(opts ...option.OperationOption) Route {
	if r.specRoute == nil {
		return r
	}
	r.specRoute.With(opts...)
	return r
}
