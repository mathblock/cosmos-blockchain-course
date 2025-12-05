# Leçon 6 : Système de Disputes et Arbitrage

## Objectifs
- Créer l'entité Dispute pour gérer les conflits
- Implémenter un système de vote pour l'arbitrage
- Utiliser EndBlock pour traiter les disputes expirées
- Gérer la redistribution des fonds en cas de litige

## Prérequis
- Leçon 5 complétée
- Système d'escrow fonctionnel

---

## 6.1 Pourquoi un système de dispute ?

Dans une marketplace, des conflits peuvent survenir :
- Le freelancer prétend avoir livré mais le client conteste la qualité
- Le client ne répond plus après la livraison
- Désaccord sur l'interprétation du cahier des charges

**Solution décentralisée :** Un système d'arbitrage où des arbitres votent pour déterminer le gagnant du litige.

```
┌────────────────────────────────────────────────────────────┐
│                    FLUX DE DISPUTE                         │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Contract (delivered)                                      │
│       │                                                    │
│       │ client ou freelancer ouvre dispute                 │
│       ▼                                                    │
│  ┌──────────────────┐                                     │
│  │     Dispute      │                                     │
│  │   status: open   │                                     │
│  └────────┬─────────┘                                     │
│           │                                                │
│           │ arbitres votent                                │
│           ▼                                                │
│  ┌──────────────────┐                                     │
│  │  Votes collectés │                                     │
│  │  (client/free)   │                                     │
│  └────────┬─────────┘                                     │
│           │                                                │
│     ┌─────┴─────┐                                         │
│     ▼           ▼                                         │
│ [client]    [freelancer]                                  │
│ gagne       gagne                                         │
│     │           │                                         │
│     ▼           ▼                                         │
│ Refund      Paiement                                      │
│ client      freelancer                                    │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## 6.2 Créer l'entité Dispute

```bash
ignite scaffold list dispute \
  contract_id:uint \
  initiator:string \
  reason:string \
  client_evidence:string \
  freelancer_evidence:string \
  status:string \
  votes_client:uint \
  votes_freelancer:uint \
  resolution:string \
  created_at:int \
  deadline:int \
  --module marketplace \
  --no-message
```

**Statuts d'une Dispute :**
- `open` : Dispute ouverte, en attente de votes
- `voting` : Phase de vote active
- `resolved_client` : Résolu en faveur du client
- `resolved_freelancer` : Résolu en faveur du freelancer
- `expired` : Délai dépassé sans résolution

---

## 6.3 Créer l'entité Vote

```bash
ignite scaffold map dispute-vote \
  dispute_id:uint \
  vote:string \
  voted_at:int \
  --index arbiter \
  --module marketplace \
  --no-message
```

Chaque arbitre ne peut voter qu'une fois par dispute.

---

## 6.4 Ajouter les paramètres de dispute

**Modifier proto/skillchain/marketplace/params.proto :**
```protobuf
message Params {
  option (amino.name) = "skillchain/x/marketplace/Params";
  option (gogoproto.equal) = true;
  
  uint64 platform_fee_percent = 1;
  uint64 min_contract_duration = 2;
  string min_gig_price = 3 [(cosmos_proto.scalar) = "cosmos.Int"];
  
  // Nouveaux paramètres pour les disputes
  uint64 dispute_duration = 4;        // Durée en secondes pour voter (ex: 7 jours)
  uint64 min_arbiters_required = 5;   // Nombre minimum d'arbitres pour résoudre
  uint64 arbiter_stake_required = 6;  // Stake minimum pour être arbitre (en uskill)
}
```

**Mettre à jour x/marketplace/types/params.go :**
```go
package types

import (
    "fmt"
    
    "cosmossdk.io/math"
)

var (
    DefaultPlatformFeePercent    = uint64(5)
    DefaultMinContractDuration   = uint64(86400)
    DefaultMinGigPrice           = math.NewInt(10000)
    DefaultDisputeDuration       = uint64(604800)     // 7 jours en secondes
    DefaultMinArbitersRequired   = uint64(3)          // Minimum 3 arbitres
    DefaultArbiterStakeRequired  = uint64(1000000)    // 1 SKILL minimum staked
)

func NewParams(
    feePercent, minDuration uint64,
    minPrice math.Int,
    disputeDuration, minArbiters, arbiterStake uint64,
) Params {
    return Params{
        PlatformFeePercent:    feePercent,
        MinContractDuration:   minDuration,
        MinGigPrice:           minPrice.String(),
        DisputeDuration:       disputeDuration,
        MinArbitersRequired:   minArbiters,
        ArbiterStakeRequired:  arbiterStake,
    }
}

func DefaultParams() Params {
    return NewParams(
        DefaultPlatformFeePercent,
        DefaultMinContractDuration,
        DefaultMinGigPrice,
        DefaultDisputeDuration,
        DefaultMinArbitersRequired,
        DefaultArbiterStakeRequired,
    )
}

func (p Params) Validate() error {
    if p.PlatformFeePercent > 100 {
        return fmt.Errorf("platform fee cannot exceed 100%%")
    }
    if p.MinArbitersRequired == 0 {
        return fmt.Errorf("min arbiters required must be at least 1")
    }
    if p.DisputeDuration < 86400 {
        return fmt.Errorf("dispute duration must be at least 1 day")
    }
    return nil
}
```

Régénérer :
```bash
ignite generate proto-go
```

---

## 6.5 Créer les messages pour les disputes

```bash
# Ouvrir une dispute
ignite scaffold message open-dispute \
  contract_id:uint \
  reason:string \
  evidence:string \
  --module marketplace

# Soumettre une preuve (pour la partie adverse)
ignite scaffold message submit-evidence \
  dispute_id:uint \
  evidence:string \
  --module marketplace

# Voter sur une dispute (arbitres)
ignite scaffold message vote-dispute \
  dispute_id:uint \
  vote:string \
  --module marketplace

# Résoudre manuellement (admin/governance)
ignite scaffold message resolve-dispute \
  dispute_id:uint \
  winner:string \
  --module marketplace
```

---

## 6.6 Implémenter OpenDispute

**x/marketplace/keeper/msg_server_open_dispute.go :**
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

func (k msgServer) OpenDispute(goCtx context.Context, msg *types.MsgOpenDispute) (*types.MsgOpenDisputeResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // 1. Récupérer le contrat
    contract, found := k.GetContract(ctx, msg.ContractId)
    if !found {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract %d not found", msg.ContractId)
    }
    
    // 2. Vérifier que le caller est partie au contrat
    isClient := contract.Client == msg.Creator
    isFreelancer := contract.Freelancer == msg.Creator
    
    if !isClient && !isFreelancer {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only client or freelancer can open dispute")
    }
    
    // 3. Vérifier le statut du contrat (doit être active ou delivered)
    if contract.Status != "active" && contract.Status != "delivered" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "cannot dispute contract with status %s",
            contract.Status,
        )
    }
    
    // 4. Vérifier qu'il n'y a pas déjà une dispute ouverte
    allDisputes := k.GetAllDispute(ctx)
    for _, d := range allDisputes {
        if d.ContractId == msg.ContractId && (d.Status == "open" || d.Status == "voting") {
            return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "dispute already exists for this contract")
        }
    }
    
    // 5. Calculer la deadline
    params := k.GetParams(ctx)
    deadline := ctx.BlockTime().Unix() + int64(params.DisputeDuration)
    
    // 6. Créer la dispute
    dispute := types.Dispute{
        ContractId:         msg.ContractId,
        Initiator:          msg.Creator,
        Reason:             msg.Reason,
        Status:             "open",
        VotesClient:        0,
        VotesFreelancer:    0,
        Resolution:         "",
        CreatedAt:          ctx.BlockTime().Unix(),
        Deadline:           deadline,
    }
    
    // Attribuer l'evidence au bon champ
    if isClient {
        dispute.ClientEvidence = msg.Evidence
    } else {
        dispute.FreelancerEvidence = msg.Evidence
    }
    
    disputeId := k.AppendDispute(ctx, dispute)
    
    // 7. Mettre à jour le statut du contrat
    contract.Status = "disputed"
    k.SetContract(ctx, contract)
    
    // 8. Événement
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "dispute_opened",
            sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", disputeId)),
            sdk.NewAttribute("contract_id", fmt.Sprintf("%d", msg.ContractId)),
            sdk.NewAttribute("initiator", msg.Creator),
            sdk.NewAttribute("deadline", fmt.Sprintf("%d", deadline)),
        ),
    )
    
    return &types.MsgOpenDisputeResponse{
        DisputeId: disputeId,
    }, nil
}
```

**Modifier le response :**
```protobuf
message MsgOpenDisputeResponse {
  uint64 dispute_id = 1;
}
```

---

## 6.7 Implémenter SubmitEvidence

**x/marketplace/keeper/msg_server_submit_evidence.go :**
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

func (k msgServer) SubmitEvidence(goCtx context.Context, msg *types.MsgSubmitEvidence) (*types.MsgSubmitEvidenceResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // 1. Récupérer la dispute
    dispute, found := k.GetDispute(ctx, msg.DisputeId)
    if !found {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "dispute %d not found", msg.DisputeId)
    }
    
    // 2. Vérifier le statut
    if dispute.Status != "open" && dispute.Status != "voting" {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "dispute is not open for evidence")
    }
    
    // 3. Vérifier le deadline
    if ctx.BlockTime().Unix() > dispute.Deadline {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "dispute deadline has passed")
    }
    
    // 4. Récupérer le contrat pour identifier le rôle
    contract, _ := k.GetContract(ctx, dispute.ContractId)
    
    isClient := contract.Client == msg.Creator
    isFreelancer := contract.Freelancer == msg.Creator
    
    if !isClient && !isFreelancer {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "only parties can submit evidence")
    }
    
    // 5. Ajouter l'evidence
    if isClient {
        if dispute.ClientEvidence != "" {
            return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "client evidence already submitted")
        }
        dispute.ClientEvidence = msg.Evidence
    } else {
        if dispute.FreelancerEvidence != "" {
            return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "freelancer evidence already submitted")
        }
        dispute.FreelancerEvidence = msg.Evidence
    }
    
    // 6. Si les deux parties ont soumis, passer en voting
    if dispute.ClientEvidence != "" && dispute.FreelancerEvidence != "" {
        dispute.Status = "voting"
    }
    
    k.SetDispute(ctx, dispute)
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "evidence_submitted",
            sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", msg.DisputeId)),
            sdk.NewAttribute("submitter", msg.Creator),
        ),
    )
    
    return &types.MsgSubmitEvidenceResponse{}, nil
}
```

---

## 6.8 Implémenter VoteDispute

**x/marketplace/keeper/msg_server_vote_dispute.go :**
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

func (k msgServer) VoteDispute(goCtx context.Context, msg *types.MsgVoteDispute) (*types.MsgVoteDisputeResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    
    // 1. Récupérer la dispute
    dispute, found := k.GetDispute(ctx, msg.DisputeId)
    if !found {
        return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "dispute %d not found", msg.DisputeId)
    }
    
    // 2. Vérifier le statut (doit être open ou voting)
    if dispute.Status != "open" && dispute.Status != "voting" {
        return nil, errorsmod.Wrapf(
            sdkerrors.ErrInvalidRequest,
            "dispute is not open for voting (status: %s)",
            dispute.Status,
        )
    }
    
    // 3. Vérifier le deadline
    if ctx.BlockTime().Unix() > dispute.Deadline {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "voting deadline has passed")
    }
    
    // 4. Vérifier que le voter n'est pas partie au contrat
    contract, _ := k.GetContract(ctx, dispute.ContractId)
    if contract.Client == msg.Creator || contract.Freelancer == msg.Creator {
        return nil, errorsmod.Wrap(types.ErrUnauthorized, "parties cannot vote on their own dispute")
    }
    
    // 5. Vérifier que le voter a le stake requis (simplifié: vérifie le solde)
    params := k.GetParams(ctx)
    voterAddr, _ := sdk.AccAddressFromBech32(msg.Creator)
    balance := k.bankKeeper.GetBalance(ctx, voterAddr, "stake")
    
    if balance.Amount.LT(sdk.NewIntFromUint64(params.ArbiterStakeRequired)) {
        return nil, errorsmod.Wrapf(
            types.ErrInsufficientFunds,
            "arbiter must have at least %d stake (has %s)",
            params.ArbiterStakeRequired,
            balance.String(),
        )
    }
    
    // 6. Vérifier qu'il n'a pas déjà voté
    voteKey := fmt.Sprintf("%d-%s", msg.DisputeId, msg.Creator)
    _, alreadyVoted := k.GetDisputeVote(ctx, msg.Creator)
    
    // Vérification plus précise: chercher un vote pour cette dispute
    allVotes := k.GetAllDisputeVote(ctx)
    for _, v := range allVotes {
        if v.DisputeId == msg.DisputeId && v.Arbiter == msg.Creator {
            return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "already voted on this dispute")
        }
    }
    
    // 7. Valider le vote
    if msg.Vote != "client" && msg.Vote != "freelancer" {
        return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "vote must be 'client' or 'freelancer'")
    }
    
    // 8. Enregistrer le vote
    vote := types.DisputeVote{
        Arbiter:   msg.Creator,
        DisputeId: msg.DisputeId,
        Vote:      msg.Vote,
        VotedAt:   ctx.BlockTime().Unix(),
    }
    k.SetDisputeVote(ctx, vote)
    
    // 9. Mettre à jour les compteurs
    if msg.Vote == "client" {
        dispute.VotesClient++
    } else {
        dispute.VotesFreelancer++
    }
    
    // Passer en voting si pas déjà fait
    if dispute.Status == "open" {
        dispute.Status = "voting"
    }
    
    k.SetDispute(ctx, dispute)
    
    // 10. Vérifier si on peut résoudre
    totalVotes := dispute.VotesClient + dispute.VotesFreelancer
    if totalVotes >= params.MinArbitersRequired {
        // Résoudre automatiquement
        k.resolveDispute(ctx, &dispute)
    }
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "dispute_vote_cast",
            sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", msg.DisputeId)),
            sdk.NewAttribute("arbiter", msg.Creator),
            sdk.NewAttribute("vote", msg.Vote),
            sdk.NewAttribute("total_votes", fmt.Sprintf("%d", totalVotes)),
        ),
    )
    
    return &types.MsgVoteDisputeResponse{}, nil
}
```

---

## 6.9 Fonction de résolution

**Ajouter dans x/marketplace/keeper/dispute_resolution.go :**
```go
package keeper

import (
    "fmt"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "skillchain/x/marketplace/types"
)

// resolveDispute résout une dispute et redistribue les fonds
func (k Keeper) resolveDispute(ctx sdk.Context, dispute *types.Dispute) error {
    contract, found := k.GetContract(ctx, dispute.ContractId)
    if !found {
        return types.ErrContractNotFound
    }
    
    // Déterminer le gagnant
    var winner string
    var winnerAddr sdk.AccAddress
    var err error
    
    if dispute.VotesClient > dispute.VotesFreelancer {
        winner = "client"
        dispute.Status = "resolved_client"
        dispute.Resolution = "Client wins by majority vote"
        winnerAddr, err = sdk.AccAddressFromBech32(contract.Client)
    } else if dispute.VotesFreelancer > dispute.VotesClient {
        winner = "freelancer"
        dispute.Status = "resolved_freelancer"
        dispute.Resolution = "Freelancer wins by majority vote"
        winnerAddr, err = sdk.AccAddressFromBech32(contract.Freelancer)
    } else {
        // Égalité - favorise le freelancer (travail fourni)
        winner = "freelancer"
        dispute.Status = "resolved_freelancer"
        dispute.Resolution = "Tie resolved in favor of freelancer"
        winnerAddr, err = sdk.AccAddressFromBech32(contract.Freelancer)
    }
    
    if err != nil {
        return err
    }
    
    // Transférer les fonds
    escrowAmount := sdk.NewCoins(sdk.NewCoin("uskill", sdk.NewIntFromUint64(contract.Price)))
    
    err = k.bankKeeper.SendCoinsFromModuleToAccount(
        ctx,
        types.ModuleName,
        winnerAddr,
        escrowAmount,
    )
    if err != nil {
        return err
    }
    
    // Mettre à jour les entités
    k.SetDispute(ctx, *dispute)
    
    contract.Status = "resolved_" + winner
    contract.CompletedAt = ctx.BlockTime().Unix()
    k.SetContract(ctx, contract)
    
    gig, _ := k.GetGig(ctx, contract.GigId)
    gig.Status = "closed"
    k.SetGig(ctx, gig)
    
    // Si le freelancer gagne, mettre à jour son profil
    if winner == "freelancer" {
        profile, found := k.GetProfile(ctx, contract.Freelancer)
        if found {
            profile.TotalJobs++
            profile.TotalEarned += contract.Price
            k.SetProfile(ctx, profile)
        }
    }
    
    // Événement
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "dispute_resolved",
            sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", dispute.Id)),
            sdk.NewAttribute("contract_id", fmt.Sprintf("%d", contract.Id)),
            sdk.NewAttribute("winner", winner),
            sdk.NewAttribute("amount", escrowAmount.String()),
        ),
    )
    
    return nil
}
```

---

## 6.10 EndBlock pour les disputes expirées

Traiter automatiquement les disputes qui dépassent leur deadline.

**x/marketplace/module/abci.go :**
```go
package marketplace

import (
    "fmt"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "skillchain/x/marketplace/keeper"
)

// EndBlock is called at the end of every block
func (am AppModule) EndBlock(ctx sdk.Context) error {
    return am.keeper.ProcessExpiredDisputes(ctx)
}
```

**x/marketplace/keeper/dispute_expiry.go :**
```go
package keeper

import (
    "fmt"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "skillchain/x/marketplace/types"
)

// ProcessExpiredDisputes traite les disputes qui ont dépassé leur deadline
func (k Keeper) ProcessExpiredDisputes(ctx sdk.Context) error {
    currentTime := ctx.BlockTime().Unix()
    allDisputes := k.GetAllDispute(ctx)
    
    for _, dispute := range allDisputes {
        // Ne traiter que les disputes actives et expirées
        if (dispute.Status != "open" && dispute.Status != "voting") {
            continue
        }
        
        if currentTime <= dispute.Deadline {
            continue
        }
        
        // La dispute est expirée
        params := k.GetParams(ctx)
        totalVotes := dispute.VotesClient + dispute.VotesFreelancer
        
        if totalVotes >= params.MinArbitersRequired {
            // Assez de votes: résoudre normalement
            err := k.resolveDispute(ctx, &dispute)
            if err != nil {
                k.Logger(ctx).Error("failed to resolve dispute", "dispute_id", dispute.Id, "error", err)
                continue
            }
        } else {
            // Pas assez de votes: résoudre en faveur du freelancer par défaut
            // (car le travail a été fourni)
            dispute.VotesFreelancer = 1 // Force le résultat
            dispute.Resolution = fmt.Sprintf(
                "Expired with insufficient votes (%d/%d required). Defaulting to freelancer.",
                totalVotes,
                params.MinArbitersRequired,
            )
            
            err := k.resolveDispute(ctx, &dispute)
            if err != nil {
                k.Logger(ctx).Error("failed to resolve expired dispute", "dispute_id", dispute.Id, "error", err)
                continue
            }
            
            ctx.EventManager().EmitEvent(
                sdk.NewEvent(
                    "dispute_expired",
                    sdk.NewAttribute("dispute_id", fmt.Sprintf("%d", dispute.Id)),
                    sdk.NewAttribute("total_votes", fmt.Sprintf("%d", totalVotes)),
                    sdk.NewAttribute("required_votes", fmt.Sprintf("%d", params.MinArbitersRequired)),
                ),
            )
        }
    }
    
    return nil
}
```

---

## 6.11 Ajouter le EndBlock au module

**Modifier x/marketplace/module/module.go** pour inclure EndBlock :
```go
// Dans le fichier module.go, s'assurer que AppModule implémente l'interface

// EndBlock executes all ABCI EndBlock logic respective to the marketplace module
func (am AppModule) EndBlock(ctx context.Context) error {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    return am.keeper.ProcessExpiredDisputes(sdkCtx)
}
```

---

## 6.12 Tests du système de dispute

```bash
# Relancer la chaîne
ignite chain serve --reset-once
```

**Dans un nouveau terminal :**

```bash
# === SETUP COMPLET ===

# Alice = freelancer
skillchaind tx marketplace create-profile \
  "Alice Dev" "Developer" "golang" 50000 \
  --from alice --yes

# Bob = client, crée un gig
skillchaind tx marketplace create-gig \
  "Build API" "Need a REST API" 200000 "dev" 7 \
  --from bob --yes

# Alice postule
skillchaind tx marketplace apply-to-gig 0 "I can do it" 200000 7 --from alice --yes

# Bob accepte (fonds verrouillés)
skillchaind tx marketplace accept-application 0 --from bob --yes

# Alice livre
skillchaind tx marketplace deliver-contract 0 "Done" --from alice --yes

# === OUVRIR UNE DISPUTE ===

# Bob n'est pas satisfait et ouvre une dispute
skillchaind tx marketplace open-dispute \
  0 \
  "Work does not meet requirements" \
  "Screenshots showing missing features" \
  --from bob --yes

# Vérifier la dispute
skillchaind query marketplace list-dispute
skillchaind query marketplace show-dispute 0
# status: open
# client_evidence: "Screenshots..."

# Alice soumet sa preuve
skillchaind tx marketplace submit-evidence \
  0 \
  "Git commits showing all requirements implemented" \
  --from alice --yes

# Vérifier que le status est passé à "voting"
skillchaind query marketplace show-dispute 0
# status: voting

# === VOTES ===

# Charlie vote pour le freelancer
skillchaind tx marketplace vote-dispute 0 freelancer --from charlie --yes

# Vérifier les votes (il faut 3 arbitres par défaut)
skillchaind query marketplace show-dispute 0
# votes_freelancer: 1

# On aurait besoin de 2 autres arbitres pour résoudre automatiquement
# Pour les tests, on peut modifier les paramètres ou utiliser resolve-dispute

# === RÉSOLUTION MANUELLE (si admin/governance) ===
# Note: Dans une vraie implémentation, ceci nécessiterait une proposition de gouvernance

# Vérifier le solde de l'escrow avant résolution
skillchaind query marketplace escrow-balance
# 200000uskill

# Après que suffisamment d'arbitres aient voté, la dispute est résolue
# Les fonds sont transférés au gagnant
```

---

## 6.13 Test du EndBlock (disputes expirées)

Pour tester le EndBlock, on peut réduire temporairement la durée des disputes.

**Modifier la config pour des tests rapides :**
```yaml
# Dans genesis de config.yml, ajouter:
genesis:
  app_state:
    marketplace:
      params:
        dispute_duration: "60"  # 60 secondes pour les tests
        min_arbiters_required: "1"
```

```bash
# Relancer avec les nouveaux paramètres
ignite chain serve --reset-once

# Répéter le setup et ouvrir une dispute
# Attendre 60+ secondes sans voter
# La dispute sera résolue automatiquement au prochain bloc
# en faveur du freelancer (défaut quand pas assez de votes)
```

---

## 6.14 Diagramme récapitulatif

```
┌────────────────────────────────────────────────────────────────────┐
│                    CYCLE DE VIE D'UNE DISPUTE                       │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  Contract (active/delivered)                                       │
│         │                                                          │
│         │ open_dispute                                             │
│         ▼                                                          │
│  ┌─────────────────┐                                              │
│  │  Dispute: open  │◄──── initiator soumet evidence               │
│  └────────┬────────┘                                              │
│           │                                                        │
│           │ submit_evidence (autre partie)                         │
│           ▼                                                        │
│  ┌─────────────────┐                                              │
│  │ Dispute: voting │◄──── arbitres votent                         │
│  └────────┬────────┘                                              │
│           │                                                        │
│     ┌─────┴─────────────────────┐                                 │
│     │                           │                                 │
│     ▼                           ▼                                 │
│  [3+ votes]                 [deadline]                            │
│     │                           │                                 │
│     ▼                           ▼                                 │
│  Résolution               EndBlock traite                         │
│  par majorité             votes insuffisants                      │
│     │                           │                                 │
│     └───────────┬───────────────┘                                 │
│                 ▼                                                  │
│        ┌───────────────┐                                          │
│        │ FONDS LIBÉRÉS │                                          │
│        │  au gagnant   │                                          │
│        └───────────────┘                                          │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## Questions de révision

1. **Pourquoi les parties au contrat ne peuvent-elles pas voter sur leur propre dispute ?**

2. **Que se passe-t-il si une dispute expire sans suffisamment de votes ?**

3. **Quel est le rôle du EndBlock dans le module marketplace ?**

4. **Comment s'assure-t-on qu'un arbitre a le droit de voter (stake minimum) ?**

5. **Pourquoi y a-t-il deux statuts distincts "open" et "voting" pour une dispute ?**

6. **En cas d'égalité des votes, qui gagne et pourquoi ?**

---

## Récapitulatif des commandes

```bash
# Gestion des disputes
skillchaind tx marketplace open-dispute <contract_id> <reason> <evidence> --from <party>
skillchaind tx marketplace submit-evidence <dispute_id> <evidence> --from <party>
skillchaind tx marketplace vote-dispute <dispute_id> <client|freelancer> --from <arbiter>

# Queries
skillchaind query marketplace list-dispute
skillchaind query marketplace show-dispute <id>
skillchaind query marketplace list-dispute-vote
```

---

**Prochaine leçon** : Nous allons créer le frontend React/TypeScript et commencer l'intégration avec CosmJS.
