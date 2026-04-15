package commands

import (
	"dx-api/app/consts"
	"dx-api/app/models"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
)

type BackfillMetas struct{}

func (c *BackfillMetas) Signature() string {
	return "app:backfill-metas"
}

func (c *BackfillMetas) Description() string {
	return "1:1 backfill content_metas and game_metas for imported content_items (source_from=import)"
}

func (c *BackfillMetas) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.IntFlag{
				Name:  "batch-size",
				Value: 5000,
				Usage: "Rows per transaction",
			},
			&command.IntFlag{
				Name:  "limit",
				Value: 0,
				Usage: "Process at most N rows total (0 = no limit)",
			},
			&command.BoolFlag{
				Name:  "dry-run",
				Usage: "Count affected rows without writing",
			},
		},
	}
}

func (c *BackfillMetas) Handle(ctx console.Context) error {
	start := time.Now()
	batchSize := ctx.OptionInt("batch-size")
	limit := ctx.OptionInt("limit")
	dryRun := ctx.OptionBool("dry-run")

	if batchSize <= 0 {
		batchSize = 5000
	}

	total, err := countBackfillCandidates()
	if err != nil {
		return fmt.Errorf("failed to count candidates: %w", err)
	}
	if limit > 0 && int64(limit) < total {
		total = int64(limit)
	}
	ctx.Info(fmt.Sprintf("backfill candidates: %d", total))
	if total == 0 {
		ctx.Info("nothing to backfill")
		return nil
	}
	if dryRun {
		ctx.Info("dry-run — no writes")
		return nil
	}

	// Placeholder — filled in by Task 8.
	ctx.Info(fmt.Sprintf("batch-size=%d (not yet implemented; elapsed %s)", batchSize, time.Since(start)))
	return nil
}

// countBackfillCandidates returns the number of content_items still needing a meta.
func countBackfillCandidates() (int64, error) {
	return facades.Orm().Query().Model(&models.ContentItem{}).
		Where("content_meta_id IS NULL").
		Count()
}

// deriveSourceType maps a content_items.content_type to the corresponding
// content_metas.source_type per the backfill rule:
//
//	sentence → sentence (complete sentence)
//	word, phrase, block → vocab (all non-complete units)
func deriveSourceType(contentType string) string {
	if contentType == consts.ContentTypeSentence {
		return consts.SourceTypeSentence
	}
	return consts.SourceTypeVocab
}
