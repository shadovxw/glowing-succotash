CREATE TABLE IF NOT EXISTS game_events (
  id         BIGSERIAL   PRIMARY KEY,
  game_id    UUID        NOT NULL REFERENCES games(id) ON DELETE CASCADE,
  seq        INT         NOT NULL,
  actor_id   TEXT,
  type       TEXT        NOT NULL,
  payload    JSONB       NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (game_id, seq)
);

CREATE INDEX IF NOT EXISTS idx_game_events_game_id ON game_events(game_id);
