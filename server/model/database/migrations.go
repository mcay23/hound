package database

import (
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

// Only postgresql is supported for now
// IDs Should follow YYYYMMDD_id_description
var addForeignKeys = &migrate.Migration{
	// xorm doesn't support foreign keys automatically
	ID: "20250101_01_add_foreign_keys_test",
	Migrate: func(tx *xorm.Engine) error {
		query :=
			`
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
			`
		_, err := tx.Exec(query)
		return err
	},
	Rollback: func(tx *xorm.Engine) error {
		sql := `
		ALTER TABLE collections DROP CONSTRAINT fk_collection_user;
		ALTER TABLE comments DROP CONSTRANT fk_comment_user;
		`
		_, err := tx.Exec(sql)
		return err
	},
}

func runMigrations() error {
	m := migrate.New(databaseEngine, migrate.DefaultOptions, []*migrate.Migration{addForeignKeys})
	if err := m.Migrate(); err != nil {
		return err
	}
	return nil
}
