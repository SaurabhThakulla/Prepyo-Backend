-- EPS-TOPIK becomes a third exam.
--
-- The Employment Permit System Test of Proficiency in Korean is what a worker
-- sits for the E-9 visa. HRD Korea's published standard (출제기준) and its
-- 2025 sample book fix the shape: reading then listening, 20 four-option
-- questions each, 25 minutes each, 100 points in all. There is no speaking or
-- writing section.
--
-- Every exam column is guarded by a CHECK that lists PTE and IELTS. Postgres
-- stores each one as ARRAY['PTE'::text, 'IELTS'::text], so they are rewritten
-- by finding that exact array rather than by naming thirteen constraints: a
-- guard added by an earlier migration and missed here would otherwise reject
-- the first EPS row long after this file ran. Checks that mention IELTS on its
-- own (writing_mock_sessions is IELTS-only by design) do not contain the array
-- and are left alone.

DO $$
DECLARE
    c record;
BEGIN
    FOR c IN
        SELECT conrelid::regclass AS tbl, conname, pg_get_constraintdef(oid) AS def
        FROM pg_constraint
        WHERE contype = 'c'
          AND connamespace = 'public'::regnamespace
          AND pg_get_constraintdef(oid) LIKE '%ARRAY[''PTE''::text, ''IELTS''::text]%'
    LOOP
        EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I', c.tbl, c.conname);
        EXECUTE format('ALTER TABLE %s ADD CONSTRAINT %I %s', c.tbl, c.conname,
            replace(c.def,
                'ARRAY[''PTE''::text, ''IELTS''::text]',
                'ARRAY[''PTE''::text, ''IELTS''::text, ''EPS_TOPIK''::text]'));
    END LOOP;
END $$;

-- The scale is the published one: 40 items at 2.5 points. A practice estimate
-- lands on the same steps a real score can take.
INSERT INTO exam_versions (id, exam, label, description, min_score, max_score, score_step, is_current) VALUES
    ('eps-topik-2026-01', 'EPS_TOPIK', 'EPS-TOPIK 2026.1',
     'Employment Permit System Test of Proficiency in Korean.', 0, 100, 2.5, TRUE)
ON CONFLICT (id) DO NOTHING;
