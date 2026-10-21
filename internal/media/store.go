// Package media keeps learners' files outside Postgres: a private Backblaze B2
// bucket in staging and production, a local folder in development. Speaking
// recordings are the first thing stored here.
package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/prepyo/backend/pkg/config"
)

// ErrNotFound is returned for a key the store does not hold.
var ErrNotFound = errors.New("media: not found")

// Store holds files by key, such as "recordings/<user>/<evaluation>.wav".
type Store interface {
	Put(ctx context.Context, key, contentType string, body []byte) error
	Get(ctx context.Context, key string) (body io.ReadCloser, contentType string, err error)
	Delete(ctx context.Context, key string) error
}

// New picks the B2 bucket, then MEDIA_DIR, and nil (nothing stored) when
// neither is set.
func New(cfg *config.Config) Store {
	switch {
	case cfg.S3Enabled():
		return NewS3(cfg.S3Endpoint, cfg.S3Region, cfg.S3Bucket, cfg.S3AccessKeyID, cfg.S3SecretAccessKey)
	case cfg.MediaDir != "":
		return Dir(cfg.MediaDir)
	}
	return nil
}

// Dir stores files in a local folder, for development without a bucket.
type Dir string

func (d Dir) path(key string) (string, error) {
	clean := path.Clean("/" + key)
	if clean == "/" || strings.Contains(key, "..") {
		return "", fmt.Errorf("media: bad key %q", key)
	}
	return filepath.Join(string(d), filepath.FromSlash(clean[1:])), nil
}

func (d Dir) Put(_ context.Context, key, _ string, body []byte) error {
	p, err := d.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	return os.WriteFile(p, body, 0o600)
}

func (d Dir) Get(_ context.Context, key string) (io.ReadCloser, string, error) {
	p, err := d.path(key)
	if err != nil {
		return nil, "", err
	}
	f, err := os.Open(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}
	return f, mime.TypeByExtension(path.Ext(key)), nil
}

func (d Dir) Delete(_ context.Context, key string) error {
	p, err := d.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
