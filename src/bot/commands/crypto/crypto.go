package crypto

import (
	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/bwmarrin/discordgo"
)

/*
	Buy and sell real crypto using in-game currency.

	https://api.coinbase.com/v2/prices/BTC-USD/buy
	https://api.coinbase.com/v2/prices/BTC-USD/sell

	'BTC' can be changed to any crypto on Coinbase
*/

// https://developers.coinbase.com/api/v2#get-buy-price

// Crypto-Buy - Buy crypto using in-game currency.
func Crypto(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

}
