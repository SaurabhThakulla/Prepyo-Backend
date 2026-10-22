package questions

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/prepyo/backend/internal/media"
)

// AssetKey is where a question file lives in the media bucket.
func AssetKey(id string) string { return "question-assets/" + id }

// MoveAssetsToStore copies every question file still held in the database to
// the media bucket, reads it back to check it arrived whole, and only then
// clears the bytes from the database. With write false it only counts.
// It can be stopped and run again: each file is moved on its own.
func MoveAssetsToStore(ctx context.Context, db *pgxpool.Pool, store media.Store, write bool, log *slog.Logger) (files int, size int64, err error) {
	after := ""
	for {
		rows, err := db.Query(ctx, `
			SELECT id, content_type, data FROM question_assets
			 WHERE data IS NOT NULL AND id > $1 ORDER BY id LIMIT 50`, after)
		if err != nil {
			return files, size, err
		}
		type asset struct {
			id, contentType string
			data            []byte
		}
		var batch []asset
		for rows.Next() {
			var a asset
			if err := rows.Scan(&a.id, &a.contentType, &a.data); err != nil {
				rows.Close()
				return files, size, err
			}
			batch = append(batch, a)
		}
		rows.Close()
		if len(batch) == 0 {
			return files, size, rows.Err()
		}
		for _, a := range batch {
			after = a.id
			files++
			size += int64(len(a.data))
			if !write {
				continue
			}
			key := AssetKey(a.id)
			if err := store.Put(ctx, key, a.contentType, a.data); err != nil {
				return files, size, fmt.Errorf("upload %s: %w", a.id, err)
			}
			if err := sameAsStored(ctx, store, key, a.data); err != nil {
				return files, size, fmt.Errorf("check %s: %w", a.id, err)
			}
			if _, err := db.Exec(ctx, `
				UPDATE question_assets SET storage_key = $2, data = NULL
				 WHERE id = $1 AND data IS NOT NULL`, a.id, key); err != nil {
				return files, size, fmt.Errorf("record %s: %w", a.id, err)
			}
		}
		log.Info("question assets", "moved", files, "megabytes", size/1e6, "write", write)
	}
}

// RestoreAssetsFromStore is the undo: it copies files back from the bucket into
// the database (the bucket copy is kept). With write false it only counts.
func RestoreAssetsFromStore(ctx context.Context, db *pgxpool.Pool, store media.Store, write bool, log *slog.Logger) (files int, err error) {
	rows, err := db.Query(ctx, `SELECT id, storage_key FROM question_assets WHERE data IS NULL AND storage_key IS NOT NULL ORDER BY id`)
	if err != nil {
		return 0, err
	}
	type moved struct{ id, key string }
	var all []moved
	for rows.Next() {
		var m moved
		if err := rows.Scan(&m.id, &m.key); err != nil {
			rows.Close()
			return 0, err
		}
		all = append(all, m)
	}
	rows.Close()
	for _, m := range all {
		files++
		if !write {
			continue
		}
		body, _, err := store.Get(ctx, m.key)
		if err != nil {
			return files, fmt.Errorf("download %s: %w", m.id, err)
		}
		data, err := io.ReadAll(body)
		body.Close()
		if err != nil {
			return files, fmt.Errorf("download %s: %w", m.id, err)
		}
		tag, err := db.Exec(ctx, `
			UPDATE question_assets SET data = $2, storage_key = NULL
			 WHERE id = $1 AND byte_size = $3`, m.id, data, len(data))
		if err != nil {
			return files, fmt.Errorf("restore %s: %w", m.id, err)
		}
		if tag.RowsAffected() == 0 {
			return files, fmt.Errorf("restore %s: bucket copy is %d bytes, not the recorded size", m.id, len(data))
		}
		if files%50 == 0 {
			log.Info("question assets restored", "files", files)
		}
	}
	return files, nil
}

// sameAsStored reads a file back from the bucket and compares it byte for byte.
func sameAsStored(ctx context.Context, store media.Store, key string, want []byte) error {
	body, _, err := store.Get(ctx, key)
	if err != nil {
		return err
	}
	defer body.Close()
	got, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("bucket copy differs (%d bytes, want %d)", len(got), len(want))
	}
	return nil
}
