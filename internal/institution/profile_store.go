package institution

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// saveProfile stores the setup answers; the first save marks the setup finished.
func (s *store) saveProfile(ctx context.Context, institutionID string, p Profile) (time.Time, error) {
	var onboardedAt time.Time
	err := s.pool.QueryRow(ctx, `
		UPDATE institutions
		SET exams = $2, city = $3, branches = $4, monthly_students = $5, brand_color = $6,
			onboarded_at = COALESCE(onboarded_at, now())
		WHERE id = $1
		RETURNING onboarded_at`,
		institutionID, p.Exams, p.City, p.Branches, p.MonthlyStudents, p.BrandColor).Scan(&onboardedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("save institution profile: %w", err)
	}
	return onboardedAt, nil
}

// setLogo stores or replaces the institution's logo.
func (s *store) setLogo(ctx context.Context, institutionID, contentType string, data []byte) (time.Time, error) {
	var updatedAt time.Time
	err := s.pool.QueryRow(ctx, `
		INSERT INTO institution_logos (institution_id, content_type, data) VALUES ($1, $2, $3)
		ON CONFLICT (institution_id) DO UPDATE
			SET content_type = EXCLUDED.content_type, data = EXCLUDED.data, updated_at = now()
		RETURNING updated_at`,
		institutionID, contentType, data).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("save institution logo: %w", err)
	}
	return updatedAt, nil
}

// logo finds the logo of the institution at /<code>.
func (s *store) logo(ctx context.Context, code string) (string, []byte, error) {
	var contentType string
	var data []byte
	err := s.pool.QueryRow(ctx, `
		SELECT l.content_type, l.data
		FROM institution_logos l
		JOIN institutions i ON i.id = l.institution_id
		WHERE i.code = $1`, code).Scan(&contentType, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, errNotFound
	}
	if err != nil {
		return "", nil, fmt.Errorf("find institution logo: %w", err)
	}
	return contentType, data, nil
}

// brand is what the student sign-in page shows: name, colour and logo.
func (s *store) brand(ctx context.Context, code string) (name, color string, logoUpdatedAt *time.Time, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT i.name, i.brand_color, l.updated_at
		FROM institutions i
		LEFT JOIN institution_logos l ON l.institution_id = i.id
		WHERE i.code = $1`, code).Scan(&name, &color, &logoUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil, errNotFound
	}
	if err != nil {
		return "", "", nil, fmt.Errorf("find institution: %w", err)
	}
	return name, color, logoUpdatedAt, nil
}
