package main

import (
	"testing"

	"github.com/trojsten/ksp-proboj/client"
)

// ===== Test Helpers =====

// createTestGame creates a minimal Game instance for testing
func createTestGame() *Game {
	return &Game{
		Map: &Map{
			Width:  10,
			Height: 10,
			Tiles:  make([][]Tile, 10),
		},
		Players:   make([]Player, 2),
		Ships:     make(map[int]*Ship),
		MaxShipId: 0,
		Harbors:   []Harbor{},
		Bases:     []Base{},
		Runner:    client.NewRunner(),
	}
}

// initializeTestMapTiles initializes the map tiles
func initializeTestMapTiles(g *Game) {
	for y := 0; y < g.Map.Height; y++ {
		g.Map.Tiles[y] = make([]Tile, g.Map.Width)
		for x := 0; x < g.Map.Width; x++ {
			g.Map.Tiles[y][x] = Tile{Type: TILE_WATER, Index: -1}
		}
	}
}

// createTestPlayer creates a Player instance for testing
func createTestPlayer(g *Game, index int, name string) Player {
	return Player{
		Index: index,
		Name:  name,
		Gold:  100,
		game:  g,
		Score: Score{},
		Statistics: Statistics{
			Kills:           map[string]int{},
			Damage:          map[string]int{},
			SellsByType:     map[int]int{},
			PurchasesByType: map[int]int{},
			TimeByShip:      map[string]int{},
			TimeOfResponses: 0,
		},
	}
}

// createTestShip creates a Ship instance for testing
func createTestShip(id, playerIndex, x, y int, shipType ShipType) *Ship {
	return &Ship{
		Id:          id,
		PlayerIndex: playerIndex,
		Type:        shipType,
		X:           x,
		Y:           y,
		Health:      shipType.Stats().MaxHealth,
		IsWreck:     false,
		Resources: Resources{
			Wood:      0,
			Stone:     0,
			Iron:      0,
			Gem:       0,
			Wool:      0,
			Hide:      0,
			Wheat:     0,
			Pineapple: 0,
			Gold:      50,
		},
	}
}

// createTestHarbor creates a Harbor instance for testing
// productionPositive: true = harbor produces (positive production), false = harbor consumes (negative production)
func createTestHarbor(x, y int, productionPositive bool, resourceType ResourceType) Harbor {
	h := Harbor{
		X: x,
		Y: y,
		Storage: Resources{
			Wood:      10,
			Stone:     10,
			Iron:      10,
			Gem:       10,
			Wool:      10,
			Hide:      10,
			Wheat:     10,
			Pineapple: 10,
			Gold:      0,
		},
	}

	if productionPositive {
		// Harbor produces this resource
		h.Production = Resources{
			Wood: 0, Stone: 0, Iron: 0, Gem: 0,
			Wool: 0, Hide: 0, Wheat: 0, Pineapple: 0, Gold: 0,
		}
		switch resourceType {
		case RESOURCE_WOOD:
			h.Production.Wood = 5
		case RESOURCE_STONE:
			h.Production.Stone = 5
		case RESOURCE_IRON:
			h.Production.Iron = 5
		case RESOURCE_GEM:
			h.Production.Gem = 5
		case RESOURCE_WOOL:
			h.Production.Wool = 5
		case RESOURCE_HIDE:
			h.Production.Hide = 5
		case RESOURCE_WHEAT:
			h.Production.Wheat = 5
		case RESOURCE_PINEAPPLE:
			h.Production.Pineapple = 5
		}
	} else {
		// Harbor consumes this resource
		h.Production = Resources{
			Wood: 0, Stone: 0, Iron: 0, Gem: 0,
			Wool: 0, Hide: 0, Wheat: 0, Pineapple: 0, Gold: 0,
		}
		switch resourceType {
		case RESOURCE_WOOD:
			h.Production.Wood = -5
		case RESOURCE_STONE:
			h.Production.Stone = -5
		case RESOURCE_IRON:
			h.Production.Iron = -5
		case RESOURCE_GEM:
			h.Production.Gem = -5
		case RESOURCE_WOOL:
			h.Production.Wool = -5
		case RESOURCE_HIDE:
			h.Production.Hide = -5
		case RESOURCE_WHEAT:
			h.Production.Wheat = -5
		case RESOURCE_PINEAPPLE:
			h.Production.Pineapple = -5
		}
	}

	return h
}

// ===== Test Cases =====

// TestTradeInvalidFormat tests parsing an invalid command format
func TestTradeInvalidFormat(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	commandedShips := make(map[int]bool)

	err := trade(g, &p, "invalid", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid command format, got nil")
	}
}

// TestTradeMultipleCommandsToSameShip tests that multiple commands to the same ship are rejected
func TestTradeMultipleCommandsToSameShip(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Mark ship as already commanded
	commandedShips := make(map[int]bool)
	commandedShips[1] = true

	err := trade(g, &p, "1 0 5", commandedShips)

	if err == nil {
		t.Errorf("Expected error for multiple commands to same ship, got nil")
	}
}

// TestTradeInvalidShipId tests that invalid ship ID returns error
func TestTradeInvalidShipId(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	commandedShips := make(map[int]bool)

	err := trade(g, &p, "999 0 5", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid ship ID, got nil")
	}
}

// TestTradeInvalidResourceType tests that RESOURCE_GOLD is rejected
func TestTradeInvalidResourceType(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Add harbor at ship location
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// RESOURCE_GOLD = 8
	err := trade(g, &p, "1 8 5", commandedShips)

	if err == nil {
		t.Errorf("Expected error for RESOURCE_GOLD, got nil")
	}
}

// TestTradeShipNotInHarbor tests that trade fails when ship is not at a harbor
func TestTradeShipNotInHarbor(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Add harbor at different location
	harbor := createTestHarbor(0, 0, true, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	err := trade(g, &p, "1 0 5", commandedShips)

	if err == nil {
		t.Errorf("Expected error when ship not in harbor, got nil")
	}
}

// TestTradeBuyWhenHarborDoesNotProduce tests that buying fails when harbor doesn't produce the resource
func TestTradeBuyWhenHarborDoesNotProduce(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Harbor consumes wood (Production = -5), doesn't produce it
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// Try to buy wood (amount > 0) from harbor that consumes wood
	err := trade(g, &p, "1 0 5", commandedShips)

	if err == nil {
		t.Errorf("Expected error when buying from harbor that doesn't produce, got nil")
	}
}

// TestTradeBuyInsufficientGold tests that buying fails when ship doesn't have enough gold
func TestTradeBuyInsufficientGold(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	p.Gold = 0
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Gold = 0 // No gold on ship either
	g.Ships[1] = ship

	// Harbor produces wood with storage
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 100
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// Try to buy 50 wood (will cost gold)
	err := trade(g, &p, "1 0 50", commandedShips)

	if err == nil {
		t.Errorf("Expected error when ship has insufficient gold, got nil")
	}
}

// TestTradeBuyInsufficientCargoSpace tests that buying fails when ship doesn't have cargo space
func TestTradeBuyInsufficientCargoSpace(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 50 // MaxCargo is 50, so ship is full
	g.Ships[1] = ship

	// Harbor produces wood with storage
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 100
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// Try to buy more wood
	err := trade(g, &p, "1 0 10", commandedShips)

	if err == nil {
		t.Errorf("Expected error when ship has insufficient cargo space, got nil")
	}
}

// TestTradeBuySuccess tests successful buying from harbor
func TestTradeBuySuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Harbor produces wood with storage
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 20
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	initialGold := ship.Resources.Gold
	initialHarborStorage := g.Harbors[0].Storage.Wood
	initialScore := p.Score.PurchasesFromHarbor

	// Buy 10 wood
	err := trade(g, &p, "1 0 10", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify ship received wood
	if ship.Resources.Wood != 10 {
		t.Errorf("Expected ship to have 10 wood, got %d", ship.Resources.Wood)
	}

	// Verify ship gold decreased
	if ship.Resources.Gold >= initialGold {
		t.Errorf("Expected ship gold to decrease, was %d, now %d", initialGold, ship.Resources.Gold)
	}

	// Verify harbor storage decreased
	if g.Harbors[0].Storage.Wood >= initialHarborStorage {
		t.Errorf("Expected harbor storage to decrease, was %d, now %d", initialHarborStorage, g.Harbors[0].Storage.Wood)
	}

	// Verify score was updated
	if p.Score.PurchasesFromHarbor <= initialScore {
		t.Errorf("Expected purchase score to increase, was %d, now %d", initialScore, p.Score.PurchasesFromHarbor)
	}

	// Verify statistics were updated
	if p.Statistics.PurchasesByType[int(RESOURCE_WOOD)] != 10 {
		t.Errorf("Expected purchase statistics for wood to be 10, got %d", p.Statistics.PurchasesByType[int(RESOURCE_WOOD)])
	}

	// Verify ship was marked as commanded
	if !commandedShips[1] {
		t.Errorf("Expected ship to be marked as commanded")
	}
}

// TestTradeSellWhenHarborDoesNotConsume tests that selling fails when harbor doesn't consume the resource
func TestTradeSellWhenHarborDoesNotConsume(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 10 // Ship has wood to sell
	g.Ships[1] = ship

	// Harbor produces wood (Production = 5), doesn't consume it
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// Try to sell wood (amount < 0) to harbor that produces wood
	err := trade(g, &p, "1 0 -10", commandedShips)

	if err == nil {
		t.Errorf("Expected error when selling to harbor that doesn't consume, got nil")
	}
}

// TestTradeSellSuccess tests successful selling to harbor
func TestTradeSellSuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 20 // Ship has wood to sell
	g.Ships[1] = ship

	// Harbor consumes wood
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	initialGold := ship.Resources.Gold
	initialWood := ship.Resources.Wood
	initialHarborStorage := g.Harbors[0].Storage.Wood
	initialScore := p.Score.SellsToHarbor
	initialGoldEarned := p.Score.GoldEarned

	// Sell 10 wood (negative amount)
	err := trade(g, &p, "1 0 -10", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify ship lost wood
	if ship.Resources.Wood != initialWood-10 {
		t.Errorf("Expected ship to have %d wood, got %d", initialWood-10, ship.Resources.Wood)
	}

	// Verify ship gold increased
	if ship.Resources.Gold <= initialGold {
		t.Errorf("Expected ship gold to increase, was %d, now %d", initialGold, ship.Resources.Gold)
	}

	// Verify harbor storage increased
	if g.Harbors[0].Storage.Wood <= initialHarborStorage {
		t.Errorf("Expected harbor storage to increase, was %d, now %d", initialHarborStorage, g.Harbors[0].Storage.Wood)
	}

	// Verify score was updated
	if p.Score.SellsToHarbor <= initialScore {
		t.Errorf("Expected sell score to increase, was %d, now %d", initialScore, p.Score.SellsToHarbor)
	}

	if p.Score.GoldEarned <= initialGoldEarned {
		t.Errorf("Expected gold earned score to increase, was %d, now %d", initialGoldEarned, p.Score.GoldEarned)
	}

	// Verify statistics were updated
	if p.Statistics.SellsByType[int(RESOURCE_WOOD)] != 10 {
		t.Errorf("Expected sell statistics for wood to be 10, got %d", p.Statistics.SellsByType[int(RESOURCE_WOOD)])
	}

	// Verify ship was marked as commanded
	if !commandedShips[1] {
		t.Errorf("Expected ship to be marked as commanded")
	}
}

// TestTradeBuyAmountLimitedByStorage tests that buy amount is limited by harbor storage
func TestTradeBuyAmountLimitedByStorage(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Harbor produces wood with limited storage
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 5 // Only 5 wood available
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// Try to buy 50 wood, but only 5 available
	err := trade(g, &p, "1 0 50", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should only get 5 wood (limited by storage)
	if ship.Resources.Wood != 5 {
		t.Errorf("Expected ship to receive 5 wood (limited by storage), got %d", ship.Resources.Wood)
	}

	// Harbor should be empty
	if g.Harbors[0].Storage.Wood != 0 {
		t.Errorf("Expected harbor storage to be 0, got %d", g.Harbors[0].Storage.Wood)
	}
}

// TestTradeSellAmountLimitedByShipResources tests that sell amount is limited by ship resources
func TestTradeSellAmountLimitedByShipResources(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 5 // Only 5 wood on ship
	g.Ships[1] = ship

	// Harbor consumes wood
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// Try to sell 50 wood, but only 5 available
	err := trade(g, &p, "1 0 -50", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should only sell 5 wood (limited by ship resources)
	if ship.Resources.Wood != 0 {
		t.Errorf("Expected ship to have 0 wood after selling all, got %d", ship.Resources.Wood)
	}

	// Harbor should receive 5 wood
	if g.Harbors[0].Storage.Wood != 15 { // Initial 10 + 5 sold
		t.Errorf("Expected harbor storage to be 15, got %d", g.Harbors[0].Storage.Wood)
	}
}

// TestTradeBuyZeroAmount tests that buying zero amount returns error
func TestTradeBuyZeroAmount(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Harbor produces wood but has no storage
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 0
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// Try to buy when harbor has no storage (amount will be 0)
	err := trade(g, &p, "1 0 10", commandedShips)

	if err == nil {
		t.Errorf("Expected error when resulting amount is 0, got nil")
	}
}

// TestTradeSellZeroAmount tests that selling zero amount returns error
func TestTradeSellZeroAmount(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 0 // No wood on ship
	g.Ships[1] = ship

	// Harbor consumes wood
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// Try to sell when ship has no wood (amount will be 0)
	err := trade(g, &p, "1 0 -10", commandedShips)

	if err == nil {
		t.Errorf("Expected error when resulting amount is 0, got nil")
	}
}

// TestTradeNilResource tests that nil resource (invalid resource type) returns error
func TestTradeNilResource(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Harbor produces wood
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	commandedShips := make(map[int]bool)

	// Use invalid resource type (100 - out of range)
	err := trade(g, &p, "1 100 5", commandedShips)

	if err == nil {
		t.Errorf("Expected error for nil resource type, got nil")
	}
}

// ===== Tests for move() =====

// TestMoveInvalidFormat tests parsing an invalid command format
func TestMoveInvalidFormat(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	commandedShips := make(map[int]bool)

	err := move(g, &p, "invalid", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid command format, got nil")
	}
}

// TestMoveMultipleCommandsToSameShip tests that multiple commands to the same ship are rejected
func TestMoveMultipleCommandsToSameShip(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	commandedShips := make(map[int]bool)
	commandedShips[1] = true

	err := move(g, &p, "1 6 6", commandedShips)

	if err == nil {
		t.Errorf("Expected error for multiple commands to same ship, got nil")
	}
}

// TestMoveInvalidShipId tests that invalid ship ID returns error
func TestMoveInvalidShipId(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	commandedShips := make(map[int]bool)

	err := move(g, &p, "999 6 6", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid ship ID, got nil")
	}
}

// TestMoveOutOfRange tests that moving out of range fails
func TestMoveOutOfRange(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	g.Ships[1] = ship

	commandedShips := make(map[int]bool)

	// Try to move to (10, 0) which is out of range (MaxMoveRange = 3)
	err := move(g, &p, "1 10 0", commandedShips)

	if err == nil {
		t.Errorf("Expected error for out of range move, got nil")
	}
}

// TestMoveToBlockedTile tests that moving to a tile with a non-wreck ship fails
func TestMoveToBlockedTile(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	g.Ships[1] = ship

	// Add another ship at target location
	enemyShip := createTestShip(2, 1, 1, 0, SmallMerchantShip{})
	g.Ships[2] = enemyShip

	commandedShips := make(map[int]bool)

	err := move(g, &p, "1 1 0", commandedShips)

	if err == nil {
		t.Errorf("Expected error when moving to tile with non-wreck ship, got nil")
	}
}

// TestMoveSuccess tests successful movement within range
func TestMoveSuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	g.Ships[1] = ship

	commandedShips := make(map[int]bool)

	err := move(g, &p, "1 3 0", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify ship moved
	if g.Ships[1].X != 3 || g.Ships[1].Y != 0 {
		t.Errorf("Expected ship at (3, 0), got (%d, %d)", g.Ships[1].X, g.Ships[1].Y)
	}

	// Verify ship was marked as commanded
	if !commandedShips[1] {
		t.Errorf("Expected ship to be marked as commanded")
	}
}

// TestMoveToWreck tests successful movement to tile with wreck
func TestMoveToWreck(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	g.Ships[1] = ship

	// Add wreck at target location
	wreck := createTestShip(2, 1, 1, 0, SmallMerchantShip{})
	wreck.IsWreck = true
	g.Ships[2] = wreck

	commandedShips := make(map[int]bool)

	err := move(g, &p, "1 1 0", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify ship moved to wreck position
	if g.Ships[1].X != 1 || g.Ships[1].Y != 0 {
		t.Errorf("Expected ship at (1, 0), got (%d, %d)", g.Ships[1].X, g.Ships[1].Y)
	}
}

// ===== Tests for price() =====

// TestPriceSmallAmount tests price calculation with small amount
func TestPriceSmallAmount(t *testing.T) {
	// price = min(100/(amount+3)+1, 4) * BASE_PRICE[resourceType]
	result := price(RESOURCE_WOOD, 5)
	expected := min(100/(5+3)+1, 4) * BASE_PRICE[RESOURCE_WOOD]

	if result != expected {
		t.Errorf("Expected price %d, got %d", expected, result)
	}
}

// TestPriceZeroAmount tests price calculation with zero amount
func TestPriceZeroAmount(t *testing.T) {
	result := price(RESOURCE_WOOD, 0)
	expected := min(100/(0+3)+1, 4) * BASE_PRICE[RESOURCE_WOOD]

	if result != expected {
		t.Errorf("Expected price %d, got %d", expected, result)
	}
}

// TestPriceLargeAmount tests price calculation with large amount (capped at 4)
func TestPriceLargeAmount(t *testing.T) {
	result := price(RESOURCE_GEM, 100)
	expected := min(100/(100+3)+1, 4) * BASE_PRICE[RESOURCE_GEM]

	if result != expected {
		t.Errorf("Expected price %d, got %d", expected, result)
	}
}

// TestPriceDifferentResources tests price calculation for different resources
func TestPriceDifferentResources(t *testing.T) {
	woodPrice := price(RESOURCE_WOOD, 10)
	gemPrice := price(RESOURCE_GEM, 10)

	// Should be different due to BASE_PRICE
	if woodPrice == gemPrice {
		t.Errorf("Expected different prices for different resource types")
	}
}

// ===== Tests for loot() =====

// TestLootInvalidFormat tests parsing an invalid command format
func TestLootInvalidFormat(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	commandedShips := make(map[int]bool)

	err := loot(g, &p, "invalid", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid command format, got nil")
	}
}

// TestLootMultipleCommandsToSameShip tests that multiple commands to the same ship are rejected
func TestLootMultipleCommandsToSameShip(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, LooterScooter{})
	g.Ships[1] = ship

	commandedShips := make(map[int]bool)
	commandedShips[1] = true

	err := loot(g, &p, "1 2", commandedShips)

	if err == nil {
		t.Errorf("Expected error for multiple commands to same ship, got nil")
	}
}

// TestLootInvalidShipId tests that invalid ship ID returns error
func TestLootInvalidShipId(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	commandedShips := make(map[int]bool)

	err := loot(g, &p, "999 2", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid ship ID, got nil")
	}
}

// TestLootTargetNotWreck tests that looting non-wreck ship fails
func TestLootTargetNotWreck(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, LooterScooter{})
	g.Ships[1] = ship

	// Add non-wreck ship
	enemyShip := createTestShip(2, 1, 6, 6, SmallMerchantShip{})
	g.Ships[2] = enemyShip

	commandedShips := make(map[int]bool)

	err := loot(g, &p, "1 2", commandedShips)

	if err == nil {
		t.Errorf("Expected error when looting non-wreck ship, got nil")
	}
}

// TestLootTargetDoesNotExist tests that looting non-existent ship fails
func TestLootTargetDoesNotExist(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, LooterScooter{})
	g.Ships[1] = ship

	commandedShips := make(map[int]bool)

	err := loot(g, &p, "1 999", commandedShips)

	if err == nil {
		t.Errorf("Expected error when looting non-existent ship, got nil")
	}
}

// TestLootSuccess tests successful looting from wreck
func TestLootSuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	lootingShip := createTestShip(1, 0, 5, 5, LooterScooter{})
	g.Ships[1] = lootingShip

	// Add wreck with resources
	wreck := createTestShip(2, 1, 6, 6, SmallMerchantShip{})
	wreck.IsWreck = true
	wreck.Resources.Gold = 100
	wreck.Resources.Wood = 50
	g.Ships[2] = wreck

	commandedShips := make(map[int]bool)

	initialGold := lootingShip.Resources.Gold

	err := loot(g, &p, "1 2", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify resources transferred (limited by cargo space and yield)
	if lootingShip.Resources.Gold <= initialGold {
		t.Errorf("Expected ship gold to increase, was %d, now %d", initialGold, lootingShip.Resources.Gold)
	}

	// Verify wreck resources emptied
	if g.Ships[2].Resources.Gold != 0 {
		t.Errorf("Expected wreck gold to be 0, got %d", g.Ships[2].Resources.Gold)
	}

	// Verify ship was marked as commanded
	if !commandedShips[1] {
		t.Errorf("Expected ship to be marked as commanded")
	}
}

// TestLootCargoSpaceLimited tests that looting is limited by cargo space
func TestLootCargoSpaceLimited(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	lootingShip := createTestShip(1, 0, 5, 5, LooterScooter{})
	lootingShip.Resources.Wood = 20 // Use some cargo space
	g.Ships[1] = lootingShip

	// Add wreck with lots of resources
	wreck := createTestShip(2, 1, 6, 6, SmallMerchantShip{})
	wreck.IsWreck = true
	wreck.Resources.Gold = 1000
	wreck.Resources.Wood = 1000
	g.Ships[2] = wreck

	commandedShips := make(map[int]bool)

	err := loot(g, &p, "1 2", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify cargo not exceeded (MaxCargo = 30)
	totalResources := lootingShip.Resources.countResources()
	if totalResources > lootingShip.Type.Stats().MaxCargo {
		t.Errorf("Expected cargo not to exceed %d, got %d", lootingShip.Type.Stats().MaxCargo, totalResources)
	}
}

// ===== Tests for minDistanceToHarborAndBase() =====

// TestMinDistanceToHarborOnly tests distance with only harbors
func TestMinDistanceToHarborOnly(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)

	// Add harbor at (0, 0)
	harbor := createTestHarbor(0, 0, true, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	distance := minDistanceToHarborAndBase(g, 5, 5)
	expected := 10

	if distance != expected {
		t.Errorf("Expected distance %d, got %d", expected, distance)
	}
}

// TestMinDistanceToBaseOnly tests distance with only bases
func TestMinDistanceToBaseOnly(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)

	// Add base at (0, 0)
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	distance := minDistanceToHarborAndBase(g, 5, 5)
	expected := 10

	if distance != expected {
		t.Errorf("Expected distance %d, got %d", expected, distance)
	}
}

// TestMinDistanceToBoth tests distance with both harbors and bases
func TestMinDistanceToBoth(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)

	// Add harbor at (0, 0)
	harbor := createTestHarbor(0, 0, true, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	// Add base at (10, 10)
	base := Base{X: 10, Y: 10, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	distance := minDistanceToHarborAndBase(g, 5, 5)
	expected := 10 // Distance to harbor at (0, 0)

	if distance != expected {
		t.Errorf("Expected distance %d, got %d", expected, distance)
	}
}

// TestMinDistanceNearBoth tests distance when position is near both
func TestMinDistanceNearBoth(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)

	// Add harbor at (5, 5)
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	// Add base at (6, 6)
	base := Base{X: 6, Y: 6, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	distance := minDistanceToHarborAndBase(g, 5, 6)
	expected := 1 // Distance to harbor at (5, 5) is 1, to base (6, 6) is 2

	if distance != expected {
		t.Errorf("Expected distance %d, got %d", expected, distance)
	}
}

// ===== Tests for shoot() =====

// TestShootInvalidFormat tests parsing an invalid command format
func TestShootInvalidFormat(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	commandedShips := make(map[int]bool)

	err := shoot(g, &p, "invalid", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid command format, got nil")
	}
}

// TestShootMultipleCommandsToSameShip tests that multiple commands to the same ship are rejected
func TestShootMultipleCommandsToSameShip(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SomalianPirateShip{})
	g.Ships[1] = ship

	commandedShips := make(map[int]bool)
	commandedShips[1] = true

	err := shoot(g, &p, "1 2", commandedShips)

	if err == nil {
		t.Errorf("Expected error for multiple commands to same ship, got nil")
	}
}

// TestShootInvalidShipId tests that invalid ship ID returns error
func TestShootInvalidShipId(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	commandedShips := make(map[int]bool)

	err := shoot(g, &p, "999 2", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid ship ID, got nil")
	}
}

// TestShootTargetDoesNotExist tests that shooting non-existent ship fails
func TestShootTargetDoesNotExist(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SomalianPirateShip{})
	g.Ships[1] = ship

	commandedShips := make(map[int]bool)

	err := shoot(g, &p, "1 999", commandedShips)

	if err == nil {
		t.Errorf("Expected error when shooting non-existent ship, got nil")
	}
}

// TestShootOutOfRange tests that shooting out of range fails
func TestShootOutOfRange(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	attacker := createTestShip(1, 0, 0, 0, SomalianPirateShip{})
	g.Ships[1] = attacker

	// Add enemy ship far away
	enemy := createTestShip(2, 1, 10, 0, SmallMerchantShip{})
	g.Ships[2] = enemy

	commandedShips := make(map[int]bool)

	err := shoot(g, &p, "1 2", commandedShips)

	if err == nil {
		t.Errorf("Expected error when shooting out of range, got nil")
	}
}

// TestShootTargetInHarborProtection tests that shooting ship in harbor protection radius fails
func TestShootTargetInHarborProtection(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	attacker := createTestShip(1, 0, 0, 0, SomalianPirateShip{})
	g.Ships[1] = attacker

	// Add harbor at (5, 0)
	harbor := createTestHarbor(5, 0, true, RESOURCE_WOOD)
	g.Harbors = append(g.Harbors, harbor)

	// Add enemy ship at (5, 2) which is within HARBOUR_DAMAGE_RADIUS/2 = 4 of harbor
	enemy := createTestShip(2, 1, 5, 2, SmallMerchantShip{})
	g.Ships[2] = enemy

	commandedShips := make(map[int]bool)

	err := shoot(g, &p, "1 2", commandedShips)

	if err == nil {
		t.Errorf("Expected error when shooting ship in harbor protection, got nil")
	}
}

// TestShootSuccess tests successful attack
func TestShootSuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	p2 := createTestPlayer(g, 1, "EnemyPlayer")
	g.Players[0] = p
	g.Players[1] = p2

	attacker := createTestShip(1, 0, 0, 0, SomalianPirateShip{})
	g.Ships[1] = attacker

	// Add enemy ship within range (Range = 2)
	enemy := createTestShip(2, 1, 2, 0, SmallMerchantShip{})
	g.Ships[2] = enemy

	commandedShips := make(map[int]bool)

	initialEnemyHealth := enemy.Health
	initialDamage := p.Statistics.Damage[p2.Name]

	err := shoot(g, &p, "1 2", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify enemy health decreased
	if g.Ships[2].Health >= initialEnemyHealth {
		t.Errorf("Expected enemy health to decrease, was %d, now %d", initialEnemyHealth, g.Ships[2].Health)
	}

	// Verify damage statistics updated
	if p.Statistics.Damage[p2.Name] <= initialDamage {
		t.Errorf("Expected damage statistics to increase, was %d, now %d", initialDamage, p.Statistics.Damage[p2.Name])
	}

	// Verify ship was marked as commanded
	if !commandedShips[1] {
		t.Errorf("Expected ship to be marked as commanded")
	}
}

// TestShootKill tests that destroying a ship updates kill score
func TestShootKill(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	p2 := createTestPlayer(g, 1, "EnemyPlayer")
	g.Players[0] = p
	g.Players[1] = p2

	attacker := createTestShip(1, 0, 0, 0, SomalianPirateShip{})
	g.Ships[1] = attacker

	// Add enemy ship with low health that will be destroyed
	enemy := createTestShip(2, 1, 2, 0, SmallMerchantShip{})
	enemy.Health = 2 // Will be killed by damage of 3
	g.Ships[2] = enemy

	commandedShips := make(map[int]bool)

	initialKills := p.Score.Kills

	err := shoot(g, &p, "1 2", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify enemy destroyed (health <= 0)
	if g.Ships[2].Health > 0 {
		t.Errorf("Expected enemy ship to be destroyed, health is %d", g.Ships[2].Health)
	}

	// Verify kill score updated
	if p.Score.Kills <= initialKills {
		t.Errorf("Expected kill score to increase, was %d, now %d", initialKills, p.Score.Kills)
	}
}

// ===== Tests for buy() =====

// TestBuyInvalidFormat tests parsing an invalid command format
func TestBuyInvalidFormat(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	// Add base for player
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	err := buy(g, &p, "invalid", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid command format, got nil")
	}
}

// TestBuyInvalidShipTypeId tests that invalid ship type ID returns error
func TestBuyInvalidShipTypeId(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	// Add base for player
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	// Ship type ID -1 is invalid
	err := buy(g, &p, "-1", commandedShips)

	if err == nil {
		t.Errorf("Expected error for negative ship type ID, got nil")
	}

	// Ship type ID 100 is out of range (ships has 8 elements)
	err = buy(g, &p, "100", commandedShips)

	if err == nil {
		t.Errorf("Expected error for out of range ship type ID, got nil")
	}
}

// TestBuyInsufficientGold tests that buying with insufficient gold fails
func TestBuyInsufficientGold(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	p.Gold = 5 // Not enough to buy a ship
	g.Players[0] = p

	// Add base for player
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	// Try to buy ship type 0 (Cln, price = 10)
	err := buy(g, &p, "0", commandedShips)

	if err == nil {
		t.Errorf("Expected error for insufficient gold, got nil")
	}
}

// TestBuySuccess tests successful ship purchase
func TestBuySuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	p.Gold = 100
	g.Players[0] = p

	// Add base for player
	base := Base{X: 5, Y: 5, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	initialGold := p.Gold
	initialMaxShipId := g.MaxShipId

	// Buy SmallMerchantShip (type 2, price = 100)
	err := buy(g, &p, "2", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify player gold decreased
	if p.Gold != initialGold-100 {
		t.Errorf("Expected player gold to be %d, got %d", initialGold-100, p.Gold)
	}

	// Verify new ship created
	if len(g.Ships) != 1 {
		t.Errorf("Expected 1 ship in game, got %d", len(g.Ships))
	}

	newShip := g.Ships[initialMaxShipId]
	if newShip == nil {
		t.Fatalf("Expected new ship to be created")
	}

	// Verify ship stats
	if newShip.PlayerIndex != p.Index {
		t.Errorf("Expected ship player index to be %d, got %d", p.Index, newShip.PlayerIndex)
	}

	if newShip.X != base.X || newShip.Y != base.Y {
		t.Errorf("Expected ship at base location (%d, %d), got (%d, %d)", base.X, base.Y, newShip.X, newShip.Y)
	}

	// Verify MaxShipId incremented
	if g.MaxShipId != initialMaxShipId+1 {
		t.Errorf("Expected MaxShipId to be %d, got %d", initialMaxShipId+1, g.MaxShipId)
	}
}

// TestBuyNoBase tests that buying fails when player has no base
func TestBuyNoBase(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	p.Gold = 100
	g.Players[0] = p

	// No base added for player

	commandedShips := make(map[int]bool)

	// Try to buy ship
	err := buy(g, &p, "2", commandedShips)

	// Should succeed but not create a ship (no base)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify no ship created
	if len(g.Ships) != 0 {
		t.Errorf("Expected no ship to be created when player has no base, got %d ships", len(g.Ships))
	}
}

// ===== Tests for store() =====

// TestStoreInvalidFormat tests parsing an invalid command format
func TestStoreInvalidFormat(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	// Add base for player
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	err := store(g, &p, "invalid", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid command format, got nil")
	}
}

// TestStoreMultipleCommandsToSameShip tests that multiple commands to the same ship are rejected
func TestStoreMultipleCommandsToSameShip(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	g.Ships[1] = ship

	// Add base at ship location
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)
	commandedShips[1] = true

	err := store(g, &p, "1 10", commandedShips)

	if err == nil {
		t.Errorf("Expected error for multiple commands to same ship, got nil")
	}
}

// TestStoreInvalidShipId tests that invalid ship ID returns error
func TestStoreInvalidShipId(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	// Add base for player
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	err := store(g, &p, "999 10", commandedShips)

	if err == nil {
		t.Errorf("Expected error for invalid ship ID, got nil")
	}
}

// TestStoreShipNotAtBase tests that storing fails when ship is not at base
func TestStoreShipNotAtBase(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Add base at different location
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	err := store(g, &p, "1 10", commandedShips)

	if err == nil {
		t.Errorf("Expected error when ship not at base, got nil")
	}
}

// TestStoreNoBase tests that storing fails when player has no base
func TestStoreNoBase(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	g.Ships[1] = ship

	// No base added

	commandedShips := make(map[int]bool)

	err := store(g, &p, "1 10", commandedShips)

	if err == nil {
		t.Errorf("Expected error when player has no base, got nil")
	}
}

// TestStoreGoldSuccess tests successful gold storage
func TestStoreGoldSuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	ship.Resources.Gold = 30
	g.Ships[1] = ship

	// Add base at ship location
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	initialPlayerGold := p.Gold
	initialShipGold := ship.Resources.Gold

	// Store 20 gold
	err := store(g, &p, "1 20", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify player gold increased
	if p.Gold != initialPlayerGold+20 {
		t.Errorf("Expected player gold to be %d, got %d", initialPlayerGold+20, p.Gold)
	}

	// Verify ship gold decreased
	if g.Ships[1].Resources.Gold != initialShipGold-20 {
		t.Errorf("Expected ship gold to be %d, got %d", initialShipGold-20, g.Ships[1].Resources.Gold)
	}

	// Verify ship was marked as commanded
	if !commandedShips[1] {
		t.Errorf("Expected ship to be marked as commanded")
	}
}

// TestStoreGoldLimitedByShip tests that storage is limited by ship gold
func TestStoreGoldLimitedByShip(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	ship.Resources.Gold = 10
	g.Ships[1] = ship

	// Add base at ship location
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	initialPlayerGold := p.Gold

	// Try to store 50 gold, but ship only has 10
	err := store(g, &p, "1 50", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify only 10 gold was stored
	if p.Gold != initialPlayerGold+10 {
		t.Errorf("Expected player gold to be %d, got %d", initialPlayerGold+10, p.Gold)
	}

	// Verify ship gold is now 0
	if g.Ships[1].Resources.Gold != 0 {
		t.Errorf("Expected ship gold to be 0, got %d", g.Ships[1].Resources.Gold)
	}
}

// TestWithdrawGoldSuccess tests successful gold withdrawal
func TestWithdrawGoldSuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	p.Gold = 50
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	ship.Resources.Gold = 10
	g.Ships[1] = ship

	// Add base at ship location
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	initialPlayerGold := p.Gold
	initialShipGold := ship.Resources.Gold

	// Withdraw 30 gold (negative amount)
	err := store(g, &p, "1 -30", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify player gold decreased
	if p.Gold != initialPlayerGold-30 {
		t.Errorf("Expected player gold to be %d, got %d", initialPlayerGold-30, p.Gold)
	}

	// Verify ship gold increased
	if g.Ships[1].Resources.Gold != initialShipGold+30 {
		t.Errorf("Expected ship gold to be %d, got %d", initialShipGold+30, g.Ships[1].Resources.Gold)
	}
}

// TestWithdrawGoldLimitedByPlayer tests that withdrawal is limited by player gold
func TestWithdrawGoldLimitedByPlayer(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	p.Gold = 10
	g.Players[0] = p

	ship := createTestShip(1, 0, 0, 0, SmallMerchantShip{})
	ship.Resources.Gold = 0
	g.Ships[1] = ship

	// Add base at ship location
	base := Base{X: 0, Y: 0, PlayerIndex: 0}
	g.Bases = append(g.Bases, base)

	commandedShips := make(map[int]bool)

	// Try to withdraw 50 gold, but player only has 10
	err := store(g, &p, "1 -50", commandedShips)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify only 10 gold was withdrawn
	if p.Gold != 0 {
		t.Errorf("Expected player gold to be 0, got %d", p.Gold)
	}

	// Verify ship gold is now 10
	if g.Ships[1].Resources.Gold != 10 {
		t.Errorf("Expected ship gold to be 10, got %d", g.Ships[1].Resources.Gold)
	}
}
