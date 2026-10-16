DROP TABLE IF EXISTS institution_logos;
ALTER TABLE institutions
    DROP COLUMN IF EXISTS exams,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS branches,
    DROP COLUMN IF EXISTS monthly_students,
    DROP COLUMN IF EXISTS brand_color,
    DROP COLUMN IF EXISTS onboarded_at;
