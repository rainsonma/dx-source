package commands

import (
	"dx-api/app/consts"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
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
	// Placeholder — filled in by Task 5.
	ctx.Info("backfill-metas: not implemented yet")
	return nil
}

// deriveSourceType maps a content_items.content_type to the corresponding
// content_metas.source_type per the backfill rule:
//   sentence → sentence (complete sentence)
//   word, phrase, block → vocab (all non-complete units)
func deriveSourceType(contentType string) string {
	if contentType == consts.ContentTypeSentence {
		return consts.SourceTypeSentence
	}
	return consts.SourceTypeVocab
}
