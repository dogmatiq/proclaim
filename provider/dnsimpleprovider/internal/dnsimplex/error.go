package dnsimplex

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
)

// IsNotFound returns true if err is an error response from dnsimple.com that
// indicates that the requested resource does not exist.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

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

// Errorf returns an error that formats according to a format specifier.
func Errorf(format string, args ...any) error {
	for i, arg := range args {
		if err, ok := arg.(error); ok {
			args[i] = enrichErrorMessage(err)
		}
	}

	return fmt.Errorf(format, args...)
}

// enrichErrorMessage extracts more detailed information from an
// [dnsimple.ErrorResponse] and returns a new error that includes the original
// error message and detailed information about invalid attributes.
func enrichErrorMessage(err error) error {
	var res *dnsimple.ErrorResponse
	if !errors.As(err, &res) {
		return err
	}

	var m strings.Builder

	first := true
	for name, errors := range res.AttributeErrors {
		if first {
			m.WriteString(": ")
			first = false
		} else {
			m.WriteString(", ")
		}

		m.WriteString(name)
		m.WriteString("=")
		m.WriteString(strings.Join(errors, ","))
	}

	return fmt.Errorf("%w: %s", err, m.String())
}
