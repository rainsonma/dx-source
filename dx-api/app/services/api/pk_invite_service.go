package api

import (
	"context"
	"fmt"
	"time"

	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
	"dx-api/app/realtime"
)

func levelItemsCount(gameLevelID, degree string) int {
	n, _ := countLevelItems(facades.Orm().Query(), gameLevelID, degree)
	return int(n)
}

type PkInviteResult struct {
	PkID        string `json:"pk_id"`
	SessionID   string `json:"session_id"`
	GameLevelID string `json:"game_level_id"`
}

type PkAcceptResult struct {
	SessionID   string  `json:"session_id"`
	GameID      string  `json:"game_id"`
	GameLevelID string  `json:"game_level_id"`
	Degree      string  `json:"degree"`
	Pattern     *string `json:"pattern"`
}

type PkDetailsResult struct {
	PkID             string  `json:"pk_id"`
	SessionID        string  `json:"session_id"`
	GameID           string  `json:"game_id"`
	GameName         string  `json:"game_name"`
	GameMode         string  `json:"game_mode"`
	LevelID          string  `json:"level_id"`
	LevelName        string  `json:"level_name"`
	Degree           string  `json:"degree"`
	Pattern          *string `json:"pattern"`
	InitiatorID      string  `json:"initiator_id"`
	InitiatorName    string  `json:"initiator_name"`
	OpponentID       string  `json:"opponent_id"`
	OpponentName     string  `json:"opponent_name"`
	InvitationStatus string  `json:"invitation_status"`
}

// InvitePk creates a specified PK and sends an invitation to the opponent.
func InvitePk(userID, gameID, gameLevelID, degree string, pattern *string, opponentID string) (*PkInviteResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	// Verify opponent is still online and VIP
	if !helpers.UserHub.IsOnline(opponentID) {
		return nil, ErrOpponentOffline
	}
	opponentVip, err := IsVipActive(opponentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check opponent VIP: %w", err)
	}
	if !opponentVip {
		return nil, ErrOpponentNotVip
	}

	// End any stale active PK for this user on this game before creating a new one
	cleanupStalePk(facades.Orm().Query(), userID, gameID)

	// Verify game exists and is published
	var game models.Game
	if err := facades.Orm().Query().Where("id", gameID).First(&game); err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}
	if game.ID == "" {
		return nil, ErrGameNotFound
	}
	if game.Status != "published" {
		return nil, ErrGameNotPublished
	}

	// Verify level exists
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, ErrLevelNotFound
	}

	// Get opponent name for SSE event
	var opponent models.User
	if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", opponentID).First(&opponent); err != nil {
		return nil, fmt.Errorf("failed to find opponent: %w", err)
	}

	// Get initiator name for SSE event
	var initiator models.User
	if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&initiator); err != nil {
		return nil, fmt.Errorf("failed to find initiator: %w", err)
	}

	pkID := newID()
	statusPending := consts.PkInvitationPending

	pk := models.GamePk{
		ID:               pkID,
		UserID:           userID,
		OpponentID:       opponentID,
		GameID:           gameID,
		GameLevelID:      gameLevelID,
		Degree:           degree,
		Pattern:          pattern,
		RobotDifficulty:  "",
		IsPlaying:        true,
		PkType:           consts.PkTypeSpecified,
		InvitationStatus: &statusPending,
	}
	if err := facades.Orm().Query().Create(&pk); err != nil {
		return nil, fmt.Errorf("failed to create PK: %w", err)
	}

	// Create initiator's session
	sessionID := newID()
	now := time.Now()
	session := models.GameSession{
		ID:              sessionID,
		UserID:          userID,
		GameID:          gameID,
		GameLevelID:     gameLevelID,
		Degree:          degree,
		Pattern:         pattern,
		GamePkID:        &pkID,
		StartedAt:       now,
		TotalItemsCount: levelItemsCount(gameLevelID, degree),
	}
	if err := facades.Orm().Query().Create(&session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Push invitation SSE event to opponent
	helpers.UserHub.SendToUser(opponentID, "pk_invitation", map[string]string{
		"pk_id":          pkID,
		"game_id":        gameID,
		"game_name":      game.Name,
		"game_mode":      game.Mode,
		"level_name":     level.Name,
		"initiator_id":   userID,
		"initiator_name": nickname(initiator),
	})
	_ = realtime.Publish(context.Background(), realtime.UserTopic(opponentID), realtime.Event{Type: "pk_invitation", Data: map[string]string{
		"pk_id":          pkID,
		"game_id":        gameID,
		"game_name":      game.Name,
		"game_mode":      game.Mode,
		"level_name":     level.Name,
		"initiator_id":   userID,
		"initiator_name": nickname(initiator),
	}})

	return &PkInviteResult{
		PkID:        pkID,
		SessionID:   sessionID,
		GameLevelID: gameLevelID,
	}, nil
}

// AcceptPkInvite accepts an invitation and creates the opponent's session.
func AcceptPkInvite(userID, pkID string) (*PkAcceptResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil {
		return nil, fmt.Errorf("failed to find PK: %w", err)
	}
	if pk.ID == "" {
		return nil, ErrPkNotFound
	}
	if pk.OpponentID != userID {
		return nil, ErrForbidden
	}
	if pk.InvitationStatus == nil || *pk.InvitationStatus != consts.PkInvitationPending {
		return nil, ErrInvitationNotPending
	}

	// Update invitation status
	statusAccepted := consts.PkInvitationAccepted
	if _, err := facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).
		Update("invitation_status", statusAccepted); err != nil {
		return nil, fmt.Errorf("failed to update invitation status: %w", err)
	}

	// Create opponent's session
	sessionID := newID()
	now := time.Now()
	session := models.GameSession{
		ID:              sessionID,
		UserID:          userID,
		GameID:          pk.GameID,
		GameLevelID:     pk.GameLevelID,
		Degree:          pk.Degree,
		Pattern:         pk.Pattern,
		GamePkID:        &pkID,
		StartedAt:       now,
		TotalItemsCount: levelItemsCount(pk.GameLevelID, pk.Degree),
	}
	if err := facades.Orm().Query().Create(&session); err != nil {
		return nil, fmt.Errorf("failed to create opponent session: %w", err)
	}

	// Get opponent name for SSE event
	var opponent models.User
	facades.Orm().Query().Select("id", "username", "nickname").Where("id", userID).First(&opponent)

	// Broadcast accepted event to PK room via PkHub
	helpers.PkHub.Broadcast(pkID, "pk_invitation_accepted", map[string]string{
		"pk_id":         pkID,
		"opponent_id":   userID,
		"opponent_name": nickname(opponent),
	})
	_ = realtime.Publish(context.Background(), realtime.PkTopic(pkID), realtime.Event{Type: "pk_invitation_accepted", Data: map[string]string{
		"pk_id":         pkID,
		"opponent_id":   userID,
		"opponent_name": nickname(opponent),
	}})
	_ = realtime.Publish(context.Background(), realtime.UserTopic(pk.UserID), realtime.Event{Type: "pk_invitation_accepted", Data: map[string]string{
		"pk_id":         pkID,
		"opponent_id":   userID,
		"opponent_name": nickname(opponent),
	}})

	return &PkAcceptResult{
		SessionID:   sessionID,
		GameID:      pk.GameID,
		GameLevelID: pk.GameLevelID,
		Degree:      pk.Degree,
		Pattern:     pk.Pattern,
	}, nil
}

// DeclinePkInvite declines an invitation and ends the initiator's session.
func DeclinePkInvite(userID, pkID string) error {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil {
		return fmt.Errorf("failed to find PK: %w", err)
	}
	if pk.ID == "" {
		return ErrPkNotFound
	}
	if pk.OpponentID != userID {
		return ErrForbidden
	}
	if pk.InvitationStatus == nil || *pk.InvitationStatus != consts.PkInvitationPending {
		return ErrInvitationNotPending
	}

	statusDeclined := consts.PkInvitationDeclined
	now := time.Now()

	// Update PK status
	facades.Orm().Query().Model(&models.GamePk{}).
		Where("id", pkID).
		Update(map[string]any{
			"invitation_status": statusDeclined,
			"is_playing":        false,
		})

	// End initiator's session
	facades.Orm().Query().Exec(
		"UPDATE game_sessions SET ended_at = ? WHERE game_pk_id = ? AND user_id = ? AND ended_at IS NULL",
		now, pkID, pk.UserID)

	// Broadcast declined event
	helpers.PkHub.Broadcast(pkID, "pk_invitation_declined", map[string]string{
		"pk_id": pkID,
	})
	_ = realtime.Publish(context.Background(), realtime.PkTopic(pkID), realtime.Event{Type: "pk_invitation_declined", Data: map[string]string{
		"pk_id": pkID,
	}})
	_ = realtime.Publish(context.Background(), realtime.UserTopic(pk.UserID), realtime.Event{Type: "pk_invitation_declined", Data: map[string]string{
		"pk_id": pkID,
	}})

	return nil
}

// GetPkDetails returns PK information for the room page.
func GetPkDetails(userID, pkID string) (*PkDetailsResult, error) {
	var pk models.GamePk
	if err := facades.Orm().Query().Where("id", pkID).First(&pk); err != nil {
		return nil, fmt.Errorf("failed to find PK: %w", err)
	}
	if pk.ID == "" {
		return nil, ErrPkNotFound
	}
	if pk.UserID != userID && pk.OpponentID != userID {
		return nil, ErrForbidden
	}

	var game models.Game
	facades.Orm().Query().Select("id", "name", "mode").Where("id", pk.GameID).First(&game)

	var level models.GameLevel
	facades.Orm().Query().Select("id", "name").Where("id", pk.GameLevelID).First(&level)

	var initiator, opponent models.User
	facades.Orm().Query().Select("id", "username", "nickname").Where("id", pk.UserID).First(&initiator)
	facades.Orm().Query().Select("id", "username", "nickname").Where("id", pk.OpponentID).First(&opponent)

	// Find the calling user's session for this PK
	var session models.GameSession
	facades.Orm().Query().Select("id").Where("game_pk_id", pk.ID).Where("user_id", userID).First(&session)

	status := ""
	if pk.InvitationStatus != nil {
		status = *pk.InvitationStatus
	}

	return &PkDetailsResult{
		PkID:             pk.ID,
		SessionID:        session.ID,
		GameID:           pk.GameID,
		GameName:         game.Name,
		GameMode:         game.Mode,
		LevelID:          pk.GameLevelID,
		LevelName:        level.Name,
		Degree:           pk.Degree,
		Pattern:          pk.Pattern,
		InitiatorID:      pk.UserID,
		InitiatorName:    nickname(initiator),
		OpponentID:       pk.OpponentID,
		OpponentName:     nickname(opponent),
		InvitationStatus: status,
	}, nil
}
