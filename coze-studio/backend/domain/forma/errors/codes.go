/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
)

const (
	CodeOK = 0

	CodeUnauthenticated     int32 = 40101 // HTTP 401
	CodeTenantRequired      int32 = 40001 // HTTP 400
	CodeTenantNotFound      int32 = 40401 // HTTP 404
	CodeTenantForbidden     int32 = 40301 // HTTP 403
	CodeTenantSuspended     int32 = 40302 // HTTP 403
	CodeMembershipRequired  int32 = 40303 // HTTP 403
	CodeMembershipForbidden int32 = 40304 // HTTP 403
	CodeSpaceNotFound       int32 = 40402 // HTTP 404
	CodeSpaceForbidden      int32 = 40305 // HTTP 403
	CodeVersionConflict     int32 = 40901 // HTTP 409
	CodeInternal            int32 = 50001 // HTTP 500
)

const (
	KeyUnauthenticated     = "FORMA_UNAUTHENTICATED"
	KeyTenantRequired      = "FORMA_TENANT_REQUIRED"
	KeyTenantNotFound      = "FORMA_TENANT_NOT_FOUND"
	KeyTenantForbidden     = "FORMA_TENANT_FORBIDDEN"
	KeyTenantSuspended     = "FORMA_TENANT_SUSPENDED"
	KeyMembershipRequired  = "FORMA_MEMBERSHIP_REQUIRED"
	KeyMembershipForbidden = "FORMA_MEMBERSHIP_FORBIDDEN"
	KeySpaceNotFound       = "FORMA_SPACE_NOT_FOUND"
	KeySpaceForbidden      = "FORMA_SPACE_FORBIDDEN"
	KeyVersionConflict     = "FORMA_VERSION_CONFLICT"
	KeyInternal            = "FORMA_INTERNAL"
)

// FormaError is the typed API/domain error for Forma endpoints.
type FormaError struct {
	Code       int32
	HTTPStatus int
	Msg        string
	Key        string
}

func (e *FormaError) Error() string {
	if e == nil {
		return ""
	}
	if e.Key != "" {
		return fmt.Sprintf("%s: %s", e.Key, e.Msg)
	}
	return e.Msg
}

func New(code int32, httpStatus int, key, msg string) *FormaError {
	return &FormaError{Code: code, HTTPStatus: httpStatus, Key: key, Msg: msg}
}

func Unauthenticated(msg string) *FormaError {
	if msg == "" {
		msg = "authentication required"
	}
	return New(CodeUnauthenticated, http.StatusUnauthorized, KeyUnauthenticated, msg)
}

func TenantRequired(msg string) *FormaError {
	if msg == "" {
		msg = "tenant selection required"
	}
	return New(CodeTenantRequired, http.StatusBadRequest, KeyTenantRequired, msg)
}

func TenantNotFound(msg string) *FormaError {
	if msg == "" {
		msg = "tenant not found"
	}
	return New(CodeTenantNotFound, http.StatusNotFound, KeyTenantNotFound, msg)
}

func TenantForbidden(msg string) *FormaError {
	if msg == "" {
		msg = "tenant access forbidden"
	}
	return New(CodeTenantForbidden, http.StatusForbidden, KeyTenantForbidden, msg)
}

func TenantSuspended(msg string) *FormaError {
	if msg == "" {
		msg = "tenant is suspended"
	}
	return New(CodeTenantSuspended, http.StatusForbidden, KeyTenantSuspended, msg)
}

func MembershipRequired(msg string) *FormaError {
	if msg == "" {
		msg = "tenant membership required"
	}
	return New(CodeMembershipRequired, http.StatusForbidden, KeyMembershipRequired, msg)
}

func MembershipForbidden(msg string) *FormaError {
	if msg == "" {
		msg = "insufficient membership role"
	}
	return New(CodeMembershipForbidden, http.StatusForbidden, KeyMembershipForbidden, msg)
}

func SpaceNotFound(msg string) *FormaError {
	if msg == "" {
		msg = "space not found"
	}
	return New(CodeSpaceNotFound, http.StatusNotFound, KeySpaceNotFound, msg)
}

func SpaceForbidden(msg string) *FormaError {
	if msg == "" {
		msg = "space access forbidden"
	}
	return New(CodeSpaceForbidden, http.StatusForbidden, KeySpaceForbidden, msg)
}

func VersionConflict(msg string) *FormaError {
	if msg == "" {
		msg = "revision conflict"
	}
	return New(CodeVersionConflict, http.StatusConflict, KeyVersionConflict, msg)
}

func Internal(msg string) *FormaError {
	if msg == "" {
		msg = "internal error"
	}
	return New(CodeInternal, http.StatusInternalServerError, KeyInternal, msg)
}

// AsFormaError extracts a FormaError from err.
func AsFormaError(err error) (*FormaError, bool) {
	var fe *FormaError
	if errors.As(err, &fe) {
		return fe, true
	}
	return nil, false
}

// MapDomainError maps known domain errors to FormaError.
func MapDomainError(err error) *FormaError {
	if err == nil {
		return nil
	}
	if fe, ok := AsFormaError(err); ok {
		return fe
	}
	if errors.Is(err, entity.ErrRevisionConflict) {
		return VersionConflict(err.Error())
	}
	if errors.Is(err, entity.ErrNotFound) {
		return TenantNotFound(err.Error())
	}
	return Internal(err.Error())
}
