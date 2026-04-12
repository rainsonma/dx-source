package realtime

import (
	"context"

	"dx-api/app/consts"

	"github.com/goravel/framework/facades"
)

// Authorizer decides whether a user may subscribe to a topic.
type Authorizer struct {
	// isPkParticipant returns true if userID is a participant in pkID.
	// Separated for testability; production wires it to a DB query.
	isPkParticipant func(ctx context.Context, userID, pkID string) (bool, error)

	// isGroupMember returns true if userID is a current member of groupID.
	isGroupMember func(ctx context.Context, userID, groupID string) (bool, error)
}

// NewAuthorizer returns an Authorizer wired to production DB queries.
func NewAuthorizer() *Authorizer {
	return &Authorizer{
		isPkParticipant: pkParticipantQuery,
		isGroupMember:   groupMemberQuery,
	}
}

// AuthorizeSubscribe checks whether userID may subscribe to topic.
// Returns nil on success or a realtimeError with consts.Code* and message
// on failure.
func (a *Authorizer) AuthorizeSubscribe(ctx context.Context, userID, topic string) error {
	parsed, err := ParseTopic(topic)
	if err != nil {
		return realtimeError{Code: consts.CodeInvalidTopic, Message: "invalid topic"}
	}
	switch parsed.Kind {
	case KindUser, KindUserKick:
		if parsed.ID != userID {
			return realtimeError{Code: consts.CodeForbidden, Message: "forbidden"}
		}
		return nil
	case KindPk:
		ok, err := a.isPkParticipant(ctx, userID, parsed.ID)
		if err != nil {
			return realtimeError{Code: consts.CodeForbidden, Message: "forbidden"}
		}
		if !ok {
			return realtimeError{Code: consts.CodeForbidden, Message: "not a participant in this PK"}
		}
		return nil
	case KindGroup, KindGroupNotify:
		ok, err := a.isGroupMember(ctx, userID, parsed.ID)
		if err != nil {
			return realtimeError{Code: consts.CodeGroupForbidden, Message: "您不在该群组中"}
		}
		if !ok {
			return realtimeError{Code: consts.CodeGroupForbidden, Message: "您不在该群组中"}
		}
		return nil
	default:
		return realtimeError{Code: consts.CodeInvalidTopic, Message: "unknown topic kind"}
	}
}

// pkParticipantQuery checks whether userID is the initiator (user_id) or
// opponent (opponent_id) of the given pkID in the game_pks table.
func pkParticipantQuery(ctx context.Context, userID, pkID string) (bool, error) {
	count, err := facades.Orm().WithContext(ctx).Query().
		Table("game_pks").
		Where("id = ?", pkID).
		Where("(user_id = ? OR opponent_id = ?)", userID, userID).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// groupMemberQuery checks whether userID is a current member of game_group_id
// in the game_group_members table.
func groupMemberQuery(ctx context.Context, userID, groupID string) (bool, error) {
	count, err := facades.Orm().WithContext(ctx).Query().
		Table("game_group_members").
		Where("game_group_id = ?", groupID).
		Where("user_id = ?", userID).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
