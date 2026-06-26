package game

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ---- RuleSet (loaded from DB once per game, treated as immutable) ----

type RuleSet struct {
	Config    GameConfig
	Tiles     [40]Tile
	Props     map[int]*Property // key: tile position
	Chance    []Card
	Community []Card
}

type GameConfig struct {
	CurrencyName   string `json:"currencyName"`
	CurrencySymbol string `json:"currencySymbol"`
	CurrencyIcon   string `json:"currencyIcon"`

	StartingBalance int `json:"startingBalance"`
	GoAmount        int `json:"goAmount"`
	IncomeTax       int `json:"incomeTax"`
	LuxuryTax       int `json:"luxuryTax"`
	HouseLimit      int `json:"houseLimit"`
	HotelLimit      int `json:"hotelLimit"`
	MaxPlayers      int `json:"maxPlayers"`
	TurnTimerSecs   int `json:"turnTimerSecs"`

	FreeParkingJackpot bool `json:"freeParkingJackpot"`
	AuctionOnDecline   bool `json:"auctionOnDecline"`
	NoRentInJail       bool `json:"noRentInJail"`
	DoubleSalaryOnGo   bool `json:"doubleSalaryOnGo"`
	BankruptcyToBank   bool `json:"bankruptcyToBank"`
}

type TileType = string

const (
	TileGO             TileType = "go"
	TileStreet         TileType = "street"
	TileRailroad       TileType = "railroad"
	TileUtility        TileType = "utility"
	TileTax            TileType = "tax"
	TileChance         TileType = "chance"
	TileCommunityChest TileType = "community_chest"
	TileJail           TileType = "jail"
	TileGoToJail       TileType = "go_to_jail"
	TileFreeParking    TileType = "free_parking"
)

type Tile struct {
	Position    int    `json:"position"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ColorGroup  string `json:"colorGroup"`
	Icon        string `json:"icon"`
}

type Property struct {
	ID            uuid.UUID `json:"id"`
	TilePosition  int       `json:"tilePosition"`
	Price         int       `json:"price"`
	MortgageValue int       `json:"mortgageValue"`
	HouseCost     int       `json:"houseCost"` // 0 for railroads/utilities
	// streets:   [base, 1h, 2h, 3h, 4h, hotel]
	// railroads: [1rr, 2rr, 3rr, 4rr]
	// utilities: [mult_1owned, mult_2owned]
	Rent []int `json:"rent"`
}

type Card struct {
	ID            uuid.UUID       `json:"id"`
	Deck          string          `json:"deck"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	EffectType    string          `json:"effectType"`
	EffectPayload json.RawMessage `json:"effectPayload"`
}

// Effect type constants — the game engine switches on these.
const (
	EffectMoveTo          = "move_to"           // {"position":0,"collect_go":true}
	EffectMoveRelative    = "move_relative"      // {"steps":-3}
	EffectNearestRailroad = "nearest_railroad"   // {"double_rent":true}
	EffectNearestUtility  = "nearest_utility"    // {"dice_mult":10}
	EffectCollect         = "collect"            // {"amount":50}
	EffectPay             = "pay"                // {"amount":150}
	EffectCollectPerPlayer = "collect_per_player" // {"amount":50}
	EffectPayPerPlayer    = "pay_per_player"     // {"amount":50}
	EffectPayPerBuilding  = "pay_per_building"   // {"house":25,"hotel":100}
	EffectCollectPerBuilding = "collect_per_building"
	EffectGoToJail        = "go_to_jail"
	EffectGetOutOfJail    = "get_out_of_jail"
	EffectBackToGo        = "back_to_go"
)

// ---- GameState (mutable, serialised to DB after every action) ----

type GameState struct {
	ID           uuid.UUID   `json:"id"`
	RoomID       string      `json:"roomId"`
	Status       string      `json:"status"`
	Players      []Player    `json:"players"`
	Tiles        []TileState `json:"tiles"` // always 40
	ActiveIdx    int         `json:"activeIdx"`
	Phase        string      `json:"phase"`
	Dice         [2]int      `json:"dice"`
	DoublesCount int         `json:"doublesCount"`

	// Deck order: shuffled indices into RuleSet.Chance / RuleSet.Community.
	// Circular: idx wraps around, deck reshuffled when exhausted.
	ChanceDeck []int `json:"chanceDeck"`
	CommDeck   []int `json:"commDeck"`
	ChanceIdx  int   `json:"chanceIdx"`
	CommIdx    int   `json:"commIdx"`

	HousesLeft     int `json:"housesLeft"`
	HotelsLeft     int `json:"hotelsLeft"`
	FreeParkingPot int `json:"freeParkingPot"`
	TurnNumber     int `json:"turnNumber"`
	EventSeq       int `json:"eventSeq"`

	ActiveAuction *Auction `json:"activeAuction,omitempty"`
	PendingTrades []Trade  `json:"pendingTrades"`

	StartedAt time.Time `json:"startedAt"`
}

type TileState struct {
	Position  int        `json:"position"`
	OwnerID   *uuid.UUID `json:"ownerId"`   // nil = bank
	Houses    int        `json:"houses"`    // 0-4
	Hotel     bool       `json:"hotel"`
	Mortgaged bool       `json:"mortgaged"`
}

type Player struct {
	ID          uuid.UUID `json:"id"`
	BastionID   string    `json:"bastionId"`
	DisplayName string    `json:"displayName"`
	Avatar      string    `json:"avatar"`
	Balance     int       `json:"balance"`
	Position    int       `json:"position"`
	InJail      bool      `json:"inJail"`
	JailTurns   int       `json:"jailTurns"`
	JailCards   int       `json:"jailCards"` // get-out-of-jail-free cards held
	Bankrupt    bool      `json:"bankrupt"`
	Token       string    `json:"token"`
	SeatIndex   int       `json:"seatIndex"`
	IsActive    bool      `json:"isActive"`
}

type Auction struct {
	TilePosition  int                `json:"tilePosition"`
	Bids          map[string]int     `json:"bids"` // bastionID -> bid
	HighestBid    int                `json:"highestBid"`
	HighestBidder string             `json:"highestBidder"`
}

type Trade struct {
	ID         uuid.UUID  `json:"id"`
	ProposerID string     `json:"proposerId"` // bastionID
	ReceiverID string     `json:"receiverId"`
	Offer      TradeOffer `json:"offer"`
	Request    TradeOffer `json:"request"`
	Status     string     `json:"status"` // pending | accepted | rejected
}

type TradeOffer struct {
	Cash       int   `json:"cash"`
	Properties []int `json:"properties"` // tile positions
	JailCards  int   `json:"jailCards"`
}

// Event is returned by every game action and broadcast to all players.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Phase constants
const (
	PhaseRoll     = "roll"
	PhasePostRoll = "post_roll" // awaiting buy/decline decision
	PhaseAuction  = "auction"
	PhaseTrade    = "trade"
)

// Status constants
const (
	StatusWaiting   = "waiting"
	StatusActive    = "active"
	StatusPaused    = "paused"
	StatusEnded     = "ended"
	StatusAbandoned = "abandoned"
)

// Fixed board positions
const (
	PosGO         = 0
	PosJail       = 10
	PosFreeParking = 20
	PosGoToJail   = 30
)
