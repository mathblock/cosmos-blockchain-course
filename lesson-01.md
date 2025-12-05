# Leçon 1 : Introduction à Cosmos et Installation d'Ignite CLI

## Objectifs
- Comprendre l'écosystème Cosmos et son architecture
- Installer Go et Ignite CLI
- Créer et lancer sa première blockchain

## Prérequis
- Connaissance de base en programmation
- Terminal Linux/macOS ou WSL2 sur Windows
- 8 Go RAM minimum

---

## 1.1 L'écosystème Cosmos en bref

Cosmos est un réseau de blockchains interopérables. Chaque blockchain est souveraine et communique avec les autres via **IBC** (Inter-Blockchain Communication).

**Composants clés :**
- **Cosmos SDK** : Framework Go pour construire des blockchains application-specific
- **CometBFT** (ex-Tendermint) : Moteur de consensus Byzantine Fault Tolerant
- **IBC** : Protocole de communication inter-chaînes
- **Ignite CLI** : Outil de scaffolding et développement rapide

**Architecture d'une blockchain Cosmos :**
```
┌─────────────────────────────────────┐
│           Application               │  ← Cosmos SDK (logique métier)
├─────────────────────────────────────┤
│              ABCI                   │  ← Interface Application-Consensus
├─────────────────────────────────────┤
│            CometBFT                 │  ← Consensus + Networking
└─────────────────────────────────────┘
```

---

## 1.2 Installation de Go

Ignite CLI requiert **Go 1.23+** (recommandé: 1.24.1).

```bash
# Télécharger Go (Linux/WSL)
wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz

# Extraire dans /usr/local
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz

# Configurer les variables d'environnement
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

# Vérifier l'installation
go version
# Output attendu: go version go1.24.1 linux/amd64
```

**macOS avec Homebrew :**
```bash
brew install go
```

---

## 1.3 Installation d'Ignite CLI

```bash
# Installation via curl (recommandé)
curl https://get.ignite.com/cli! | bash

# Vérifier l'installation
ignite version
# Output attendu: Ignite CLI version v29.x.x
```

**Alternative - Installation depuis les sources :**
```bash
git clone https://github.com/ignite/cli.git
cd cli
make install
```

---

## 1.4 Première blockchain de test

Créons une blockchain minimale pour valider notre installation.

```bash
# Créer une nouvelle blockchain
ignite scaffold chain testchain --skip-git

# Structure générée
cd testchain
ls -la
# app/          - Configuration de l'application
# cmd/          - Point d'entrée du binaire
# proto/        - Définitions Protobuf
# x/            - Modules custom
# config.yml    - Configuration Ignite
```

**Lancer la blockchain en mode développement :**
```bash
# Compile et lance avec hot-reload
ignite chain serve

# Output attendu:
# 🌍 Tendermint node: http://localhost:26657
# 🌍 Blockchain API: http://localhost:1317
# 🌍 Token faucet:   http://localhost:4500
```

La commande `ignite chain serve` :
- Compile le code Go
- Génère les fichiers Protobuf
- Initialise la chaîne avec un validateur
- Lance le nœud avec hot-reload automatique

---

## 1.5 Explorer la blockchain

**Ouvrir un nouveau terminal** pendant que la chaîne tourne :

```bash
cd testchain

# Lister les comptes créés par défaut
testchaind keys list

# Output:
# - address: cosmos1...
#   name: alice
# - address: cosmos1...
#   name: bob

# Vérifier le solde d'Alice
testchaind query bank balances $(testchaind keys show alice -a)

# Output:
# balances:
# - amount: "200000000"
#   denom: stake
# - amount: "20000"
#   denom: token
```

**Tester une transaction :**
```bash
# Envoyer des tokens d'Alice à Bob
testchaind tx bank send alice $(testchaind keys show bob -a) 1000token --yes

# Vérifier le nouveau solde de Bob
testchaind query bank balances $(testchaind keys show bob -a)
```

---

## 1.6 Structure du fichier config.yml

```yaml
version: 1

# Comptes initiaux avec leurs soldes
accounts:
  - name: alice
    coins: ['20000token', '200000000stake']
  - name: bob
    coins: ['10000token', '100000000stake']

# Configuration du validateur
validators:
  - name: alice
    bonded: '100000000stake'

# Faucet pour obtenir des tokens en dev
faucet:
  name: bob
  coins: ['5token', '100000stake']
  port: 4500

# Configuration genesis
genesis:
  chain_id: "hello"

# Génération client TypeScript (optionnel)
client:
  typescript:
    path: "ts-client"
```

---

## 1.7 Commandes Ignite essentielles

| Commande | Description |
|----------|-------------|
| `ignite scaffold chain <name>` | Créer une nouvelle blockchain |
| `ignite chain serve` | Lancer en mode dev avec hot-reload |
| `ignite chain serve --reset-once` | Reset l'état et relancer |
| `ignite chain build` | Compiler sans lancer |
| `ignite generate proto-go` | Régénérer le code Go depuis Proto |
| `ignite generate ts-client` | Générer le client TypeScript |
| `ignite docs` | Ouvrir la documentation |

---

## 1.8 Test pratique

Effectuez les manipulations suivantes pour valider votre compréhension :

```bash
# 1. Arrêter la chaîne (Ctrl+C dans le terminal ignite chain serve)

# 2. Modifier config.yml pour ajouter un compte "charlie"
# accounts:
#   - name: charlie
#     coins: ['5000token', '50000000stake']

# 3. Relancer avec reset pour appliquer les changements
ignite chain serve --reset-once

# 4. Vérifier que charlie existe
testchaind keys list
testchaind query bank balances $(testchaind keys show charlie -a)

# 5. Effectuer un transfert de charlie vers alice
testchaind tx bank send charlie $(testchaind keys show alice -a) 500token --yes
```

---

## Questions de révision

1. **Quel est le rôle de CometBFT dans l'architecture Cosmos ?**

2. **Quelle commande permet de lancer une blockchain en mode développement avec hot-reload ?**

3. **Dans quel fichier configure-t-on les comptes initiaux et leurs soldes ?**

4. **Quelle est la différence entre `ignite chain serve` et `ignite chain serve --reset-once` ?**

5. **Quel port expose l'API REST de la blockchain par défaut ?**

---

## Ressources complémentaires

- Documentation Ignite CLI : https://docs.ignite.com
- Cosmos SDK Docs : https://docs.cosmos.network
- Tutoriels Cosmos : https://tutorials.cosmos.network

---

**Prochaine leçon** : Nous allons scaffolder le projet SkillChain avec ses premiers modules et comprendre la structure générée en détail.
