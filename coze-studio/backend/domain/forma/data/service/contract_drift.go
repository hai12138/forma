/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
)

func snapshotMatchesBinding(snap *entity.SchemaSnapshot, binding entity.ContractBinding, tenantID string) bool {
	if snap == nil {
		return false
	}
	return snap.TenantID == tenantID &&
		snap.SourceID == binding.SourceID &&
		snap.ConnectionID == binding.ConnectionID &&
		snap.AssetID == binding.AssetID
}

func (s *contractService) EvaluateDrift(ctx context.Context, in *EvaluateDriftInput) (*entity.DataDriftResult, *entity.DataContractRevision, error) {
	if !s.configured() {
		return nil, nil, entity.ErrNotConfigured
	}
	if in == nil || strings.TrimSpace(in.TenantID) == "" || strings.TrimSpace(in.RevisionID) == "" {
		return nil, nil, entity.ErrContractDriftInvalid
	}
	rev, err := s.contracts.GetRevision(ctx, in.TenantID, in.RevisionID)
	if err != nil {
		return nil, nil, err
	}
	if rev.Status != entity.ContractStatusActive && rev.Status != entity.ContractStatusStale {
		return nil, nil, entity.ErrContractInvalidState
	}
	if in.NewSnapshotIDs == nil {
		in.NewSnapshotIDs = map[string]string{}
	}

	findings := []entity.DriftFinding{}
	compared := map[string]string{}
	overall := entity.DriftSeverityNoChange
	fieldsByReq := map[string][]entity.LogicalField{}
	for _, field := range rev.LogicalSchema.Fields {
		reqID := strings.TrimSpace(field.RequirementID)
		fieldsByReq[reqID] = append(fieldsByReq[reqID], field)
	}

	for _, binding := range rev.BindingRefs {
		newSnapID, ok := in.NewSnapshotIDs[binding.SchemaSnapshotID]
		if !ok || strings.TrimSpace(newSnapID) == "" {
			return nil, nil, fmt.Errorf("%w: missing new snapshot for pinned %q", entity.ErrContractDriftInvalid, binding.SchemaSnapshotID)
		}
		compared[binding.SchemaSnapshotID] = newSnapID

		pinned, err := s.sources.GetSnapshot(ctx, in.TenantID, binding.SchemaSnapshotID)
		if err != nil {
			return nil, nil, err
		}
		fresh, err := s.sources.GetSnapshot(ctx, in.TenantID, newSnapID)
		if err != nil {
			return nil, nil, err
		}
		if !snapshotMatchesBinding(pinned, binding, in.TenantID) || !snapshotMatchesBinding(fresh, binding, in.TenantID) {
			return nil, nil, fmt.Errorf("%w: snapshot lineage mismatch for binding %q", entity.ErrContractDriftInvalid, binding.MappingID)
		}
		if pinned.Fingerprint != "" && pinned.Fingerprint == fresh.Fingerprint {
			continue
		}

		m, err := s.mappings.GetMapping(ctx, in.TenantID, binding.MappingID)
		if err != nil {
			return nil, nil, err
		}
		var newSchema entity.PhysicalSchema
		if json.Unmarshal([]byte(fresh.SchemaJSON), &newSchema) != nil {
			return nil, nil, entity.ErrContractDriftInvalid
		}
		newIndex := buildFieldPathIndex(&newSchema)
		bindingSeverity := entity.DriftSeverityCompatible

		for _, path := range m.TargetFieldPaths {
			path = strings.TrimSpace(path)
			if _, ok := newIndex[path]; !ok {
				findings = append(findings, entity.DriftFinding{
					Code: "FIELD_MISSING", Message: fmt.Sprintf("mapped path %q missing in new schema", path),
					BindingMappingID: binding.MappingID, FieldPath: path,
				})
				bindingSeverity = entity.DriftSeverityBreaking
			}
		}

		for _, field := range fieldsByReq[binding.RequirementID] {
			resolved, err := ResolveMappingOutputContractType(m, &newSchema)
			if err != nil || resolved != normalizedType(field.LogicalType) {
				msg := fmt.Sprintf("logical field %q type guarantee lost", field.LogicalKey)
				if err != nil {
					msg = err.Error()
				} else {
					msg = fmt.Sprintf("logical field %q type guarantee lost: resolved %q != %q", field.LogicalKey, resolved, field.LogicalType)
				}
				findings = append(findings, entity.DriftFinding{
					Code: "TYPE_GUARANTEE_LOST", Message: msg,
					BindingMappingID: binding.MappingID, FieldPath: field.LogicalKey,
				})
				bindingSeverity = entity.DriftSeverityBreaking
			}
			if !field.Nullable {
				for _, path := range m.TargetFieldPaths {
					pf, ok := newIndex[strings.TrimSpace(path)]
					if !ok {
						continue
					}
					if pf.Nullable {
						findings = append(findings, entity.DriftFinding{
							Code: "NULLABILITY_GUARANTEE_LOST", Message: fmt.Sprintf("logical field %q nullability guarantee lost on %q", field.LogicalKey, path),
							BindingMappingID: binding.MappingID, FieldPath: path,
						})
						bindingSeverity = entity.DriftSeverityBreaking
					}
				}
			}
		}

		if m.MappingType == entity.MappingTypeJoinRef {
			if err := ValidateJoinRef(m.TransformSpec, &newSchema); err != nil {
				findings = append(findings, entity.DriftFinding{
					Code: "JOIN_REF_BROKEN", Message: err.Error(),
					BindingMappingID: binding.MappingID,
				})
				bindingSeverity = entity.DriftSeverityBreaking
			}
		}
		overall = maxDriftSeverity(overall, bindingSeverity)
	}

	now := time.Now().UTC()
	result := &entity.DataDriftResult{
		DriftResultID: newID("cdrift"), TenantID: in.TenantID, BusinessID: rev.BusinessID,
		ContractID: rev.ContractID, RevisionID: rev.RevisionID, Version: rev.Version,
		Severity: overall, Findings: findings, ComparedSnapshotIDs: compared,
		EvaluatedBy: in.ActorID, EvaluatedAt: now, CreatedAt: now,
	}

	err = s.contracts.Transaction(ctx, func(tx repository.ContractRepository) error {
		if err := tx.CreateDriftResult(ctx, result); err != nil {
			return err
		}
		if overall == entity.DriftSeverityBreaking && rev.Status == entity.ContractStatusActive {
			if _, err := tx.GetContractForUpdate(ctx, in.TenantID, rev.ContractID); err != nil {
				return err
			}
			if err := tx.UpdateRevisionStatus(ctx, in.TenantID, rev.RevisionID, entity.ContractStatusActive, entity.ContractStatusStale); err != nil {
				return err
			}
			if err := tx.ClearContractActiveRevisionIfMatch(ctx, in.TenantID, rev.ContractID, rev.RevisionID); err != nil {
				return err
			}
			if err := tx.CreateLifecycleEvent(ctx, &entity.DataContractLifecycleEvent{
				EventID: newID("clife"), TenantID: in.TenantID, BusinessID: rev.BusinessID,
				ContractID: rev.ContractID, RevisionID: rev.RevisionID, Version: rev.Version,
				Action: entity.LifecycleActionMarkStale, ActorPrincipalID: in.ActorID,
				Reason: "breaking schema drift", CreatedAt: now,
			}); err != nil {
				return err
			}
			rev.Status = entity.ContractStatusStale
			rev.UpdatedAt = now
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return result, rev, nil
}

func maxDriftSeverity(a, b entity.DriftSeverity) entity.DriftSeverity {
	rank := map[entity.DriftSeverity]int{
		entity.DriftSeverityNoChange:   0,
		entity.DriftSeverityCompatible: 1,
		entity.DriftSeverityBreaking:   2,
	}
	if rank[b] > rank[a] {
		return b
	}
	return a
}
