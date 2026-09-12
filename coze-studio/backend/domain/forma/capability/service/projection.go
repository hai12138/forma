/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"fmt"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

// ProjectionResult is the deterministic AssetRef header projection (§3.4.1).
type ProjectionResult struct {
	Name            string
	SemanticVersion string
	ContentDigest   string
	Status          assetentity.AssetStatus
}

// ProjectCapabilityAssetRef selects the projected revision and AssetRef fields.
// On consistency error returns entity.ErrConsistency — caller must not write AssetRef.
func ProjectCapabilityAssetRef(cap *entity.BusinessCapability, revisions []*entity.BusinessCapabilityRevision) (*ProjectionResult, error) {
	if cap == nil {
		return nil, entity.ErrConsistency
	}
	if len(revisions) == 0 {
		return nil, entity.ErrConsistency
	}

	seenVersion := map[int32]struct{}{}
	byID := map[string]*entity.BusinessCapabilityRevision{}
	var activeSet []*entity.BusinessCapabilityRevision
	for _, r := range revisions {
		if r == nil {
			return nil, entity.ErrConsistency
		}
		if r.CapabilityID != "" && r.CapabilityID != cap.CapabilityID {
			return nil, entity.ErrConsistency
		}
		if r.TenantID != "" && r.TenantID != cap.TenantID {
			return nil, entity.ErrConsistency
		}
		if _, dup := seenVersion[r.Version]; dup {
			return nil, entity.ErrConsistency
		}
		seenVersion[r.Version] = struct{}{}
		byID[r.RevisionID] = r
		if r.Status == entity.RevisionActive {
			activeSet = append(activeSet, r)
		}
	}

	if cap.ActiveRevisionID == "" {
		if len(activeSet) != 0 {
			return nil, entity.ErrConsistency
		}
	} else {
		if len(activeSet) != 1 {
			return nil, entity.ErrConsistency
		}
		pointed, ok := byID[cap.ActiveRevisionID]
		if !ok || pointed.Status != entity.RevisionActive || pointed.RevisionID != activeSet[0].RevisionID {
			return nil, entity.ErrConsistency
		}
	}

	chosen, assetStatus, err := selectProjectedRevision(cap, revisions)
	if err != nil {
		return nil, err
	}
	digest, err := CapabilityContentDigest(chosen.ToSemanticPayload())
	if err != nil {
		return nil, err
	}
	return &ProjectionResult{
		Name:            chosen.Name,
		SemanticVersion: fmt.Sprintf("0.%d.0", chosen.Version),
		ContentDigest:   digest,
		Status:          assetStatus,
	}, nil
}

func selectProjectedRevision(cap *entity.BusinessCapability, revisions []*entity.BusinessCapabilityRevision) (*entity.BusinessCapabilityRevision, assetentity.AssetStatus, error) {
	if cap.ActiveRevisionID != "" {
		for _, r := range revisions {
			if r.RevisionID == cap.ActiveRevisionID && r.Status == entity.RevisionActive {
				return r, assetentity.AssetStatusReleased, nil
			}
		}
		return nil, "", entity.ErrConsistency
	}

	if r := maxVersionWithStatus(revisions, entity.RevisionValidated); r != nil {
		return r, assetentity.AssetStatusVerified, nil
	}
	if r := maxVersionWithStatus(revisions, entity.RevisionDraft); r != nil {
		return r, assetentity.AssetStatusDraft, nil
	}
	if r := maxVersionWithStatus(revisions, entity.RevisionStale); r != nil {
		return r, assetentity.AssetStatusInReview, nil
	}
	if r := maxVersionWithStatus(revisions, entity.RevisionDeprecated); r != nil {
		return r, assetentity.AssetStatusDeprecated, nil
	}
	return nil, "", entity.ErrConsistency
}

func maxVersionWithStatus(revisions []*entity.BusinessCapabilityRevision, status entity.RevisionStatus) *entity.BusinessCapabilityRevision {
	var best *entity.BusinessCapabilityRevision
	for _, r := range revisions {
		if r.Status != status {
			continue
		}
		if best == nil || r.Version > best.Version {
			best = r
		}
	}
	return best
}
