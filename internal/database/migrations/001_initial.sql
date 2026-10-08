CREATE TABLE question_sets (
 id TEXT PRIMARY KEY, title TEXT NOT NULL, description TEXT NOT NULL,
 document TEXT NOT NULL, imported_at TEXT NOT NULL
);
CREATE TABLE questions (
 id TEXT PRIMARY KEY, set_id TEXT NOT NULL REFERENCES question_sets(id),
 position INTEGER NOT NULL, body TEXT NOT NULL, fingerprint TEXT NOT NULL,
 UNIQUE(set_id, position)
);
CREATE TABLE quiz_sessions (
 id TEXT PRIMARY KEY, title TEXT NOT NULL, mode TEXT NOT NULL CHECK(mode IN ('study','exam')),
 status TEXT NOT NULL CHECK(status IN ('active','completed')), started_at TEXT NOT NULL, completed_at TEXT NOT NULL DEFAULT ''
);
CREATE TABLE quiz_answers (
 session_id TEXT NOT NULL REFERENCES quiz_sessions(id), question_id TEXT NOT NULL REFERENCES questions(id),
 position INTEGER NOT NULL, selected TEXT NOT NULL DEFAULT '' CHECK(selected IN ('','A','B','C','D')),
 flagged INTEGER NOT NULL DEFAULT 0 CHECK(flagged IN (0,1)),
 scored INTEGER NOT NULL DEFAULT 0 CHECK(scored IN (0,1)), correct INTEGER CHECK(correct IN (0,1)),
 scored_at TEXT NOT NULL DEFAULT '',
 PRIMARY KEY(session_id,question_id), UNIQUE(session_id,position),
 CHECK((scored=0 AND correct IS NULL AND scored_at='') OR (scored=1 AND correct IS NOT NULL AND scored_at<>''))
);
CREATE TABLE question_flags (question_id TEXT PRIMARY KEY REFERENCES questions(id), favorite INTEGER NOT NULL CHECK(favorite IN (0,1)));
CREATE TABLE app_settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);
INSERT INTO app_settings VALUES ('language','zh-CN');
CREATE INDEX answers_question ON quiz_answers(question_id,scored_at);
PRAGMA user_version = 1;
