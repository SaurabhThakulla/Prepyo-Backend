-- Sub-test metering: one daily allowance that covers the whole product.
--
-- Why this exists. `ai_evaluations_per_day` metered writing and speaking only,
-- because those are the two skills that cost money per provider call. Reading
-- and listening practice was not metered at all — internal/practice/handler.go
-- had no quota check — so a free learner could practise reading forever while
-- the pricing table advertised "3 AI evaluations per day" as though it were a
-- cap on practice.
--
-- The allowance now counts **sub-tests**: one attempted task set, in any skill.
-- A True/False set of six statements is one, a Fill in the Blanks text is one,
-- an essay is one, a speaking recording is one. Mocks are not sub-tests; they
-- keep their own allowance.
--
-- The column is renamed rather than reused because the meaning genuinely
-- changed. Leaving it called ai_evaluations_per_day while it governs all
-- practice would mislead every future reader.

ALTER TABLE plans RENAME COLUMN ai_evaluations_per_day TO sub_tests_per_day;

-- ---------------------------------------------------------------------------
-- Plan copy and quotas
-- ---------------------------------------------------------------------------
--
-- Prices, ids, duration_days, bonus_days, duration_months and sort_order are
-- deliberately untouched. Ids in particular are foreign-keyed from users.plan_id
-- and subscription_payments.plan_id.
--
-- The names are the learner's own journey: Suru (start), Abhyas (practice),
-- Taiyari (preparation), Udaan (take-off).
--
-- features[] had drifted badly. 000003 repriced and re-quota'd every plan but
-- never rewrote the strings seeded in 000002, so `pro` still advertised 30
-- evaluations and 5 mocks against an enforced 15 and 3, and `elite` advertised
-- 100 and 15 against 25 and 10. Those strings are rendered verbatim as the
-- feature bullets on the page where a learner buys. The guard at the bottom of
-- this file is what stops it happening again.
--
-- Mock periods are stated explicitly because the two plans count differently:
-- free counts mocks lifetime, paid plans count them per calendar month. See
-- internal/billing/billing.go State().

UPDATE plans SET
    name              = 'Suru',
    sub_tests_per_day = 5,
    features          = ARRAY[
        'Core practice tasks',
        '5 practice sub-tests per day',
        '1 full mock exam (lifetime)'
    ]
WHERE id = 'free';

UPDATE plans SET
    name                = 'Abhyas',
    sub_tests_per_day   = 40,
    mock_tests_included = 2,
    features            = ARRAY[
        '7 days intensive sprint',
        '40 practice sub-tests per day',
        '2 full mock exams per calendar month',
        'All practice question types'
    ]
WHERE id = 'weekly';

UPDATE plans SET
    name                = 'Taiyari',
    sub_tests_per_day   = 50,
    mock_tests_included = 5,
    features            = ARRAY[
        '33 days full access',
        '50 practice sub-tests per day',
        '5 full mock exams per calendar month',
        'Sentence-level rewrites'
    ]
WHERE id = 'pro';

UPDATE plans SET
    name              = 'Udaan',
    sub_tests_per_day = 60,
    features          = ARRAY[
        '97 days full access',
        '60 practice sub-tests per day',
        '10 full mock exams per calendar month',
        'Priority evaluation queue'
    ]
WHERE id = 'elite';

-- ---------------------------------------------------------------------------
-- Guard
-- ---------------------------------------------------------------------------
--
-- Semantic, not "does the number appear anywhere". The line that *describes*
-- sub-tests must state that plan's sub_tests_per_day, and the line that
-- describes mocks must state its mock_tests_included. A number that happens to
-- appear in an unrelated bullet does not satisfy it.
--
-- \m and \M are Postgres word boundaries, so the enforced 5 of a mock allowance
-- is not matched by the 50 in "50 practice sub-tests per day".

DO $$
DECLARE
    p         RECORD;
    sub_line  TEXT;
    mock_line TEXT;
BEGIN
    FOR p IN SELECT id, sub_tests_per_day, mock_tests_included, features FROM plans LOOP
        SELECT f INTO sub_line  FROM unnest(p.features) f WHERE f ILIKE '%sub-test%' LIMIT 1;
        SELECT f INTO mock_line FROM unnest(p.features) f WHERE f ILIKE '%mock%'     LIMIT 1;

        IF sub_line IS NULL THEN
            RAISE EXCEPTION 'plan %: no feature describes sub-tests', p.id;
        END IF;
        IF mock_line IS NULL THEN
            RAISE EXCEPTION 'plan %: no feature describes mock exams', p.id;
        END IF;

        IF sub_line !~ ('\m' || p.sub_tests_per_day::text || '\M') THEN
            RAISE EXCEPTION 'plan %: sub-test feature "%" does not state the enforced limit of %',
                p.id, sub_line, p.sub_tests_per_day;
        END IF;
        IF mock_line !~ ('\m' || p.mock_tests_included::text || '\M') THEN
            RAISE EXCEPTION 'plan %: mock feature "%" does not state the enforced allowance of %',
                p.id, mock_line, p.mock_tests_included;
        END IF;
    END LOOP;
END $$;
