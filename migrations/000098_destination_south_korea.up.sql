-- South Korea becomes a destination learners can pick at onboarding. It is
-- where EPS-TOPIK takes them, and the exam itself is now offered.
ALTER TABLE users DROP CONSTRAINT users_destination_check;
ALTER TABLE users ADD CONSTRAINT users_destination_check
    CHECK (destination IN ('australia', 'uk', 'canada', 'usa', 'new-zealand', 'south-korea', 'other'));
