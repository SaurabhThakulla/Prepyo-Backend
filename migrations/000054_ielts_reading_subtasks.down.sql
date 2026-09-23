-- Remove only the rows introduced by 000054, in foreign-key order.
DELETE FROM questions WHERE id IN (
 'q-ir54-single-1','q-ir54-single-2','q-ir54-single-3',
 'q-ir54-multiple-1','q-ir54-multiple-2',
 'q-ir54-tfng-1','q-ir54-tfng-2','q-ir54-tfng-3','q-ir54-tfng-4','q-ir54-tfng-5',
 'q-ir54-ynng-1','q-ir54-ynng-2','q-ir54-ynng-3','q-ir54-ynng-4','q-ir54-ynng-5',
 'q-ir54-matching-1','q-ir54-matching-2','q-ir54-matching-3','q-ir54-matching-4','q-ir54-matching-5',
 'q-ir54-completion-1','q-ir54-completion-2','q-ir54-completion-3','q-ir54-completion-4','q-ir54-completion-5');
DELETE FROM reading_question_groups WHERE id IN (
 'g-ir54-single','g-ir54-multiple','g-ir54-tfng','g-ir54-ynng','g-ir54-matching','g-ir54-completion');
DELETE FROM reading_passages WHERE id = 'rp-ir54-library';
