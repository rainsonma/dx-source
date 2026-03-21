package constants

// Game status values.
const (
	GameStatusDraft     = "draft"
	GameStatusPublished = "published"
	GameStatusWithdraw  = "withdraw"
)

// GameStatusLabels maps each status to its Chinese label.
var GameStatusLabels = map[string]string{
	GameStatusDraft:     "草稿",
	GameStatusPublished: "已发布",
	GameStatusWithdraw:  "已撤回",
}

// GameStatusOption represents a selectable game status.
type GameStatusOption struct {
	Value string
	Label string
}

// GameStatusOptions returns all game statuses as an ordered slice.
func GameStatusOptions() []GameStatusOption {
	return []GameStatusOption{
		{Value: GameStatusDraft, Label: "草稿"},
		{Value: GameStatusPublished, Label: "已发布"},
		{Value: GameStatusWithdraw, Label: "已撤回"},
	}
}
