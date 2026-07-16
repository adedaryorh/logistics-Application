ALTER TABLE identity_.users
    ADD COLUMN full_name TEXT;

UPDATE identity_.users
SET full_name = split_part(email, '@', 1)
WHERE full_name IS NULL;

ALTER TABLE identity_.users
    ALTER COLUMN full_name SET NOT NULL;

ALTER TABLE identity_.users
    ADD CONSTRAINT users_full_name_length_check
    CHECK (char_length(full_name) BETWEEN 2 AND 120);
