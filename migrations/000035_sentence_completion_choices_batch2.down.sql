UPDATE questions SET options = '[]'::jsonb
WHERE type_id = 'reading-sentence-completion'
  AND passage_id IN ('rp-container-01', 'rp-icecore-01', 'rp-seedvault-01');
