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

	CodeDataRequirementNotFound                int32 = 40430
	CodeDataRequirementAlreadyDecided          int32 = 40930
	CodeDataRequirementInvalidState            int32 = 40931
	CodeDataAnalysisNotFound                   int32 = 40431
	CodeDataAnalysisIdempotencyConflict        int32 = 40932
	CodeDataAnalysisNotFailed                  int32 = 40933
	CodeDataBusinessRevisionNotFound           int32 = 40432
	CodeDataBusinessElementRefInvalid          int32 = 40030
	CodeDataRequirementInvalid                 int32 = 40031
	CodeDataForbidden                          int32 = 40330
	CodeDataNotConfigured                      int32 = 50030
	CodeDataModelFailed                        int32 = 50031
	CodeDataSemanticMappingNotFound            int32 = 40450
	CodeDataSemanticMappingAlreadyDecided      int32 = 40950
	CodeDataSemanticMappingInvalidState        int32 = 40951
	CodeDataSemanticMappingAlreadyConfirmed    int32 = 40952
	CodeDataSemanticMappingAnalysisNotFound    int32 = 40451
	CodeDataSemanticMappingIdempotencyConflict int32 = 40953
	CodeDataSemanticMappingAnalysisNotFailed   int32 = 40954
	CodeDataSemanticMappingTargetInvalid       int32 = 40050
	CodeDataSemanticMappingTransformInvalid    int32 = 40051
	CodeDataSemanticMappingLineageInvalid      int32 = 40052
	CodeDataSemanticMappingRequirementInvalid  int32 = 40955

	CodeDataContractNotFound             int32 = 40460
	CodeDataContractRevisionNotFound     int32 = 40461
	CodeDataContractInvalidState         int32 = 40960
	CodeDataContractInvalidPayload       int32 = 40060
	CodeDataContractVersionConflict      int32 = 40961
	CodeDataContractValidationFailed     int32 = 40962
	CodeDataContractBindingInvalid       int32 = 40061
	CodeDataContractLogicalSchemaInvalid int32 = 40062
	CodeDataContractDriftInvalid         int32 = 40063
	CodeDataContractGapInvalid           int32 = 40064
	CodeDataContractNotActive            int32 = 40963

	CodeDataSourceNotFound            int32 = 40440
	CodeDataConnectionNotFound        int32 = 40441
	CodeDataCredentialNotFound        int32 = 40442
	CodeDataAdapterNotSupported       int32 = 40040
	CodeDataConnectionFailed          int32 = 50240
	CodeDataDiscoveryFailed           int32 = 50241
	CodeDataAssetNotFound             int32 = 40443
	CodeDataSchemaSnapshotNotFound    int32 = 40444
	CodeDataSecretProviderUnavailable int32 = 50340
	CodeDataSecretInvalid             int32 = 40041
	CodeDataPublicConfigInvalid       int32 = 40042
	CodeSourceTypeNotSupported        int32 = 40043
	CodeDataContractNotAvailable      int32 = 40940
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

	KeyDataRequirementNotFound                = "FORMA_DATA_REQUIREMENT_NOT_FOUND"
	KeyDataRequirementAlreadyDecided          = "FORMA_DATA_REQUIREMENT_ALREADY_DECIDED"
	KeyDataRequirementInvalidState            = "FORMA_DATA_REQUIREMENT_INVALID_STATE"
	KeyDataAnalysisNotFound                   = "FORMA_DATA_ANALYSIS_NOT_FOUND"
	KeyDataAnalysisIdempotencyConflict        = "FORMA_DATA_ANALYSIS_IDEMPOTENCY_CONFLICT"
	KeyDataAnalysisNotFailed                  = "FORMA_DATA_ANALYSIS_NOT_FAILED"
	KeyDataBusinessRevisionNotFound           = "FORMA_DATA_BUSINESS_REVISION_NOT_FOUND"
	KeyDataBusinessElementRefInvalid          = "FORMA_DATA_BUSINESS_ELEMENT_REF_INVALID"
	KeyDataRequirementInvalid                 = "FORMA_DATA_REQUIREMENT_INVALID"
	KeyDataForbidden                          = "FORMA_DATA_FORBIDDEN"
	KeyDataNotConfigured                      = "FORMA_DATA_NOT_CONFIGURED"
	KeyDataModelFailed                        = "FORMA_DATA_MODEL_FAILED"
	KeyDataSemanticMappingNotFound            = "FORMA_DATA_SEMANTIC_MAPPING_NOT_FOUND"
	KeyDataSemanticMappingAlreadyDecided      = "FORMA_DATA_SEMANTIC_MAPPING_ALREADY_DECIDED"
	KeyDataSemanticMappingInvalidState        = "FORMA_DATA_SEMANTIC_MAPPING_INVALID_STATE"
	KeyDataSemanticMappingAlreadyConfirmed    = "FORMA_DATA_SEMANTIC_MAPPING_ALREADY_CONFIRMED"
	KeyDataSemanticMappingAnalysisNotFound    = "FORMA_DATA_SEMANTIC_MAPPING_ANALYSIS_NOT_FOUND"
	KeyDataSemanticMappingIdempotencyConflict = "FORMA_DATA_SEMANTIC_MAPPING_ANALYSIS_IDEMPOTENCY_CONFLICT"
	KeyDataSemanticMappingAnalysisNotFailed   = "FORMA_DATA_SEMANTIC_MAPPING_ANALYSIS_NOT_FAILED"
	KeyDataSemanticMappingTargetInvalid       = "FORMA_DATA_SEMANTIC_MAPPING_TARGET_INVALID"
	KeyDataSemanticMappingTransformInvalid    = "FORMA_DATA_SEMANTIC_MAPPING_TRANSFORM_INVALID"
	KeyDataSemanticMappingLineageInvalid      = "FORMA_DATA_SEMANTIC_MAPPING_LINEAGE_INVALID"
	KeyDataSemanticMappingRequirementInvalid  = "FORMA_DATA_SEMANTIC_MAPPING_REQUIREMENT_NOT_CONFIRMED"

	KeyDataContractNotFound             = "FORMA_DATA_CONTRACT_NOT_FOUND"
	KeyDataContractRevisionNotFound     = "FORMA_DATA_CONTRACT_REVISION_NOT_FOUND"
	KeyDataContractInvalidState         = "FORMA_DATA_CONTRACT_INVALID_STATE"
	KeyDataContractInvalidPayload       = "FORMA_DATA_CONTRACT_INVALID_PAYLOAD"
	KeyDataContractVersionConflict      = "FORMA_DATA_CONTRACT_VERSION_CONFLICT"
	KeyDataContractValidationFailed     = "FORMA_DATA_CONTRACT_VALIDATION_FAILED"
	KeyDataContractBindingInvalid       = "FORMA_DATA_CONTRACT_BINDING_INVALID"
	KeyDataContractLogicalSchemaInvalid = "FORMA_DATA_CONTRACT_LOGICAL_SCHEMA_INVALID"
	KeyDataContractDriftInvalid         = "FORMA_DATA_CONTRACT_DRIFT_INVALID"
	KeyDataContractGapInvalid           = "FORMA_DATA_CONTRACT_GAP_INVALID"
	KeyDataContractNotActive            = "FORMA_DATA_CONTRACT_NOT_ACTIVE"

	KeyDataSourceNotFound            = "FORMA_DATA_SOURCE_NOT_FOUND"
	KeyDataConnectionNotFound        = "FORMA_DATA_CONNECTION_NOT_FOUND"
	KeyDataCredentialNotFound        = "FORMA_DATA_CREDENTIAL_NOT_FOUND"
	KeyDataAdapterNotSupported       = "FORMA_DATA_ADAPTER_NOT_SUPPORTED"
	KeyDataConnectionFailed          = "FORMA_DATA_CONNECTION_FAILED"
	KeyDataDiscoveryFailed           = "FORMA_DATA_DISCOVERY_FAILED"
	KeyDataAssetNotFound             = "FORMA_DATA_ASSET_NOT_FOUND"
	KeyDataSchemaSnapshotNotFound    = "FORMA_DATA_SCHEMA_SNAPSHOT_NOT_FOUND"
	KeyDataSecretProviderUnavailable = "FORMA_DATA_SECRET_PROVIDER_UNAVAILABLE"
	KeyDataSecretInvalid             = "FORMA_DATA_SECRET_INVALID"
	KeyDataPublicConfigInvalid       = "FORMA_DATA_PUBLIC_CONFIG_INVALID"
	KeySourceTypeNotSupported        = "FORMA_DATA_SOURCE_TYPE_NOT_SUPPORTED"
	KeyDataContractNotAvailable      = "FORMA_DATA_CONTRACT_NOT_AVAILABLE"
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

func dataMappingError(code int32, status int, key, msg, fallback string) *FormaError {
	return New(code, status, key, defaultMessage(msg, fallback))
}
func DataSemanticMappingNotFound(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingNotFound, http.StatusNotFound, KeyDataSemanticMappingNotFound, msg, "semantic mapping not found")
}
func DataSemanticMappingAlreadyDecided(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingAlreadyDecided, http.StatusConflict, KeyDataSemanticMappingAlreadyDecided, msg, "semantic mapping already decided")
}
func DataSemanticMappingInvalidState(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingInvalidState, http.StatusConflict, KeyDataSemanticMappingInvalidState, msg, "semantic mapping invalid state")
}
func DataSemanticMappingAlreadyConfirmed(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingAlreadyConfirmed, http.StatusConflict, KeyDataSemanticMappingAlreadyConfirmed, msg, "semantic mapping already confirmed")
}
func DataSemanticMappingAnalysisNotFound(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingAnalysisNotFound, http.StatusNotFound, KeyDataSemanticMappingAnalysisNotFound, msg, "semantic mapping analysis run not found")
}
func DataSemanticMappingIdempotencyConflict(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingIdempotencyConflict, http.StatusConflict, KeyDataSemanticMappingIdempotencyConflict, msg, "semantic mapping analysis idempotency conflict")
}
func DataSemanticMappingAnalysisNotFailed(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingAnalysisNotFailed, http.StatusConflict, KeyDataSemanticMappingAnalysisNotFailed, msg, "semantic mapping analysis run is not failed")
}
func DataSemanticMappingTargetInvalid(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingTargetInvalid, http.StatusBadRequest, KeyDataSemanticMappingTargetInvalid, msg, "semantic mapping target invalid")
}
func DataSemanticMappingTransformInvalid(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingTransformInvalid, http.StatusBadRequest, KeyDataSemanticMappingTransformInvalid, msg, "semantic mapping transform invalid")
}
func DataSemanticMappingLineageInvalid(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingLineageInvalid, http.StatusBadRequest, KeyDataSemanticMappingLineageInvalid, msg, "semantic mapping lineage invalid")
}
func DataSemanticMappingRequirementInvalid(msg string) *FormaError {
	return dataMappingError(CodeDataSemanticMappingRequirementInvalid, http.StatusConflict, KeyDataSemanticMappingRequirementInvalid, msg, "semantic mapping requirement is not confirmed")
}

func dataContractError(code int32, status int, key, msg, fallback string) *FormaError {
	return New(code, status, key, defaultMessage(msg, fallback))
}
func DataContractNotFound(msg string) *FormaError {
	return dataContractError(CodeDataContractNotFound, http.StatusNotFound, KeyDataContractNotFound, msg, "data contract not found")
}
func DataContractRevisionNotFound(msg string) *FormaError {
	return dataContractError(CodeDataContractRevisionNotFound, http.StatusNotFound, KeyDataContractRevisionNotFound, msg, "data contract revision not found")
}
func DataContractInvalidState(msg string) *FormaError {
	return dataContractError(CodeDataContractInvalidState, http.StatusConflict, KeyDataContractInvalidState, msg, "data contract invalid state")
}
func DataContractInvalidPayload(msg string) *FormaError {
	return dataContractError(CodeDataContractInvalidPayload, http.StatusBadRequest, KeyDataContractInvalidPayload, msg, "data contract invalid payload")
}
func DataContractVersionConflict(msg string) *FormaError {
	return dataContractError(CodeDataContractVersionConflict, http.StatusConflict, KeyDataContractVersionConflict, msg, "data contract version conflict")
}
func DataContractValidationFailed(msg string) *FormaError {
	return dataContractError(CodeDataContractValidationFailed, http.StatusConflict, KeyDataContractValidationFailed, msg, "data contract validation failed")
}
func DataContractBindingInvalid(msg string) *FormaError {
	return dataContractError(CodeDataContractBindingInvalid, http.StatusBadRequest, KeyDataContractBindingInvalid, msg, "data contract binding invalid")
}
func DataContractLogicalSchemaInvalid(msg string) *FormaError {
	return dataContractError(CodeDataContractLogicalSchemaInvalid, http.StatusBadRequest, KeyDataContractLogicalSchemaInvalid, msg, "data contract logical schema invalid")
}
func DataContractDriftInvalid(msg string) *FormaError {
	return dataContractError(CodeDataContractDriftInvalid, http.StatusBadRequest, KeyDataContractDriftInvalid, msg, "data contract drift invalid")
}
func DataContractGapInvalid(msg string) *FormaError {
	return dataContractError(CodeDataContractGapInvalid, http.StatusBadRequest, KeyDataContractGapInvalid, msg, "data contract gap invalid")
}
func DataContractNotActive(msg string) *FormaError {
	return dataContractError(CodeDataContractNotActive, http.StatusConflict, KeyDataContractNotActive, msg, "data contract not active")
}

func DataSourceNotFound(msg string) *FormaError {
	return New(CodeDataSourceNotFound, http.StatusNotFound, KeyDataSourceNotFound, defaultMessage(msg, "data source not found"))
}
func DataConnectionNotFound(msg string) *FormaError {
	return New(CodeDataConnectionNotFound, http.StatusNotFound, KeyDataConnectionNotFound, defaultMessage(msg, "data connection not found"))
}
func DataCredentialNotFound(msg string) *FormaError {
	return New(CodeDataCredentialNotFound, http.StatusNotFound, KeyDataCredentialNotFound, defaultMessage(msg, "data credential not found"))
}
func DataAdapterNotSupported(msg string) *FormaError {
	return New(CodeDataAdapterNotSupported, http.StatusBadRequest, KeyDataAdapterNotSupported, defaultMessage(msg, "data adapter not supported"))
}
func DataConnectionFailed(msg string) *FormaError {
	return New(CodeDataConnectionFailed, http.StatusBadGateway, KeyDataConnectionFailed, defaultMessage(msg, "data connection failed"))
}
func DataDiscoveryFailed(msg string) *FormaError {
	return New(CodeDataDiscoveryFailed, http.StatusBadGateway, KeyDataDiscoveryFailed, defaultMessage(msg, "data discovery failed"))
}
func DataAssetNotFound(msg string) *FormaError {
	return New(CodeDataAssetNotFound, http.StatusNotFound, KeyDataAssetNotFound, defaultMessage(msg, "data asset not found"))
}
func DataSchemaSnapshotNotFound(msg string) *FormaError {
	return New(CodeDataSchemaSnapshotNotFound, http.StatusNotFound, KeyDataSchemaSnapshotNotFound, defaultMessage(msg, "data schema snapshot not found"))
}
func DataSecretProviderUnavailable(msg string) *FormaError {
	return New(CodeDataSecretProviderUnavailable, http.StatusServiceUnavailable, KeyDataSecretProviderUnavailable, defaultMessage(msg, "data secret provider unavailable"))
}
func DataSecretInvalid(msg string) *FormaError {
	return New(CodeDataSecretInvalid, http.StatusBadRequest, KeyDataSecretInvalid, defaultMessage(msg, "data secret invalid"))
}
func DataPublicConfigInvalid(msg string) *FormaError {
	return New(CodeDataPublicConfigInvalid, http.StatusBadRequest, KeyDataPublicConfigInvalid, defaultMessage(msg, "data public config invalid"))
}
func SourceTypeNotSupported(msg string) *FormaError {
	return New(CodeSourceTypeNotSupported, http.StatusBadRequest, KeySourceTypeNotSupported, defaultMessage(msg, "data source type not supported"))
}
func DataContractNotAvailable(msg string) *FormaError {
	return New(CodeDataContractNotAvailable, http.StatusConflict, KeyDataContractNotAvailable, defaultMessage(msg, "data contract not available"))
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
	if errors.Is(err, dataentity.ErrMappingNotFound) {
		return DataSemanticMappingNotFound(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingAlreadyDecided) {
		return DataSemanticMappingAlreadyDecided(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingInvalidState) {
		return DataSemanticMappingInvalidState(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingAlreadyConfirmed) {
		return DataSemanticMappingAlreadyConfirmed(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingAnalysisNotFound) {
		return DataSemanticMappingAnalysisNotFound(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingAnalysisIdempotencyConflict) {
		return DataSemanticMappingIdempotencyConflict(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingAnalysisNotFailed) {
		return DataSemanticMappingAnalysisNotFailed(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingTargetInvalid) {
		return DataSemanticMappingTargetInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingTransformInvalid) {
		return DataSemanticMappingTransformInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingLineageInvalid) {
		return DataSemanticMappingLineageInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrMappingRequirementNotConfirmed) {
		return DataSemanticMappingRequirementInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractNotFound) {
		return DataContractNotFound(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractRevisionNotFound) {
		return DataContractRevisionNotFound(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractInvalidState) {
		return DataContractInvalidState(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractInvalidPayload) {
		return DataContractInvalidPayload(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractVersionConflict) {
		return DataContractVersionConflict(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractValidationFailed) {
		return DataContractValidationFailed(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractBindingInvalid) {
		return DataContractBindingInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractLogicalSchemaInvalid) {
		return DataContractLogicalSchemaInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractDriftInvalid) {
		return DataContractDriftInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractGapInvalid) {
		return DataContractGapInvalid(err.Error())
	}
	if errors.Is(err, dataentity.ErrContractNotActive) {
		return DataContractNotActive(err.Error())
	}
	if errors.Is(err, dataentity.ErrDataSourceNotFound) {
		return DataSourceNotFound("")
	}
	if errors.Is(err, dataentity.ErrDataConnectionNotFound) {
		return DataConnectionNotFound("")
	}
	if errors.Is(err, dataentity.ErrDataCredentialNotFound) {
		return DataCredentialNotFound("")
	}
	if errors.Is(err, dataentity.ErrDataAdapterNotSupported) {
		return DataAdapterNotSupported("")
	}
	if errors.Is(err, dataentity.ErrDataConnectionFailed) {
		return DataConnectionFailed("")
	}
	if errors.Is(err, dataentity.ErrDataDiscoveryFailed) {
		return DataDiscoveryFailed("")
	}
	if errors.Is(err, dataentity.ErrDataAssetNotFound) {
		return DataAssetNotFound("")
	}
	if errors.Is(err, dataentity.ErrDataSchemaSnapshotNotFound) {
		return DataSchemaSnapshotNotFound("")
	}
	if errors.Is(err, dataentity.ErrSecretProviderUnavailable) {
		return DataSecretProviderUnavailable("")
	}
	if errors.Is(err, dataentity.ErrSecretInvalid) {
		return DataSecretInvalid("")
	}
	if errors.Is(err, dataentity.ErrPublicConfigInvalid) {
		return DataPublicConfigInvalid("")
	}
	if errors.Is(err, dataentity.ErrSourceTypeNotSupported) {
		return SourceTypeNotSupported("")
	}
	if errors.Is(err, dataentity.ErrDataContractNotAvailable) {
		return DataContractNotAvailable("")
	}
	return Internal(err.Error())
}
