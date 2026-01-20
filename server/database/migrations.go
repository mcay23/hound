package database

import (
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

// Only postgresql is supported for now
// IDs Should follow YYYYMMDD_id_description
var addForeignKeys = &migrate.Migration{
	// xorm doesn't support foreign keys automatically
	ID: "20250101_02_add_foreign_keys",
	Migrate: func(tx *xorm.Engine) error {
		// drop first if exists, safer
		query :=
			`
			ALTER TABLE collections DROP CONSTRAINT IF EXISTS fk_collections_user_id;
			ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_user_id;
			ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_record_id;
			ALTER TABLE collection_relations DROP CONSTRAINT IF EXISTS fk_collection_relations_user_id;
			ALTER TABLE collection_relations DROP CONSTRAINT IF EXISTS fk_collection_relations_record_id;
			ALTER TABLE collection_relations DROP CONSTRAINT IF EXISTS fk_collection_relations_collection_id;
			ALTER TABLE media_files DROP CONSTRAINT IF EXISTS fk_media_files_record_id;

			ALTER TABLE collections
			ADD CONSTRAINT fk_collections_user_id
				FOREIGN KEY (owner_user_id) REFERENCES users (user_id)
				ON UPDATE CASCADE ON DELETE SET NULL;

			ALTER TABLE comments
			ADD CONSTRAINT fk_comments_user_id
				FOREIGN KEY (user_id) REFERENCES users (user_id)
				ON UPDATE CASCADE ON DELETE CASCADE,
			ADD CONSTRAINT fk_comments_record_id
				FOREIGN KEY (record_id) REFERENCES media_records (record_id)
				ON UPDATE CASCADE ON DELETE SET NULL;

			ALTER TABLE collection_relations
			ADD CONSTRAINT fk_collection_relations_user_id
				FOREIGN KEY (user_id) REFERENCES users (user_id)
				ON UPDATE CASCADE ON DELETE CASCADE,
			ADD CONSTRAINT fk_collection_relations_record_id
				FOREIGN KEY (record_id) REFERENCES media_records (record_id)
				ON UPDATE CASCADE ON DELETE CASCADE,
			ADD CONSTRAINT fk_collection_relations_collection_id
				FOREIGN KEY (collection_id) REFERENCES collections (collection_id)
				ON UPDATE CASCADE ON DELETE CASCADE;
				
			ALTER TABLE media_files
			ADD CONSTRAINT fk_media_files_record_id
				FOREIGN KEY (record_id) REFERENCES media_records (record_id)
				ON UPDATE CASCADE ON DELETE CASCADE;
			`
		_, err := tx.Exec(query)
		return err
	},
	Rollback: func(tx *xorm.Engine) error {
		sql := `
		ALTER TABLE collections DROP CONSTRAINT IF EXISTS fk_collections_user_id;
		ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_user_id;
		ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_record_id;
		ALTER TABLE collection_relations DROP CONSTRAINT IF EXISTS fk_collection_relations_user_id;
		ALTER TABLE collection_relations DROP CONSTRAINT IF EXISTS fk_collection_relations_record_id;
		ALTER TABLE collection_relations DROP CONSTRAINT IF EXISTS fk_collection_relations_collection_id;
		ALTER TABLE media_files DROP CONSTRAINT IF EXISTS fk_media_files_record_id;
		`
		_, err := tx.Exec(sql)
		return err
	},
}

// Indexes with asc/desc specified
var addComplexIndexes = &migrate.Migration{
	ID: "20250101_02_add_complex_indexes",
	Migrate: func(tx *xorm.Engine) error {
		query := `
			CREATE INDEX IF NOT EXISTS idx_watch_events_history ON watch_events (rewatch_id, watched_at DESC);
			CREATE INDEX IF NOT EXISTS idx_rewatches_user_record ON rewatches (user_id, record_id, started_at DESC);
			CREATE INDEX IF NOT EXISTS idx_ingest_tasks_status_id ON ingest_tasks (status, ingest_task_id ASC);
			CREATE INDEX IF NOT EXISTS idx_collection_relations_user_record_created ON collection_relations (user_id, record_id, created_at DESC);
		`
		_, err := tx.Exec(query)
		return err
	},
	Rollback: func(tx *xorm.Engine) error {
		query := `
			DROP INDEX IF EXISTS idx_watch_events_history;
			DROP INDEX IF EXISTS idx_rewatches_user_record;
			DROP INDEX IF EXISTS idx_ingest_tasks_status_id;
			DROP INDEX IF EXISTS idx_collection_relations_user_record_created;
		`
		_, err := tx.Exec(query)
		return err
	},
}

func runMigrations() error {
	m := migrate.New(databaseEngine, migrate.DefaultOptions, []*migrate.Migration{
		addForeignKeys,
		addComplexIndexes,
	})
	if err := m.Migrate(); err != nil {
		return err
	}
	return nil
}
