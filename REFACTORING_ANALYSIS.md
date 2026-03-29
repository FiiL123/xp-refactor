# Príležitosti na refaktorovanie - Server Kód

Tento dokument obsahuje 3 súbory s konkrétnymi príležitosťami na refaktorovanie serverového kódu s ukážkami existujúceho kódu a dôvodmi pre zmenu.



## 1. Refaktorovanie súboru `comms_cmds.go`

Súbor `comms_cmds.go` obsahuje viacero príležitostí na refaktorovanie, ktoré súvisia s duplicítou kódu a porušením princípov čistého kódu.

#### Existujúci kód (opakuje sa na viacerých miestach)

**1. Validácia lode** - opakuje sa 5-krát (move, trade, loot, shoot, store):
```go
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

**2. Funkcia trade()** - 75 riadkov so zložitou logikou:
```go
func trade(g *Game, p *Player, line string, commandedShips map[int]bool) error {
    // Parsovanie a validácia...
    if amount > 0 { // Nákup z prístavu
        // výpočet ceny, kontrola zlata, kontrola cargo miesta
        // prenos zdrojov, odpočítanie zlata, aktualizácia skóre
    } else { // Predaj do prístavu
        // výpočet ceny, kontrola zdrojov na lodi
        // prenos zdrojov, pridanie zlata, aktualizácia skóre
    }
}
```

**3. Prenos zdrojov v loot()** - opakuje sa 9-krát:
```go
remainingSpace := ship.Type.Stats().MaxCargo - ship.Resources.countResources()
ship.Resources.Gold += min(int(ship.Type.Stats().Yield*float32(wreckShip.Resources.Gold)), remainingSpace)

remainingSpace = ship.Type.Stats().MaxCargo - ship.Resources.countResources()
ship.Resources.Gem += min(int(ship.Type.Stats().Yield*float32(wreckShip.Resources.Gem)), remainingSpace)

// ... toto isté pre: Iron, Hide, Wool, Pineapple, Wheat, Stone, Wood
```

#### Porušenie princípov čistého kódu

- **DRY (Don't Repeat Yourself)** - rovnaký kód je zkopírovaný na viacerých miestach (5× validácia, 9× prenos zdrojov)
- **Single Responsibility Principle** - funkcia trade() robí 2 veci (nákup AND predaj)
- **Cyclomatic Complexity** - príliš mnoho vetvení v trade() (if/else)
- **Otvorený/Zatvorený princíp (OCP)** - pridanie nového zdroja vyžaduje úpravu repetitívneho kódu
- **Údržba** - ak sa zmení logika, musí sa upraviť na viacerých miestach
- **Čitateľnosť** - 75+ riadkov repetitívneho kódu je ťažké pochopiť
- **Testovateľnosť** - ťažké otestovať jednotlivé časti samostatne
- **Chybovosť** - pri kopírovaní sa ľahko urobí chyba

#### Navrhované riešenie

1. **Extrahovať validáciu lode** - vytvoriť pomocnú funkciu `validateAndMarkShipCommanded()`, ktorá skontroluje, či lode nebola už použité, získa lode a označí ju ako použité

2. **Rozdeliť funkciu trade()** - vytvoriť dve samostatné funkcie:
   - `buyFromHarbor()` - len nákup z prístavu (validácia produkcie, výpočet ceny, prenos zdrojov)
   - `sellToHarbor()` - len predaj do prístavu (validácia spotreby, výpočet ceny, prenos zdrojov)
   - `trade()` - len parsovanie a volanie správnej funkcie

3. **Zovšeobecniť prenos zdrojov v loot()** - vytvoriť funkciu `transferYieldToShip()`, ktorá:
   - vytvorí zoznam všetkých zdrojov ako slice
   - prejde cez ne v cykle
   - pre každý zdroj vypočíta remainingSpace a transferAmount
   - aktualizuje zdroj na lodi



## 2. Extrahovanie inkrementácie mapy

### Súbor: `statistics.go`

#### Existujúci kód (opakuje sa 5-krát)

```go
// V newKill() - riadky 12-19
func (statistics *Statistics) newKill(playerName string) {
    _, ok := statistics.Kills[playerName]
    if ok {
        statistics.Kills[playerName] += 1
    } else {
        statistics.Kills[playerName] = 1
    }
}

// To isté v: addDamage(), newSell(), newPurchase(), addTimeByShip()
```

Rovnaký vzor sa nachádza v: `newKill()`, `addDamage()`, `newSell()`, `newPurchase()`, `addTimeByShip()`

#### Porušenie princípov čistého kódu

- **DRY (Don't Repeat Yourself)** - rovnaký vzor 5-krát (check exists → increment or set)
- **Single Responsibility** - logika mapovania by mala byť zovšeobecnená
- **Údržba** - ak sa zmení logika inkrementácie, musí sa upraviť na 5 miestach
- **Čitateľnosť** - zbytočne rozšírený kód pre jednoduchú operáciu

#### Navrhované riešenie

1. **Vytvoriť pomocnú funkciu** `incrementMap(m map[string]int, key string, value int)`, ktorá:
   - skontroluje, či kľúč existuje
   - ak existuje → pripočíta hodnotu
   - ak neexistuje → nastaví hodnotu

2. **Nahradiť repetitívny kód** vo všetkých 5 funkciách volaním tejto pomocnej funkcie so správnymi parametrami


## 3. Zovšeobecnenie aplikovania poškodenia lodiám (Ship Damage Application)

### Súbor: `mainloop.go`

#### Existujúci kód (opakuje sa 2-krát)

```go
// Aplikovanie poškodenia z prístavov - riadky 67-77
for _, harbor := range g.Harbors {
    for _, ship := range g.Ships {
        if ship.Type.Stats().Class == SHIP_ATTACK {
            if dist(harbor.X, harbor.Y, ship.X, ship.Y) <= HARBOUR_DAMAGE_RADIUS {
                ship.Health -= HARBOUR_DAMAGE
                g.Runner.Log(fmt.Sprintf("attack ship %d was near harbour, so applying HARBOUR_DAMAGE", ship.Id))
            }
        }
    }
}

// Aplikovanie poškodenia zo základní - riadky 78-88 (takmer identické)
for _, base := range g.Bases {
    for _, ship := range g.Ships {
        if ship.PlayerIndex != base.PlayerIndex {
            if dist(base.X, base.Y, ship.X, ship.Y) <= BASE_DAMAGE_RADIUS {
                ship.Health -= BASE_DAMAGE
                g.Runner.Log(fmt.Sprintf("ship %d from player \"%s\" was near base of player \"%s\", so applying BASE_DAMAGE", ship.Id, g.Players[ship.PlayerIndex].Name, g.Players[base.PlayerIndex].Name))
            }
        }
    }
}
```

#### Porušenie princípov čistého kódu

- **DRY (Don't Repeat Yourself)** - rovnaký vzor dvakrát (nested loops + distance check + damage)
- **Single Responsibility** - logika aplikovania poškodenia by mala byť zovšeobecnená
- **Otvorený/Zatvorený princíp (OCP)** - pridanie nového zdroja poškodenia vyžaduje duplicitný kód
- **Testovateľnosť** - ťažké otestovať logiku poškodenia pre každý zdroj samostatne

#### Navrhované riešenie

1. **Vytvoriť štruktúru** `DamageSource` (X, Y, Radius, Damage, Name) pre reprezentáciu zdroja poškodenia

2. **Vytvoriť zovšeobecnenú funkciu** `applyDamageFromSources(g *Game, sources []DamageSource, filter func(*Ship, DamageSource) bool)`:
   - prejde všetky zdroje poškodenia
   - pre každý zdroj prejde všetky lode
   - aplikuje filter (napr. iba útočné lode pre prístavy, cudzie lode pre základne)
   - ak lode je v dosahu → aplikuje poškodenie

3. **Použitie v hlavnej slučke**:
   - vytvoriť `harborSources` z `g.Harbors` s filterom iba útočných lodí
   - vytvoriť `baseSources` z `g.Bases` s filterom cudzích lodí
   - zavolať `applyDamageFromSources()` pre obe skupiny
