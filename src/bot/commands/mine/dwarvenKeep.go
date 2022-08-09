package mine

import (
	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/bwmarrin/discordgo"
)

/*
Construct building
* [Initial building] Dwarfen Keep (Limited storage, houses limited dwarfs)
* Tavern (Houses dwarfs)
* Mine
* Storage
* Smith & Forge

# Constructing takes time and cost money

Recuit dwarfs though the Tavern that will work in the mines.
Each additional recruit costs more money

X amount of workers per mine
Limit on amount of mines one can construct. Each additional mine costs more money

Ore can gained based on num of works and time passed until the 'storage' is full
Ore in storage can be sold or refined to increase their value
2x ore -> 1x ingot

# Ore or ingot can be manualy sold for money

X amount of workers per Smith & Forge
Limit on amount of Smith & Forges one can construct. Each additional mine costs more money

Able to upgrade buildings.
A normal tavern can only recruit X amount of dwarfs.
A normal mine can only allow x amount of dwarfs to work there.
A normal Smith&Forge can only allow x amount of dwarfs to work there.

Upgrading building increases (doubles) the output
*/
func Dwarvenkeep(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

}
