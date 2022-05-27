CREATE TABLE books (
	id SERIAL NOT NULL, 
	name varchar(30),
	PRIMARY KEY (id)
);

INSERT INTO books (name) VALUES ('Book 1');
INSERT INTO books (name) VALUES ('Book 2');