CREATE TABLE IF NOT EXISTS players (
  bastion_user_id TEXT        PRIMARY KEY,
  display_name    TEXT        NOT NULL,
  avatar          TEXT        NOT NULL DEFAULT '',
  games_played    INT         NOT NULL DEFAULT 0,
  games_won       INT         NOT NULL DEFAULT 0,
  last_seen       TIMESTAMPTZ NOT NULL DEFAULT now()
);
