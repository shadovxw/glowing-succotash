CREATE TABLE IF NOT EXISTS game_config (
  id               INT     PRIMARY KEY DEFAULT 1,
  currency_name    TEXT    NOT NULL DEFAULT 'Dollars',
  currency_symbol  TEXT    NOT NULL DEFAULT '$',
  currency_icon    TEXT    NOT NULL DEFAULT '💰',
  starting_balance INT     NOT NULL DEFAULT 1500,
  go_amount        INT     NOT NULL DEFAULT 200,
  income_tax       INT     NOT NULL DEFAULT 200,
  luxury_tax       INT     NOT NULL DEFAULT 100,
  house_limit      INT     NOT NULL DEFAULT 32,
  hotel_limit      INT     NOT NULL DEFAULT 12,
  max_players      INT     NOT NULL DEFAULT 6,
  turn_timer_secs  INT     NOT NULL DEFAULT 0,
  free_parking_jackpot  BOOLEAN NOT NULL DEFAULT false,
  auction_on_decline    BOOLEAN NOT NULL DEFAULT true,
  no_rent_in_jail       BOOLEAN NOT NULL DEFAULT false,
  double_salary_on_go   BOOLEAN NOT NULL DEFAULT false,
  bankruptcy_to_bank    BOOLEAN NOT NULL DEFAULT true
);

INSERT INTO game_config (id) VALUES (1) ON CONFLICT DO NOTHING;
