-- Questions go first: they reference their groups, and a paper that was dealt
-- one of them keeps the id in reading_mock_sessions.question_ids, where a
-- missing question is simply skipped on hydration.
DELETE FROM questions WHERE id IN (
    'q-choc-101','q-choc-102','q-choc-103','q-choc-104','q-choc-105','q-choc-106',
    'q-bees-101','q-bees-102','q-bees-103','q-bees-104',
    'q-paper-101','q-paper-102','q-paper-103','q-paper-104','q-paper-105',
    'q-paper-106','q-paper-107','q-paper-108',
    'q-time-101','q-time-102',
    'q-time-201','q-time-202','q-time-203','q-time-204','q-time-205','q-time-206',
    'q-time-207','q-time-208','q-time-209','q-time-210');

DELETE FROM reading_question_groups WHERE id IN (
    'g-choc-9','g-choc-10',
    'g-bees-9','g-bees-10',
    'g-paper-9','g-paper-10','g-paper-11',
    'g-time-5','g-time-6','g-time-7','g-time-8');
