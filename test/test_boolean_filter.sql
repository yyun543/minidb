-- Test BOOLEAN filter
CREATE DATABASE booltest;
USE booltest;

CREATE TABLE flags (id INTEGER, active BOOLEAN);

INSERT INTO flags VALUES (1, 1);
INSERT INTO flags VALUES (2, 0);
INSERT INTO flags VALUES (3, 1);

-- Should return rows 1 and 3
SELECT * FROM flags WHERE active = 1;

-- Should return row 2
SELECT * FROM flags WHERE active = 0;

-- All rows
SELECT * FROM flags;
