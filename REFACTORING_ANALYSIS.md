# Príležitosti na refaktorovanie - Server Kód

Tento dokument obsahuje 3 konkrétne príležitosti na refaktorovanie serverového kódu s ukážkami existujúceho kódu a dôvodmi pre zmenu.

---

## 1. Extrahovanie validácie lode (Ship Command Validation)

### Súbor: `comms_cmds.go`

#### Existujúci kód (opakuje sa 5-krát)

```go
// V move() - riadky 14-22
_, exist := commandedShips[shipId]
if exist {
    return fmt.Errorf("multiple commands to ship %d. ignoring", shipId)
}
ship, err := p.Ship(g, shipId)
if err != nil {
    return err
}
commandedShips[shipId] = true
```

Rovnaký kód sa nachádza v: `move()`, `trade()`, `loot()`, `shoot()`, `store()`

#### Porušenie princípov čistého kódu

- **DRY (Don't Repeat Yourself)** - rovnaký kód je zkopírovaný na 5 miest
- **Single Responsibility** - validácia by mala byť v jednej funkcii
- **Údržba** - ak sa zmení logika validácie, musí sa upraviť na 5 miestach

#### Navrhované riešenie

```go
func validateAndMarkShipCommanded(g *Game, p *Player, shipId int, commandedShips map[int]bool) (*Ship, error) {
    _, exist := commandedShips[shipId]
    if exist {
        return nil, fmt.Errorf("multiple commands to ship %d. ignoring", shipId)
    }
    ship, err := p.Ship(g, shipId)
    if err != nil {
        return nil, err
    }
    commandedShips[shipId] = true
    return ship, nil
}
```

---

## 2. Rozdelenie `trade()` na `buyFromHarbor()` a `sellToHarbor()`

### Súbor: `comms_cmds.go` (riadky 44-119)

#### Existujúci kód (75 riadkov)

```go
func trade(g *Game, p *Player, line string, commandedShips map[int]bool) error {
    // Parsovanie...
    var shipId, resourceId, amount int
    _, err := fmt.Sscanf(line, "%d %d %d", &shipId, &resourceId, &amount)
    // ... validácia lode ...

    if amount > 0 { // Nákup z prístavu
        if *harbor.Production.Resource(ResourceType(resourceId)) <= 0 {
            return fmt.Errorf("cannot take resource...")
        }
        // Výpočet ceny, kontrola zlata, kontrola cargo miesta
        // Prenos zdrojov, odpočítanie zlata
        // Aktualizácia skóre
    } else { // Predaj do prístavu
        if *harbor.Production.Resource(ResourceType(resourceId)) >= 0 {
            return fmt.Errorf("cannot take resource...")
        }
        // Výpočet ceny, kontrola zdrojov na lodi
        // Prenos zdrojov, pridanie zlata
        // Aktualizácia skóre
    }
    return nil
}
```

#### Porušenie princípov čistého kódu

- **Single Responsibility Principle** - funkcia robí 2 veci: nákup AND predaj
- **Cyclomatic Complexity** - príliš mnoho vetvení (if/else)
- **Čitateľnosť** - 75 riadkov je ťažké pochopiť na jeden pohľad
- **Testovateľnosť** - ťažké otestovať nákup a predaj samostatne

#### Navrhované riešenie

```go
func buyFromHarbor(g *Game, p *Player, ship *Ship, harbor *Harbor, resourceId ResourceType, amount int) error {
    // Iba nákup z prístavu
    // - validácia produkcie prístavu
    // - výpočet a kontrola ceny
    // - kontrola cargo miesta
    // - prenos zdrojov a odpočítanie zlata
}

func sellToHarbor(g *Game, p *Player, ship *Ship, harbor *Harbor, resourceId ResourceType, amount int) error {
    // Iba predaj do prístavu
    // - validácia spotreby prístavu
    // - výpočet ceny
    // - kontrola zdrojov na lodi
    // - prenos zdrojov a pridanie zlata
}

func trade(g *Game, p *Player, line string, commandedShips map[int]bool) error {
    // Parsovanie, nájdenie prístavu
    // Volanie buyFromHarbor() alebo sellToHarbor()
}
```

---

## 3. Zovšeobecnenie prenosu zdrojov v `loot()`

### Súbor: `comms_cmds.go` (riadky 139-164)

#### Existujúci kód (opakuje sa 9-krát)

```go
remainingSpace := ship.Type.Stats().MaxCargo - ship.Resources.countResources()
ship.Resources.Gold += min(int(ship.Type.Stats().Yield*float32(wreckShip.Resources.Gold)), remainingSpace)

remainingSpace = ship.Type.Stats().MaxCargo - ship.Resources.countResources()
ship.Resources.Gem += min(int(ship.Type.Stats().Yield*float32(wreckShip.Resources.Gem)), remainingSpace)

remainingSpace = ship.Type.Stats().MaxCargo - ship.Resources.countResources()
ship.Resources.Iron += min(int(ship.Type.Stats().Yield*float32(wreckShip.Resources.Iron)), remainingSpace)

// ... toto isté sa opakuje pre: Hide, Wool, Pineapple, Wheat, Stone, Wood
```

#### Porušenie princípov čistého kódu

- **DRY (Don't Repeat Yourself)** - rovnaký vzor sa opakuje 9-krát
- **Otvorený/Zatvorený princíp (OCP)** - pridanie nového zdroja vyžaduje úpravu kódu
- **Chybovosť** - pri kopírovaní sa ľahko urobí chyba
- **Čitateľnosť** - 50+ riadkov repetitívneho kódu

#### Navrhované riešenie

```go
func transferYieldToShip(lootingShip *Ship, wreckShip *Ship) {
    // Získanie všetkých zdrojových polí ako slice
    resourceFields := []*int{
        &lootingShip.Resources.Gold, &lootingShip.Resources.Gem,
        &lootingShip.Resources.Iron, &lootingShip.Resources.Hide,
        &lootingShip.Resources.Wool, &lootingShip.Resources.Pineapple,
        &lootingShip.Resources.Wheat, &lootingShip.Resources.Stone,
        &lootingShip.Resources.Wood,
    }
    wreckFields := []int{
        wreckShip.Resources.Gold, wreckShip.Resources.Gem,
        wreckShip.Resources.Iron, wreckShip.Resources.Hide,
        wreckShip.Resources.Wool, wreckShip.Resources.Pineapple,
        wreckShip.Resources.Wheat, wreckShip.Resources.Stone,
        wreckShip.Resources.Wood,
    }

    // Iterácia cez všetky zdroje
    for i, dest := range resourceFields {
        remainingSpace := lootingShip.Type.Stats().MaxCargo - lootingShip.Resources.countResources()
        transferAmount := min(int(lootingShip.Type.Stats().Yield*float32(wreckFields[i])), remainingSpace)
        *dest += transferAmount
    }
}
```

---

## Zhrnutie

| # | Príležitosť | Riadky kódu | Zníženie duplicity |
|---|-------------|-------------|-------------------|
| 1 | Validácia lode | ~40 riadkov | 5× → 1× |
| 2 | Rozdelenie trade() | 75 riadkov | 1 funkcia → 3 funkcie |
| 3 | Generalizácia loot() | ~50 riadkov | 9× opakovanie → 1 cyklus |

### Spoločné benefity

✅ Lepšia testovateľnosť - každá funkcia sa dá otestovať samostatne
✅ Jednoduchšia údržba - zmena na jednom mieste
✅ Vyššia čitateľnosť - menšie funkcie s jasným účelom
✅ Menej chýb - menej kopírovania kódu
