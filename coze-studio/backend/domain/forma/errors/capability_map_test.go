/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package errors_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

func TestMapDomainError_CapabilityStableKeys(t *testing.T) {
	cases := []struct {
		err  error
		code int32
		key  string
	}{
		{capentity.ErrNotFound, formaerrors.CodeCapabilityNotFound, formaerrors.KeyCapabilityNotFound},
		{capentity.ErrRevisionNotFound, formaerrors.CodeCapabilityRevisionNotFound, formaerrors.KeyCapabilityRevisionNotFound},
		{capentity.ErrProposalNotFound, formaerrors.CodeCapabilityProposalNotFound, formaerrors.KeyCapabilityProposalNotFound},
		{capentity.ErrAnalysisNotFound, formaerrors.CodeCapabilityAnalysisNotFound, formaerrors.KeyCapabilityAnalysisNotFound},
		{capentity.ErrDecisionNotFound, formaerrors.CodeCapabilityDecisionNotFound, formaerrors.KeyCapabilityDecisionNotFound},
		{capentity.ErrForbidden, formaerrors.CodeCapabilityForbidden, formaerrors.KeyCapabilityForbidden},
		{capentity.ErrCrossTenant, formaerrors.CodeCapabilityForbidden, formaerrors.KeyCapabilityForbidden},
		{capentity.ErrInvalidPayload, formaerrors.CodeCapabilityInvalidPayload, formaerrors.KeyCapabilityInvalidPayload},
		{capentity.ErrIllegalTransition, formaerrors.CodeCapabilityIllegalTransition, formaerrors.KeyCapabilityIllegalTransition},
		{capentity.ErrIdempotencyConflict, formaerrors.CodeCapabilityIdempotencyConflict, formaerrors.KeyCapabilityIdempotencyConflict},
		{capentity.ErrActiveConflict, formaerrors.CodeCapabilityActiveConflict, formaerrors.KeyCapabilityActiveConflict},
		{capentity.ErrValidationFailed, formaerrors.CodeCapabilityValidationFailed, formaerrors.KeyCapabilityValidationFailed},
		{capentity.ErrConflict, formaerrors.CodeCapabilityConflict, formaerrors.KeyCapabilityConflict},
		{capentity.ErrConsistency, formaerrors.CodeCapabilityConflict, formaerrors.KeyCapabilityConflict},
		{capentity.ErrNotConfigured, formaerrors.CodeCapabilityNotConfigured, formaerrors.KeyCapabilityNotConfigured},
		{capentity.ErrPortsNotConfigured, formaerrors.CodeCapabilityNotConfigured, formaerrors.KeyCapabilityNotConfigured},
	}
	for _, tc := range cases {
		fe := formaerrors.MapDomainError(tc.err)
		require.Equal(t, tc.code, fe.Code, "%v", tc.err)
		require.Equal(t, tc.key, fe.Key, "%v", tc.err)
		low := strings.ToLower(fe.Msg)
		require.NotContains(t, low, "password")
		require.NotContains(t, low, "token")
		require.NotContains(t, low, "authorization")
	}
}
