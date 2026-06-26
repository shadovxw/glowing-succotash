CREATE TABLE IF NOT EXISTS game_seats (
  game_id         UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
  seat_index      INT  NOT NULL,
  bastion_user_id TEXT NOT NULL REFERENCES players(bastion_user_id),
  PRIMARY KEY (game_id, seat_index)
);
