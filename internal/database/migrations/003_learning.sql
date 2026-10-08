CREATE TABLE learning_units (
 id TEXT PRIMARY KEY,
 document TEXT NOT NULL,
 fingerprint TEXT NOT NULL,
 imported_at TEXT NOT NULL
);
CREATE TABLE learning_progress (
 unit_id TEXT NOT NULL REFERENCES learning_units(id),
 concept_id TEXT NOT NULL,
 status TEXT NOT NULL CHECK(status IN ('in_progress','understood','needs_review')),
 last_opened_at TEXT,
 updated_at TEXT NOT NULL,
 PRIMARY KEY(unit_id,concept_id)
);
PRAGMA user_version=3;
