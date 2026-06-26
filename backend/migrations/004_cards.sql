CREATE TABLE IF NOT EXISTS cards (
  id             UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  deck           TEXT    NOT NULL CHECK (deck IN ('chance', 'community_chest')),
  name           TEXT    NOT NULL,
  description    TEXT    NOT NULL,
  effect_type    TEXT    NOT NULL,
  effect_payload JSONB   NOT NULL DEFAULT '{}',
  is_active      BOOLEAN NOT NULL DEFAULT true,
  sort_order     INT     NOT NULL DEFAULT 0
);
