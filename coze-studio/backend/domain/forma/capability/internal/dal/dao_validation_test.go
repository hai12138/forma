/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

func TestToValidationMalformedIssueCodesJSON(t *testing.T) {
	row := &validationResultRow{
		ValidationID: "cval_1", TenantID: "t1", BusinessID: "biz", CapabilityID: "cap",
		RevisionID: "crev_1", RevisionContentDigest: "r", BusinessModelRevision: 1,
		BusinessModelContentDigest: "b", ContractEvidenceDigest: "c", EvidenceDigest: "e",
		Status: string(entity.ValidationFail), IssueCodesJSON: "{not-json",
		ValidatedBy: "a", ValidatedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(),
	}
	got, err := toValidation(row)
	require.ErrorIs(t, err, entity.ErrConsistency)
	require.Nil(t, got)
}

func TestToValidationValidIssueCodesJSON(t *testing.T) {
	row := &validationResultRow{
		ValidationID: "cval_1", TenantID: "t1", BusinessID: "biz", CapabilityID: "cap",
		RevisionID: "crev_1", RevisionContentDigest: "r", BusinessModelRevision: 1,
		BusinessModelContentDigest: "b", ContractEvidenceDigest: "c", EvidenceDigest: "e",
		Status: string(entity.ValidationFail), IssueCodesJSON: `["FORMA_CAPABILITY_TYPE_MISMATCH"]`,
		ValidatedBy: "a", ValidatedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(),
	}
	got, err := toValidation(row)
	require.NoError(t, err)
	require.Equal(t, []string{entity.IssueTypeMismatch}, got.IssueCodes)
}

func TestValidationFromMarshal(t *testing.T) {
	row, err := validationFrom(&entity.CapabilityValidationResult{
		ValidationID: "cval_1", TenantID: "t1", IssueCodes: []string{entity.IssueMappingMissing},
		Status: entity.ValidationFail,
	})
	require.NoError(t, err)
	require.Equal(t, `["FORMA_CAPABILITY_MAPPING_MISSING"]`, row.IssueCodesJSON)
}
