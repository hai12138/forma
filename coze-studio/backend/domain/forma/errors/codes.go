/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package errors

import (
	"errors"
	"fmt"
	"net/http"

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	tenancyentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
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

	CodeBusinessNotFound        int32 = 40410
	CodeBusinessForbidden       int32 = 40310
	CodeBusinessInvalidModel    int32 = 40010
	CodeBusinessRevisionMissing int32 = 40411
	CodeBusinessNoChange        int32 = 20010 // mapped as 200 with key, or 409-ish — use 200 OK with key in app; code for envelope
	CodeBusinessLayoutConflict  int32 = 40911
	CodeBusinessModelConflict   int32 = 40912
	CodeBusinessInvalidRelation            int32 = 40011
	CodeBusinessLayoutModelRevisionMissing int32 = 40412
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

	KeyBusinessNotFound        = "FORMA_BUSINESS_NOT_FOUND"
	KeyBusinessForbidden       = "FORMA_BUSINESS_FORBIDDEN"
	KeyBusinessInvalidModel    = "FORMA_BUSINESS_INVALID_MODEL"
	KeyBusinessRevisionMissing = "FORMA_BUSINESS_REVISION_NOT_FOUND"
	KeyBusinessNoChange        = "FORMA_BUSINESS_NO_CHANGE"
	KeyBusinessLayoutConflict  = "FORMA_BUSINESS_LAYOUT_CONFLICT"
	KeyBusinessModelConflict   = "FORMA_BUSINESS_MODEL_CONFLICT"
	KeyBusinessInvalidRelation            = "FORMA_BUSINESS_INVALID_RELATION"
	KeyBusinessLayoutModelRevisionMissing = "FORMA_BUSINESS_LAYOUT_MODEL_REVISION_NOT_FOUND"
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

func BusinessNotFound(msg string) *FormaError {
	if msg == "" {
		msg = "business not found"
	}
	return New(CodeBusinessNotFound, http.StatusNotFound, KeyBusinessNotFound, msg)
}

func BusinessForbidden(msg string) *FormaError {
	if msg == "" {
		msg = "business access forbidden"
	}
	return New(CodeBusinessForbidden, http.StatusForbidden, KeyBusinessForbidden, msg)
}

func BusinessInvalidModel(msg string) *FormaError {
	if msg == "" {
		msg = "invalid business model"
	}
	return New(CodeBusinessInvalidModel, http.StatusBadRequest, KeyBusinessInvalidModel, msg)
}

func BusinessRevisionNotFound(msg string) *FormaError {
	if msg == "" {
		msg = "business revision not found"
	}
	return New(CodeBusinessRevisionMissing, http.StatusNotFound, KeyBusinessRevisionMissing, msg)
}

func BusinessNoChange(msg string) *FormaError {
	if msg == "" {
		msg = "no semantic change"
	}
	return New(CodeBusinessNoChange, http.StatusOK, KeyBusinessNoChange, msg)
}

func BusinessLayoutConflict(msg string) *FormaError {
	if msg == "" {
		msg = "layout revision conflict"
	}
	return New(CodeBusinessLayoutConflict, http.StatusConflict, KeyBusinessLayoutConflict, msg)
}

func BusinessModelConflict(msg string) *FormaError {
	if msg == "" {
		msg = "business model revision conflict"
	}
	return New(CodeBusinessModelConflict, http.StatusConflict, KeyBusinessModelConflict, msg)
}

func BusinessInvalidRelation(msg string) *FormaError {
	if msg == "" {
		msg = "invalid business relation"
	}
	return New(CodeBusinessInvalidRelation, http.StatusBadRequest, KeyBusinessInvalidRelation, msg)
}

func BusinessLayoutModelRevisionNotFound(msg string) *FormaError {
	if msg == "" {
		msg = "layout based_on_model_revision not found"
	}
	return New(CodeBusinessLayoutModelRevisionMissing, http.StatusNotFound, KeyBusinessLayoutModelRevisionMissing, msg)
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
	if errors.Is(err, tenancyentity.ErrRevisionConflict) {
		return VersionConflict(err.Error())
	}
	if errors.Is(err, tenancyentity.ErrNotFound) {
		return TenantNotFound(err.Error())
	}
	if errors.Is(err, tenancyentity.ErrLastOwner) ||
		errors.Is(err, tenancyentity.ErrPrimaryOwnerImmutable) ||
		errors.Is(err, tenancyentity.ErrInvalidRole) {
		return MembershipForbidden(err.Error())
	}
	if errors.Is(err, tenancyentity.ErrSpaceOwned) {
		return SpaceForbidden(err.Error())
	}
	if errors.Is(err, businessentity.ErrNotFound) {
		return BusinessNotFound(err.Error())
	}
	if errors.Is(err, businessentity.ErrRevisionNotFound) {
		return BusinessRevisionNotFound(err.Error())
	}
	if errors.Is(err, businessentity.ErrRevisionConflict) {
		return BusinessModelConflict(err.Error())
	}
	if errors.Is(err, businessentity.ErrLayoutConflict) {
		return BusinessLayoutConflict(err.Error())
	}
	if errors.Is(err, businessentity.ErrNoChange) {
		return BusinessNoChange(err.Error())
	}
	if errors.Is(err, businessentity.ErrInvalidModel) {
		return BusinessInvalidModel(err.Error())
	}
	if errors.Is(err, businessentity.ErrInvalidRelation) {
		return BusinessInvalidRelation(err.Error())
	}
	if errors.Is(err, businessentity.ErrLayoutModelRevisionNotFound) {
		return BusinessLayoutModelRevisionNotFound(err.Error())
	}
	return Internal(err.Error())
}
