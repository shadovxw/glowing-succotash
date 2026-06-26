// Package ws handles the WebSocket upgrade and per-client read loops.
// Message type definitions live in internal/proto to avoid import cycles.
package ws

import "github.com/shadovxw/monopoly/internal/proto"

// Re-export proto types so callers only need to import ws if they prefer.
type Envelope = proto.Envelope

const (
	MsgRollDice          = proto.MsgRollDice
	MsgBuyProperty       = proto.MsgBuyProperty
	MsgDeclineProperty   = proto.MsgDeclineProperty
	MsgBuildHouse        = proto.MsgBuildHouse
	MsgSellHouse         = proto.MsgSellHouse
	MsgBuildHotel        = proto.MsgBuildHotel
	MsgSellHotel         = proto.MsgSellHotel
	MsgMortgage          = proto.MsgMortgage
	MsgUnmortgage        = proto.MsgUnmortgage
	MsgTradePropose      = proto.MsgTradePropose
	MsgTradeAccept       = proto.MsgTradeAccept
	MsgTradeReject       = proto.MsgTradeReject
	MsgAuctionBid        = proto.MsgAuctionBid
	MsgAuctionResolve    = proto.MsgAuctionResolve
	MsgJailPay           = proto.MsgJailPay
	MsgJailUseCard       = proto.MsgJailUseCard
	MsgJailRoll          = proto.MsgJailRoll
	MsgEndTurn           = proto.MsgEndTurn
	MsgDeclareBankruptcy = proto.MsgDeclareBankruptcy
	MsgChat              = proto.MsgChat
	MsgWebRTCOffer       = proto.MsgWebRTCOffer
	MsgWebRTCAnswer      = proto.MsgWebRTCAnswer
	MsgWebRTCIce         = proto.MsgWebRTCIce
	MsgGameState         = proto.MsgGameState
	MsgPlayerJoined      = proto.MsgPlayerJoined
	MsgPlayerLeft        = proto.MsgPlayerLeft
	MsgGameStarted       = proto.MsgGameStarted
	MsgGameEnded         = proto.MsgGameEnded
	MsgYourTurn          = proto.MsgYourTurn
	MsgActionError       = proto.MsgActionError
	MsgChatMessage       = proto.MsgChatMessage
	MsgSystemMessage     = proto.MsgSystemMessage
)

var Encode = proto.Encode
