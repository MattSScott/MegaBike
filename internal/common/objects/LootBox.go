package objects

import (
	utils "SOMAS2023/internal/common/utils"
)

type ILootBox interface {
	IPhysicsObject
	GetTotalResources() float64
	GetColour() utils.Colour
}

type LootBox struct {
	*PhysicsObject
	colour    utils.Colour
	totalLoot float64				
}

// constructor
func GetLootBox() *LootBox {
	return &LootBox{
		PhysicsObject: GetPhysicsObject(0),
		colour:        utils.GenerateRandomColour(),    
		totalLoot:     utils.GenerateRandomFloat(2, 4), 
	}
}

// returns: the total loot of the lootbox
func (lb *LootBox) GetTotalResources() float64 {
	return lb.totalLoot
}

// returns: the color of the lootbox
func (lb *LootBox) GetColour() utils.Colour {
	return lb.colour
}
