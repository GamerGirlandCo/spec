package validate

import (
	"net/http"
	"strings"

	"github.com/oaswrap/spec/openapi"
)

func AllowsOperationMethod(version, method string) bool {
	if strings.EqualFold(method, http.MethodConnect) {
		return version == openapi.Version320
	}

	return true
}
