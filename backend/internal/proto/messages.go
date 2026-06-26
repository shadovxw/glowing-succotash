// Package proto defines the WebSocket message types shared between the room
// and ws packages to avoid import cycles.
package proto

import "encoding/json"

type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Inbound message types (client → server)
const (
	MsgRollDice          = "ROLL_DICE"
	MsgBuyProperty       = "BUY_PROPERTY"
	MsgDeclineProperty   = "DECLINE_PROPERTY"
	MsgBuildHouse        = "BUILD_HOUSE"
	MsgSellHouse         = "SELL_HOUSE"
	MsgBuildHotel        = "BUILD_HOTEL"
	MsgSellHotel         = "SELL_HOTEL"
	MsgMortgage          = "MORTGAGE"
	MsgUnmortgage        = "UNMORTGAGE"
	MsgTradePropose      = "TRADE_PROPOSE"
	MsgTradeAccept       = "TRADE_ACCEPT"
	MsgTradeReject       = "TRADE_REJECT"
	MsgAuctionBid        = "AUCTION_BID"
	MsgAuctionResolve    = "AUCTION_RESOLVE"
	MsgJailPay           = "JAIL_PAY"
	MsgJailUseCard       = "JAIL_USE_CARD"
	MsgJailRoll          = "JAIL_ROLL"
	MsgEndTurn           = "END_TURN"
	MsgDeclareBankruptcy = "DECLARE_BANKRUPTCY"
	MsgChat              = "CHAT_SEND"
	MsgWebRTCOffer       = "WEBRTC_OFFER"
	MsgWebRTCAnswer      = "WEBRTC_ANSWER"
	MsgWebRTCIce         = "WEBRTC_ICE"
)

// Outbound message types (server → client)
const (
	MsgGameState     = "GAME_STATE"
	MsgPlayerJoined  = "PLAYER_JOINED"
	MsgPlayerLeft    = "PLAYER_LEFT"
	MsgGameStarted   = "GAME_STARTED"
	MsgGameEnded     = "GAME_ENDED"
	MsgYourTurn      = "YOUR_TURN"
	MsgActionError   = "ACTION_ERROR"
	MsgChatMessage   = "CHAT_MESSAGE"
	MsgSystemMessage = "SYSTEM_MESSAGE"
)

// Inbound payload structs

type TilePositionPayload struct {
	TilePosition int `json:"tilePosition"`
}

type TradePayload struct {
	TradeID    string             `json:"tradeId,omitempty"`
	ReceiverID string             `json:"receiverId,omitempty"`
	Offer      *TradeOfferPayload `json:"offer,omitempty"`
	Request    *TradeOfferPayload `json:"request,omitempty"`
}

type TradeOfferPayload struct {
	Cash       int   `json:"cash"`
	Properties []int `json:"properties"`
	JailCards  int   `json:"jailCards"`
}

type AuctionBidPayload struct {
	Amount int `json:"amount"`
}

type ChatPayload struct {
	Message string `json:"message"`
}

type WebRTCPayload struct {
	TargetID string          `json:"targetId"`
	Data     json.RawMessage `json:"data"`
}

type BankruptcyPayload struct {
	CreditorID string `json:"creditorId"`
}

func Encode(msgType string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: msgType, Payload: raw})
}
