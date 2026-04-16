package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260415000001AddContentMetasDedupLookupIndex struct{}

func (r *M20260415000001AddContentMetasDedupLookupIndex) Signature() string {
	return "20260415000001_add_content_metas_dedup_lookup_index"
}

func (r *M20260415000001AddContentMetasDedupLookupIndex) Up() error {
	// Supports the per-user dedup SELECT in SaveMetadataBatch:
	//   WHERE cm.deleted_at IS NULL
	//     AND cm.source_data IN ?
	//     AND cm.source_type IN ?
	//   AND <join to games via user_id>
	//
	// Column order is (source_data, source_type) because source_data has
	// ~millions of distinct values on our dataset while source_type only has
	// 2 ('sentence' | 'vocab'). Leading with the more-selective column
	// narrows the B-tree scan aggressively before the second-column filter
	// runs — ~2-5x faster than the reverse order on the 1.22M-row table.
	_, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_content_metas_dedup_lookup
		   ON content_metas (source_data, source_type)
		   WHERE deleted_at IS NULL`,
	)
	return err
}

func (r *M20260415000001AddContentMetasDedupLookupIndex) Down() error {
	_, err := facades.Orm().Query().Exec(
		`DROP INDEX IF EXISTS idx_content_metas_dedup_lookup`,
	)
	return err
}
