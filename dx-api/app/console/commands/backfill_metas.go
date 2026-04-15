package commands

import (
	"dx-api/app/consts"
	"dx-api/app/models"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/orm"
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

// backfillRow is a single (content_item, game_item) pair we need to process.
type backfillRow struct {
	CIID        string  `gorm:"column:ci_id"`
	Content     string  `gorm:"column:content"`
	ContentType string  `gorm:"column:content_type"`
	Translation *string `gorm:"column:translation"`
	GameID      string  `gorm:"column:game_id"`
	GameLevelID string  `gorm:"column:game_level_id"`
	GIOrder     float64 `gorm:"column:gi_order"`
}

// loadBackfillChunk selects up to `size` content_items that still need a meta,
// joined with their game_item so we know the target game/level/order.
// Rows are ordered by content_items.id (UUIDv7, time-sortable) so every run
// processes the oldest unlinked rows first.
func loadBackfillChunk(size int) ([]backfillRow, error) {
	var rows []backfillRow
	if err := facades.Orm().Query().Raw(`
		SELECT ci.id AS ci_id,
		       ci.content,
		       ci.content_type,
		       ci.translation,
		       gi.game_id,
		       gi.game_level_id,
		       gi."order" AS gi_order
		FROM content_items ci
		JOIN game_items gi
		  ON gi.content_item_id = ci.id
		 AND gi.deleted_at IS NULL
		WHERE ci.content_meta_id IS NULL
		  AND ci.deleted_at IS NULL
		ORDER BY ci.id
		LIMIT ?
	`, size).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to load chunk: %w", err)
	}
	return rows, nil
}

// bulkLinkItems issues a single UPDATE that sets content_meta_id for an entire
// chunk of rows via a VALUES list. Far faster than per-row UPDATEs.
//
// Goravel uses `?` placeholders (GORM-style) rather than PostgreSQL-native `$N`.
// String values bind as SQL text, so we cast both sides of the join back to
// uuid inside the query.
func bulkLinkItems(tx orm.Query, rows []backfillRow, metaIDs []string) error {
	if len(rows) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(`UPDATE content_items AS ci
	SET content_meta_id = v.meta_id::uuid
	FROM (VALUES `)

	args := make([]any, 0, len(rows)*2)
	for i, r := range rows {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(?, ?)")
		args = append(args, r.CIID, metaIDs[i])
	}
	sb.WriteString(`) AS v(ci_id, meta_id) WHERE ci.id = v.ci_id::uuid`)

	if _, err := tx.Exec(sb.String(), args...); err != nil {
		return fmt.Errorf("bulk link update: %w", err)
	}
	return nil
}

// backfillChunk loads up to `size` unprocessed rows, writes metas/game_metas,
// and links them back on content_items. All three writes run inside a single
// transaction so the state is consistent per chunk. Returns the number of
// rows processed (0 when nothing left to do).
func backfillChunk(size int) (int, error) {
	rows, err := loadBackfillChunk(size)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	// Pre-generate UUIDs outside the transaction so the work on the critical
	// path is purely I/O.
	metaIDs := make([]string, len(rows))
	gameMetaIDs := make([]string, len(rows))
	for i := range rows {
		metaIDs[i] = uuid.Must(uuid.NewV7()).String()
		gameMetaIDs[i] = uuid.Must(uuid.NewV7()).String()
	}

	err = facades.Orm().Transaction(func(tx orm.Query) error {
		// 1. Bulk insert content_metas.
		metas := make([]models.ContentMeta, len(rows))
		for i, r := range rows {
			metas[i] = models.ContentMeta{
				ID:          metaIDs[i],
				SourceFrom:  consts.SourceFromImport,
				SourceType:  deriveSourceType(r.ContentType),
				SourceData:  r.Content,
				Translation: r.Translation,
				IsBreakDone: true,
			}
		}
		if err := tx.Create(&metas); err != nil {
			return fmt.Errorf("insert content_metas: %w", err)
		}

		// 2. Bulk insert game_metas, each pointing at the matching content_meta
		//    with the same (game_id, game_level_id, order) as its game_item.
		gameMetas := make([]models.GameMeta, len(rows))
		for i, r := range rows {
			gameMetas[i] = models.GameMeta{
				ID:            gameMetaIDs[i],
				GameID:        r.GameID,
				GameLevelID:   r.GameLevelID,
				ContentMetaID: metaIDs[i],
				Order:         r.GIOrder,
			}
		}
		if err := tx.Create(&gameMetas); err != nil {
			return fmt.Errorf("insert game_metas: %w", err)
		}

		// 3. Bulk update content_items.content_meta_id in a single UPDATE.
		if err := bulkLinkItems(tx, rows, metaIDs); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}
