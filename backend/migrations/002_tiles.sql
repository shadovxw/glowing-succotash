CREATE TABLE IF NOT EXISTS tiles (
  position    INT  PRIMARY KEY CHECK (position >= 0 AND position <= 39),
  type        TEXT NOT NULL,
  name        TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  color_group TEXT,
  icon        TEXT
);
