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
		Index:  index,
		Name:   name,
		Gold:   100,
		game:   g,
		Score:  Score{},
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
			Wood:      0, Stone: 0, Iron: 0, Gem: 0,
			Wool:      0, Hide: 0, Wheat: 0, Pineapple: 0, Gold: 0,
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
			Wood:      0, Stone: 0, Iron: 0, Gem: 0,
			Wool:      0, Hide: 0, Wheat: 0, Pineapple: 0, Gold: 0,
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

// ===== Tests for buyFromHarbor() =====

// TestBuyFromHarborHarborDoesNotProduce tests error when harbor doesn't produce the resource
func TestBuyFromHarborHarborDoesNotProduce(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Harbor consumes wood (Production = -5), doesn't produce it
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)

	err := buyFromHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 10)

	if err == nil {
		t.Errorf("Expected error when harbor doesn't produce resource, got nil")
	}
}

// TestBuyFromHarborInsufficientGold tests error when ship doesn't have enough gold
func TestBuyFromHarborInsufficientGold(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	p.Gold = 0
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Gold = 0
	g.Ships[1] = ship

	// Harbor produces wood with storage
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 100

	err := buyFromHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 50)

	if err == nil {
		t.Errorf("Expected error when ship has insufficient gold, got nil")
	}
}

// TestBuyFromHarborInsufficientCargoSpace tests error when ship doesn't have cargo space
func TestBuyFromHarborInsufficientCargoSpace(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 50 // Full cargo (MaxCargo = 50)
	g.Ships[1] = ship

	// Harbor produces wood with storage
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 100

	err := buyFromHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 10)

	if err == nil {
		t.Errorf("Expected error when ship has insufficient cargo space, got nil")
	}
}

// TestBuyFromHarborAmountLimitedByStorage tests that amount is limited by harbor storage
func TestBuyFromHarborAmountLimitedByStorage(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Harbor produces wood with limited storage
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 5 // Only 5 available

	initialHarborStorage := harbor.Storage.Wood
	initialStorageWood := g.Ships[1].Resources.Wood

	err := buyFromHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 50)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should only get 5 wood (limited by storage)
	if g.Ships[1].Resources.Wood-initialStorageWood != 5 {
		t.Errorf("Expected ship to receive 5 wood, got %d", g.Ships[1].Resources.Wood-initialStorageWood)
	}

	// Harbor should be empty
	if harbor.Storage.Wood != 0 {
		t.Errorf("Expected harbor storage to be 0, got %d", harbor.Storage.Wood)
	}

	// Harbor storage should have decreased
	if harbor.Storage.Wood >= initialHarborStorage {
		t.Errorf("Expected harbor storage to decrease, was %d, now %d", initialHarborStorage, harbor.Storage.Wood)
	}
}

// TestBuyFromHarborSuccess tests successful purchase from harbor
func TestBuyFromHarborSuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Harbor produces wood with storage
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 20

	initialGold := g.Ships[1].Resources.Gold
	initialHarborStorage := harbor.Storage.Wood
	initialScore := p.Score.PurchasesFromHarbor

	err := buyFromHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 10)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify ship received wood
	if g.Ships[1].Resources.Wood != 10 {
		t.Errorf("Expected ship to have 10 wood, got %d", g.Ships[1].Resources.Wood)
	}

	// Verify ship gold decreased
	if g.Ships[1].Resources.Gold >= initialGold {
		t.Errorf("Expected ship gold to decrease, was %d, now %d", initialGold, g.Ships[1].Resources.Gold)
	}

	// Verify harbor storage decreased
	if harbor.Storage.Wood >= initialHarborStorage {
		t.Errorf("Expected harbor storage to decrease, was %d, now %d", initialHarborStorage, harbor.Storage.Wood)
	}

	// Verify score was updated
	if p.Score.PurchasesFromHarbor <= initialScore {
		t.Errorf("Expected purchase score to increase, was %d, now %d", initialScore, p.Score.PurchasesFromHarbor)
	}

	// Verify statistics were updated
	if p.Statistics.PurchasesByType[int(RESOURCE_WOOD)] != 10 {
		t.Errorf("Expected purchase statistics for wood to be 10, got %d", p.Statistics.PurchasesByType[int(RESOURCE_WOOD)])
	}
}

// TestBuyFromHarborUpdatesGameShip tests that g.Ships map is updated, not just local ship variable
func TestBuyFromHarborUpdatesGameShip(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	g.Ships[1] = ship

	// Harbor produces wood
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)
	harbor.Storage.Wood = 20

	err := buyFromHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 10)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The local 'ship' variable might not be updated, but g.Ships[1] should be
	if g.Ships[1].Resources.Wood != 10 {
		t.Errorf("Expected g.Ships[1] to have 10 wood, got %d", g.Ships[1].Resources.Wood)
	}

	if g.Ships[1].Resources.Gold == ship.Resources.Gold {
		t.Errorf("Expected g.Ships[1] gold to differ from local ship (should be updated)")
	}
}

// ===== Tests for sellToHarbor() =====

// TestSellToHarborHarborDoesNotConsume tests error when harbor doesn't consume the resource
func TestSellToHarborHarborDoesNotConsume(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 10
	g.Ships[1] = ship

	// Harbor produces wood (Production = 5), doesn't consume it
	harbor := createTestHarbor(5, 5, true, RESOURCE_WOOD)

	err := sellToHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 10)

	if err == nil {
		t.Errorf("Expected error when harbor doesn't consume resource, got nil")
	}
}

// TestSellToHarborInsufficientResources tests error when ship doesn't have enough resources
func TestSellToHarborInsufficientResources(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 5 // Only 5 wood
	g.Ships[1] = ship

	// Harbor consumes wood
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)

	err := sellToHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 10)

	if err == nil {
		t.Errorf("Expected error when ship has insufficient resources, got nil")
	}
}

// TestSellToHarborAmountLimitedByShipResources tests that amount is limited by ship resources
func TestSellToHarborAmountLimitedByShipResources(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 5 // Only 5 wood available
	g.Ships[1] = ship

	// Harbor consumes wood
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)
	initialHarborStorage := harbor.Storage.Wood

	err := sellToHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 50)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should only sell 5 wood (limited by ship resources)
	if g.Ships[1].Resources.Wood != 0 {
		t.Errorf("Expected ship to have 0 wood after selling all, got %d", g.Ships[1].Resources.Wood)
	}

	// Harbor should receive 5 wood
	if harbor.Storage.Wood != initialHarborStorage+5 {
		t.Errorf("Expected harbor storage to be %d, got %d", initialHarborStorage+5, harbor.Storage.Wood)
	}
}

// TestSellToHarborSuccess tests successful sale to harbor
func TestSellToHarborSuccess(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 20
	g.Ships[1] = ship

	// Harbor consumes wood
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)

	initialGold := g.Ships[1].Resources.Gold
	initialWood := g.Ships[1].Resources.Wood
	initialHarborStorage := harbor.Storage.Wood
	initialScore := p.Score.SellsToHarbor
	initialGoldEarned := p.Score.GoldEarned

	err := sellToHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 10)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify ship lost wood
	if g.Ships[1].Resources.Wood != initialWood-10 {
		t.Errorf("Expected ship to have %d wood, got %d", initialWood-10, g.Ships[1].Resources.Wood)
	}

	// Verify ship gold increased
	if g.Ships[1].Resources.Gold <= initialGold {
		t.Errorf("Expected ship gold to increase, was %d, now %d", initialGold, g.Ships[1].Resources.Gold)
	}

	// Verify harbor storage increased
	if harbor.Storage.Wood <= initialHarborStorage {
		t.Errorf("Expected harbor storage to increase, was %d, now %d", initialHarborStorage, harbor.Storage.Wood)
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
}

// TestSellToHarborUpdatesGameShip tests that g.Ships map is updated
func TestSellToHarborUpdatesGameShip(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 20
	g.Ships[1] = ship

	// Harbor consumes wood
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)

	err := sellToHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 10)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The local 'ship' variable might not be updated, but g.Ships[1] should be
	if g.Ships[1].Resources.Wood != 10 {
		t.Errorf("Expected g.Ships[1] to have 10 wood, got %d", g.Ships[1].Resources.Wood)
	}

	if g.Ships[1].Resources.Gold == ship.Resources.Gold {
		t.Errorf("Expected g.Ships[1] gold to differ from local ship (should be updated)")
	}
}

// TestSellToHarborAllResources tests selling all resources of a type
func TestSellToHarborAllResources(t *testing.T) {
	g := createTestGame()
	initializeTestMapTiles(g)
	p := createTestPlayer(g, 0, "TestPlayer")
	g.Players[0] = p

	ship := createTestShip(1, 0, 5, 5, SmallMerchantShip{})
	ship.Resources.Wood = 25
	g.Ships[1] = ship

	// Harbor consumes wood
	harbor := createTestHarbor(5, 5, false, RESOURCE_WOOD)

	err := sellToHarbor(g, &p, ship, &harbor, RESOURCE_WOOD, 25)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// All wood should be gone
	if g.Ships[1].Resources.Wood != 0 {
		t.Errorf("Expected ship to have 0 wood, got %d", g.Ships[1].Resources.Wood)
	}

	// Gold should have increased
	if g.Ships[1].Resources.Gold <= 50 {
		t.Errorf("Expected ship gold to be greater than 50, got %d", g.Ships[1].Resources.Gold)
	}
}
