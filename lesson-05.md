# Leçon 5 : Système d'Escrow pour les paiements

## Objectifs
- Comprendre le pattern escrow sur blockchain
- Utiliser le BankKeeper pour gérer les transferts
- Implémenter le verrouillage des fonds à l'acceptation
- Gérer la libération des fonds à la complétion

## Prérequis
- Leçon 4 complétée
- Workflow Application/Contract fonctionnel

---

## 5.1 Qu'est-ce qu'un Escrow ?

L'escrow est un mécanisme où les fonds sont verrouillés par un tiers de confiance jusqu'à ce que certaines conditions soient remplies.

**Flux SkillChain :**
```
┌────────────┐     lock funds      ┌─────────────────┐
│   CLIENT   │ ──────────────────► │  MODULE ACCOUNT │
│            │                     │    (escrow)     │
└────────────┘                     └────────┬────────┘
                                            │
                                            │ work completed
                                            │ + client approves
                                            ▼
                                   ┌─────────────────┐
                                   │   FREELANCER    │
                                   │  (- platform fee)│
                                   └─────────────────┘
```

**Avantages :**
- Le client ne peut pas partir sans payer
- Le freelancer est garanti d'être payé si le travail est validé
- La plateforme peut prélever des frais automatiquement
- Les litiges peuvent être arbitrés avec les fonds toujours disponibles

---

## 5.2 Module Account dans Cosmos SDK

Chaque module peut avoir son propre compte pour stocker des fonds. Le module `bank` gère les transferts vers/depuis ces comptes.

**Configurer le module account pour marketplace :**

Modifier **x/marketplace/types/expected_keepers.go** :
```go
package types

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface for the Bank module
type BankKeeper interface {
    // Envoyer des coins d'un compte utilisateur vers un autre
    SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
    
    // Envoyer des coins d'un compte utilisateur vers le module
    SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
    
    // Envoyer des coins du module vers un compte utilisateur
    SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
    
    // Vérifier le solde d'un compte
    GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
    
    // Vérifier tous les soldes d'un compte
    GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
    
    // Vérifier le solde d'un module
    GetModuleBalance(ctx context.Context, moduleName string, denom string) sdk.Coin
}

// AccountKeeper defines the expected interface for the Account module
type AccountKeeper interface {
    GetModuleAddress(moduleName string) sdk.AccAddress
    GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}
```

---

## 5.3 Enregistrer le Module Account

Le module doit déclarer qu'il peut détenir des fonds.

**Modifier x/marketplace/types/keys.go** :
```go
package types

const (
    // ModuleName defines the module name
    ModuleName = "marketplace"
    
    // StoreKey defines the primary module store key
    StoreKey = ModuleName
    
    // MemStoreKey defines the in-memory store key
    MemStoreKey = "mem_marketplace"
    
    // EscrowAccountName is the name of the escrow module account
    EscrowAccountName = "marketplace_escrow"
)
```

**Ajouter le module account dans app/app.go** :

Chercher la section `maccPerms` (module account permissions) et ajouter :
```go
// app/app.go
var maccPerms = map[string][]string{
    // ... autres modules ...
    marketplacetypes.ModuleName: nil,  // peut recevoir/envoyer des coins
}
```

---

## 5.4 Modifier AcceptApplication pour verrouiller les fonds

Quand un client accepte une application, les fonds doivent être transférés vers l'escrow.

**x/marketplace/keeper/msg_server_accept_application.go** (version complète avec escrow) :
```go
package keeper

import (
    "context"
    "fmt"
    
    errorsmod "cosmossdk.io/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "skillchain/x/marketplace/types"
)

func (k msgServer) AcceptApplication(goCtx context.Context, msg *types.MsgAcceptApplication) (*types.MsgAcceptApplicationResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // 1. Récupérer l'application
    application, found := k.GetApplication(ctx, msg.ApplicationId)
    if !found {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "application %d not found", msg.ApplicationId)
    }
    
    if application.Status != "pending" {
        return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "application is not pending")
    }
    
    // 2. Récupérer le gig
    gig, found := k.GetGig(ctx, application.GigId)
    if !found {
        return nil, errorsmod.Wrapf(types.ErrGigNotFound, "gig %d not found", application.GigId)
    }
    
    if gig.Owner != msg.Creator {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only gig owner can accept applications")
    }
    
    if gig.Status != "open" {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "gig is no longer open")
    }
    
    // 3. Convertir l'adresse du client
    clientAddr, err := sdk.AccAddressFromBech32(gig.Owner)
    if err != nil {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid client address")
    }
    
    // 4. Préparer le montant à verrouiller
    escrowAmount := sdk.NewCoins(sdk.NewCoin("uskill", sdk.NewIntFromUint64(application.ProposedPrice)))
    
    // 5. Vérifier que le client a suffisamment de fonds
    clientBalance := k.bankKeeper.GetBalance(ctx, clientAddr, "uskill")
    if clientBalance.Amount.LT(sdk.NewIntFromUint64(application.ProposedPrice)) {
        return nil, errorsmod.Wrapf(
            types.ErrInsufficientFunds,
            "client has %s but needs %s",
            clientBalance.String(),
            escrowAmount.String(),
        )
    }
    
    // 6. Transférer les fonds vers le module escrow
    err = k.bankKeeper.SendCoinsFromAccountToModule(
        ctx,
        clientAddr,
        types.ModuleName,  // Le module marketplace détient l'escrow
        escrowAmount,
    )
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to lock funds in escrow")
    }
    
    // 7. Mettre à jour l'application
    application.Status = "accepted"
    k.SetApplication(ctx, application)
    
    // 8. Mettre à jour le gig
    gig.Status = "in_progress"
    k.SetGig(ctx, gig)
    
    // 9. Rejeter les autres candidatures
    allApplications := k.GetAllApplication(ctx)
    for _, app := range allApplications {
        if app.GigId == application.GigId && app.Id != application.Id && app.Status == "pending" {
            app.Status = "rejected"
            k.SetApplication(ctx, app)
        }
    }
    
    // 10. Créer le contrat
    deliveryDeadline := ctx.BlockTime().Unix() + int64(application.ProposedDays*86400)
    
    contract := types.Contract{
        GigId:            application.GigId,
        ApplicationId:    application.Id,
        Client:           gig.Owner,
        Freelancer:       application.Freelancer,
        Price:            application.ProposedPrice,
        DeliveryDeadline: deliveryDeadline,
        Status:           "active",
        CreatedAt:        ctx.BlockTime().Unix(),
        CompletedAt:      0,
    }
    
    contractId := k.AppendContract(ctx, contract)
    
    // 11. Événements
    ctx.EventManager().EmitEvents(sdk.Events{
        sdk.NewEvent(
            "funds_locked_in_escrow",
            sdk.NewAttribute("contract_id", fmt.Sprintf("%d", contractId)),
            sdk.NewAttribute("client", gig.Owner),
            sdk.NewAttribute("amount", escrowAmount.String()),
        ),
        sdk.NewEvent(
            "contract_created",
            sdk.NewAttribute("contract_id", fmt.Sprintf("%d", contractId)),
            sdk.NewAttribute("client", gig.Owner),
            sdk.NewAttribute("freelancer", application.Freelancer),
            sdk.NewAttribute("price", fmt.Sprintf("%d", application.ProposedPrice)),
        ),
    })
    
    return &types.MsgAcceptApplicationResponse{
        ContractId: contractId,
    }, nil
}
```

---

## 5.5 Implémenter DeliverContract

Le freelancer marque le travail comme livré.

**x/marketplace/keeper/msg_server_deliver_contract.go** :
```go
package keeper

import (
    "context"
    "fmt"
    
    errorsmod "cosmossdk.io/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "skillchain/x/marketplace/types"
)

func (k msgServer) DeliverContract(goCtx context.Context, msg *types.MsgDeliverContract) (*types.MsgDeliverContractResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // 1. Récupérer le contrat
    contract, found := k.GetContract(ctx, msg.ContractId)
    if !found {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", msg.ContractId)
    }
    
    // 2. Vérifier que le caller est le freelancer
    if contract.Freelancer != msg.Creator {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only freelancer can deliver")
    }
    
    // 3. Vérifier le statut
    if contract.Status != "active" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "contract must be active to deliver (current: %s)",
            contract.Status,
        )
    }
    
    // 4. Mettre à jour le statut
    contract.Status = "delivered"
    k.SetContract(ctx, contract)
    
    // 5. Événement
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "contract_delivered",
            sdk.NewAttribute("contract_id", fmt.Sprintf("%d", msg.ContractId)),
            sdk.NewAttribute("freelancer", msg.Creator),
            sdk.NewAttribute("delivery_note", msg.DeliveryNote),
        ),
    )
    
    return &types.MsgDeliverContractResponse{}, nil
}
```

---

## 5.6 Implémenter CompleteContract avec paiement

Quand le client valide, les fonds sont libérés au freelancer (moins les frais plateforme).

**x/marketplace/keeper/msg_server_complete_contract.go** :
```go
package keeper

import (
    "context"
    "fmt"
    
    errorsmod "cosmossdk.io/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "skillchain/x/marketplace/types"
)

func (k msgServer) CompleteContract(goCtx context.Context, msg *types.MsgCompleteContract) (*types.MsgCompleteContractResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // 1. Récupérer le contrat
    contract, found := k.GetContract(ctx, msg.ContractId)
    if !found {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", msg.ContractId)
    }
    
    // 2. Vérifier que le caller est le client
    if contract.Client != msg.Creator {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only client can complete the contract")
    }
    
    // 3. Vérifier le statut (doit être "delivered")
    if contract.Status != "delivered" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "contract must be delivered to complete (current: %s)",
            contract.Status,
        )
    }
    
    // 4. Calculer les montants
    params := k.GetParams(ctx)
    totalAmount := sdk.NewIntFromUint64(contract.Price)
    
    // Frais plateforme = prix * feePercent / 100
    platformFee := totalAmount.Mul(sdk.NewIntFromUint64(params.PlatformFeePercent)).Quo(sdk.NewInt(100))
    freelancerAmount := totalAmount.Sub(platformFee)
    
    // 5. Convertir l'adresse du freelancer
    freelancerAddr, err := sdk.AccAddressFromBech32(contract.Freelancer)
    if err != nil {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid freelancer address")
    }
    
    // 6. Transférer les fonds de l'escrow vers le freelancer
    freelancerCoins := sdk.NewCoins(sdk.NewCoin("uskill", freelancerAmount))
    err = k.bankKeeper.SendCoinsFromModuleToAccount(
        ctx,
        types.ModuleName,
        freelancerAddr,
        freelancerCoins,
    )
    if err != nil {
        return nil, errorsmod.Wrap(err, "failed to release funds to freelancer")
    }
    
    // 7. Les frais plateforme restent dans le module account
    // (ou peuvent être envoyés vers un compte treasury)
    
    // 8. Mettre à jour le contrat
    contract.Status = "completed"
    contract.CompletedAt = ctx.BlockTime().Unix()
    k.SetContract(ctx, contract)
    
    // 9. Mettre à jour le gig
    gig, _ := k.GetGig(ctx, contract.GigId)
    gig.Status = "completed"
    k.SetGig(ctx, gig)
    
    // 10. Mettre à jour les stats du freelancer
    profile, found := k.GetProfile(ctx, contract.Freelancer)
    if found {
        profile.TotalJobs++
        profile.TotalEarned += freelancerAmount.Uint64()
        k.SetProfile(ctx, profile)
    }
    
    // 11. Événements
    ctx.EventManager().EmitEvents(sdk.Events{
        sdk.NewEvent(
            "contract_completed",
            sdk.NewAttribute("contract_id", fmt.Sprintf("%d", msg.ContractId)),
            sdk.NewAttribute("client", contract.Client),
            sdk.NewAttribute("freelancer", contract.Freelancer),
        ),
        sdk.NewEvent(
            "payment_released",
            sdk.NewAttribute("contract_id", fmt.Sprintf("%d", msg.ContractId)),
            sdk.NewAttribute("freelancer", contract.Freelancer),
            sdk.NewAttribute("amount", freelancerCoins.String()),
            sdk.NewAttribute("platform_fee", platformFee.String()),
        ),
    })
    
    return &types.MsgCompleteContractResponse{}, nil
}
```

---

## 5.7 Ajouter une query pour le solde escrow

```bash
ignite scaffold query escrow-balance --response balance:cosmos.base.v1beta1.Coin --module marketplace
```

**x/marketplace/keeper/query_escrow_balance.go** :
```go
package keeper

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    
    "skillchain/x/marketplace/types"
)

func (k Keeper) EscrowBalance(goCtx context.Context, req *types.QueryEscrowBalanceRequest) (*types.QueryEscrowBalanceResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }
    
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // Récupérer le solde du module account
    balance := k.bankKeeper.GetModuleBalance(ctx, types.ModuleName, "uskill")
    
    return &types.QueryEscrowBalanceResponse{
        Balance: &balance,
    }, nil
}
```

---

## 5.8 Tests complets du système d'escrow

```bash
# Relancer la chaîne
ignite chain serve --reset-once
```

**Tests :**

```bash
# === SETUP ===

# Vérifier les soldes initiaux
skillchaind query bank balances $(skillchaind keys show alice -a)
# alice: 1000000uskill

skillchaind query bank balances $(skillchaind keys show bob -a)
# bob: 500000uskill

# Vérifier le solde escrow initial (devrait être 0)
skillchaind query marketplace escrow-balance
# balance: "0uskill"

# === CRÉER LE WORKFLOW ===

# Alice crée son profil freelancer
skillchaind tx marketplace create-profile \
  "Alice Dev" "Expert developer" "golang,react" 50000 \
  --from alice --yes

# Bob crée un gig (500000 uskill = 0.5 SKILL)
skillchaind tx marketplace create-gig \
  "Build a Web App" \
  "Need a full-stack web application" \
  500000 "development" 14 \
  --from bob --yes

# Alice postule au gig
skillchaind tx marketplace apply-to-gig \
  0 "I can build this app" 500000 14 \
  --from alice --yes

# === ESCROW: LOCK ===

# Avant d'accepter - vérifier le solde de Bob
skillchaind query bank balances $(skillchaind keys show bob -a)
# 500000uskill

# Bob accepte l'application d'Alice
skillchaind tx marketplace accept-application 0 --from bob --yes

# Vérifier que les fonds sont verrouillés
skillchaind query bank balances $(skillchaind keys show bob -a)
# 0uskill (les 500000 sont dans l'escrow)

skillchaind query marketplace escrow-balance
# balance: "500000uskill"

# Le contrat est créé
skillchaind query marketplace list-contract

# === DELIVERY ===

# Alice livre le travail
skillchaind tx marketplace deliver-contract 0 "Work completed, check the repo" --from alice --yes

# Vérifier le statut
skillchaind query marketplace show-contract 0
# status: delivered

# === COMPLETION & PAYMENT ===

# Solde d'Alice avant paiement
skillchaind query bank balances $(skillchaind keys show alice -a)
# 1000000uskill (solde initial)

# Bob valide et complète le contrat
skillchaind tx marketplace complete-contract 0 --from bob --yes

# === VÉRIFICATIONS FINALES ===

# Solde d'Alice après paiement (prix - 5% frais plateforme)
# 500000 - 5% = 475000 uskill reçus
skillchaind query bank balances $(skillchaind keys show alice -a)
# 1475000uskill (1000000 + 475000)

# L'escrow est vide pour ce contrat
# Note: les 25000 uskill de frais restent dans le module
skillchaind query marketplace escrow-balance
# balance: "25000uskill" (frais plateforme accumulés)

# Profil d'Alice mis à jour
skillchaind query marketplace show-profile $(skillchaind keys show alice -a)
# total_jobs: 1
# total_earned: 475000

# Contrat et gig marqués comme completed
skillchaind query marketplace show-contract 0
# status: completed
# completed_at: <timestamp>

skillchaind query marketplace show-gig 0
# status: completed
```

---

## 5.9 Gérer les cas d'erreur

**Test: Fonds insuffisants**

```bash
# Charlie crée un gig à 1000000 uskill mais n'a que 500000
skillchaind tx marketplace create-gig \
  "Expensive Project" "Very complex work" \
  1000000 "consulting" 30 \
  --from charlie --yes

# Alice postule
skillchaind tx marketplace apply-to-gig \
  1 "I can do this" 1000000 30 \
  --from alice --yes

# Charlie essaie d'accepter mais n'a pas assez de fonds
skillchaind tx marketplace accept-application 1 --from charlie --yes
# Erreur: client has 500000uskill but needs 1000000uskill
```

**Test: Transitions de statut invalides**

```bash
# Essayer de compléter un contrat qui n'est pas "delivered"
# (Créer un nouveau workflow d'abord)
skillchaind tx marketplace create-gig "Test Gig" "Test" 50000 "test" 7 --from bob --yes
skillchaind tx marketplace apply-to-gig 2 "Test" 50000 7 --from alice --yes
skillchaind tx marketplace accept-application 2 --from bob --yes

# Essayer de compléter directement (sans delivery)
skillchaind tx marketplace complete-contract 1 --from bob --yes
# Erreur: contract must be delivered to complete
```

---

## 5.10 Diagramme du flux financier

```
┌─────────────────────────────────────────────────────────────────────┐
│                      FLUX FINANCIER SKILLCHAIN                       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────┐                                                       │
│  │  CLIENT  │                                                       │
│  │ (500,000)│                                                       │
│  └────┬─────┘                                                       │
│       │                                                             │
│       │ accept_application                                          │
│       │ (lock 500,000 uskill)                                       │
│       ▼                                                             │
│  ┌──────────────────────┐                                          │
│  │   MODULE ESCROW      │                                          │
│  │   (500,000 uskill)   │                                          │
│  └──────────┬───────────┘                                          │
│             │                                                       │
│             │ complete_contract                                     │
│             │                                                       │
│       ┌─────┴─────┐                                                │
│       ▼           ▼                                                │
│  ┌──────────┐  ┌──────────────────┐                                │
│  │FREELANCER│  │ PLATFORM FEES    │                                │
│  │(475,000) │  │ (25,000 = 5%)    │                                │
│  │  95%     │  │ reste en module  │                                │
│  └──────────┘  └──────────────────┘                                │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Questions de révision

1. **Quelle méthode du BankKeeper permet de transférer des fonds vers un module ?**

2. **Pourquoi vérifie-t-on le solde du client AVANT d'appeler `SendCoinsFromAccountToModule` ?**

3. **Comment calcule-t-on les frais de plateforme dans `CompleteContract` ?**

4. **Que se passe-t-il si on essaie de compléter un contrat qui n'est pas en statut "delivered" ?**

5. **Où sont stockés les frais de plateforme après la complétion d'un contrat ?**

6. **Pourquoi le statut "delivered" existe-t-il entre "active" et "completed" ?**

---

## Récapitulatif des commandes

```bash
# Vérifier les soldes
skillchaind query bank balances <address>
skillchaind query marketplace escrow-balance

# Workflow complet
skillchaind tx marketplace accept-application <app_id> --from <client>  # Lock funds
skillchaind tx marketplace deliver-contract <contract_id> --from <freelancer>
skillchaind tx marketplace complete-contract <contract_id> --from <client>  # Release funds

# Vérifier les mises à jour
skillchaind query marketplace show-contract <id>
skillchaind query marketplace show-profile <address>
```

---

**Prochaine leçon** : Nous allons implémenter le système de disputes et d'arbitrage pour gérer les conflits entre clients et freelancers.
