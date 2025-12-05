# Leçon 10 : Déploiement Testnet et Monitoring

## Objectifs
- Préparer la blockchain pour un déploiement production
- Déployer SkillChain sur un testnet avec Docker
- Configurer le monitoring avec Prometheus et Grafana
- Mettre en place un explorateur de blocs

## Prérequis
- Leçon 9 complétée
- Docker et Docker Compose installés
- Serveur avec 4 CPU, 8 Go RAM minimum (pour le testnet)

---

## 10.1 Architecture de déploiement

Pour un testnet public, nous avons besoin de plusieurs composants :

```
┌─────────────────────────────────────────────────────────────────────┐
│                    ARCHITECTURE TESTNET                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │                         VALIDATEURS                            │ │
│  │  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   │ │
│  │  │Validator1│   │Validator2│   │Validator3│   │Validator4│   │ │
│  │  │ (genesis)│   │  (peer)  │   │  (peer)  │   │  (peer)  │   │ │
│  │  └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘   │ │
│  │       │              │              │              │          │ │
│  │       └──────────────┼──────────────┼──────────────┘          │ │
│  │                      │              │                          │ │
│  └──────────────────────┼──────────────┼──────────────────────────┘ │
│                         │              │                            │
│  ┌──────────────────────┼──────────────┼──────────────────────────┐ │
│  │                    SERVICES                                    │ │
│  │  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   │ │
│  │  │  Faucet  │   │ Explorer │   │Prometheus│   │ Grafana  │   │ │
│  │  │  :4500   │   │  :8080   │   │  :9090   │   │  :3000   │   │ │
│  │  └──────────┘   └──────────┘   └──────────┘   └──────────┘   │ │
│  └──────────────────────────────────────────────────────────────────┘ │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 10.2 Préparer le build de production

**Créer un Dockerfile optimisé :**

**Dockerfile :**
```dockerfile
# ==================== BUILD STAGE ====================
FROM golang:1.24-alpine AS builder

# Installer les dépendances de build
RUN apk add --no-cache make git gcc musl-dev linux-headers

WORKDIR /app

# Copier les fichiers de dépendances d'abord (pour le cache Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copier le reste du code
COPY . .

# Build le binaire avec optimisations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.version=$(git describe --tags 2>/dev/null || echo 'dev')" \
    -o /app/build/skillchaind ./cmd/skillchaind

# ==================== RUNTIME STAGE ====================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl jq bash

# Créer un utilisateur non-root
RUN addgroup -S skillchain && adduser -S skillchain -G skillchain

WORKDIR /home/skillchain

# Copier le binaire depuis le builder
COPY --from=builder /app/build/skillchaind /usr/local/bin/

# Créer les répertoires nécessaires
RUN mkdir -p /home/skillchain/.skillchain && \
    chown -R skillchain:skillchain /home/skillchain

USER skillchain

# Ports exposés
# P2P: 26656, RPC: 26657, REST API: 1317, gRPC: 9090, Prometheus: 26660
EXPOSE 26656 26657 1317 9090 26660

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:26657/health || exit 1

ENTRYPOINT ["skillchaind"]
CMD ["start"]
```

**Build l'image :**
```bash
cd skillchain
docker build -t skillchain:latest .
```

---

## 10.3 Configuration Docker Compose

**docker-compose.yml :**
```yaml
version: '3.8'

services:
  # ==================== VALIDATOR NODE ====================
  skillchain-node:
    image: skillchain:latest
    container_name: skillchain-validator
    restart: unless-stopped
    ports:
      - "26656:26656"  # P2P
      - "26657:26657"  # RPC
      - "1317:1317"    # REST API
      - "9090:9090"    # gRPC
      - "26660:26660"  # Prometheus metrics
    volumes:
      - skillchain-data:/home/skillchain/.skillchain
      - ./config/genesis.json:/home/skillchain/.skillchain/config/genesis.json:ro
      - ./config/config.toml:/home/skillchain/.skillchain/config/config.toml:ro
      - ./config/app.toml:/home/skillchain/.skillchain/config/app.toml:ro
    environment:
      - SKILLCHAIN_MONIKER=validator-1
      - SKILLCHAIN_CHAIN_ID=skillchain-testnet-1
    command: start --pruning=nothing --minimum-gas-prices=0.025uskill
    networks:
      - skillchain-network
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"

  # ==================== FAUCET ====================
  faucet:
    image: skillchain:latest
    container_name: skillchain-faucet
    restart: unless-stopped
    ports:
      - "4500:4500"
    depends_on:
      - skillchain-node
    environment:
      - FAUCET_MNEMONIC=${FAUCET_MNEMONIC}
      - FAUCET_AMOUNT=10000uskill
      - FAUCET_RATE_LIMIT=3600  # 1 request per hour
    command: >
      sh -c "
        sleep 30 &&
        skillchaind faucet --address=0.0.0.0 --port=4500
      "
    networks:
      - skillchain-network

  # ==================== PROMETHEUS ====================
  prometheus:
    image: prom/prometheus:v2.48.0
    container_name: skillchain-prometheus
    restart: unless-stopped
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.enable-lifecycle'
    networks:
      - skillchain-network

  # ==================== GRAFANA ====================
  grafana:
    image: grafana/grafana:10.2.0
    container_name: skillchain-grafana
    restart: unless-stopped
    ports:
      - "3001:3000"
    volumes:
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
      - grafana-data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
    networks:
      - skillchain-network

  # ==================== BLOCK EXPLORER ====================
  explorer:
    image: ping-pub/explorer:latest
    container_name: skillchain-explorer
    restart: unless-stopped
    ports:
      - "8080:80"
    volumes:
      - ./explorer/config.json:/app/public/chain-config.json:ro
    environment:
      - VUE_APP_API_URL=http://localhost:1317
      - VUE_APP_RPC_URL=http://localhost:26657
    networks:
      - skillchain-network

networks:
  skillchain-network:
    driver: bridge

volumes:
  skillchain-data:
  prometheus-data:
  grafana-data:
```

---

## 10.4 Configuration du nœud

**config/config.toml** (extraits importants) :
```toml
# Nom du validateur
moniker = "skillchain-validator-1"

# Configuration réseau
[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = ""  # Mettre l'IP publique en production
seeds = ""
persistent_peers = ""
max_num_inbound_peers = 40
max_num_outbound_peers = 10

# Configuration RPC
[rpc]
laddr = "tcp://0.0.0.0:26657"
cors_allowed_origins = ["*"]

# Activer les métriques Prometheus
[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
namespace = "skillchain"

# Configuration consensus
[consensus]
timeout_commit = "5s"
```

**config/app.toml** (extraits) :
```toml
# Configuration API REST
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"
enabled-unsafe-cors = true

# Configuration gRPC
[grpc]
enable = true
address = "0.0.0.0:9090"

# Configuration gRPC-Web
[grpc-web]
enable = true
address = "0.0.0.0:9091"

# Minimum gas prices
minimum-gas-prices = "0.025uskill"

# Pruning (garder moins de données pour économiser l'espace)
pruning = "custom"
pruning-keep-recent = "100"
pruning-keep-every = "500"
pruning-interval = "10"
```

---

## 10.5 Initialisation du testnet

**Script d'initialisation - init-testnet.sh :**
```bash
#!/bin/bash
set -e

CHAIN_ID="skillchain-testnet-1"
MONIKER="validator-1"
HOME_DIR="./testnet-data"
DENOM="uskill"

echo "=== Initialisation du testnet SkillChain ==="

# Nettoyer les anciennes données
rm -rf $HOME_DIR
mkdir -p $HOME_DIR

# Initialiser la chaîne
skillchaind init $MONIKER --chain-id $CHAIN_ID --home $HOME_DIR

# Créer les comptes
echo "Création des comptes..."

# Validateur principal
VAL_MNEMONIC=$(skillchaind keys add validator --keyring-backend test --home $HOME_DIR --output json 2>&1 | tail -n 1)
VAL_ADDRESS=$(skillchaind keys show validator -a --keyring-backend test --home $HOME_DIR)

# Compte faucet
FAUCET_MNEMONIC=$(skillchaind keys add faucet --keyring-backend test --home $HOME_DIR --output json 2>&1 | tail -n 1)
FAUCET_ADDRESS=$(skillchaind keys show faucet -a --keyring-backend test --home $HOME_DIR)

# Sauvegarder les mnemonics (IMPORTANT: sécuriser en production!)
echo "Validator mnemonic: $VAL_MNEMONIC" > $HOME_DIR/validator_mnemonic.txt
echo "Faucet mnemonic: $FAUCET_MNEMONIC" > $HOME_DIR/faucet_mnemonic.txt

# Modifier le genesis
echo "Configuration du genesis..."

# Ajouter les comptes au genesis
skillchaind genesis add-genesis-account $VAL_ADDRESS 1000000000$DENOM,100000000stake --home $HOME_DIR
skillchaind genesis add-genesis-account $FAUCET_ADDRESS 10000000000$DENOM --home $HOME_DIR

# Créer la transaction de genèse du validateur
skillchaind genesis gentx validator 100000000stake \
  --chain-id $CHAIN_ID \
  --moniker $MONIKER \
  --commission-rate 0.05 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1 \
  --keyring-backend test \
  --home $HOME_DIR

# Collecter les gentx
skillchaind genesis collect-gentxs --home $HOME_DIR

# Valider le genesis
skillchaind genesis validate-genesis --home $HOME_DIR

# Copier les fichiers de config
cp $HOME_DIR/config/genesis.json ./config/genesis.json

echo ""
echo "=== Testnet initialisé avec succès ==="
echo "Validator: $VAL_ADDRESS"
echo "Faucet: $FAUCET_ADDRESS"
echo ""
echo "Pour démarrer:"
echo "  docker-compose up -d"
```

---

## 10.6 Configuration Prometheus

**monitoring/prometheus.yml :**
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  # Métriques du nœud SkillChain
  - job_name: 'skillchain-node'
    static_configs:
      - targets: ['skillchain-node:26660']
    metrics_path: /metrics
    scheme: http

  # Métriques de Prometheus lui-même
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

# Règles d'alerting (optionnel)
rule_files:
  - /etc/prometheus/alerts/*.yml

alerting:
  alertmanagers:
    - static_configs:
        - targets: []
```

---

## 10.7 Dashboard Grafana

**monitoring/grafana/provisioning/datasources/prometheus.yml :**
```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
```

**monitoring/grafana/provisioning/dashboards/dashboards.yml :**
```yaml
apiVersion: 1

providers:
  - name: 'SkillChain'
    orgId: 1
    folder: 'SkillChain'
    folderUid: 'skillchain'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 30
    options:
      path: /var/lib/grafana/dashboards
```

**monitoring/grafana/dashboards/skillchain.json** (simplifié) :
```json
{
  "dashboard": {
    "title": "SkillChain Testnet",
    "uid": "skillchain-main",
    "panels": [
      {
        "title": "Block Height",
        "type": "stat",
        "gridPos": { "x": 0, "y": 0, "w": 6, "h": 4 },
        "targets": [
          {
            "expr": "tendermint_consensus_height",
            "legendFormat": "Height"
          }
        ]
      },
      {
        "title": "Transactions per Block",
        "type": "graph",
        "gridPos": { "x": 6, "y": 0, "w": 12, "h": 8 },
        "targets": [
          {
            "expr": "rate(tendermint_consensus_total_txs[5m])",
            "legendFormat": "TX Rate"
          }
        ]
      },
      {
        "title": "Validators",
        "type": "stat",
        "gridPos": { "x": 18, "y": 0, "w": 6, "h": 4 },
        "targets": [
          {
            "expr": "tendermint_consensus_validators",
            "legendFormat": "Active Validators"
          }
        ]
      },
      {
        "title": "Consensus Round Duration",
        "type": "graph",
        "gridPos": { "x": 0, "y": 8, "w": 12, "h": 8 },
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(tendermint_consensus_round_duration_seconds_bucket[5m]))",
            "legendFormat": "p95 Round Duration"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "gridPos": { "x": 12, "y": 8, "w": 12, "h": 8 },
        "targets": [
          {
            "expr": "process_resident_memory_bytes{job='skillchain-node'}",
            "legendFormat": "Memory"
          }
        ]
      }
    ]
  }
}
```

---

## 10.8 Configuration de l'explorateur

**explorer/config.json :**
```json
{
  "chain_name": "SkillChain Testnet",
  "coingecko": "",
  "api": ["http://localhost:1317"],
  "rpc": ["http://localhost:26657"],
  "snapshot_provider": "",
  "sdk_version": "0.50.2",
  "coin_type": "118",
  "min_tx_fee": "5000",
  "addr_prefix": "skill",
  "logo": "/logos/skillchain.png",
  "keplr_features": ["ibc-transfer"],
  "assets": [
    {
      "base": "uskill",
      "symbol": "SKILL",
      "exponent": "6",
      "coingecko_id": "",
      "logo": "/logos/skill.png"
    }
  ]
}
```

---

## 10.9 Lancement du testnet

```bash
# 1. Initialiser le testnet
chmod +x init-testnet.sh
./init-testnet.sh

# 2. Créer le fichier .env
cat > .env << EOF
FAUCET_MNEMONIC="your faucet mnemonic here"
GRAFANA_PASSWORD=secure_password_here
EOF

# 3. Démarrer tous les services
docker-compose up -d

# 4. Vérifier les logs
docker-compose logs -f skillchain-node

# 5. Vérifier le statut
curl http://localhost:26657/status | jq '.result.sync_info'

# 6. Accéder aux interfaces
# - API REST: http://localhost:1317
# - RPC: http://localhost:26657
# - Explorer: http://localhost:8080
# - Grafana: http://localhost:3001 (admin/admin)
# - Prometheus: http://localhost:9091
```

---

## 10.10 Ajouter des validateurs supplémentaires

Pour un testnet robuste, ajoutons d'autres validateurs.

**Sur un second serveur :**
```bash
# 1. Copier le genesis depuis le premier validateur
scp validator1:/path/to/config/genesis.json ./config/genesis.json

# 2. Initialiser le nouveau nœud
skillchaind init validator-2 --chain-id skillchain-testnet-1

# 3. Remplacer le genesis
cp ./config/genesis.json ~/.skillchain/config/genesis.json

# 4. Configurer le peer du premier validateur
# Dans config.toml:
# persistent_peers = "node_id@validator1_ip:26656"

# 5. Démarrer le nœud (en mode full node d'abord)
skillchaind start

# 6. Une fois synchronisé, créer une clé et devenir validateur
skillchaind keys add validator2 --keyring-backend test

# 7. Obtenir des tokens depuis le faucet
curl -X POST http://validator1_ip:4500/faucet \
  -d '{"address": "skill1...", "denom": "uskill"}'

# 8. Créer la transaction de staking
skillchaind tx staking create-validator \
  --amount=10000000stake \
  --pubkey=$(skillchaind tendermint show-validator) \
  --moniker="validator-2" \
  --chain-id=skillchain-testnet-1 \
  --commission-rate="0.05" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --from=validator2 \
  --keyring-backend=test \
  --yes
```

---

## 10.11 Scripts de maintenance

**scripts/backup.sh :**
```bash
#!/bin/bash
# Backup des données du validateur

BACKUP_DIR="/backups/skillchain"
DATE=$(date +%Y%m%d_%H%M%S)
DATA_DIR="/home/skillchain/.skillchain"

mkdir -p $BACKUP_DIR

# Arrêter le nœud temporairement pour un backup cohérent
docker-compose stop skillchain-node

# Backup des données
tar -czvf $BACKUP_DIR/skillchain_$DATE.tar.gz \
  --exclude='*/cs.wal' \
  $DATA_DIR/data \
  $DATA_DIR/config

# Redémarrer
docker-compose start skillchain-node

# Nettoyer les vieux backups (garder 7 jours)
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_DIR/skillchain_$DATE.tar.gz"
```

**scripts/health-check.sh :**
```bash
#!/bin/bash
# Script de santé pour alerting

RPC_URL="http://localhost:26657"
ALERT_WEBHOOK="https://hooks.slack.com/services/xxx"

# Vérifier si le nœud répond
if ! curl -s -f "$RPC_URL/health" > /dev/null; then
    curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"⚠️ SkillChain node is not responding!"}' \
        $ALERT_WEBHOOK
    exit 1
fi

# Vérifier si le nœud est synchronisé
CATCHING_UP=$(curl -s "$RPC_URL/status" | jq -r '.result.sync_info.catching_up')
if [ "$CATCHING_UP" = "true" ]; then
    curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"⚠️ SkillChain node is catching up (not synced)"}' \
        $ALERT_WEBHOOK
fi

# Vérifier l'espace disque
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"⚠️ SkillChain disk usage is at ${DISK_USAGE}%\"}" \
        $ALERT_WEBHOOK
fi

echo "Health check passed"
```

---

## 10.12 Métriques importantes à surveiller

| Métrique | Description | Seuil d'alerte |
|----------|-------------|----------------|
| `tendermint_consensus_height` | Hauteur du bloc | Pas d'augmentation pendant 1min |
| `tendermint_consensus_validators` | Nombre de validateurs | < 2/3 du set |
| `tendermint_consensus_missing_validators` | Validateurs absents | > 1/3 du set |
| `tendermint_p2p_peers` | Nombre de peers | < 3 |
| `process_resident_memory_bytes` | Mémoire utilisée | > 80% de la RAM |
| `tendermint_consensus_total_txs` | Total de transactions | Monitoring uniquement |
| `tendermint_consensus_block_size_bytes` | Taille des blocs | > 80% du max |

---

## 10.13 Mise à jour de la chaîne (Upgrade)

Pour les mises à jour de la chaîne, Cosmos SDK utilise le module `upgrade`.

**Proposer une mise à jour via gouvernance :**
```bash
# 1. Soumettre une proposition de mise à jour
skillchaind tx gov submit-proposal software-upgrade v2.0.0 \
  --title "Upgrade to v2.0.0" \
  --description "This upgrade includes new marketplace features" \
  --upgrade-height 100000 \
  --deposit 10000000uskill \
  --from validator \
  --yes

# 2. Voter sur la proposition
skillchaind tx gov vote 1 yes --from validator --yes

# 3. Une fois votée, préparer le nouveau binaire
# Le nœud s'arrêtera automatiquement à la hauteur spécifiée
# Remplacer le binaire et redémarrer
```

**Avec Cosmovisor (automatisation des upgrades) :**
```bash
# Structure des dossiers Cosmovisor
.skillchain/
└── cosmovisor/
    ├── genesis/
    │   └── bin/
    │       └── skillchaind  # Version actuelle
    └── upgrades/
        └── v2.0.0/
            └── bin/
                └── skillchaind  # Nouvelle version

# Cosmovisor gère automatiquement le switch de binaire
```

---

## 10.14 Checklist de déploiement production

**Avant le lancement :**
```
□ Genesis validé avec tous les validateurs
□ Paramètres de gouvernance appropriés
□ Minimum gas prices configurés sur tous les nœuds
□ TLS/HTTPS configuré pour les endpoints publics
□ Firewall configuré (ports nécessaires seulement)
□ Backups automatiques en place
□ Monitoring et alerting configurés
□ Documentation pour les utilisateurs
□ Faucet avec rate limiting
□ Explorer fonctionnel
```

**Sécurité :**
```
□ Clés des validateurs sécurisées (HSM recommandé en mainnet)
□ Pas de clés en clair dans les fichiers de config
□ SSH avec clés uniquement (pas de mot de passe)
□ Fail2ban ou équivalent installé
□ Mises à jour de sécurité automatiques pour l'OS
□ Nœuds sentinelles configurés (pour les validateurs mainnet)
```

---

## 10.15 Résumé du cours

Félicitations ! Vous avez complété le cours de développement blockchain Cosmos avec SkillChain.

**Ce que vous avez appris :**

1. **Fondamentaux Cosmos** : Architecture, CometBFT, modules
2. **Ignite CLI** : Scaffolding rapide de blockchain et modules
3. **Développement de modules** : Keeper, messages, queries
4. **Gestion d'état** : Stockage avec Collections, types de données
5. **Logique métier** : Workflow complet d'une marketplace
6. **Escrow** : Paiements sécurisés avec module accounts
7. **Disputes** : Système d'arbitrage décentralisé
8. **Frontend** : React, CosmJS, intégration Keplr
9. **IBC** : Communication inter-chaînes, tokens cross-chain
10. **Déploiement** : Docker, monitoring, maintenance

**Prochaines étapes suggérées :**

1. Ajouter plus de fonctionnalités (système de reviews, badges, etc.)
2. Implémenter des smart contracts CosmWasm
3. Participer à un testnet public Cosmos
4. Contribuer à des projets open source de l'écosystème
5. Explorer les solutions de scaling (rollups, data availability)

---

## Questions de révision finales

1. **Quels ports doit-on exposer pour un nœud validateur public ?**

2. **Pourquoi utilise-t-on Prometheus plutôt que de simplement regarder les logs ?**

3. **Quelle est la différence entre un full node et un validateur ?**

4. **Comment Cosmovisor facilite-t-il les mises à jour de la chaîne ?**

5. **Pourquoi est-il important d'avoir au moins 4 validateurs pour un testnet ?**

6. **Quelles métriques indiquent un problème de consensus imminent ?**

---

## Récapitulatif des commandes de déploiement

```bash
# Build
docker build -t skillchain:latest .

# Initialisation
./init-testnet.sh

# Démarrage
docker-compose up -d

# Logs
docker-compose logs -f skillchain-node

# Statut
curl http://localhost:26657/status | jq

# Backup
./scripts/backup.sh

# Health check
./scripts/health-check.sh

# Ajouter un validateur
skillchaind tx staking create-validator ...

# Proposition d'upgrade
skillchaind tx gov submit-proposal software-upgrade ...
```

---

## Ressources supplémentaires

- **Documentation Cosmos SDK** : https://docs.cosmos.network
- **Ignite CLI** : https://docs.ignite.com
- **Tutoriels Cosmos** : https://tutorials.cosmos.network
- **Forum Cosmos** : https://forum.cosmos.network
- **Discord Cosmos** : https://discord.gg/cosmosnetwork
- **IBC Protocol** : https://ibc.cosmos.network
- **CometBFT** : https://docs.cometbft.com

---

**Merci d'avoir suivi ce cours ! Bonne continuation dans l'écosystème Cosmos !** 🚀
