// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: infoblox-training-task-3/portal/pkg/pb/service.proto

package pb

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
)

// Validate checks the field values on VersionResponse with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *VersionResponse) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Version

	return nil
}

// VersionResponseValidationError is the validation error returned by
// VersionResponse.Validate if the designated constraints aren't met.
type VersionResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e VersionResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e VersionResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e VersionResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e VersionResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e VersionResponseValidationError) ErrorName() string { return "VersionResponseValidationError" }

// Error satisfies the builtin error interface
func (e VersionResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVersionResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = VersionResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = VersionResponseValidationError{}
