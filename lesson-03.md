# Leçon 3 : CRUD des entités Profile et Gig

## Objectifs
- Scaffolder les entités Profile (freelancer) et Gig (mission)
- Comprendre la différence entre `list`, `map` et `single`
- Personnaliser la logique métier dans les handlers
- Tester les opérations CRUD via CLI

## Prérequis
- Leçon 2 complétée
- Projet SkillChain avec module marketplace créé

---

## 3.1 Les types de scaffold pour les données

Ignite propose trois types de scaffold pour stocker des données :

| Type | Clé | Usage | Exemple SkillChain |
|------|-----|-------|-------------------|
| `list` | Auto-incrémentée (uint64) | Collections numérotées | Gigs, Applications |
| `map` | Définie par l'utilisateur | Lookup par clé unique | Profile par adresse |
| `single` | Unique (singleton) | Configuration globale | Statistiques plateforme |

---

## 3.2 Créer l'entité Profile (map)

Un Profile est lié à une adresse unique. Utilisons `map` avec l'adresse comme clé.

```bash
cd skillchain

ignite scaffold map profile \
  name:string \
  bio:string \
  skills:strings \
  hourly_rate:uint \
  total_jobs:uint \
  total_earned:uint \
  rating_sum:uint \
  rating_count:uint \
  --index owner \
  --module marketplace \
  --no-message
```

**Explications des options :**
- `name:string` : Nom du freelancer
- `skills:strings` : Liste de compétences (array de strings)
- `hourly_rate:uint` : Taux horaire en uskill
- `--index owner` : La clé de map sera l'adresse du propriétaire
- `--module marketplace` : Ajoute au module marketplace
- `--no-message` : Ne génère pas les messages CRUD (on les fera custom)

**Fichiers générés :**
```
proto/skillchain/marketplace/profile.proto    # Type Protobuf
x/marketplace/keeper/profile.go               # Méthodes CRUD du keeper
x/marketplace/keeper/query_profile.go         # Queries
x/marketplace/types/key_profile.go            # Clés de stockage
```

---

## 3.3 Créer l'entité Gig (list)

Un Gig (mission) a un ID auto-incrémenté.

```bash
ignite scaffold list gig \
  title:string \
  description:string \
  owner:string \
  price:uint \
  category:string \
  delivery_days:uint \
  status:string \
  created_at:int \
  --module marketplace \
  --no-message
```

**Champs :**
- `owner` : Adresse du client qui crée la mission
- `price` : Prix en uskill
- `status` : "open", "in_progress", "completed", "cancelled"
- `created_at` : Timestamp Unix de création

---

## 3.4 Examiner les types générés

**proto/skillchain/marketplace/profile.proto :**
```protobuf
syntax = "proto3";
package skillchain.marketplace;

option go_package = "skillchain/x/marketplace/types";

message Profile {
  string owner = 1;           // Clé d'index (adresse)
  string name = 2;
  string bio = 3;
  repeated string skills = 4; // Array de strings
  uint64 hourly_rate = 5;
  uint64 total_jobs = 6;
  uint64 total_earned = 7;
  uint64 rating_sum = 8;
  uint64 rating_count = 9;
}
```

**proto/skillchain/marketplace/gig.proto :**
```protobuf
syntax = "proto3";
package skillchain.marketplace;

option go_package = "skillchain/x/marketplace/types";

message Gig {
  uint64 id = 1;              // Auto-incrémenté
  string title = 2;
  string description = 3;
  string owner = 4;
  uint64 price = 5;
  string category = 6;
  uint64 delivery_days = 7;
  string status = 8;
  int64 created_at = 9;
}
```

---

## 3.5 Créer les messages personnalisés

Nous allons créer des messages custom pour contrôler la logique métier.

### Message CreateProfile

```bash
ignite scaffold message create-profile \
  name:string \
  bio:string \
  skills:strings \
  hourly_rate:uint \
  --module marketplace
```

### Message UpdateProfile

```bash
ignite scaffold message update-profile \
  name:string \
  bio:string \
  skills:strings \
  hourly_rate:uint \
  --module marketplace
```

### Message CreateGig

```bash
ignite scaffold message create-gig \
  title:string \
  description:string \
  price:uint \
  category:string \
  delivery_days:uint \
  --module marketplace
```

### Message UpdateGigStatus

```bash
ignite scaffold message update-gig-status \
  gig_id:uint \
  status:string \
  --module marketplace
```

---

## 3.6 Implémenter la logique CreateProfile

Modifions le handler pour ajouter notre logique métier.

**x/marketplace/keeper/msg_server_create_profile.go :**
```go
package keeper

import (
    "context"
    
    errorsmod "cosmossdk.io/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "skillchain/x/marketplace/types"
)

func (k msgServer) CreateProfile(goCtx context.Context, msg *types.MsgCreateProfile) (*types.MsgCreateProfileResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // Vérifier que le profil n'existe pas déjà
    _, found := k.GetProfile(ctx, msg.Creator)
    if found {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "profile already exists for this address")
    }
    
    // Valider le taux horaire (minimum 1000 uskill = 0.001 SKILL)
    if msg.HourlyRate < 1000 {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "hourly rate must be at least 1000 uskill")
    }
    
    // Valider les skills (au moins une compétence requise)
    if len(msg.Skills) == 0 {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one skill is required")
    }
    
    // Créer le profil
    profile := types.Profile{
        Owner:       msg.Creator,
        Name:        msg.Name,
        Bio:         msg.Bio,
        Skills:      msg.Skills,
        HourlyRate:  msg.HourlyRate,
        TotalJobs:   0,
        TotalEarned: 0,
        RatingSum:   0,
        RatingCount: 0,
    }
    
    // Sauvegarder dans le state
    k.SetProfile(ctx, profile)
    
    // Émettre un événement
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "profile_created",
            sdk.NewAttribute("owner", msg.Creator),
            sdk.NewAttribute("name", msg.Name),
        ),
    )
    
    return &types.MsgCreateProfileResponse{}, nil
}
```

---

## 3.7 Implémenter la logique UpdateProfile

**x/marketplace/keeper/msg_server_update_profile.go :**
```go
package keeper

import (
    "context"
    
    errorsmod "cosmossdk.io/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "skillchain/x/marketplace/types"
)

func (k msgServer) UpdateProfile(goCtx context.Context, msg *types.MsgUpdateProfile) (*types.MsgUpdateProfileResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // Récupérer le profil existant
    profile, found := k.GetProfile(ctx, msg.Creator)
    if !found {
        return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "profile not found")
    }
    
    // Seul le propriétaire peut modifier son profil
    // (déjà garanti car Creator = signer du message)
    
    // Valider le taux horaire
    if msg.HourlyRate < 1000 {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "hourly rate must be at least 1000 uskill")
    }
    
    // Mettre à jour les champs modifiables
    profile.Name = msg.Name
    profile.Bio = msg.Bio
    profile.Skills = msg.Skills
    profile.HourlyRate = msg.HourlyRate
    // Note: TotalJobs, TotalEarned, Rating ne sont pas modifiables par l'utilisateur
    
    k.SetProfile(ctx, profile)
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "profile_updated",
            sdk.NewAttribute("owner", msg.Creator),
        ),
    )
    
    return &types.MsgUpdateProfileResponse{}, nil
}
```

---

## 3.8 Implémenter la logique CreateGig

**x/marketplace/keeper/msg_server_create_gig.go :**
```go
package keeper

import (
    "context"
    "time"
    
    errorsmod "cosmossdk.io/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "skillchain/x/marketplace/types"
)

func (k msgServer) CreateGig(goCtx context.Context, msg *types.MsgCreateGig) (*types.MsgCreateGigResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // Récupérer les paramètres du module
    params := k.GetParams(ctx)
    
    // Valider le prix minimum
    minPrice, _ := sdk.NewIntFromString(params.MinGigPrice)
    if sdk.NewIntFromUint64(msg.Price).LT(minPrice) {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "price must be at least %s uskill",
            params.MinGigPrice,
        )
    }
    
    // Valider la durée de livraison
    if msg.DeliveryDays == 0 || msg.DeliveryDays > 365 {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "delivery days must be between 1 and 365")
    }
    
    // Valider le titre
    if len(msg.Title) < 10 || len(msg.Title) > 100 {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "title must be between 10 and 100 characters")
    }
    
    // Créer le gig avec un nouvel ID
    gig := types.Gig{
        Title:        msg.Title,
        Description:  msg.Description,
        Owner:        msg.Creator,
        Price:        msg.Price,
        Category:     msg.Category,
        DeliveryDays: msg.DeliveryDays,
        Status:       "open",
        CreatedAt:    ctx.BlockTime().Unix(),
    }
    
    // AppendGig génère automatiquement l'ID et sauvegarde
    id := k.AppendGig(ctx, gig)
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "gig_created",
            sdk.NewAttribute("id", fmt.Sprintf("%d", id)),
            sdk.NewAttribute("owner", msg.Creator),
            sdk.NewAttribute("title", msg.Title),
            sdk.NewAttribute("price", fmt.Sprintf("%d", msg.Price)),
        ),
    )
    
    return &types.MsgCreateGigResponse{
        Id: id,
    }, nil
}
```

**Modifier le response pour inclure l'ID :**

**proto/skillchain/marketplace/tx.proto** (ajouter dans MsgCreateGigResponse) :
```protobuf
message MsgCreateGigResponse {
  uint64 id = 1;  // ID du gig créé
}
```

Puis régénérer :
```bash
ignite generate proto-go
```

---

## 3.9 Implémenter UpdateGigStatus

**x/marketplace/keeper/msg_server_update_gig_status.go :**
```go
package keeper

import (
    "context"
    
    errorsmod "cosmossdk.io/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "skillchain/x/marketplace/types"
)

// Statuts valides et transitions autorisées
var validStatusTransitions = map[string][]string{
    "open":        {"cancelled", "in_progress"},
    "in_progress": {"completed", "disputed"},
    "disputed":    {"completed", "cancelled"},
}

func (k msgServer) UpdateGigStatus(goCtx context.Context, msg *types.MsgUpdateGigStatus) (*types.MsgUpdateGigStatusResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // Récupérer le gig
    gig, found := k.GetGig(ctx, msg.GigId)
    if !found {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "gig %d not found", msg.GigId)
    }
    
    // Vérifier que le caller est le owner du gig
    if gig.Owner != msg.Creator {
        return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only gig owner can update status")
    }
    
    // Vérifier que la transition est valide
    allowedTransitions, exists := validStatusTransitions[gig.Status]
    if !exists {
        return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "gig status %s cannot be changed", gig.Status)
    }
    
    isValidTransition := false
    for _, allowed := range allowedTransitions {
        if allowed == msg.Status {
            isValidTransition = true
            break
        }
    }
    
    if !isValidTransition {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "cannot transition from %s to %s",
            gig.Status,
            msg.Status,
        )
    }
    
    // Mettre à jour le statut
    gig.Status = msg.Status
    k.SetGig(ctx, gig)
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "gig_status_updated",
            sdk.NewAttribute("id", fmt.Sprintf("%d", msg.GigId)),
            sdk.NewAttribute("old_status", gig.Status),
            sdk.NewAttribute("new_status", msg.Status),
        ),
    )
    
    return &types.MsgUpdateGigStatusResponse{}, nil
}
```

---

## 3.10 Ajouter les erreurs custom

**x/marketplace/types/errors.go :**
```go
package types

import (
    errorsmod "cosmossdk.io/errors"
)

var (
    ErrProfileNotFound    = errorsmod.Register(ModuleName, 1100, "profile not found")
    ErrProfileExists      = errorsmod.Register(ModuleName, 1101, "profile already exists")
    ErrGigNotFound        = errorsmod.Register(ModuleName, 1200, "gig not found")
    ErrInvalidGigStatus   = errorsmod.Register(ModuleName, 1201, "invalid gig status transition")
    ErrUnauthorized       = errorsmod.Register(ModuleName, 1300, "unauthorized")
    ErrInsufficientFunds  = errorsmod.Register(ModuleName, 1400, "insufficient funds")
    ErrInvalidPrice       = errorsmod.Register(ModuleName, 1401, "invalid price")
)
```

---

## 3.11 Tester les opérations CRUD

```bash
# Relancer la chaîne
ignite chain serve --reset-once
```

**Nouveau terminal - Tests :**

```bash
# === PROFILS ===

# Créer un profil pour Alice (freelancer)
skillchaind tx marketplace create-profile \
  "Alice Dev" \
  "Full-stack developer with 5 years experience" \
  "golang,react,cosmos-sdk" \
  50000 \
  --from alice \
  --yes

# Query le profil d'Alice
skillchaind query marketplace show-profile $(skillchaind keys show alice -a)

# Output attendu:
# profile:
#   bio: Full-stack developer with 5 years experience
#   hourly_rate: "50000"
#   name: Alice Dev
#   owner: skill1...
#   rating_count: "0"
#   rating_sum: "0"
#   skills:
#   - golang
#   - react
#   - cosmos-sdk
#   total_earned: "0"
#   total_jobs: "0"

# Lister tous les profils
skillchaind query marketplace list-profile

# Mettre à jour le profil
skillchaind tx marketplace update-profile \
  "Alice Developer" \
  "Senior full-stack developer" \
  "golang,react,cosmos-sdk,rust" \
  75000 \
  --from alice \
  --yes

# Vérifier la mise à jour
skillchaind query marketplace show-profile $(skillchaind keys show alice -a)

# === GIGS ===

# Créer un gig (Bob est le client)
skillchaind tx marketplace create-gig \
  "Build a DeFi Dashboard" \
  "Need a React dashboard to display DeFi metrics from Cosmos chains" \
  500000 \
  "development" \
  14 \
  --from bob \
  --yes

# Query le gig créé
skillchaind query marketplace show-gig 0

# Output attendu:
# gig:
#   category: development
#   created_at: "1234567890"
#   delivery_days: "14"
#   description: Need a React dashboard...
#   id: "0"
#   owner: skill1... (Bob's address)
#   price: "500000"
#   status: open
#   title: Build a DeFi Dashboard

# Lister tous les gigs
skillchaind query marketplace list-gig

# Créer un second gig
skillchaind tx marketplace create-gig \
  "Smart Contract Audit" \
  "Security audit for a CosmWasm contract" \
  1000000 \
  "security" \
  7 \
  --from charlie \
  --yes

# Lister avec pagination
skillchaind query marketplace list-gig --limit 1 --offset 0
skillchaind query marketplace list-gig --limit 1 --offset 1

# Mettre à jour le statut d'un gig
skillchaind tx marketplace update-gig-status 0 cancelled --from bob --yes

# Vérifier le nouveau statut
skillchaind query marketplace show-gig 0
# status: cancelled
```

---

## 3.12 Test des validations

```bash
# Test: Créer un profil avec taux horaire trop bas
skillchaind tx marketplace create-profile \
  "Test" "Bio" "skill1" 100 \
  --from charlie --yes
# Erreur attendue: hourly rate must be at least 1000 uskill

# Test: Créer un gig avec prix trop bas
skillchaind tx marketplace create-gig \
  "Short title" "Desc" 100 "cat" 7 \
  --from bob --yes
# Erreur attendue: price must be at least 10000 uskill

# Test: Créer un profil qui existe déjà
skillchaind tx marketplace create-profile \
  "Alice 2" "Bio" "skill" 50000 \
  --from alice --yes
# Erreur attendue: profile already exists for this address

# Test: Transition de statut invalide
# D'abord recréer un gig ouvert
skillchaind tx marketplace create-gig \
  "Another Gig Task" "Description here" 50000 "dev" 5 \
  --from bob --yes

# Essayer de passer directement à "completed" (invalide depuis "open")
skillchaind tx marketplace update-gig-status 2 completed --from bob --yes
# Erreur attendue: cannot transition from open to completed
```

---

## Questions de révision

1. **Quelle est la différence entre `scaffold list` et `scaffold map` ?**

2. **Pourquoi utilise-t-on `--no-message` lors du scaffold des entités Profile et Gig ?**

3. **Comment s'assure-t-on qu'un utilisateur ne peut modifier que son propre profil ?**

4. **Quelle méthode du keeper génère automatiquement un ID pour une entité list ?**

5. **Comment émet-on un événement dans un handler de message ?**

6. **Où définit-on les codes d'erreur personnalisés du module ?**

---

## Récapitulatif des commandes

```bash
# Scaffold map (clé custom)
ignite scaffold map profile ... --index owner --module marketplace

# Scaffold list (ID auto)
ignite scaffold list gig ... --module marketplace

# Scaffold message
ignite scaffold message create-profile ... --module marketplace

# Régénérer Protobuf
ignite generate proto-go

# Tester les transactions
skillchaind tx marketplace create-profile ...
skillchaind tx marketplace create-gig ...

# Queries
skillchaind query marketplace show-profile <address>
skillchaind query marketplace list-gig
```

---

**Prochaine leçon** : Nous allons créer les entités Application et Contract pour gérer les candidatures aux missions et les contrats entre clients et freelancers.
