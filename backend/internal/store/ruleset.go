package store

import (
	"encoding/json"
	"fmt"

	"github.com/shadovxw/monopoly/internal/game"
	"gorm.io/gorm"
)

// LoadRuleSet queries the database and assembles a RuleSet ready for the game engine.
func LoadRuleSet(db *gorm.DB) (*game.RuleSet, error) {
	cfg, err := loadConfig(db)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	tiles, err := loadTiles(db)
	if err != nil {
		return nil, fmt.Errorf("load tiles: %w", err)
	}

	props, err := loadProperties(db)
	if err != nil {
		return nil, fmt.Errorf("load properties: %w", err)
	}

	chance, community, err := loadCards(db)
	if err != nil {
		return nil, fmt.Errorf("load cards: %w", err)
	}

	rs := &game.RuleSet{
		Config:    cfg,
		Props:     props,
		Chance:    chance,
		Community: community,
	}
	for i, t := range tiles {
		if i < 40 {
			rs.Tiles[i] = t
		}
	}

	return rs, nil
}

// ---- DB row types (separate from game types to avoid coupling) ----

type dbConfig struct {
	CurrencyName        string `gorm:"column:currency_name"`
	CurrencySymbol      string `gorm:"column:currency_symbol"`
	CurrencyIcon        string `gorm:"column:currency_icon"`
	StartingBalance     int    `gorm:"column:starting_balance"`
	GoAmount            int    `gorm:"column:go_amount"`
	IncomeTax           int    `gorm:"column:income_tax"`
	LuxuryTax           int    `gorm:"column:luxury_tax"`
	HouseLimit          int    `gorm:"column:house_limit"`
	HotelLimit          int    `gorm:"column:hotel_limit"`
	MaxPlayers          int    `gorm:"column:max_players"`
	TurnTimerSecs       int    `gorm:"column:turn_timer_secs"`
	FreeParkingJackpot  bool   `gorm:"column:free_parking_jackpot"`
	AuctionOnDecline    bool   `gorm:"column:auction_on_decline"`
	NoRentInJail        bool   `gorm:"column:no_rent_in_jail"`
	DoubleSalaryOnGo    bool   `gorm:"column:double_salary_on_go"`
	BankruptcyToBank    bool   `gorm:"column:bankruptcy_to_bank"`
}

func (dbConfig) TableName() string { return "game_config" }

type dbTile struct {
	Position    int    `gorm:"column:position"`
	Type        string `gorm:"column:type"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	ColorGroup  string `gorm:"column:color_group"`
	Icon        string `gorm:"column:icon"`
}

func (dbTile) TableName() string { return "tiles" }

type dbProperty struct {
	ID            string `gorm:"column:id"`
	TilePosition  int    `gorm:"column:tile_position"`
	Price         int    `gorm:"column:price"`
	MortgageValue int    `gorm:"column:mortgage_value"`
	HouseCost     int    `gorm:"column:house_cost"`
	Rent          []int64 `gorm:"column:rent;type:int[]"`
}

func (dbProperty) TableName() string { return "properties" }

type dbCard struct {
	ID            string          `gorm:"column:id"`
	Deck          string          `gorm:"column:deck"`
	Name          string          `gorm:"column:name"`
	Description   string          `gorm:"column:description"`
	EffectType    string          `gorm:"column:effect_type"`
	EffectPayload json.RawMessage `gorm:"column:effect_payload;type:jsonb"`
	SortOrder     int             `gorm:"column:sort_order"`
}

func (dbCard) TableName() string { return "cards" }

func loadConfig(db *gorm.DB) (game.GameConfig, error) {
	var row dbConfig
	if err := db.First(&row).Error; err != nil {
		return game.GameConfig{}, err
	}
	return game.GameConfig{
		CurrencyName:       row.CurrencyName,
		CurrencySymbol:     row.CurrencySymbol,
		CurrencyIcon:       row.CurrencyIcon,
		StartingBalance:    row.StartingBalance,
		GoAmount:           row.GoAmount,
		IncomeTax:          row.IncomeTax,
		LuxuryTax:          row.LuxuryTax,
		HouseLimit:         row.HouseLimit,
		HotelLimit:         row.HotelLimit,
		MaxPlayers:         row.MaxPlayers,
		TurnTimerSecs:      row.TurnTimerSecs,
		FreeParkingJackpot: row.FreeParkingJackpot,
		AuctionOnDecline:   row.AuctionOnDecline,
		NoRentInJail:       row.NoRentInJail,
		DoubleSalaryOnGo:   row.DoubleSalaryOnGo,
		BankruptcyToBank:   row.BankruptcyToBank,
	}, nil
}

func loadTiles(db *gorm.DB) ([]game.Tile, error) {
	var rows []dbTile
	if err := db.Order("position").Find(&rows).Error; err != nil {
		return nil, err
	}
	tiles := make([]game.Tile, len(rows))
	for i, r := range rows {
		tiles[i] = game.Tile{
			Position:    r.Position,
			Type:        r.Type,
			Name:        r.Name,
			Description: r.Description,
			ColorGroup:  r.ColorGroup,
			Icon:        r.Icon,
		}
	}
	return tiles, nil
}

func loadProperties(db *gorm.DB) (map[int]*game.Property, error) {
	var rows []dbProperty
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	props := make(map[int]*game.Property, len(rows))
	for _, r := range rows {
		rent := make([]int, len(r.Rent))
		for i, v := range r.Rent {
			rent[i] = int(v)
		}
		props[r.TilePosition] = &game.Property{
			TilePosition:  r.TilePosition,
			Price:         r.Price,
			MortgageValue: r.MortgageValue,
			HouseCost:     r.HouseCost,
			Rent:          rent,
		}
	}
	return props, nil
}

func loadCards(db *gorm.DB) (chance, community []game.Card, err error) {
	var rows []dbCard
	if err = db.Where("is_active = true").Order("deck, sort_order").Find(&rows).Error; err != nil {
		return
	}
	for _, r := range rows {
		c := game.Card{
			Deck:          r.Deck,
			Name:          r.Name,
			Description:   r.Description,
			EffectType:    r.EffectType,
			EffectPayload: r.EffectPayload,
		}
		if r.Deck == "chance" {
			chance = append(chance, c)
		} else {
			community = append(community, c)
		}
	}
	return
}
