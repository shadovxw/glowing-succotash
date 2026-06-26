package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/shadovxw/monopoly/internal/config"
	"github.com/shadovxw/monopoly/internal/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	cfg := config.Load()
	if err := db.Connect(cfg.DatabaseURL); err != nil {
		log.Fatal("connect db:", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	if err := db.Migrate(filepath.Join(repoRoot, "migrations")); err != nil {
		log.Fatal("migrate:", err)
	}

	seedTiles(db.DB)
	seedProperties(db.DB)
	seedCards(db.DB)
	log.Println("seed complete")
}

// ---- Tiles ----

type tileRow struct {
	Position    int    `gorm:"column:position"`
	Type        string `gorm:"column:type"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	ColorGroup  string `gorm:"column:color_group"`
	Icon        string `gorm:"column:icon"`
}

func (tileRow) TableName() string { return "tiles" }

func seedTiles(db *gorm.DB) {
	tiles := []tileRow{
		{0, "go", "GO", "Collect $200 salary as you pass", "", "🏁"},
		{1, "street", "Mediterranean Avenue", "The cheapest property on the board", "brown", "🟤"},
		{2, "community_chest", "Community Chest", "Draw a Community Chest card", "", "📦"},
		{3, "street", "Baltic Avenue", "A bargain if you can get the set", "brown", "🟤"},
		{4, "tax", "Income Tax", "Pay $200 or 10% of your total worth", "", "💸"},
		{5, "railroad", "Reading Railroad", "All aboard!", "railroad", "🚂"},
		{6, "street", "Oriental Avenue", "Gateway to the light blues", "light_blue", "🔵"},
		{7, "chance", "Chance", "Draw a Chance card — good or bad?", "", "❓"},
		{8, "street", "Vermont Avenue", "A solid light blue", "light_blue", "🔵"},
		{9, "street", "Connecticut Avenue", "Completing the light blues", "light_blue", "🔵"},
		{10, "jail", "Jail / Just Visiting", "Just passing through... or not", "", "⛓️"},
		{11, "street", "St. Charles Place", "First of the pinks", "pink", "🩷"},
		{12, "utility", "Electric Company", "Pay 4× or 10× dice roll", "utility", "⚡"},
		{13, "street", "States Avenue", "A pink in progress", "pink", "🩷"},
		{14, "street", "Virginia Avenue", "Completing the pink set", "pink", "🩷"},
		{15, "railroad", "Pennsylvania Railroad", "A key transit hub", "railroad", "🚂"},
		{16, "street", "St. James Place", "Orange and valuable", "orange", "🟠"},
		{17, "community_chest", "Community Chest", "Draw a Community Chest card", "", "📦"},
		{18, "street", "Tennessee Avenue", "Heart of the orange group", "orange", "🟠"},
		{19, "street", "New York Avenue", "The most-landed orange", "orange", "🟠"},
		{20, "free_parking", "Free Parking", "Rest your weary tokens here", "", "🅿️"},
		{21, "street", "Kentucky Avenue", "Start of the reds", "red", "🔴"},
		{22, "chance", "Chance", "Draw a Chance card", "", "❓"},
		{23, "street", "Indiana Avenue", "Red and lucrative", "red", "🔴"},
		{24, "street", "Illinois Avenue", "Most-landed property on the board", "red", "🔴"},
		{25, "railroad", "B. & O. Railroad", "Third railroad — power builds", "railroad", "🚂"},
		{26, "street", "Atlantic Avenue", "Yellow and prosperous", "yellow", "🟡"},
		{27, "street", "Ventnor Avenue", "A strong yellow", "yellow", "🟡"},
		{28, "utility", "Water Works", "Pay 4× or 10× dice roll", "utility", "💧"},
		{29, "street", "Marvin Gardens", "Completing the yellows", "yellow", "🟡"},
		{30, "go_to_jail", "Go To Jail", "Do not pass GO. Do not collect $200.", "", "👮"},
		{31, "street", "Pacific Avenue", "First of the greens", "green", "🟢"},
		{32, "street", "North Carolina Avenue", "Solid green investment", "green", "🟢"},
		{33, "community_chest", "Community Chest", "Draw a Community Chest card", "", "📦"},
		{34, "street", "Pennsylvania Avenue", "Expensive green", "green", "🟢"},
		{35, "railroad", "Short Line Railroad", "The fourth and final railroad", "railroad", "🚂"},
		{36, "chance", "Chance", "Draw a Chance card", "", "❓"},
		{37, "street", "Park Place", "Elite real estate", "dark_blue", "🔷"},
		{38, "tax", "Luxury Tax", "Pay $100 — the price of luxury", "", "💸"},
		{39, "street", "Boardwalk", "The most prestigious address", "dark_blue", "🔷"},
	}

	for _, t := range tiles {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&t) //nolint:errcheck
	}
	log.Printf("seeded %d tiles", len(tiles))
}

// ---- Properties ----

type propertyRow struct {
	TilePosition  int     `gorm:"column:tile_position"`
	Price         int     `gorm:"column:price"`
	MortgageValue int     `gorm:"column:mortgage_value"`
	HouseCost     int     `gorm:"column:house_cost"`
	Rent          []int64 `gorm:"column:rent;type:int[]"`
}

func (propertyRow) TableName() string { return "properties" }

func seedProperties(db *gorm.DB) {
	// rent arrays:
	// streets:   [base, 1h, 2h, 3h, 4h, hotel]
	// railroads: [1rr, 2rr, 3rr, 4rr]
	// utilities: [mult_1owned, mult_2owned]
	props := []propertyRow{
		// Brown
		{1, 60, 30, 50, i64s(2, 10, 30, 90, 160, 250)},
		{3, 60, 30, 50, i64s(4, 20, 60, 180, 320, 450)},
		// Light Blue
		{6, 100, 50, 50, i64s(6, 30, 90, 270, 400, 550)},
		{8, 100, 50, 50, i64s(6, 30, 90, 270, 400, 550)},
		{9, 120, 60, 50, i64s(8, 40, 100, 300, 450, 600)},
		// Pink
		{11, 140, 70, 100, i64s(10, 50, 150, 450, 625, 750)},
		{13, 140, 70, 100, i64s(10, 50, 150, 450, 625, 750)},
		{14, 160, 80, 100, i64s(12, 60, 180, 500, 700, 900)},
		// Orange
		{16, 180, 90, 100, i64s(14, 70, 200, 550, 750, 950)},
		{18, 180, 90, 100, i64s(14, 70, 200, 550, 750, 950)},
		{19, 200, 100, 100, i64s(16, 80, 220, 600, 800, 1000)},
		// Red
		{21, 220, 110, 150, i64s(18, 90, 250, 700, 875, 1050)},
		{23, 220, 110, 150, i64s(18, 90, 250, 700, 875, 1050)},
		{24, 240, 120, 150, i64s(20, 100, 300, 750, 925, 1100)},
		// Yellow
		{26, 260, 130, 150, i64s(22, 110, 330, 800, 975, 1150)},
		{27, 260, 130, 150, i64s(22, 110, 330, 800, 975, 1150)},
		{29, 280, 140, 150, i64s(24, 120, 360, 850, 1025, 1200)},
		// Green
		{31, 300, 150, 200, i64s(26, 130, 390, 900, 1100, 1275)},
		{32, 300, 150, 200, i64s(26, 130, 390, 900, 1100, 1275)},
		{34, 320, 160, 200, i64s(28, 150, 450, 1000, 1200, 1400)},
		// Dark Blue
		{37, 350, 175, 200, i64s(35, 175, 500, 1100, 1300, 1500)},
		{39, 400, 200, 200, i64s(50, 200, 600, 1400, 1700, 2000)},
		// Railroads: rent scale by count owned
		{5, 200, 100, 0, i64s(25, 50, 100, 200)},
		{15, 200, 100, 0, i64s(25, 50, 100, 200)},
		{25, 200, 100, 0, i64s(25, 50, 100, 200)},
		{35, 200, 100, 0, i64s(25, 50, 100, 200)},
		// Utilities: multipliers
		{12, 150, 75, 0, i64s(4, 10)},
		{28, 150, 75, 0, i64s(4, 10)},
	}

	for _, p := range props {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&p) //nolint:errcheck
	}
	log.Printf("seeded %d properties", len(props))
}

// ---- Cards ----

type cardRow struct {
	Deck          string `gorm:"column:deck"`
	Name          string `gorm:"column:name"`
	Description   string `gorm:"column:description"`
	EffectType    string `gorm:"column:effect_type"`
	EffectPayload string `gorm:"column:effect_payload;type:jsonb"`
	SortOrder     int    `gorm:"column:sort_order"`
}

func (cardRow) TableName() string { return "cards" }

func seedCards(db *gorm.DB) {
	cards := []cardRow{
		// ---- Chance (16 cards) ----
		{
			"chance", "Advance to Boardwalk",
			"Move directly to Boardwalk.",
			"move_to", `{"position":39,"collect_go":true}`, 1,
		},
		{
			"chance", "Advance to GO",
			"Move to GO and collect $200.",
			"move_to", `{"position":0,"collect_go":false}`, 2,
		},
		{
			"chance", "Advance to Illinois Avenue",
			"If you pass GO, collect $200.",
			"move_to", `{"position":24,"collect_go":true}`, 3,
		},
		{
			"chance", "Advance to St. Charles Place",
			"If you pass GO, collect $200.",
			"move_to", `{"position":11,"collect_go":true}`, 4,
		},
		{
			"chance", "Advance to Nearest Railroad",
			"Advance to nearest railroad. If unowned, buy it. If owned, pay double rent.",
			"nearest_railroad", `{"double_rent":true}`, 5,
		},
		{
			"chance", "Advance to Nearest Railroad (2)",
			"Advance to nearest railroad. If unowned, buy it. If owned, pay double rent.",
			"nearest_railroad", `{"double_rent":true}`, 6,
		},
		{
			"chance", "Advance to Nearest Utility",
			"Advance to nearest utility. If unowned, buy it. If owned, pay 10× dice.",
			"nearest_utility", `{"dice_mult":10}`, 7,
		},
		{
			"chance", "Bank Pays Dividend",
			"The bank pays you a dividend of $50.",
			"collect", `{"amount":50}`, 8,
		},
		{
			"chance", "Get Out of Jail Free",
			"Keep this card until needed, then use it to exit jail free.",
			"get_out_of_jail", `{}`, 9,
		},
		{
			"chance", "Go Back 3 Spaces",
			"Move your token back 3 spaces.",
			"move_relative", `{"steps":-3}`, 10,
		},
		{
			"chance", "Go to Jail",
			"Go directly to Jail. Do not pass GO. Do not collect $200.",
			"go_to_jail", `{}`, 11,
		},
		{
			"chance", "General Repairs",
			"Make general repairs on all your property: pay $25 per house, $100 per hotel.",
			"pay_per_building", `{"house":25,"hotel":100}`, 12,
		},
		{
			"chance", "Pay Poor Tax",
			"Pay a poor tax of $15.",
			"pay", `{"amount":15}`, 13,
		},
		{
			"chance", "Take a Trip to Reading Railroad",
			"If you pass GO, collect $200.",
			"move_to", `{"position":5,"collect_go":true}`, 14,
		},
		{
			"chance", "Elected Chairman of the Board",
			"You have been elected Chairman of the Board. Pay each player $50.",
			"pay_per_player", `{"amount":50}`, 15,
		},
		{
			"chance", "Building Loan Matures",
			"Your building loan matures. Collect $150.",
			"collect", `{"amount":150}`, 16,
		},

		// ---- Community Chest (16 cards) ----
		{
			"community_chest", "Advance to GO",
			"Move to GO and collect $200.",
			"move_to", `{"position":0,"collect_go":false}`, 1,
		},
		{
			"community_chest", "Bank Error in Your Favour",
			"Bank error in your favour. Collect $200.",
			"collect", `{"amount":200}`, 2,
		},
		{
			"community_chest", "Doctor's Fee",
			"Doctor's fee. Pay $50.",
			"pay", `{"amount":50}`, 3,
		},
		{
			"community_chest", "From Sale of Stock",
			"From sale of stock you get $50.",
			"collect", `{"amount":50}`, 4,
		},
		{
			"community_chest", "Get Out of Jail Free",
			"Keep this card until needed, then use it to exit jail free.",
			"get_out_of_jail", `{}`, 5,
		},
		{
			"community_chest", "Go to Jail",
			"Go directly to Jail. Do not pass GO. Do not collect $200.",
			"go_to_jail", `{}`, 6,
		},
		{
			"community_chest", "Grand Opera Night",
			"Grand Opera Night. Collect $50 from each player for opening night seats.",
			"collect_per_player", `{"amount":50}`, 7,
		},
		{
			"community_chest", "Holiday Fund Matures",
			"Holiday fund matures. Collect $100.",
			"collect", `{"amount":100}`, 8,
		},
		{
			"community_chest", "Income Tax Refund",
			"Income tax refund. Collect $20.",
			"collect", `{"amount":20}`, 9,
		},
		{
			"community_chest", "It's Your Birthday",
			"It is your birthday. Collect $10 from each player.",
			"collect_per_player", `{"amount":10}`, 10,
		},
		{
			"community_chest", "Life Insurance Matures",
			"Life insurance matures. Collect $100.",
			"collect", `{"amount":100}`, 11,
		},
		{
			"community_chest", "Pay Hospital Fees",
			"Pay hospital fees of $100.",
			"pay", `{"amount":100}`, 12,
		},
		{
			"community_chest", "Pay School Fees",
			"Pay school fees of $50.",
			"pay", `{"amount":50}`, 13,
		},
		{
			"community_chest", "Receive Consultancy Fee",
			"Receive consultancy fee of $25.",
			"collect", `{"amount":25}`, 14,
		},
		{
			"community_chest", "Street Repairs",
			"You are assessed for street repairs: pay $40 per house and $115 per hotel.",
			"pay_per_building", `{"house":40,"hotel":115}`, 15,
		},
		{
			"community_chest", "Won Second Prize in Beauty Contest",
			"You have won second prize in a beauty contest. Collect $10.",
			"collect", `{"amount":10}`, 16,
		},
	}

	for _, c := range cards {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&c) //nolint:errcheck
	}
	log.Printf("seeded %d cards", len(cards))
}

func i64s(vals ...int64) []int64 { return vals }
