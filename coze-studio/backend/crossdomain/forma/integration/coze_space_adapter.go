/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import (
	"context"
	"fmt"

	crossuser "github.com/coze-dev/coze-studio/backend/crossdomain/user"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

type cozeSpaceAdapter struct{}

func NewCozeSpaceAdapter() CozeSpaceAdapter {
	return &cozeSpaceAdapter{}
}

func (a *cozeSpaceAdapter) DescribeSpace(ctx context.Context, spaceID int64) (*SpaceDescribeResult, error) {
	if spaceID == 0 {
		return nil, formaerrors.SpaceNotFound("space_id is required")
	}
	svc := crossuser.DefaultSVC()
	if svc == nil {
		return &SpaceDescribeResult{
			SpaceID:   spaceID,
			Available: false,
			Message:   "coze user crossdomain service not initialized",
		}, nil
	}
	spaces, err := svc.GetUserSpaceBySpaceID(ctx, []int64{spaceID})
	if err != nil {
		return nil, err
	}
	if len(spaces) == 0 || spaces[0] == nil {
		return &SpaceDescribeResult{
			SpaceID:   spaceID,
			Available: false,
			Message:   "space not found",
		}, nil
	}
	sp := spaces[0]
	return &SpaceDescribeResult{
		SpaceID:     sp.ID,
		Name:        sp.Name,
		Description: sp.Description,
		OwnerID:     sp.OwnerID,
		Available:   true,
		Message:     "ok",
	}, nil
}

func (a *cozeSpaceAdapter) ValidateSpaceAccess(ctx context.Context, cozeUserID, spaceID int64) error {
	if cozeUserID == 0 {
		return formaerrors.Unauthenticated("coze user id is required")
	}
	if spaceID == 0 {
		return formaerrors.SpaceNotFound("space_id is required")
	}
	svc := crossuser.DefaultSVC()
	if svc == nil {
		return fmt.Errorf("coze user crossdomain service not initialized")
	}
	spaces, err := svc.GetUserSpaceList(ctx, cozeUserID)
	if err != nil {
		return err
	}
	for _, sp := range spaces {
		if sp != nil && sp.ID == spaceID {
			return nil
		}
	}
	return formaerrors.SpaceForbidden("user does not have access to space")
}
