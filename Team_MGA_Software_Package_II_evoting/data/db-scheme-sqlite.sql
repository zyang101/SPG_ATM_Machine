-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Users table
CREATE TABLE IF NOT EXISTS "users" (
  "id" TEXT PRIMARY KEY,
  "first_name" TEXT NOT NULL,
  "last_name" TEXT NOT NULL,
  "date_of_birth" DATE,
  "created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,

  UNIQUE ("first_name", "last_name", "date_of_birth")
);

-- Credentials table with role check constraint
CREATE TABLE IF NOT EXISTS "credentials" (
  "user_id" TEXT NOT NULL REFERENCES "users"("id"),
  "password_hash" TEXT NOT NULL,
  "role" TEXT NOT NULL CHECK("role" IN ('official', 'voter', 'admin'))
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_credentials_user_id" ON "credentials" ("user_id");

-- Elections table with status check constraint
CREATE TABLE IF NOT EXISTS "elections" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "name" TEXT NOT NULL,
  "district" TEXT,
  "start_date" DATETIME,
  "end_date" DATETIME,
  "district_official_id" TEXT NOT NULL REFERENCES "users"("id"),
  "status" TEXT NOT NULL DEFAULT 'not_started' CHECK("status" IN ('not_started', 'active', 'closed'))
);

-- Positions table
CREATE TABLE IF NOT EXISTS "positions" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "name" TEXT NOT NULL,
  "election_id" INTEGER NOT NULL REFERENCES "elections"("id"),
  "winner_candidate_id" INTEGER REFERENCES "candidates"("id")
);

-- Candidates table
CREATE TABLE IF NOT EXISTS "candidates" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "name" TEXT NOT NULL,
  "position_id" INTEGER NOT NULL REFERENCES "positions"("id"),
  "party_name" TEXT NOT NULL
);

-- Votes table
CREATE TABLE IF NOT EXISTS "votes" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "voter_user_id" TEXT NOT NULL REFERENCES "users"("id"),
  "position_id" INTEGER NOT NULL REFERENCES "positions"("id"),
  "candidate_id" INTEGER NOT NULL REFERENCES "candidates"("id"),
  "created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,

  UNIQUE ("voter_user_id", "position_id")
);

INSERT OR IGNORE INTO "users" ("id","first_name","last_name","date_of_birth")
VALUES ('BigBoss','Optimus','Prime','1984-05-08');

INSERT OR IGNORE INTO "credentials" ("user_id","password_hash","role")
VALUES ('BigBoss','5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5','admin');

-- Prevent votes from being sabotagoed
CREATE TRIGGER IF NOT EXISTS votes_no_update
BEFORE UPDATE ON votes
BEGIN
  SELECT RAISE(ABORT, 'Votes are immutable');
END;

CREATE TRIGGER IF NOT EXISTS votes_no_delete
BEFORE DELETE ON votes
BEGIN
  SELECT RAISE(ABORT, 'Votes are immutable');
END;


--- just for me to test admin
DELETE FROM credentials WHERE user_id = 'akwok1';
DELETE FROM users WHERE id = 'akwok1';

INSERT INTO users (id, first_name, last_name, date_of_birth)
VALUES ('akwok1', 'Amanda', 'Kwok', '2000-01-01');

INSERT INTO credentials (user_id, password_hash, role)
VALUES ('akwok1', '5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5', 'admin');

