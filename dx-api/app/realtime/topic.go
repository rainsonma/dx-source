package realtime

import (
	"errors"
	"strings"
)

type TopicKind int

const (
	KindUnknown TopicKind = iota
	KindUser
	KindUserKick
	KindPk
	KindGroup
	KindGroupNotify
)

type ParsedTopic struct {
	Kind TopicKind
	ID   string
}

var ErrBadTopic = errors.New("realtime: bad topic format")

func UserTopic(userID string) string     { return "user:" + userID }
func UserKickTopic(userID string) string { return "user:" + userID + ":kick" }
func PkTopic(pkID string) string         { return "pk:" + pkID }
func GroupTopic(groupID string) string   { return "group:" + groupID }
func GroupNotifyTopic(groupID string) string {
	return "group:" + groupID + ":notify"
}

// ParseTopic decomposes a topic string into its Kind and the entity ID.
// Rejects malformed or unknown topics.
func ParseTopic(topic string) (ParsedTopic, error) {
	parts := strings.Split(topic, ":")
	switch {
	case len(parts) == 2 && parts[0] == "user" && parts[1] != "":
		return ParsedTopic{Kind: KindUser, ID: parts[1]}, nil
	case len(parts) == 3 && parts[0] == "user" && parts[1] != "" && parts[2] == "kick":
		return ParsedTopic{Kind: KindUserKick, ID: parts[1]}, nil
	case len(parts) == 2 && parts[0] == "pk" && parts[1] != "":
		return ParsedTopic{Kind: KindPk, ID: parts[1]}, nil
	case len(parts) == 2 && parts[0] == "group" && parts[1] != "":
		return ParsedTopic{Kind: KindGroup, ID: parts[1]}, nil
	case len(parts) == 3 && parts[0] == "group" && parts[1] != "" && parts[2] == "notify":
		return ParsedTopic{Kind: KindGroupNotify, ID: parts[1]}, nil
	default:
		return ParsedTopic{}, ErrBadTopic
	}
}
