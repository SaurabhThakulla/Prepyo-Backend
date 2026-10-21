-- The imported PTE speaking and writing passages (000112) kept their source's hard line wraps.
-- Join a break that falls mid-sentence; keep one after . ? ! or : (a new speaker) and blank-line paragraphs.
UPDATE questions
SET context_passage = regexp_replace(context_passage, '([^[:space:].?!:])[ \t\r]*\n(?![ \t\r]*\n)[ \t]*', '\1 ', 'g')
WHERE id LIKE 'pte-alfa-%'
  AND skill IN ('speaking', 'writing')
  AND context_passage ~ '[^[:space:].?!:][ \t\r]*\n';
