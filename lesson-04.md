# Leçon 4 : Applications et Contracts

## Objectifs
- Créer l'entité Application (candidature à une mission)
- Créer l'entité Contract (accord client-freelancer)
- Implémenter le workflow complet de candidature
- Gérer les relations entre entités

## Prérequis
- Leçon 3 complétée
- Entités Profile et Gig fonctionnelles

---

## 4.1 Modèle de données SkillChain

Voici les relations entre les entités :

```
┌─────────────┐         ┌─────────────┐
│   Profile   │         │     Gig     │
│  (freelancer)│         │   (client)  │
└──────┬──────┘         └──────┬──────┘
       │                       │
       │ applique à            │ reçoit
       │                       │
       ▼                       ▼
┌─────────────────────────────────────┐
│            Application              │
│  (freelancer → gig)                 │
└──────────────────┬──────────────────┘
                   │
                   │ acceptée devient
                   ▼
┌─────────────────────────────────────┐
│             Contract                │
│  (client ↔ freelancer)              │
└─────────────────────────────────────┘
```

---

## 4.2 Créer l'entité Application

Une Application lie un freelancer à un Gig avec une proposition.

```bash
ignite scaffold list application \
  gig_id:uint \
  freelancer:string \
  cover_letter:string \
  proposed_price:uint \
  proposed_days:uint \
  status:string \
  created_at:int \
  --module marketplace \
  --no-message
```

**Statuts d'une Application :**
- `pending` : En attente de réponse du client
- `accepted` : Acceptée (génère un Contract)
- `rejected` : Refusée par le client
- `withdrawn` : Retirée par le freelancer

---

## 4.3 Créer l'entité Contract

Un Contract représente l'accord formel entre client et freelancer.

```bash
ignite scaffold list contract \
  gig_id:uint \
  application_id:uint \
  client:string \
  freelancer:string \
  price:uint \
  delivery_deadline:int \
  status:string \
  created_at:int \
  completed_at:int \
  --module marketplace \
  --no-message
```

**Statuts d'un Contract :**
- `active` : Travail en cours
- `delivered` : Freelancer a livré
- `completed` : Client a validé, paiement effectué
- `disputed` : Litige ouvert
- `cancelled` : Annulé

---

## 4.4 Créer les messages pour les Applications

```bash
# Freelancer postule à un gig
ignite scaffold message apply-to-gig \
  gig_id:uint \
  cover_letter:string \
  proposed_price:uint \
  proposed_days:uint \
  --module marketplace

# Freelancer retire sa candidature
ignite scaffold message withdraw-application \
  application_id:uint \
  --module marketplace

# Client accepte une candidature
ignite scaffold message accept-application \
  application_id:uint \
  --module marketplace

# Client rejette une candidature
ignite scaffold message reject-application \
  application_id:uint \
  --module marketplace
```

---

## 4.5 Créer les messages pour les Contracts

```bash
# Freelancer marque comme livré
ignite scaffold message deliver-contract \
  contract_id:uint \
  delivery_note:string \
  --module marketplace

# Client valide la livraison
ignite scaffold message complete-contract \
  contract_id:uint \
  --module marketplace

# Client ou freelancer ouvre un litige
ignite scaffold message dispute-contract \
  contract_id:uint \
  reason:string \
  --module marketplace
```

---

## 4.6 Implémenter ApplyToGig

**x/marketplace/keeper/msg_server_apply_to_gig.go :**
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

func (k msgServer) ApplyToGig(goCtx context.Context, msg *types.MsgApplyToGig) (*types.MsgApplyToGigResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // 1. Vérifier que le gig existe et est ouvert
    gig, found := k.GetGig(ctx, msg.GigId)
    if !found {
        return nil, errorsmod.Wrapf(types.ErrGigNotFound, "gig %d not found", msg.GigId)
    }
    
    if gig.Status != "open" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest, 
            "gig %d is not open for applications (status: %s)", 
            msg.GigId, 
            gig.Status,
        )
    }
    
    // 2. Vérifier que le freelancer a un profil
    _, hasProfile := k.GetProfile(ctx, msg.Creator)
    if !hasProfile {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "freelancer must have a profile to apply")
    }
    
    // 3. Vérifier que le freelancer n'est pas le owner du gig
    if gig.Owner == msg.Creator {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot apply to your own gig")
    }
    
    // 4. Vérifier qu'il n'a pas déjà postulé
    allApplications := k.GetAllApplication(ctx)
    for _, app := range allApplications {
        if app.GigId == msg.GigId && app.Freelancer == msg.Creator && app.Status == "pending" {
            return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "you already have a pending application for this gig")
        }
    }
    
    // 5. Valider le prix proposé (peut être différent du prix demandé)
    params := k.GetParams(ctx)
    minPrice, _ := sdk.NewIntFromString(params.MinGigPrice)
    if sdk.NewIntFromUint64(msg.ProposedPrice).LT(minPrice) {
        return nil, errorsmod.Wrap(types.ErrInvalidPrice, "proposed price is below minimum")
    }
    
    // 6. Créer l'application
    application := types.Application{
        GigId:         msg.GigId,
        Freelancer:    msg.Creator,
        CoverLetter:   msg.CoverLetter,
        ProposedPrice: msg.ProposedPrice,
        ProposedDays:  msg.ProposedDays,
        Status:        "pending",
        CreatedAt:     ctx.BlockTime().Unix(),
    }
    
    id := k.AppendApplication(ctx, application)
    
    // 7. Émettre l'événement
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "application_submitted",
            sdk.NewAttribute("application_id", fmt.Sprintf("%d", id)),
            sdk.NewAttribute("gig_id", fmt.Sprintf("%d", msg.GigId)),
            sdk.NewAttribute("freelancer", msg.Creator),
            sdk.NewAttribute("proposed_price", fmt.Sprintf("%d", msg.ProposedPrice)),
        ),
    )
    
    return &types.MsgApplyToGigResponse{
        ApplicationId: id,
    }, nil
}
```

**Modifier proto/skillchain/marketplace/tx.proto pour le response :**
```protobuf
message MsgApplyToGigResponse {
  uint64 application_id = 1;
}
```

---

## 4.7 Implémenter AcceptApplication

**x/marketplace/keeper/msg_server_accept_application.go :**
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
    
    // 2. Vérifier le statut
    if application.Status != "pending" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "application %d is not pending (status: %s)",
            msg.ApplicationId,
            application.Status,
        )
    }
    
    // 3. Récupérer le gig associé
    gig, found := k.GetGig(ctx, application.GigId)
    if !found {
        return nil, errorsmod.Wrapf(types.ErrGigNotFound, "gig %d not found", application.GigId)
    }
    
    // 4. Vérifier que le caller est le owner du gig
    if gig.Owner != msg.Creator {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only gig owner can accept applications")
    }
    
    // 5. Vérifier que le gig est toujours ouvert
    if gig.Status != "open" {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "gig is no longer open")
    }
    
    // 6. Mettre à jour l'application
    application.Status = "accepted"
    k.SetApplication(ctx, application)
    
    // 7. Mettre à jour le gig
    gig.Status = "in_progress"
    k.SetGig(ctx, gig)
    
    // 8. Rejeter automatiquement les autres candidatures pending
    allApplications := k.GetAllApplication(ctx)
    for _, app := range allApplications {
        if app.GigId == application.GigId && app.Id != application.Id && app.Status == "pending" {
            app.Status = "rejected"
            k.SetApplication(ctx, app)
        }
    }
    
    // 9. Calculer la deadline de livraison
    deliveryDeadline := ctx.BlockTime().Unix() + int64(application.ProposedDays*86400)
    
    // 10. Créer le contrat
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
            "application_accepted",
            sdk.NewAttribute("application_id", fmt.Sprintf("%d", msg.ApplicationId)),
            sdk.NewAttribute("gig_id", fmt.Sprintf("%d", application.GigId)),
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

**Modifier le response :**
```protobuf
message MsgAcceptApplicationResponse {
  uint64 contract_id = 1;
}
```

---

## 4.8 Implémenter WithdrawApplication

**x/marketplace/keeper/msg_server_withdraw_application.go :**
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

func (k msgServer) WithdrawApplication(goCtx context.Context, msg *types.MsgWithdrawApplication) (*types.MsgWithdrawApplicationResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // 1. Récupérer l'application
    application, found := k.GetApplication(ctx, msg.ApplicationId)
    if !found {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "application %d not found", msg.ApplicationId)
    }
    
    // 2. Vérifier que le caller est le freelancer
    if application.Freelancer != msg.Creator {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only the freelancer can withdraw their application")
    }
    
    // 3. Vérifier que l'application est pending
    if application.Status != "pending" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "can only withdraw pending applications (current status: %s)",
            application.Status,
        )
    }
    
    // 4. Mettre à jour le statut
    application.Status = "withdrawn"
    k.SetApplication(ctx, application)
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "application_withdrawn",
            sdk.NewAttribute("application_id", fmt.Sprintf("%d", msg.ApplicationId)),
            sdk.NewAttribute("freelancer", msg.Creator),
        ),
    )
    
    return &types.MsgWithdrawApplicationResponse{}, nil
}
```

---

## 4.9 Implémenter RejectApplication

**x/marketplace/keeper/msg_server_reject_application.go :**
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

func (k msgServer) RejectApplication(goCtx context.Context, msg *types.MsgRejectApplication) (*types.MsgRejectApplicationResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // 1. Récupérer l'application
    application, found := k.GetApplication(ctx, msg.ApplicationId)
    if !found {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "application %d not found", msg.ApplicationId)
    }
    
    // 2. Récupérer le gig
    gig, found := k.GetGig(ctx, application.GigId)
    if !found {
        return nil, errorsmod.Wrapf(types.ErrGigNotFound, "gig %d not found", application.GigId)
    }
    
    // 3. Vérifier que le caller est le owner du gig
    if gig.Owner != msg.Creator {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only gig owner can reject applications")
    }
    
    // 4. Vérifier que l'application est pending
    if application.Status != "pending" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "can only reject pending applications (current status: %s)",
            application.Status,
        )
    }
    
    // 5. Mettre à jour le statut
    application.Status = "rejected"
    k.SetApplication(ctx, application)
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "application_rejected",
            sdk.NewAttribute("application_id", fmt.Sprintf("%d", msg.ApplicationId)),
            sdk.NewAttribute("gig_id", fmt.Sprintf("%d", application.GigId)),
        ),
    )
    
    return &types.MsgRejectApplicationResponse{}, nil
}
```

---

## 4.10 Ajouter des queries personnalisées

Ajoutons des queries utiles pour filtrer les applications et contracts.

**Scaffold les queries :**
```bash
# Applications par gig
ignite scaffold query applications-by-gig gig_id:uint --response applications:Application --module marketplace

# Applications par freelancer
ignite scaffold query applications-by-freelancer freelancer:string --response applications:Application --module marketplace

# Contracts par utilisateur (client ou freelancer)
ignite scaffold query contracts-by-user user:string --response contracts:Contract --module marketplace

# Contract par gig
ignite scaffold query contract-by-gig gig_id:uint --response contract:Contract --module marketplace
```

**Implémenter la query ApplicationsByGig :**

**x/marketplace/keeper/query_applications_by_gig.go :**
```go
package keeper

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    
    "skillchain/x/marketplace/types"
)

func (k Keeper) ApplicationsByGig(goCtx context.Context, req *types.QueryApplicationsByGigRequest) (*types.QueryApplicationsByGigResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }
    
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    var applications []types.Application
    allApplications := k.GetAllApplication(ctx)
    
    for _, app := range allApplications {
        if app.GigId == req.GigId {
            applications = append(applications, app)
        }
    }
    
    return &types.QueryApplicationsByGigResponse{
        Applications: applications,
    }, nil
}
```

**Implémenter ContractsByUser :**

**x/marketplace/keeper/query_contracts_by_user.go :**
```go
package keeper

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    
    "skillchain/x/marketplace/types"
)

func (k Keeper) ContractsByUser(goCtx context.Context, req *types.QueryContractsByUserRequest) (*types.QueryContractsByUserResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }
    
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    var contracts []types.Contract
    allContracts := k.GetAllContract(ctx)
    
    for _, contract := range allContracts {
        // Retourne les contracts où l'utilisateur est client OU freelancer
        if contract.Client == req.User || contract.Freelancer == req.User {
            contracts = append(contracts, contract)
        }
    }
    
    return &types.QueryContractsByUserResponse{
        Contracts: contracts,
    }, nil
}
```

---

## 4.11 Tests complets du workflow

```bash
# Régénérer et relancer
ignite generate proto-go
ignite chain serve --reset-once
```

**Dans un nouveau terminal :**

```bash
# === SETUP ===

# Alice crée son profil freelancer
skillchaind tx marketplace create-profile \
  "Alice Developer" \
  "Expert Cosmos SDK developer" \
  "golang,cosmos-sdk,react" \
  50000 \
  --from alice --yes

# Bob crée un gig (il est le client)
skillchaind tx marketplace create-gig \
  "Build SkillChain Frontend" \
  "Need a React/TypeScript frontend connected to our Cosmos blockchain" \
  500000 \
  "development" \
  21 \
  --from bob --yes

# Charlie crée aussi son profil freelancer
skillchaind tx marketplace create-profile \
  "Charlie Designer" \
  "UI/UX specialist" \
  "figma,react,css" \
  40000 \
  --from charlie --yes

# === APPLICATIONS ===

# Alice postule au gig de Bob
skillchaind tx marketplace apply-to-gig \
  0 \
  "I have 3 years of experience with Cosmos SDK and React. I can deliver this project in 14 days." \
  450000 \
  14 \
  --from alice --yes

# Charlie postule aussi
skillchaind tx marketplace apply-to-gig \
  0 \
  "I specialize in beautiful UIs. I can create an amazing frontend for your project." \
  480000 \
  18 \
  --from charlie --yes

# Vérifier les applications
skillchaind query marketplace list-application

# Query les applications pour le gig 0
skillchaind query marketplace applications-by-gig 0

# === ACCEPT/REJECT ===

# Bob accepte la candidature d'Alice (application 0)
skillchaind tx marketplace accept-application 0 --from bob --yes

# Vérifier que :
# - Application 0 est "accepted"
# - Application 1 (Charlie) est "rejected" automatiquement
# - Gig 0 est "in_progress"
# - Un contract a été créé

skillchaind query marketplace show-application 0
# status: accepted

skillchaind query marketplace show-application 1
# status: rejected

skillchaind query marketplace show-gig 0
# status: in_progress

skillchaind query marketplace list-contract
# Un contract avec client=bob, freelancer=alice

# Query contracts par utilisateur
skillchaind query marketplace contracts-by-user $(skillchaind keys show alice -a)
skillchaind query marketplace contracts-by-user $(skillchaind keys show bob -a)

# === TEST ERREURS ===

# Alice essaie de postuler à nouveau (erreur: déjà une application)
skillchaind tx marketplace apply-to-gig 0 "Test" 400000 10 --from alice --yes
# Erreur: gig is not open for applications

# Bob essaie de postuler à son propre gig
skillchaind tx marketplace create-gig "Test Gig" "Description" 100000 "test" 7 --from bob --yes
skillchaind tx marketplace apply-to-gig 1 "Test" 90000 5 --from bob --yes
# Erreur: cannot apply to your own gig
```

---

## 4.12 Schéma récapitulatif du workflow

```
┌──────────────────────────────────────────────────────────────┐
│                    WORKFLOW SKILLCHAIN                        │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  1. CLIENT crée Gig          2. FREELANCER crée Profile      │
│        │                            │                        │
│        ▼                            ▼                        │
│  ┌──────────┐                ┌──────────────┐               │
│  │ Gig:open │◄───────────────│   Profile    │               │
│  └────┬─────┘   postule      └──────────────┘               │
│       │                                                      │
│       ▼                                                      │
│  ┌────────────────┐                                         │
│  │  Application   │ pending                                 │
│  │  (freelancer)  │                                         │
│  └───────┬────────┘                                         │
│          │                                                   │
│    ┌─────┴─────┐                                            │
│    ▼           ▼                                            │
│ [accept]    [reject]                                        │
│    │           │                                            │
│    ▼           ▼                                            │
│ accepted    rejected                                        │
│    │                                                        │
│    ▼                                                        │
│ ┌──────────────┐     ┌──────────┐                          │
│ │   Contract   │────►│Gig:in_   │                          │
│ │   (active)   │     │progress  │                          │
│ └──────────────┘     └──────────┘                          │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## Questions de révision

1. **Pourquoi vérifie-t-on que le freelancer a un profil avant de le laisser postuler ?**

2. **Que se passe-t-il pour les autres candidatures pending quand une est acceptée ?**

3. **Comment calcule-t-on la deadline de livraison d'un contract ?**

4. **Pourquoi interdit-on à un utilisateur de postuler à son propre gig ?**

5. **Quelle entité est créée automatiquement lors de l'acceptation d'une application ?**

6. **Comment implémenter une query qui filtre par plusieurs critères (ex: gig_id ET status) ?**

---

## Récapitulatif des commandes

```bash
# Scaffold entities
ignite scaffold list application ... --module marketplace --no-message
ignite scaffold list contract ... --module marketplace --no-message

# Scaffold messages
ignite scaffold message apply-to-gig ... --module marketplace
ignite scaffold message accept-application ... --module marketplace

# Scaffold queries
ignite scaffold query applications-by-gig gig_id:uint --response applications:Application --module marketplace

# Test transactions
skillchaind tx marketplace apply-to-gig <gig_id> <cover_letter> <price> <days> --from <account>
skillchaind tx marketplace accept-application <app_id> --from <account>

# Test queries
skillchaind query marketplace applications-by-gig <gig_id>
skillchaind query marketplace contracts-by-user <address>
```

---

**Prochaine leçon** : Nous allons implémenter le système d'escrow pour sécuriser les paiements entre clients et freelancers.
