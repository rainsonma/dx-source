package realtime

import (
	"context"
	"errors"
	"testing"

	"dx-api/app/consts"
)

func newTestAuthorizer(
	pkParticipants map[string]map[string]bool,
	groupMembers map[string]map[string]bool,
) *Authorizer {
	return &Authorizer{
		isPkParticipant: func(ctx context.Context, userID, pkID string) (bool, error) {
			return pkParticipants[pkID][userID], nil
		},
		isGroupMember: func(ctx context.Context, userID, groupID string) (bool, error) {
			return groupMembers[groupID][userID], nil
		},
	}
}

func TestAuthorize_UserTopicSelf(t *testing.T) {
	a := newTestAuthorizer(nil, nil)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "user:alice"); err != nil {
		t.Errorf("self user topic: %v", err)
	}
}

func TestAuthorize_UserTopicOther(t *testing.T) {
	a := newTestAuthorizer(nil, nil)
	err := a.AuthorizeSubscribe(context.Background(), "alice", "user:bob")
	if err == nil {
		t.Fatal("expected forbidden")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeForbidden {
		t.Errorf("want CodeForbidden, got %+v", err)
	}
}

func TestAuthorize_UserKickTopicSelf(t *testing.T) {
	a := newTestAuthorizer(nil, nil)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "user:alice:kick"); err != nil {
		t.Errorf("self kick topic: %v", err)
	}
}

func TestAuthorize_PkParticipant(t *testing.T) {
	a := newTestAuthorizer(
		map[string]map[string]bool{"pk_abc": {"alice": true}},
		nil,
	)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "pk:pk_abc"); err != nil {
		t.Errorf("pk participant: %v", err)
	}
}

func TestAuthorize_PkNonParticipant(t *testing.T) {
	a := newTestAuthorizer(
		map[string]map[string]bool{"pk_abc": {"alice": true}},
		nil,
	)
	err := a.AuthorizeSubscribe(context.Background(), "bob", "pk:pk_abc")
	if err == nil {
		t.Fatal("expected forbidden")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeForbidden {
		t.Errorf("want CodeForbidden, got %+v", err)
	}
}

func TestAuthorize_GroupMember(t *testing.T) {
	a := newTestAuthorizer(
		nil,
		map[string]map[string]bool{"grp_xyz": {"alice": true}},
	)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "group:grp_xyz"); err != nil {
		t.Errorf("group member: %v", err)
	}
}

func TestAuthorize_GroupNonMember(t *testing.T) {
	a := newTestAuthorizer(
		nil,
		map[string]map[string]bool{"grp_xyz": {"alice": true}},
	)
	err := a.AuthorizeSubscribe(context.Background(), "bob", "group:grp_xyz")
	if err == nil {
		t.Fatal("expected group forbidden")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeGroupForbidden {
		t.Errorf("want CodeGroupForbidden, got %+v", err)
	}
}

func TestAuthorize_GroupNotifyMember(t *testing.T) {
	a := newTestAuthorizer(
		nil,
		map[string]map[string]bool{"grp_xyz": {"alice": true}},
	)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "group:grp_xyz:notify"); err != nil {
		t.Errorf("group notify member: %v", err)
	}
}

func TestAuthorize_GroupNotifyNonMember(t *testing.T) {
	a := newTestAuthorizer(
		nil,
		map[string]map[string]bool{"grp_xyz": {"alice": true}},
	)
	err := a.AuthorizeSubscribe(context.Background(), "bob", "group:grp_xyz:notify")
	if err == nil {
		t.Fatal("expected group forbidden")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeGroupForbidden {
		t.Errorf("want CodeGroupForbidden, got %+v", err)
	}
}

func TestAuthorize_InvalidTopic(t *testing.T) {
	a := newTestAuthorizer(nil, nil)
	err := a.AuthorizeSubscribe(context.Background(), "alice", "garbage")
	if err == nil {
		t.Fatal("expected invalid topic")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeInvalidTopic {
		t.Errorf("want CodeInvalidTopic, got %+v", err)
	}
}

func TestAuthorize_PkQueryError(t *testing.T) {
	a := &Authorizer{
		isPkParticipant: func(ctx context.Context, userID, pkID string) (bool, error) {
			return false, errors.New("db down")
		},
		isGroupMember: func(ctx context.Context, userID, groupID string) (bool, error) {
			return false, nil
		},
	}
	err := a.AuthorizeSubscribe(context.Background(), "alice", "pk:pk_abc")
	if err == nil {
		t.Fatal("expected internal error")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeInternalError {
		t.Errorf("want CodeInternalError, got %+v", err)
	}
}

func TestAuthorize_GroupQueryError(t *testing.T) {
	a := &Authorizer{
		isPkParticipant: func(ctx context.Context, userID, pkID string) (bool, error) {
			return false, nil
		},
		isGroupMember: func(ctx context.Context, userID, groupID string) (bool, error) {
			return false, errors.New("db down")
		},
	}
	err := a.AuthorizeSubscribe(context.Background(), "alice", "group:grp_xyz")
	if err == nil {
		t.Fatal("expected internal error")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeInternalError {
		t.Errorf("want CodeInternalError, got %+v", err)
	}
}
