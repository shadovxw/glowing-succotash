CREATE TABLE IF NOT EXISTS properties (
  id             UUID  PRIMARY KEY DEFAULT gen_random_uuid(),
  tile_position  INT   UNIQUE NOT NULL REFERENCES tiles(position) ON DELETE CASCADE,
  price          INT   NOT NULL,
  mortgage_value INT   NOT NULL,
  house_cost     INT,
  -- streets:    [base, 1h, 2h, 3h, 4h, hotel]
  -- railroads:  [1rr, 2rr, 3rr, 4rr]
  -- utilities:  [mult_1owned, mult_2owned]
  rent           INT[] NOT NULL
);
