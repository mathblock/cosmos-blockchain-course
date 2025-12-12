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
```

