UPDATE questions SET options = $x$[{"id": "half", "text": "half"}, {"id": "most", "text": "most"}, {"id": "twice", "text": "twice"}, {"id": "all", "text": "all"}]$x$::jsonb WHERE id = 'q-container-001';
UPDATE questions SET options = $x$[{"id": "trucking", "text": "trucking"}, {"id": "shipping", "text": "shipping"}, {"id": "railway", "text": "railway"}, {"id": "insurance", "text": "insurance"}]$x$::jsonb WHERE id = 'q-container-002';
UPDATE questions SET options = $x$[{"id": "castings", "text": "castings"}, {"id": "fittings", "text": "fittings"}, {"id": "brackets", "text": "brackets"}, {"id": "locks", "text": "locks"}]$x$::jsonb WHERE id = 'q-container-003';
UPDATE questions SET options = $x$[{"id": "empty", "text": "empty"}, {"id": "loaded", "text": "loaded"}, {"id": "late", "text": "late"}, {"id": "damaged", "text": "damaged"}]$x$::jsonb WHERE id = 'q-container-004';
UPDATE questions SET options = $x$[{"id": "Panamax", "text": "Panamax"}, {"id": "Suezmax", "text": "Suezmax"}, {"id": "Atlantic", "text": "Atlantic"}, {"id": "Standard", "text": "Standard"}]$x$::jsonb WHERE id = 'q-container-005';
UPDATE questions SET options = $x$[{"id": "distance", "text": "distance"}, {"id": "labour", "text": "labour"}, {"id": "fuel", "text": "fuel"}, {"id": "insurance", "text": "insurance"}]$x$::jsonb WHERE id = 'q-container-006';
UPDATE questions SET options = $x$[{"id": "brittle", "text": "brittle"}, {"id": "cheaper", "text": "cheaper"}, {"id": "faster", "text": "faster"}, {"id": "safer", "text": "safer"}]$x$::jsonb WHERE id = 'q-container-007';
UPDATE questions SET options = $x$[{"id": "air", "text": "air"}, {"id": "water", "text": "water"}, {"id": "dust", "text": "dust"}, {"id": "snow", "text": "snow"}]$x$::jsonb WHERE id = 'q-icecore-001';
UPDATE questions SET options = $x$[{"id": "fluid", "text": "fluid"}, {"id": "gas", "text": "gas"}, {"id": "foam", "text": "foam"}, {"id": "cable", "text": "cable"}]$x$::jsonb WHERE id = 'q-icecore-002';
UPDATE questions SET options = $x$[{"id": "ash", "text": "ash"}, {"id": "salt", "text": "salt"}, {"id": "ice", "text": "ice"}, {"id": "sand", "text": "sand"}]$x$::jsonb WHERE id = 'q-icecore-003';
UPDATE questions SET options = $x$[{"id": "inference", "text": "inference"}, {"id": "filtering", "text": "filtering"}, {"id": "dating", "text": "dating"}, {"id": "calibration", "text": "calibration"}]$x$::jsonb WHERE id = 'q-icecore-004';
UPDATE questions SET options = $x$[{"id": "isotopes", "text": "isotopes"}, {"id": "bubbles", "text": "bubbles"}, {"id": "crystals", "text": "crystals"}, {"id": "molecules", "text": "molecules"}]$x$::jsonb WHERE id = 'q-icecore-005';
UPDATE questions SET options = $x$[{"id": "decade", "text": "decade"}, {"id": "century", "text": "century"}, {"id": "season", "text": "season"}, {"id": "year", "text": "year"}]$x$::jsonb WHERE id = 'q-icecore-006';
UPDATE questions SET options = $x$[{"id": "bedrock", "text": "bedrock"}, {"id": "surface", "text": "surface"}, {"id": "coastline", "text": "coastline"}, {"id": "borehole", "text": "borehole"}]$x$::jsonb WHERE id = 'q-icecore-007';
UPDATE questions SET options = $x$[{"id": "accession", "text": "accession"}, {"id": "archive", "text": "archive"}, {"id": "entry", "text": "entry"}, {"id": "exhibit", "text": "exhibit"}]$x$::jsonb WHERE id = 'q-seedvault-001';
UPDATE questions SET options = $x$[{"id": "eighteen", "text": "eighteen"}, {"id": "eight", "text": "eight"}, {"id": "eighty", "text": "eighty"}, {"id": "twelve", "text": "twelve"}]$x$::jsonb WHERE id = 'q-seedvault-002';
UPDATE questions SET options = $x$[{"id": "regeneration", "text": "regeneration"}, {"id": "inspection", "text": "inspection"}, {"id": "relocation", "text": "relocation"}, {"id": "replacement", "text": "replacement"}]$x$::jsonb WHERE id = 'q-seedvault-003';
UPDATE questions SET options = $x$[{"id": "permafrost", "text": "permafrost"}, {"id": "mountain", "text": "mountain"}, {"id": "glacier", "text": "glacier"}, {"id": "tundra", "text": "tundra"}]$x$::jsonb WHERE id = 'q-seedvault-004';
UPDATE questions SET options = $x$[{"id": "withdrawal", "text": "withdrawal"}, {"id": "deposit", "text": "deposit"}, {"id": "inspection", "text": "inspection"}, {"id": "closure", "text": "closure"}]$x$::jsonb WHERE id = 'q-seedvault-005';
UPDATE questions SET options = $x$[{"id": "recalcitrant", "text": "recalcitrant"}, {"id": "dormant", "text": "dormant"}, {"id": "resistant", "text": "resistant"}, {"id": "perishable", "text": "perishable"}]$x$::jsonb WHERE id = 'q-seedvault-006';
UPDATE questions SET options = $x$[{"id": "nitrogen", "text": "nitrogen"}, {"id": "oxygen", "text": "oxygen"}, {"id": "helium", "text": "helium"}, {"id": "hydrogen", "text": "hydrogen"}]$x$::jsonb WHERE id = 'q-seedvault-007';

DO $guard$
DECLARE
    bare INT;
BEGIN
    SELECT count(*) INTO bare
    FROM questions
    WHERE type_id = 'reading-sentence-completion'
      AND passage_id IN ('rp-container-01', 'rp-icecore-01', 'rp-seedvault-01')
      AND (options IS NULL OR jsonb_array_length(options) < 4);

    IF bare > 0 THEN
        RAISE EXCEPTION 'sentence completion batch 2: % gaps still have no choices', bare;
    END IF;

    SELECT count(*) INTO bare
    FROM questions q
    WHERE q.type_id = 'reading-sentence-completion'
      AND q.passage_id IN ('rp-container-01', 'rp-icecore-01', 'rp-seedvault-01')
      AND EXISTS (
          SELECT 1 FROM jsonb_array_elements_text(q.correct_answers) AS a
          WHERE NOT EXISTS (SELECT 1 FROM jsonb_array_elements(q.options) AS o WHERE o->>'id' = a)
      );

    IF bare > 0 THEN
        RAISE EXCEPTION 'sentence completion batch 2: % answers are not among their choices', bare;
    END IF;
END
$guard$;
