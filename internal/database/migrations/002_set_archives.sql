CREATE TABLE question_set_archives (
 set_id TEXT PRIMARY KEY NOT NULL REFERENCES question_sets(id)
);
PRAGMA user_version = 2;
