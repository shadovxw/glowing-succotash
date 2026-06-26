CREATE TABLE IF NOT EXISTS games (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id       TEXT        NOT NULL,
  status        TEXT        NOT NULL DEFAULT 'waiting',
  started_at    TIMESTAMPTZ,
  ended_at      TIMESTAMPTZ,
  winner_id     TEXT        REFERENCES players(bastion_user_id),
  player_ids    TEXT[]      NOT NULL DEFAULT '{}',
  current_state JSONB,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_games_room_id ON games(room_id);
CREATE INDEX IF NOT EXISTS idx_games_status  ON games(status);
