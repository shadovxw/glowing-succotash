CREATE TABLE IF NOT EXISTS chat_messages (
  id           BIGSERIAL   PRIMARY KEY,
  room_id      TEXT        NOT NULL,
  game_id      UUID        REFERENCES games(id),
  sender_id    TEXT,
  display_name TEXT        NOT NULL,
  message      TEXT        NOT NULL,
  is_system    BOOLEAN     NOT NULL DEFAULT false,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_chat_room_id    ON chat_messages(room_id);
CREATE INDEX IF NOT EXISTS idx_chat_created_at ON chat_messages(created_at);
