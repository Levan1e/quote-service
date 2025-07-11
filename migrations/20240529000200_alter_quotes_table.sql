-- +goose Up
ALTER TABLE quotes ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS quotes_id_seq;
ALTER TABLE quotes ALTER COLUMN id TYPE INTEGER;

-- +goose Down
CREATE SEQUENCE quotes_id_seq OWNED BY quotes.id;
ALTER TABLE quotes ALTER COLUMN id TYPE INTEGER;
ALTER TABLE quotes ALTER COLUMN id SET DEFAULT nextval('quotes_id_seq');
