# Leçon 9 : IBC et Paiements Cross-Chain

## Objectifs
- Comprendre le protocole IBC (Inter-Blockchain Communication)
- Configurer deux chaînes locales pour les tests IBC
- Implémenter le transfert de tokens entre chaînes
- Accepter les paiements en tokens IBC sur SkillChain

## Prérequis
- Leçon 8 complétée
- Docker installé (pour le relayer)
- Compréhension des bases de Cosmos SDK

---

## 9.1 Qu'est-ce que IBC ?

IBC (Inter-Blockchain Communication) est le protocole qui permet aux blockchains Cosmos de communiquer entre elles de manière sécurisée et trustless.

```
┌─────────────────────────────────────────────────────────────────────┐
│                      ARCHITECTURE IBC                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   SkillChain                    Relayer                 CosmosHub   │
│  ┌───────────┐                 ┌───────┐              ┌───────────┐│
│  │           │    Packets      │       │    Packets   │           ││
│  │  Module   │◄───────────────►│ Hermes│◄────────────►│  Module   ││
│  │  IBC      │                 │       │              │  IBC      ││
│  │           │                 └───────┘              │           ││
│  └───────────┘                                        └───────────┘│
│       │                                                     │      │
│       │ Light Client                           Light Client │      │
│       │ de CosmosHub                          de SkillChain │      │
│       ▼                                                     ▼      │
│  ┌───────────┐                                        ┌───────────┐│
│  │ Vérifier  │                                        │ Vérifier  ││
│  │ les preuves│                                       │ les preuves││
│  └───────────┘                                        └───────────┘│
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Composants clés :**
- **Light Clients** : Chaque chaîne maintient un light client de l'autre chaîne
- **Connections** : Liens établis entre deux chaînes via leurs light clients
- **Channels** : Canaux de communication pour des modules spécifiques
- **Packets** : Messages envoyés via les channels
- **Relayer** : Service off-chain qui transmet les packets entre les chaînes

---

## 9.2 Types de tokens IBC

Quand un token traverse IBC, il est "wrapped" avec un préfixe unique :

```
Token original sur Cosmos Hub:  uatom
                    │
                    │ IBC Transfer
                    ▼
Token sur SkillChain:  ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
                       └────────────────────────────────────────────────────────────────────────┘
                                              Hash du chemin IBC (port/channel/denom)
```

Le denom IBC est calculé comme : `ibc/` + SHA256(`transfer/channel-0/uatom`)

---

## 9.3 Configuration de deux chaînes locales

Pour tester IBC, nous avons besoin de deux chaînes. Créons une seconde chaîne "PayChain" qui simulera une source de paiement.

**Structure des dossiers :**
```
workspace/
├── skillchain/          # Notre marketplace
└── paychain/            # Chaîne de paiement (simule Cosmos Hub)
```

**Créer PayChain :**
```bash
cd workspace

# Créer une seconde chaîne simple
ignite scaffold chain paychain --address-prefix pay

cd paychain
```

**Configurer paychain/config.yml :**
```yaml
version: 1

accounts:
  - name: alice
    coins: ['100000000upay', '100000000stake']
  - name: bob
    coins: ['100000000upay', '100000000stake']
  - name: relayer
    coins: ['100000000upay', '100000000stake']

validators:
  - name: alice
    bonded: '100000000stake'

faucet:
  name: bob
  coins: ['10000upay', '100000stake']
  port: 4501  # Port différent de SkillChain

genesis:
  chain_id: "paychain-local-1"

# Ports différents de SkillChain
host:
  rpc: ":26659"      # SkillChain utilise 26657
  p2p: ":26658"      # SkillChain utilise 26656
  grpc: ":9092"      # SkillChain utilise 9090
  grpcWeb: ":9093"   # SkillChain utilise 9091
  api: ":1318"       # SkillChain utilise 1317
```

**Modifier skillchain/config.yml pour ajouter le compte relayer :**
```yaml
accounts:
  - name: alice
    coins: ['1000000uskill', '100000000stake']
  - name: bob
    coins: ['500000uskill', '100000000stake']
  - name: charlie
    coins: ['500000uskill', '100000000stake']
  - name: relayer
    coins: ['1000000uskill', '100000000stake']
```

---

## 9.4 Lancer les deux chaînes

**Terminal 1 - SkillChain :**
```bash
cd workspace/skillchain
ignite chain serve --reset-once
```

**Terminal 2 - PayChain :**
```bash
cd workspace/paychain
ignite chain serve --reset-once
```

**Vérifier que les deux chaînes tournent :**
```bash
# SkillChain
curl http://localhost:26657/status | jq '.result.node_info.network'
# "skillchain-local-1"

# PayChain
curl http://localhost:26659/status | jq '.result.node_info.network'
# "paychain-local-1"
```

---

## 9.5 Installer et configurer Hermes Relayer

Hermes est le relayer IBC le plus utilisé dans l'écosystème Cosmos.

**Installation :**
```bash
# macOS
brew install hermes

# Linux (télécharger le binaire)
wget https://github.com/informalsystems/hermes/releases/download/v1.10.0/hermes-v1.10.0-x86_64-unknown-linux-gnu.tar.gz
tar -xzf hermes-v1.10.0-x86_64-unknown-linux-gnu.tar.gz
sudo mv hermes /usr/local/bin/

# Vérifier
hermes version
```

**Créer la configuration Hermes :**

Créer `~/.hermes/config.toml` :
```toml
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = true

[mode.channels]
enabled = true

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '127.0.0.1'
port = 3000

[telemetry]
enabled = false

# Configuration SkillChain
[[chains]]
id = 'skillchain-local-1'
rpc_addr = 'http://127.0.0.1:26657'
grpc_addr = 'http://127.0.0.1:9090'
event_source = { mode = 'push', url = 'ws://127.0.0.1:26657/websocket', batch_delay = '200ms' }
rpc_timeout = '10s'
account_prefix = 'skill'
key_name = 'relayer'
store_prefix = 'ibc'
default_gas = 200000
max_gas = 1000000
gas_price = { price = 0.025, denom = 'uskill' }
gas_multiplier = 1.2
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
address_type = { derivation = 'cosmos' }

# Configuration PayChain
[[chains]]
id = 'paychain-local-1'
rpc_addr = 'http://127.0.0.1:26659'
grpc_addr = 'http://127.0.0.1:9092'
event_source = { mode = 'push', url = 'ws://127.0.0.1:26659/websocket', batch_delay = '200ms' }
rpc_timeout = '10s'
account_prefix = 'pay'
key_name = 'relayer'
store_prefix = 'ibc'
default_gas = 200000
max_gas = 1000000
gas_price = { price = 0.025, denom = 'upay' }
gas_multiplier = 1.2
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
address_type = { derivation = 'cosmos' }
```

---

## 9.6 Ajouter les clés du relayer à Hermes

```bash
# Exporter la clé relayer de SkillChain
skillchaind keys export relayer --unarmored-hex --unsafe > /tmp/skillchain_relayer.key

# Exporter la clé relayer de PayChain
paychaind keys export relayer --unarmored-hex --unsafe > /tmp/paychain_relayer.key

# Importer dans Hermes
hermes keys add --chain skillchain-local-1 --key-file /tmp/skillchain_relayer.key
hermes keys add --chain paychain-local-1 --key-file /tmp/paychain_relayer.key

# Vérifier
hermes keys list --chain skillchain-local-1
hermes keys list --chain paychain-local-1

# Nettoyer les clés temporaires
rm /tmp/skillchain_relayer.key /tmp/paychain_relayer.key
```

---

## 9.7 Créer le channel IBC

```bash
# Créer les clients, connection et channel en une commande
hermes create channel \
  --a-chain skillchain-local-1 \
  --b-chain paychain-local-1 \
  --a-port transfer \
  --b-port transfer \
  --new-client-connection --yes

# Output attendu:
# SUCCESS Channel {
#     ordering: Unordered,
#     a_side: ChannelSide {
#         chain: ChainId { id: "skillchain-local-1" },
#         client_id: ClientId("07-tendermint-0"),
#         connection_id: ConnectionId("connection-0"),
#         port_id: PortId("transfer"),
#         channel_id: ChannelId("channel-0"),
#     },
#     b_side: ChannelSide {
#         chain: ChainId { id: "paychain-local-1" },
#         ...
#         channel_id: ChannelId("channel-0"),
#     },
# }
```

**Vérifier les channels :**
```bash
# Sur SkillChain
skillchaind query ibc channel channels

# Sur PayChain
paychaind query ibc channel channels
```

---

## 9.8 Lancer le relayer

**Terminal 3 - Hermes :**
```bash
hermes start
```

Le relayer va maintenant écouter les événements sur les deux chaînes et transmettre les packets IBC.

---

## 9.9 Tester un transfert IBC

**Transférer des tokens de PayChain vers SkillChain :**
```bash
# Vérifier le solde initial sur SkillChain (alice)
skillchaind query bank balances $(skillchaind keys show alice -a)
# Seulement uskill et stake

# Transférer 1000 upay de PayChain vers SkillChain
paychaind tx ibc-transfer transfer \
  transfer \
  channel-0 \
  $(skillchaind keys show alice -a) \
  1000upay \
  --from alice \
  --chain-id paychain-local-1 \
  --yes

# Attendre quelques secondes que le relayer transmette le packet...

# Vérifier le solde sur SkillChain
skillchaind query bank balances $(skillchaind keys show alice -a)

# Output:
# balances:
# - amount: "1000"
#   denom: ibc/ABC123...  <-- Token IBC!
# - amount: "1000000"
#   denom: uskill
```

---

## 9.10 Modifier SkillChain pour accepter les tokens IBC

Nous devons permettre aux gigs d'accepter des paiements en tokens IBC.

**Modifier proto/skillchain/marketplace/gig.proto :**
```protobuf
message Gig {
  uint64 id = 1;
  string title = 2;
  string description = 3;
  string owner = 4;
  uint64 price = 5;
  string price_denom = 6;     // Nouveau: "uskill" ou "ibc/..."
  string category = 7;
  uint64 delivery_days = 8;
  string status = 9;
  int64 created_at = 10;
}
```

**Modifier le message CreateGig :**

**proto/skillchain/marketplace/tx.proto :**
```protobuf
message MsgCreateGig {
  string creator = 1;
  string title = 2;
  string description = 3;
  uint64 price = 4;
  string price_denom = 5;     // Nouveau champ
  string category = 6;
  uint64 delivery_days = 7;
}
```

**Régénérer les types :**
```bash
ignite generate proto-go
```

**Mettre à jour le handler CreateGig :**

**x/marketplace/keeper/msg_server_create_gig.go :**
```go
func (k msgServer) CreateGig(goCtx context.Context, msg *types.MsgCreateGig) (*types.MsgCreateGigResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // Valider le denom
    if !isValidDenom(msg.PriceDenom) {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid price denom")
    }
    
    // Valider le prix minimum (seulement pour uskill)
    if msg.PriceDenom == "uskill" {
        params := k.GetParams(ctx)
        minPrice, _ := sdk.NewIntFromString(params.MinGigPrice)
        if sdk.NewIntFromUint64(msg.Price).LT(minPrice) {
            return nil, errorsmod.Wrapf(
                sdkerrors.ErrInvalidRequest,
                "price must be at least %s uskill",
                params.MinGigPrice,
            )
        }
    }
    
    gig := types.Gig{
        Title:        msg.Title,
        Description:  msg.Description,
        Owner:        msg.Creator,
        Price:        msg.Price,
        PriceDenom:   msg.PriceDenom,
        Category:     msg.Category,
        DeliveryDays: msg.DeliveryDays,
        Status:       "open",
        CreatedAt:    ctx.BlockTime().Unix(),
    }
    
    id := k.AppendGig(ctx, gig)
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "gig_created",
            sdk.NewAttribute("id", fmt.Sprintf("%d", id)),
            sdk.NewAttribute("price_denom", msg.PriceDenom),
        ),
    )
    
    return &types.MsgCreateGigResponse{Id: id}, nil
}

// isValidDenom vérifie si le denom est valide
func isValidDenom(denom string) bool {
    // Denom natif
    if denom == "uskill" || denom == "stake" {
        return true
    }
    
    // Denom IBC (commence par "ibc/")
    if strings.HasPrefix(denom, "ibc/") && len(denom) == 68 {
        // ibc/ + 64 caractères hex
        return true
    }
    
    return false
}
```

---

## 9.11 Mettre à jour le système d'escrow pour les tokens IBC

**Modifier AcceptApplication pour gérer les tokens IBC :**

**x/marketplace/keeper/msg_server_accept_application.go :**
```go
func (k msgServer) AcceptApplication(goCtx context.Context, msg *types.MsgAcceptApplication) (*types.MsgAcceptApplicationResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // ... validations existantes ...
    
    // Récupérer le gig pour connaître le denom
    gig, _ := k.GetGig(ctx, application.GigId)
    
    // Préparer le montant avec le bon denom
    escrowAmount := sdk.NewCoins(sdk.NewCoin(gig.PriceDenom, sdk.NewIntFromUint64(application.ProposedPrice)))
    
    // Vérifier le solde du client dans le bon denom
    clientAddr, _ := sdk.AccAddressFromBech32(gig.Owner)
    clientBalance := k.bankKeeper.GetBalance(ctx, clientAddr, gig.PriceDenom)
    
    if clientBalance.Amount.LT(sdk.NewIntFromUint64(application.ProposedPrice)) {
        return nil, errorsmod.Wrapf(
            types.ErrInsufficientFunds,
            "client has %s but needs %s",
            clientBalance.String(),
            escrowAmount.String(),
        )
    }
    
    // Transférer vers l'escrow
    err := k.bankKeeper.SendCoinsFromAccountToModule(
        ctx,
        clientAddr,
        types.ModuleName,
        escrowAmount,
    )
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to lock funds in escrow")
    }
    
    // ... reste du code existant ...
}
```

**Modifier CompleteContract pour le paiement :**

**x/marketplace/keeper/msg_server_complete_contract.go :**
```go
func (k msgServer) CompleteContract(goCtx context.Context, msg *types.MsgCompleteContract) (*types.MsgCompleteContractResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    contract, _ := k.GetContract(ctx, msg.ContractId)
    gig, _ := k.GetGig(ctx, contract.GigId)
    
    // Calculer les montants
    params := k.GetParams(ctx)
    totalAmount := sdk.NewIntFromUint64(contract.Price)
    platformFee := totalAmount.Mul(sdk.NewIntFromUint64(params.PlatformFeePercent)).Quo(sdk.NewInt(100))
    freelancerAmount := totalAmount.Sub(platformFee)
    
    // Payer le freelancer dans le denom du gig
    freelancerAddr, _ := sdk.AccAddressFromBech32(contract.Freelancer)
    freelancerCoins := sdk.NewCoins(sdk.NewCoin(gig.PriceDenom, freelancerAmount))
    
    err := k.bankKeeper.SendCoinsFromModuleToAccount(
        ctx,
        types.ModuleName,
        freelancerAddr,
        freelancerCoins,
    )
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to release funds")
    }
    
    // Les frais plateforme restent dans le module (dans le denom IBC ou uskill)
    
    // ... reste du code ...
}
```

---

## 9.12 Ajouter un helper pour résoudre les denoms IBC

**x/marketplace/keeper/ibc_helper.go :**
```go
package keeper

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "strings"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
)

// GetIBCDenom calcule le denom IBC pour un token
func GetIBCDenom(port, channel, baseDenom string) string {
    // Le chemin est "port/channel/denom"
    path := fmt.Sprintf("%s/%s/%s", port, channel, baseDenom)
    
    // Hash SHA256 du chemin
    hash := sha256.Sum256([]byte(path))
    
    // Retourne "ibc/" + hash en hex majuscule
    return "ibc/" + strings.ToUpper(hex.EncodeToString(hash[:]))
}

// ParseIBCDenom essaie de décoder un denom IBC pour obtenir ses composants
// Retourne le baseDenom si ce n'est pas un denom IBC
func ParseIBCDenom(denom string) (baseDenom string, isIBC bool) {
    if !strings.HasPrefix(denom, "ibc/") {
        return denom, false
    }
    
    // Pour obtenir le baseDenom d'un IBC denom, il faudrait
    // interroger le module IBC transfer. Pour simplifier,
    // on retourne juste le hash.
    return denom, true
}

// GetDenomDisplayName retourne un nom lisible pour un denom
func GetDenomDisplayName(ctx sdk.Context, denom string) string {
    if denom == "uskill" {
        return "SKILL"
    }
    if denom == "stake" {
        return "STAKE"
    }
    if strings.HasPrefix(denom, "ibc/") {
        // Dans une vraie implémentation, on interrogerait les métadonnées
        return fmt.Sprintf("IBC/%s...", denom[4:10])
    }
    return denom
}
```

---

## 9.13 Query pour les denoms IBC disponibles

```bash
ignite scaffold query accepted-denoms --response denoms:string --module marketplace
```

**x/marketplace/keeper/query_accepted_denoms.go :**
```go
package keeper

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "skillchain/x/marketplace/types"
)

func (k Keeper) AcceptedDenoms(goCtx context.Context, req *types.QueryAcceptedDenomsRequest) (*types.QueryAcceptedDenomsResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // Liste des denoms natifs acceptés
    denoms := []string{"uskill", "stake"}
    
    // Ajouter les denoms IBC qui ont été utilisés dans des gigs
    allGigs := k.GetAllGig(ctx)
    seenDenoms := make(map[string]bool)
    seenDenoms["uskill"] = true
    seenDenoms["stake"] = true
    
    for _, gig := range allGigs {
        if !seenDenoms[gig.PriceDenom] {
            denoms = append(denoms, gig.PriceDenom)
            seenDenoms[gig.PriceDenom] = true
        }
    }
    
    return &types.QueryAcceptedDenomsResponse{
        Denoms: denoms,
    }, nil
}
```

---

## 9.14 Test complet du workflow IBC

```bash
# 1. Obtenir le denom IBC de upay sur SkillChain
# (après un transfert, le denom est visible dans les balances)
IBC_PAY_DENOM=$(skillchaind query bank balances $(skillchaind keys show alice -a) -o json | jq -r '.balances[] | select(.denom | startswith("ibc/")) | .denom')
echo "IBC Denom: $IBC_PAY_DENOM"

# 2. Alice (sur SkillChain) crée un gig payable en tokens IBC
skillchaind tx marketplace create-gig \
  "Cross-Chain Development" \
  "Build an IBC-enabled dApp" \
  1000 \
  "$IBC_PAY_DENOM" \
  "development" \
  14 \
  --from alice \
  --yes

# 3. Bob doit d'abord recevoir des tokens IBC
# Sur PayChain, transférer des upay vers Bob sur SkillChain
paychaind tx ibc-transfer transfer \
  transfer \
  channel-0 \
  $(skillchaind keys show bob -a) \
  5000upay \
  --from bob \
  --chain-id paychain-local-1 \
  --yes

# Attendre le relayer...
sleep 5

# Vérifier que Bob a reçu les tokens IBC
skillchaind query bank balances $(skillchaind keys show bob -a)

# 4. Charlie crée un profil freelancer
skillchaind tx marketplace create-profile \
  "Charlie Dev" "IBC Expert" "ibc,cosmos-sdk" 50000 \
  --from charlie --yes

# 5. Charlie postule au gig IBC
skillchaind tx marketplace apply-to-gig 0 "I can do this" 1000 14 --from charlie --yes

# 6. Bob (client) accepte - ses tokens IBC sont verrouillés en escrow
skillchaind tx marketplace accept-application 0 --from bob --yes

# Vérifier l'escrow (contient des tokens IBC)
skillchaind query bank balances $(skillchaind keys show -a marketplace)

# 7. Charlie livre
skillchaind tx marketplace deliver-contract 0 "Done!" --from charlie --yes

# 8. Bob valide - Charlie reçoit des tokens IBC
skillchaind tx marketplace complete-contract 0 --from bob --yes

# Vérifier que Charlie a reçu les tokens IBC
skillchaind query bank balances $(skillchaind keys show charlie -a)
```

---

## 9.15 Frontend - Sélecteur de token

**src/components/TokenSelector.tsx :**
```typescript
import { useQuery } from '@tanstack/react-query';
import * as api from '@/services/api';

interface Props {
  value: string;
  onChange: (denom: string) => void;
}

// Noms lisibles pour les denoms connus
const DENOM_NAMES: Record<string, string> = {
  uskill: 'SKILL',
  stake: 'STAKE',
};

export function TokenSelector({ value, onChange }: Props) {
  const { data: denoms } = useQuery({
    queryKey: ['marketplace', 'denoms'],
    queryFn: async () => {
      const response = await fetch('http://localhost:1317/skillchain/marketplace/accepted_denoms');
      const data = await response.json();
      return data.denoms as string[];
    },
  });
  
  const getDisplayName = (denom: string) => {
    if (DENOM_NAMES[denom]) return DENOM_NAMES[denom];
    if (denom.startsWith('ibc/')) return `IBC/${denom.slice(4, 10)}...`;
    return denom;
  };
  
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
    >
      {denoms?.map((denom) => (
        <option key={denom} value={denom}>
          {getDisplayName(denom)}
        </option>
      ))}
    </select>
  );
}
```

---

## 9.16 Diagramme du flux de paiement IBC

```
┌─────────────────────────────────────────────────────────────────────┐
│                    PAIEMENT IBC SUR SKILLCHAIN                       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   PayChain                  Hermes               SkillChain         │
│  ┌─────────┐               ┌───────┐            ┌─────────┐        │
│  │  Bob    │               │Relayer│            │  Bob    │        │
│  │ (upay)  │               │       │            │(ibc/pay)│        │
│  └────┬────┘               └───┬───┘            └────┬────┘        │
│       │                        │                     │              │
│       │ 1. IBC Transfer        │                     │              │
│       │────────────────────────┼────────────────────►│              │
│       │                        │                     │              │
│       │                        │                     │ 2. Accept    │
│       │                        │                     │    Application│
│       │                        │                     │    (lock IBC) │
│       │                        │                     ▼              │
│       │                        │              ┌───────────┐         │
│       │                        │              │  Escrow   │         │
│       │                        │              │ (ibc/pay) │         │
│       │                        │              └─────┬─────┘         │
│       │                        │                    │               │
│       │                        │                    │ 3. Complete   │
│       │                        │                    │    Contract   │
│       │                        │                    ▼               │
│       │                        │              ┌───────────┐         │
│       │                        │              │ Freelancer│         │
│       │                        │              │ (ibc/pay) │         │
│       │                        │              └───────────┘         │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Questions de révision

1. **Qu'est-ce qu'un Light Client dans le contexte IBC ?**

2. **Comment est calculé le denom d'un token IBC transféré ?**

3. **Quel est le rôle du relayer dans le protocole IBC ?**

4. **Pourquoi le denom IBC est-il un hash plutôt que le nom original du token ?**

5. **Comment vérifier qu'un channel IBC est correctement établi ?**

6. **Que se passe-t-il si le relayer s'arrête pendant un transfert IBC ?**

---

## Récapitulatif des commandes

```bash
# Configurer Hermes
hermes keys add --chain skillchain-local-1 --key-file key.hex
hermes create channel --a-chain skillchain-local-1 --b-chain paychain-local-1 ...
hermes start

# Transfert IBC
paychaind tx ibc-transfer transfer transfer channel-0 <recipient> 1000upay --from alice

# Vérifier les channels
skillchaind query ibc channel channels

# Vérifier les balances IBC
skillchaind query bank balances <address>
```

---

**Prochaine leçon** : Nous allons déployer SkillChain sur un testnet public et configurer le monitoring.
