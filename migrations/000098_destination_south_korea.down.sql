-- Learners who picked South Korea keep an answer, just a less specific one.
UPDATE users SET destination = 'other' WHERE destination = 'south-korea';
ALTER TABLE users DROP CONSTRAINT users_destination_check;
ALTER TABLE users ADD CONSTRAINT users_destination_check
    CHECK (destination IN ('australia', 'uk', 'canada', 'usa', 'new-zealand', 'other'));
