package commands

import (
	"fmt"
	"time"

	"dx-api/app/facades"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type UpdatePlayStreaks struct {
}

// Signature returns the unique signature of the command.
func (c *UpdatePlayStreaks) Signature() string {
	return "app:update-play-streaks"
}

// Description returns the console command description.
func (c *UpdatePlayStreaks) Description() string {
	return "Update play streaks for all users based on last_played_at"
}

// Extend returns the command extend options.
func (c *UpdatePlayStreaks) Extend() command.Extend {
	return command.Extend{}
}

// Handle executes the console command.
func (c *UpdatePlayStreaks) Handle(ctx console.Context) error {
	start := time.Now()

	// 1. Streak continues: played yesterday → increment streak, update max
	continued, err := updateContinuedStreaks()
	if err != nil {
		ctx.Error(fmt.Sprintf("failed to update continued streaks: %v", err))
		return err
	}

	// 2. Streak broken: played before yesterday → reset to 1 (not 0)
	reset, err := resetBrokenStreaks()
	if err != nil {
		ctx.Error(fmt.Sprintf("failed to reset broken streaks: %v", err))
		return err
	}

	// 3. Played today: skip (no update needed)

	elapsed := time.Since(start)
	ctx.Info(fmt.Sprintf("[play-streaks] done in %s — continued: %d, reset: %d", elapsed, continued, reset))
	return nil
}

func updateContinuedStreaks() (int64, error) {
	res, err := facades.Orm().Query().Exec(`
		UPDATE users
		SET current_play_streak = current_play_streak + 1,
		    max_play_streak = GREATEST(current_play_streak + 1, max_play_streak),
		    updated_at = now()
		WHERE last_played_at IS NOT NULL
		  AND last_played_at::date = CURRENT_DATE - INTERVAL '1 day'
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to update continued streaks: %w", err)
	}
	return res.RowsAffected, nil
}

func resetBrokenStreaks() (int64, error) {
	res, err := facades.Orm().Query().Exec(`
		UPDATE users
		SET current_play_streak = 1,
		    updated_at = now()
		WHERE last_played_at IS NOT NULL
		  AND last_played_at::date < CURRENT_DATE - INTERVAL '1 day'
		  AND current_play_streak != 1
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to reset broken streaks: %w", err)
	}
	return res.RowsAffected, nil
}
