# Leçon 2 : Scaffold du projet SkillChain

## Objectifs
- Créer la blockchain SkillChain avec Ignite
- Comprendre la structure détaillée d'un projet Cosmos SDK
- Scaffolder le module marketplace
- Configurer les paramètres de la chaîne

## Prérequis
- Leçon 1 complétée
- Ignite CLI installé et fonctionnel

---

## 2.1 Création du projet SkillChain

```bash
# Créer la blockchain SkillChain
ignite scaffold chain skillchain --address-prefix skill --skip-git --default-denom skill

# --address-prefix skill : les adresses commenceront par "skill1..." - Par défaut, Cosmos utilise "cosmos1..."
# --default-denom skill : le token natif sera SKILL au lieu de stake
# --skip-git : ne pas initialiser un dépôt Git

cd skillchain
```

**Structure générée :**
```
skillchain/
├── app/
│   ├── app.go              # Wiring principal de l'application
│   ├── app_config.go       # Configuration depinject
│   ├── export.go           # Export de l'état
│   └── genesis.go          # Initialisation genesis
├── cmd/
│   └── skillchaind/        # Binary principal
│       ├── main.go
│       └── cmd/
│           ├── root.go     # Commande racine CLI
│           └── commands.go # Commandes additionnelles
├── proto/
│   └── skillchain/         # Définitions Protobuf
├── x/                      # Modules custom (vide pour l'instant)
├── config.yml              # Configuration Ignite
├── go.mod                  # Dépendances Go
└── readme.md
```

---

## 2.2 Comprendre app/app.go

Le fichier `app.go` est le cœur de l'application. Il définit quels modules sont inclus.

```go
// app/app.go (extrait simplifié)
package app

import (
    // Modules Cosmos SDK standards
    "github.com/cosmos/cosmos-sdk/x/auth"
    "github.com/cosmos/cosmos-sdk/x/bank"
    "github.com/cosmos/cosmos-sdk/x/staking"
    "github.com/cosmos/cosmos-sdk/x/gov"
    // ... autres imports
)

// App définit l'application blockchain
type App struct {
    *baseapp.BaseApp
    
    // Keepers des modules (accès au state)
    AccountKeeper    authkeeper.AccountKeeper
    BankKeeper       bankkeeper.Keeper
    StakingKeeper    *stakingkeeper.Keeper
    // ... autres keepers
}
```

**Modules standards inclus par défaut :**

| Module | Rôle |
|--------|------|
| `auth` | Gestion des comptes et authentification |
| `bank` | Transferts de tokens |
| `staking` | Délégation et validateurs |
| `gov` | Gouvernance on-chain |
| `mint` | Création de nouveaux tokens |
| `distribution` | Distribution des récompenses |
| `slashing` | Pénalités pour mauvais comportement |
| `ibc` | Communication inter-chaînes |

---

## 2.3 Scaffolder le module marketplace

Le module `marketplace` contiendra toute la logique métier de SkillChain.

```bash
# Créer le module marketplace avec dépendance au module bank
ignite scaffold module marketplace --dep bank

# --dep bank : permet d'utiliser le BankKeeper pour gérer les paiements
```

**Fichiers générés dans x/marketplace/ :**
```
x/marketplace/
├── keeper/
│   ├── keeper.go           # Keeper principal (accès au state)
│   ├── msg_server.go       # Handler des transactions
│   ├── query.go            # Handler des queries
│   ├── params.go           # Gestion des paramètres
│   └── genesis.go          # Import/export genesis
├── types/
│   ├── types.go            # Types de données
│   ├── keys.go             # Clés de stockage
│   ├── params.go           # Définition des paramètres
│   ├── genesis.go          # Type GenesisState
│   ├── errors.go           # Erreurs custom
│   └── expected_keepers.go # Interfaces des keepers externes
├── module/
│   ├── module.go           # Enregistrement du module
│   └── autocli.go          # Configuration CLI automatique
└── simulation/             # Tests de simulation
```

---

## 2.4 Anatomie du Keeper

Le Keeper est le composant central qui gère l'accès au state du module.

```go
// x/marketplace/keeper/keeper.go
package keeper

import (
    "cosmossdk.io/collections"
    storetypes "cosmossdk.io/store/types"
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "skillchain/x/marketplace/types"
)

type Keeper struct {
    cdc          codec.BinaryCodec
    storeService store.KVStoreService
    logger       log.Logger
    
    // Référence au BankKeeper pour les paiements
    bankKeeper types.BankKeeper
    
    // Authority pour les messages de gouvernance
    authority string
    
    // Schema pour les collections
    Schema collections.Schema
    
    // Paramètres du module
    Params collections.Item[types.Params]
}
```

**Pourquoi un Keeper ?**
- Encapsule tout l'accès au state
- Garantit la cohérence des données
- Permet l'injection de dépendances (autres keepers)
- Facilite les tests unitaires

---

## 2.5 Configuration personnalisée de config.yml

Modifions la configuration pour SkillChain :

```yaml
# config.yml
version: 1

build:
  proto:
    path: proto

accounts:
  - name: alice
    coins: ['1000000uskill', '100000000stake']
  - name: bob
    coins: ['500000uskill', '100000000stake']
  - name: charlie
    coins: ['500000uskill', '100000000stake']

validators:
  - name: alice
    bonded: '100000000stake'

faucet:
  name: bob
  coins: ['10000uskill', '100000stake']
  port: 4500

genesis:
  chain_id: "skillchain-local-1"
  app_state:
    staking:
      params:
        bond_denom: stake
    mint:
      params:
        mint_denom: stake

client:
  typescript:
    path: "ts-client"
  openapi:
    path: "docs/static/openapi.yml"
```

**Notes sur les dénominations :**
- `uskill` : micro-SKILL (1 SKILL = 1,000,000 uskill)
- `stake` : token de staking pour les validateurs
- Convention : préfixe `u` pour les unités minimales (comme `uatom`, `uosmo`)

---

## 2.6 Lancer et tester le module vide

```bash
# Lancer la chaîne
ignite chain serve

# Dans un autre terminal, vérifier que le module est chargé
skillchaind query marketplace --help

# Output attendu:
# Querying commands for the marketplace module
# 
# Usage:
#   skillchaind query marketplace [command]
# 
# Available Commands:
#   params      Query the module parameters
```

**Tester la query des paramètres :**
```bash
skillchaind query marketplace params

# Output:
# params: {}
```

---

## 2.7 Explorer les fichiers Protobuf

Les fichiers `.proto` définissent les structures de données et les services.

```
proto/skillchain/marketplace/
├── genesis.proto    # État genesis du module
├── params.proto     # Paramètres configurables
├── query.proto      # Service de requêtes (lecture)
├── tx.proto         # Service de transactions (écriture)
└── types.proto      # Types de données (sera créé plus tard)
```

**proto/skillchain/marketplace/params.proto :**
```protobuf
syntax = "proto3";
package skillchain.marketplace;

option go_package = "skillchain/x/marketplace/types";

import "amino/amino.proto";
import "gogoproto/gogo.proto";

// Params defines the parameters for the module
message Params {
  option (amino.name) = "skillchain/x/marketplace/Params";
  option (gogoproto.equal) = true;
  
  // Paramètres à ajouter selon nos besoins
}
```

---

## 2.8 Ajouter des paramètres custom au module

Ajoutons des paramètres pour contrôler les frais de la plateforme.

**Modifier proto/skillchain/marketplace/params.proto :**
```protobuf
syntax = "proto3";
package skillchain.marketplace;

option go_package = "skillchain/x/marketplace/types";

import "amino/amino.proto";
import "gogoproto/gogo.proto";
import "cosmos_proto/cosmos.proto";

message Params {
  option (amino.name) = "skillchain/x/marketplace/Params";
  option (gogoproto.equal) = true;
  
  // Commission de la plateforme en pourcentage (ex: 5 = 5%)
  uint64 platform_fee_percent = 1;
  
  // Durée minimum d'un contrat en secondes
  uint64 min_contract_duration = 2;
  
  // Montant minimum pour créer une mission (en uskill)
  string min_gig_price = 3 [
    (cosmos_proto.scalar) = "cosmos.Int"
  ];
}
```

**Régénérer le code Go :**
```bash
ignite generate proto-go
```

**Mettre à jour les valeurs par défaut dans x/marketplace/types/params.go :**
```go
// x/marketplace/types/params.go
package types

import (
    "cosmossdk.io/math"
)

// Default parameter values
var (
    DefaultPlatformFeePercent  = uint64(5)           // 5%
    DefaultMinContractDuration = uint64(86400)       // 1 jour en secondes
    DefaultMinGigPrice         = math.NewInt(10000)  // 10000 uskill = 0.01 SKILL
)

// NewParams creates a new Params instance
func NewParams(feePercent, minDuration uint64, minPrice math.Int) Params {
    return Params{
        PlatformFeePercent:  feePercent,
        MinContractDuration: minDuration,
        MinGigPrice:         minPrice.String(),
    }
}

// DefaultParams returns default module parameters
func DefaultParams() Params {
    return NewParams(
        DefaultPlatformFeePercent,
        DefaultMinContractDuration,
        DefaultMinGigPrice,
    )
}

// Validate validates the set of params
func (p Params) Validate() error {
    if p.PlatformFeePercent > 100 {
        return fmt.Errorf("platform fee cannot exceed 100%%")
    }
    
    minPrice, ok := math.NewIntFromString(p.MinGigPrice)
    if !ok || minPrice.IsNegative() {
        return fmt.Errorf("min gig price must be a positive integer")
    }
    
    return nil
}
```

---

## 2.9 Vérifier les modifications

```bash
# Relancer la chaîne
ignite chain serve --reset-once

# Vérifier les nouveaux paramètres
skillchaind query marketplace params

# Output attendu:
# params:
#   min_contract_duration: "86400"
#   min_gig_price: "10000"
#   platform_fee_percent: "5"
```

---

## 2.10 Test pratique

```bash
# 1. Vérifier la structure du projet
tree -L 2 x/marketplace/

# 2. Examiner le keeper généré
cat x/marketplace/keeper/keeper.go

# 3. Vérifier que le module bank est bien une dépendance
grep -r "BankKeeper" x/marketplace/

# 4. Tester les comptes configurés
skillchaind keys list
skillchaind query bank balances $(skillchaind keys show alice -a)

# 5. Vérifier le chain-id
skillchaind status | jq '.node_info.network'
# Output: "skillchain-local-1"
```

---

## Questions de révision

1. **Quelle option de `ignite scaffold chain` permet de définir le préfixe des adresses ?**

2. **Quel est le rôle du Keeper dans un module Cosmos SDK ?**

3. **Pourquoi ajoute-t-on `--dep bank` lors du scaffold du module marketplace ?**

4. **Dans quel fichier définit-on les structures de données partagées (messages, requêtes) ?**

5. **Quelle commande régénère le code Go à partir des fichiers Protobuf ?**

6. **Quelle est la convention de nommage pour les unités minimales de tokens (ex: ATOM, SKILL) ?**

---

## Récapitulatif des commandes

```bash
# Créer le projet
ignite scaffold chain skillchain --address-prefix skill

# Créer un module avec dépendance
ignite scaffold module marketplace --dep bank

# Régénérer le code Protobuf
ignite generate proto-go

# Lancer avec reset de l'état
ignite chain serve --reset-once

# Query les paramètres du module
skillchaind query marketplace params
```

---

**Prochaine leçon** : Nous allons créer les entités principales de SkillChain (Profile, Gig) avec les opérations CRUD.
