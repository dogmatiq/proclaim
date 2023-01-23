package dnsimplex

import (
	"errors"
	"net/http"

	"github.com/dnsimple/dnsimple-go/dnsimple"
)

// IsNotFound returns true if err is an error response from dnsimple.com that
// indicates that the requested resource does not exist.
func IsNotFound(err error) bool {
	var res *dnsimple.ErrorResponse

	if errors.As(err, &res) {
		return res.HTTPResponse.StatusCode == http.StatusNotFound
	}

	return false
}

// IgnoreNotFound returns nil if err is a non-found error, otherwise it returns
// err unchanged.
func IgnoreNotFound(err error) error {
	if IsNotFound(err) {
		return nil
	}

	return err
}
