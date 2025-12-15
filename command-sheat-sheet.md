## Commands sheat sheet


```bash
# regenerate go code atfter changes
ignite generate proto-go

# Check tx details
skillchaind query tx <tx_hash>

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

# Gestion des disputes
skillchaind tx marketplace open-dispute <contract_id> <reason> <evidence> --from <party>
skillchaind tx marketplace submit-evidence <dispute_id> <evidence> --from <party>
skillchaind tx marketplace vote-dispute <dispute_id> <client|freelancer> --from <arbiter>

# Queries
skillchaind query marketplace list-dispute
skillchaind query marketplace show-dispute <id>
skillchaind query marketplace list-dispute-vote
```

