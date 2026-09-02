/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package errors

import (
	"errors"
	"fmt"
	"net/http"

	analystentity "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	dataentity "github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
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

	CodeBusinessNotFound                   int32 = 40410
	CodeBusinessForbidden                  int32 = 40310
	CodeBusinessInvalidModel               int32 = 40010
	CodeBusinessRevisionMissing            int32 = 40411
	CodeBusinessNoChange                   int32 = 20010 // mapped as 200 with key, or 409-ish — use 200 OK with key in app; code for envelope
	CodeBusinessLayoutConflict             int32 = 40911
	CodeBusinessModelConflict              int32 = 40912
	CodeBusinessInvalidRelation            int32 = 40011
	CodeBusinessLayoutModelRevisionMissing int32 = 40412

	CodeAnalystSessionNotFound        int32 = 40420
	CodeAnalystSessionClosed          int32 = 40020
	CodeAnalystInvalidTurn            int32 = 40021
	CodeAnalystModelFailed            int32 = 50020
	CodeAnalystInvalidExtraction      int32 = 40022
	CodeAssertionNotFound             int32 = 40421
	CodeAssertionAlreadyDecided       int32 = 40920
	CodeAssertionEvidenceRequired     int32 = 40023
	CodeAnalystProposalNotFound       int32 = 40422
	CodeAnalystProposalStale          int32 = 40921
	CodeAnalystProposalInvalid        int32 = 40024
	CodeAnalystProposalAlreadyApplied int32 = 40922
	CodeAnalystForbidden              int32 = 40320
	CodeAnalystGapNotFound            int32 = 40423

	CodeDataRequirementNotFound         int32 = 40430
	CodeDataRequirementAlreadyDecided   int32 = 40930
	CodeDataRequirementInvalidState     int32 = 40931
	CodeDataAnalysisNotFound            int32 = 40431
	CodeDataAnalysisIdempotencyConflict int32 = 40932
	CodeDataAnalysisNotFailed           int32 = 40933
	CodeDataBusinessRevisionNotFound    int32 = 40432
	CodeDataBusinessElementRefInvalid   int32 = 40030
	CodeDataRequirementInvalid          int32 = 40031
	CodeDataForbidden                   int32 = 40330
	CodeDataNotConfigured               int32 = 50030
	CodeDataModelFailed                 int32 = 50031
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

	KeyBusinessNotFound                   = "FORMA_BUSINESS_NOT_FOUND"
	KeyBusinessForbidden                  = "FORMA_BUSINESS_FORBIDDEN"
	KeyBusinessInvalidModel               = "FORMA_BUSINESS_INVALID_MODEL"
	KeyBusinessRevisionMissing            = "FORMA_BUSINESS_REVISION_NOT_FOUND"
	KeyBusinessNoChange                   = "FORMA_BUSINESS_NO_CHANGE"
	KeyBusinessLayoutConflict             = "FORMA_BUSINESS_LAYOUT_CONFLICT"
	KeyBusinessModelConflict              = "FORMA_BUSINESS_MODEL_CONFLICT"
	KeyBusinessInvalidRelation            = "FORMA_BUSINESS_INVALID_RELATION"
	KeyBusinessLayoutModelRevisionMissing = "FORMA_BUSINESS_LAYOUT_MODEL_REVISION_NOT_FOUND"

	KeyAnalystSessionNotFound        = "FORMA_ANALYST_SESSION_NOT_FOUND"
	KeyAnalystSessionClosed          = "FORMA_ANALYST_SESSION_CLOSED"
	KeyAnalystInvalidTurn            = "FORMA_ANALYST_INVALID_TURN"
	KeyAnalystModelFailed            = "FORMA_ANALYST_MODEL_FAILED"
	KeyAnalystInvalidExtraction      = "FORMA_ANALYST_INVALID_EXTRACTION"
	KeyAssertionNotFound             = "FORMA_ASSERTION_NOT_FOUND"
	KeyAssertionAlreadyDecided       = "FORMA_ASSERTION_ALREADY_DECIDED"
	KeyAssertionEvidenceRequired     = "FORMA_ASSERTION_EVIDENCE_REQUIRED"
	KeyAnalystProposalNotFound       = "FORMA_PROPOSAL_NOT_FOUND"
	KeyAnalystProposalStale          = "FORMA_PROPOSAL_STALE"
	KeyAnalystProposalInvalid        = "FORMA_PROPOSAL_INVALID"
	KeyAnalystProposalAlreadyApplied = "FORMA_PROPOSAL_ALREADY_APPLIED"
	KeyAnalystForbidden              = "FORMA_ANALYST_FORBIDDEN"
	KeyAnalystGapNotFound            = "FORMA_ANALYST_GAP_NOT_FOUND"

	KeyDataRequirementNotFound         = "FORMA_DATA_REQUIREMENT_NOT_FOUND"
	KeyDataRequirementAlreadyDecided   = "FORMA_DATA_REQUIREMENT_ALREADY_DECIDED"
	KeyDataRequirementInvalidState     = "FORMA_DATA_REQUIREMENT_INVALID_STATE"
	KeyDataAnalysisNotFound            = "FORMA_DATA_ANALYSIS_NOT_FOUND"
	KeyDataAnalysisIdempotencyConflict = "FORMA_DATA_ANALYSIS_IDEMPOTENCY_CONFLICT"
	KeyDataAnalysisNotFailed           = "FORMA_DATA_ANALYSIS_NOT_FAILED"
	KeyDataBusinessRevisionNotFound    = "FORMA_DATA_BUSINESS_REVISION_NOT_FOUND"
	KeyDataBusinessElementRefInvalid   = "FORMA_DATA_BUSINESS_ELEMENT_REF_INVALID"
	KeyDataRequirementInvalid          = "FORMA_DATA_REQUIREMENT_INVALID"
	KeyDataForbidden                   = "FORMA_DATA_FORBIDDEN"
	KeyDataNotConfigured               = "FORMA_DATA_NOT_CONFIGURED"
	KeyDataModelFailed                 = "FORMA_DATA_MODEL_FAILED"
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

func AnalystSessionNotFound(msg string) *FormaError {
	if msg == "" {
		msg = "analyst session not found"
	}
	return New(CodeAnalystSessionNotFound, http.StatusNotFound, KeyAnalystSessionNotFound, msg)
}

func AnalystSessionClosed(msg string) *FormaError {
	if msg == "" {
		msg = "analyst session closed"
	}
	return New(CodeAnalystSessionClosed, http.StatusBadRequest, KeyAnalystSessionClosed, msg)
}

func AnalystInvalidTurn(msg string) *FormaError {
	if msg == "" {
		msg = "invalid analyst turn"
	}
	return New(CodeAnalystInvalidTurn, http.StatusBadRequest, KeyAnalystInvalidTurn, msg)
}

func AnalystModelFailed(msg string) *FormaError {
	if msg == "" {
		msg = "analyst model failed"
	}
	return New(CodeAnalystModelFailed, http.StatusInternalServerError, KeyAnalystModelFailed, msg)
}

func AnalystInvalidExtraction(msg string) *FormaError {
	if msg == "" {
		msg = "invalid extraction"
	}
	return New(CodeAnalystInvalidExtraction, http.StatusBadRequest, KeyAnalystInvalidExtraction, msg)
}

func AssertionNotFound(msg string) *FormaError {
	if msg == "" {
		msg = "assertion not found"
	}
	return New(CodeAssertionNotFound, http.StatusNotFound, KeyAssertionNotFound, msg)
}

func AssertionAlreadyDecided(msg string) *FormaError {
	if msg == "" {
		msg = "assertion already decided"
	}
	return New(CodeAssertionAlreadyDecided, http.StatusConflict, KeyAssertionAlreadyDecided, msg)
}

func AssertionEvidenceRequired(msg string) *FormaError {
	if msg == "" {
		msg = "assertion evidence required"
	}
	return New(CodeAssertionEvidenceRequired, http.StatusBadRequest, KeyAssertionEvidenceRequired, msg)
}

func AnalystProposalNotFound(msg string) *FormaError {
	if msg == "" {
		msg = "proposal not found"
	}
	return New(CodeAnalystProposalNotFound, http.StatusNotFound, KeyAnalystProposalNotFound, msg)
}

func AnalystProposalStale(msg string) *FormaError {
	if msg == "" {
		msg = "proposal stale"
	}
	return New(CodeAnalystProposalStale, http.StatusConflict, KeyAnalystProposalStale, msg)
}

func AnalystProposalInvalid(msg string) *FormaError {
	if msg == "" {
		msg = "proposal invalid"
	}
	return New(CodeAnalystProposalInvalid, http.StatusBadRequest, KeyAnalystProposalInvalid, msg)
}

func AnalystProposalAlreadyApplied(msg string) *FormaError {
	if msg == "" {
		msg = "proposal already applied"
	}
	return New(CodeAnalystProposalAlreadyApplied, http.StatusConflict, KeyAnalystProposalAlreadyApplied, msg)
}

func AnalystForbidden(msg string) *FormaError {
	if msg == "" {
		msg = "analyst action forbidden"
	}
	return New(CodeAnalystForbidden, http.StatusForbidden, KeyAnalystForbidden, msg)
}

func AnalystGapNotFound(msg string) *FormaError {
	if msg == "" {
		msg = "analyst gap not found"
	}
	return New(CodeAnalystGapNotFound, http.StatusNotFound, KeyAnalystGapNotFound, msg)
}

func DataRequirementNotFound(msg string) *FormaError {
	return New(CodeDataRequirementNotFound, http.StatusNotFound, KeyDataRequirementNotFound, defaultMessage(msg, "data requirement not found"))
}

func DataRequirementAlreadyDecided(msg string) *FormaError {
	return New(CodeDataRequirementAlreadyDecided, http.StatusConflict, KeyDataRequirementAlreadyDecided, defaultMessage(msg, "data requirement already decided"))
}

func DataRequirementInvalidState(msg string) *FormaError {
	return New(CodeDataRequirementInvalidState, http.StatusConflict, KeyDataRequirementInvalidState, defaultMessage(msg, "data requirement invalid state"))
}

func DataAnalysisNotFound(msg string) *FormaError {
	return New(CodeDataAnalysisNotFound, http.StatusNotFound, KeyDataAnalysisNotFound, defaultMessage(msg, "data analysis run not found"))
}

func DataAnalysisIdempotencyConflict(msg string) *FormaError {
	return New(CodeDataAnalysisIdempotencyConflict, http.StatusConflict, KeyDataAnalysisIdempotencyConflict, defaultMessage(msg, "data analysis idempotency conflict"))
}

func DataAnalysisNotFailed(msg string) *FormaError {
	return New(CodeDataAnalysisNotFailed, http.StatusConflict, KeyDataAnalysisNotFailed, defaultMessage(msg, "data analysis run is not failed"))
}

func DataBusinessRevisionNotFound(msg string) *FormaError {
	return New(CodeDataBusinessRevisionNotFound, http.StatusNotFound, KeyDataBusinessRevisionNotFound, defaultMessage(msg, "data business revision not found"))
}

func DataBusinessElementRefInvalid(msg string) *FormaError {
	return New(CodeDataBusinessElementRefInvalid, http.StatusBadRequest, KeyDataBusinessElementRefInvalid, defaultMessage(msg, "data business element reference invalid"))
}

func DataRequirementInvalid(msg string) *FormaError {
	return New(CodeDataRequirementInvalid, http.StatusBadRequest, KeyDataRequirementInvalid, defaultMessage(msg, "data requirement proposal invalid"))
}

func DataForbidden(msg string) *FormaError {
	return New(CodeDataForbidden, http.StatusForbidden, KeyDataForbidden, defaultMessage(msg, "data action forbidden"))
}

func DataNotConfigured(msg string) *FormaError {
	return New(CodeDataNotConfigured, http.StatusInternalServerError, KeyDataNotConfigured, defaultMessage(msg, "data service not configured"))
}

func DataModelFailed(msg string) *FormaError {
	return New(CodeDataModelFailed, http.StatusInternalServerError, KeyDataModelFailed, defaultMessage(msg, "data model failed"))
}

func defaultMessage(msg, fallback string) string {
	if msg == "" {
		return fallback
	}
	return msg
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
	if errors.Is(err, analystentity.ErrSessionNotFound) {
		return AnalystSessionNotFound(err.Error())
	}
	if errors.Is(err, analystentity.ErrSessionClosed) {
		return AnalystSessionClosed(err.Error())
	}
	if errors.Is(err, analystentity.ErrInvalidTurn) {
		return AnalystInvalidTurn(err.Error())
	}
	if errors.Is(err, analystentity.ErrModelFailed) {
		return AnalystModelFailed(err.Error())
	}
	if errors.Is(err, analystentity.ErrInvalidExtraction) {
		return AnalystInvalidExtraction(err.Error())
	}
	if errors.Is(err, analystentity.ErrAssertionNotFound) {
		return AssertionNotFound(err.Error())
	}
	if errors.Is(err, analystentity.ErrAssertionAlreadyDecided) {
		return AssertionAlreadyDecided(err.Error())
	}
	if errors.Is(err, analystentity.ErrAssertionEvidenceRequired) {
		return AssertionEvidenceRequired(err.Error())
	}
	if errors.Is(err, analystentity.ErrProposalNotFound) {
		return AnalystProposalNotFound(err.Error())
	}
	if errors.Is(err, analystentity.ErrProposalStale) {
		return AnalystProposalStale(err.Error())
	}
	if errors.Is(err, analystentity.ErrProposalInvalid) {
		return AnalystProposalInvalid(err.Error())
	}
	if errors.Is(err, analystentity.ErrProposalAlreadyApplied) {
		return AnalystProposalAlreadyApplied(err.Error())
	}
	if errors.Is(err, analystentity.ErrForbidden) {
		return AnalystForbidden(err.Error())
	}
	if errors.Is(err, analystentity.ErrGapNotFound) {
		return AnalystGapNotFound(err.Error())
	}
	if errors.Is(err, dataentity.ErrRequirementNotFound) {
		return DataRequirementNotFound(err.Error())
	}
	if errors.Is(err, dataentity.ErrRequirementAlreadyDecided) {
		return DataRequirementAlreadyDecided(err.Error())
	}
	if errors.Is(err, dataentity.ErrRequirementInvalidState) {
		return DataRequirementInvalidState(err.Error())
	}
	if errors.Is(err, dataentity.ErrAnalysisNotFound) {
		return DataAnalysisNotFound(err.Error())
	}
	if errors.Is(err, dataentity.ErrAnalysisIdempotencyConflict) {
		return DataAnalysisIdempotencyConflict(err.Error())
	}
	if errors.Is(err, dataentity.ErrAnalysisNotFailed) {
		return DataAnalysisNotFailed(err.Error())
	}
	if errors.Is(err, dataentity.ErrBusinessRevisionNotFound) {
		return DataBusinessRevisionNotFound(err.Error())
	}
	if errors.Is(err, dataentity.ErrBusinessElementRefInvalid) {
		return DataBusinessElementRefInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrInvalidProposal) {
		return DataRequirementInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrForbidden) {
		return DataForbidden(err.Error())
	}
	if errors.Is(err, dataentity.ErrNotConfigured) {
		return DataNotConfigured(err.Error())
	}
	if errors.Is(err, dataentity.ErrModelFailed) {
		return DataModelFailed(err.Error())
	}
	return Internal(err.Error())
}
