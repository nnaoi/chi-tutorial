CREATE TABLE books (
	id SERIAL NOT NULL, 
	name varchar(30),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (id)
);

INSERT INTO books (name) VALUES ('Book 1');
INSERT INTO books (name) VALUES ('Book 2');