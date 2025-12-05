# Leçon 7 : Frontend React avec CosmJS

## Objectifs
- Créer une application React/TypeScript pour SkillChain
- Configurer CosmJS pour la communication blockchain
- Implémenter les queries (lecture du state)
- Préparer la structure pour les transactions

## Prérequis
- Leçon 6 complétée
- Node.js 18+ installé
- Connaissances de base React/TypeScript

---

## 7.1 Architecture Frontend

Notre frontend communiquera avec la blockchain via deux canaux :

```
┌─────────────────────────────────────────────────────────────────┐
│                      ARCHITECTURE FRONTEND                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    React Application                     │   │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │   │
│  │  │   Pages     │    │   Hooks     │    │ Components  │  │   │
│  │  └──────┬──────┘    └──────┬──────┘    └─────────────┘  │   │
│  │         │                  │                             │   │
│  │         └────────┬─────────┘                             │   │
│  │                  ▼                                       │   │
│  │         ┌─────────────────┐                             │   │
│  │         │  CosmJS Client  │                             │   │
│  │         └────────┬────────┘                             │   │
│  └──────────────────┼──────────────────────────────────────┘   │
│                     │                                           │
│            ┌────────┴────────┐                                 │
│            ▼                 ▼                                 │
│    ┌──────────────┐  ┌──────────────┐                         │
│    │  REST API    │  │   gRPC-web   │                         │
│    │ :1317        │  │   :9091      │                         │
│    └──────────────┘  └──────────────┘                         │
│            │                 │                                 │
│            └────────┬────────┘                                 │
│                     ▼                                          │
│            ┌──────────────────┐                                │
│            │   SkillChain     │                                │
│            │   Blockchain     │                                │
│            └──────────────────┘                                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 7.2 Initialisation du projet React

```bash
# Dans le dossier skillchain, créer le frontend
cd skillchain

# Créer l'application React avec Vite
npm create vite@latest frontend -- --template react-ts

cd frontend

# Installer les dépendances CosmJS
npm install @cosmjs/stargate @cosmjs/proto-signing @cosmjs/encoding @cosmjs/math

# Installer les utilitaires
npm install @tanstack/react-query axios

# Installer le state management
npm install zustand

# Installer les composants UI (optionnel mais recommandé)
npm install @headlessui/react @heroicons/react

# Installer les types
npm install -D @types/node
```

---

## 7.3 Structure du projet frontend

```
frontend/
├── src/
│   ├── components/          # Composants réutilisables
│   │   ├── Layout.tsx
│   │   ├── Navbar.tsx
│   │   ├── GigCard.tsx
│   │   └── ProfileCard.tsx
│   ├── hooks/               # Hooks custom
│   │   ├── useSkillChain.ts
│   │   ├── useWallet.ts
│   │   └── useQueries.ts
│   ├── pages/               # Pages de l'application
│   │   ├── Home.tsx
│   │   ├── Gigs.tsx
│   │   ├── Profile.tsx
│   │   └── Contracts.tsx
│   ├── services/            # Services API
│   │   ├── cosmjs.ts
│   │   └── api.ts
│   ├── stores/              # State global (Zustand)
│   │   └── walletStore.ts
│   ├── types/               # Types TypeScript
│   │   └── skillchain.ts
│   ├── App.tsx
│   └── main.tsx
├── package.json
└── vite.config.ts
```

---

## 7.4 Configuration de Vite

**vite.config.ts :**
```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  define: {
    // Nécessaire pour certaines dépendances CosmJS
    global: 'globalThis',
  },
  server: {
    port: 3000,
    proxy: {
      // Proxy vers l'API REST de la blockchain en dev
      '/api': {
        target: 'http://localhost:1317',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})
```

---

## 7.5 Types TypeScript pour SkillChain

**src/types/skillchain.ts :**
```typescript
// Types correspondant aux structures Protobuf de la blockchain

export interface Profile {
  owner: string;
  name: string;
  bio: string;
  skills: string[];
  hourlyRate: string;  // uint64 en string pour éviter les pertes de précision
  totalJobs: string;
  totalEarned: string;
  ratingSum: string;
  ratingCount: string;
}

export interface Gig {
  id: string;
  title: string;
  description: string;
  owner: string;
  price: string;
  category: string;
  deliveryDays: string;
  status: GigStatus;
  createdAt: string;
}

export type GigStatus = 'open' | 'in_progress' | 'completed' | 'cancelled' | 'disputed';

export interface Application {
  id: string;
  gigId: string;
  freelancer: string;
  coverLetter: string;
  proposedPrice: string;
  proposedDays: string;
  status: ApplicationStatus;
  createdAt: string;
}

export type ApplicationStatus = 'pending' | 'accepted' | 'rejected' | 'withdrawn';

export interface Contract {
  id: string;
  gigId: string;
  applicationId: string;
  client: string;
  freelancer: string;
  price: string;
  deliveryDeadline: string;
  status: ContractStatus;
  createdAt: string;
  completedAt: string;
}

export type ContractStatus = 'active' | 'delivered' | 'completed' | 'disputed' | 'cancelled';

export interface Dispute {
  id: string;
  contractId: string;
  initiator: string;
  reason: string;
  clientEvidence: string;
  freelancerEvidence: string;
  status: DisputeStatus;
  votesClient: string;
  votesFreelancer: string;
  resolution: string;
  createdAt: string;
  deadline: string;
}

export type DisputeStatus = 'open' | 'voting' | 'resolved_client' | 'resolved_freelancer' | 'expired';

export interface Params {
  platformFeePercent: string;
  minContractDuration: string;
  minGigPrice: string;
  disputeDuration: string;
  minArbitersRequired: string;
  arbiterStakeRequired: string;
}

// Types pour les réponses API
export interface QueryResponse<T> {
  data: T;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    nextKey: string | null;
    total: string;
  };
}

// Types pour les balances
export interface Coin {
  denom: string;
  amount: string;
}

export interface Balance {
  balances: Coin[];
}
```

---

## 7.6 Service CosmJS

**src/services/cosmjs.ts :**
```typescript
import { StargateClient, SigningStargateClient, GasPrice } from '@cosmjs/stargate';
import { OfflineSigner } from '@cosmjs/proto-signing';

// Configuration de la chaîne
export const CHAIN_CONFIG = {
  chainId: 'skillchain-local-1',
  chainName: 'SkillChain Local',
  rpc: 'http://localhost:26657',
  rest: 'http://localhost:1317',
  bech32Prefix: 'skill',
  currencies: [
    {
      coinDenom: 'SKILL',
      coinMinimalDenom: 'uskill',
      coinDecimals: 6,
    },
    {
      coinDenom: 'STAKE',
      coinMinimalDenom: 'stake',
      coinDecimals: 6,
    },
  ],
  feeCurrencies: [
    {
      coinDenom: 'SKILL',
      coinMinimalDenom: 'uskill',
      coinDecimals: 6,
      gasPriceStep: {
        low: 0.01,
        average: 0.025,
        high: 0.04,
      },
    },
  ],
  stakeCurrency: {
    coinDenom: 'STAKE',
    coinMinimalDenom: 'stake',
    coinDecimals: 6,
  },
};

// Client en lecture seule (pas besoin de wallet)
let queryClient: StargateClient | null = null;

export async function getQueryClient(): Promise<StargateClient> {
  if (!queryClient) {
    queryClient = await StargateClient.connect(CHAIN_CONFIG.rpc);
  }
  return queryClient;
}

// Client avec signature (nécessite un wallet)
export async function getSigningClient(signer: OfflineSigner): Promise<SigningStargateClient> {
  return SigningStargateClient.connectWithSigner(
    CHAIN_CONFIG.rpc,
    signer,
    {
      gasPrice: GasPrice.fromString('0.025uskill'),
    }
  );
}

// Utilitaire pour formater les montants
export function formatAmount(amount: string, decimals: number = 6): string {
  const value = parseInt(amount) / Math.pow(10, decimals);
  return value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: decimals,
  });
}

// Utilitaire pour convertir en micro-unités
export function toMicroUnits(amount: number, decimals: number = 6): string {
  return Math.floor(amount * Math.pow(10, decimals)).toString();
}

// Utilitaire pour raccourcir les adresses
export function shortenAddress(address: string, chars: number = 8): string {
  if (!address) return '';
  return `${address.slice(0, chars + 5)}...${address.slice(-chars)}`;
}
```

---

## 7.7 Service API REST

**src/services/api.ts :**
```typescript
import axios from 'axios';
import type {
  Profile,
  Gig,
  Application,
  Contract,
  Dispute,
  Params,
  Balance,
} from '@/types/skillchain';

// Base URL de l'API REST
const API_BASE = 'http://localhost:1317';

// Instance Axios configurée
const api = axios.create({
  baseURL: API_BASE,
  timeout: 10000,
});

// ============ QUERIES MARKETPLACE ============

// Paramètres du module
export async function getParams(): Promise<Params> {
  const response = await api.get('/skillchain/marketplace/params');
  return response.data.params;
}

// Profils
export async function getProfile(address: string): Promise<Profile | null> {
  try {
    const response = await api.get(`/skillchain/marketplace/profile/${address}`);
    return response.data.profile;
  } catch (error: any) {
    if (error.response?.status === 404) return null;
    throw error;
  }
}

export async function getAllProfiles(): Promise<Profile[]> {
  const response = await api.get('/skillchain/marketplace/profile');
  return response.data.profile || [];
}

// Gigs
export async function getGig(id: string): Promise<Gig | null> {
  try {
    const response = await api.get(`/skillchain/marketplace/gig/${id}`);
    return response.data.gig;
  } catch (error: any) {
    if (error.response?.status === 404) return null;
    throw error;
  }
}

export async function getAllGigs(): Promise<Gig[]> {
  const response = await api.get('/skillchain/marketplace/gig');
  return response.data.gig || [];
}

export async function getOpenGigs(): Promise<Gig[]> {
  const gigs = await getAllGigs();
  return gigs.filter(gig => gig.status === 'open');
}

// Applications
export async function getApplication(id: string): Promise<Application | null> {
  try {
    const response = await api.get(`/skillchain/marketplace/application/${id}`);
    return response.data.application;
  } catch (error: any) {
    if (error.response?.status === 404) return null;
    throw error;
  }
}

export async function getAllApplications(): Promise<Application[]> {
  const response = await api.get('/skillchain/marketplace/application');
  return response.data.application || [];
}

export async function getApplicationsByGig(gigId: string): Promise<Application[]> {
  const response = await api.get(`/skillchain/marketplace/applications_by_gig/${gigId}`);
  return response.data.applications || [];
}

export async function getApplicationsByFreelancer(address: string): Promise<Application[]> {
  const apps = await getAllApplications();
  return apps.filter(app => app.freelancer === address);
}

// Contracts
export async function getContract(id: string): Promise<Contract | null> {
  try {
    const response = await api.get(`/skillchain/marketplace/contract/${id}`);
    return response.data.contract;
  } catch (error: any) {
    if (error.response?.status === 404) return null;
    throw error;
  }
}

export async function getAllContracts(): Promise<Contract[]> {
  const response = await api.get('/skillchain/marketplace/contract');
  return response.data.contract || [];
}

export async function getContractsByUser(address: string): Promise<Contract[]> {
  const response = await api.get(`/skillchain/marketplace/contracts_by_user/${address}`);
  return response.data.contracts || [];
}

// Disputes
export async function getDispute(id: string): Promise<Dispute | null> {
  try {
    const response = await api.get(`/skillchain/marketplace/dispute/${id}`);
    return response.data.dispute;
  } catch (error: any) {
    if (error.response?.status === 404) return null;
    throw error;
  }
}

export async function getAllDisputes(): Promise<Dispute[]> {
  const response = await api.get('/skillchain/marketplace/dispute');
  return response.data.dispute || [];
}

// Escrow Balance
export async function getEscrowBalance(): Promise<string> {
  const response = await api.get('/skillchain/marketplace/escrow_balance');
  return response.data.balance?.amount || '0';
}

// ============ QUERIES BANK ============

export async function getBalance(address: string): Promise<Balance> {
  const response = await api.get(`/cosmos/bank/v1beta1/balances/${address}`);
  return response.data;
}

export async function getBalanceByDenom(address: string, denom: string): Promise<string> {
  const response = await api.get(`/cosmos/bank/v1beta1/balances/${address}/by_denom`, {
    params: { denom },
  });
  return response.data.balance?.amount || '0';
}
```

---

## 7.8 Store Zustand pour le Wallet

**src/stores/walletStore.ts :**
```typescript
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { OfflineSigner } from '@cosmjs/proto-signing';
import type { SigningStargateClient } from '@cosmjs/stargate';
import { getSigningClient, CHAIN_CONFIG } from '@/services/cosmjs';

interface WalletState {
  // State
  address: string | null;
  isConnected: boolean;
  isConnecting: boolean;
  signer: OfflineSigner | null;
  signingClient: SigningStargateClient | null;
  error: string | null;
  
  // Actions
  connect: () => Promise<void>;
  disconnect: () => void;
  clearError: () => void;
}

export const useWalletStore = create<WalletState>()(
  persist(
    (set, get) => ({
      // Initial state
      address: null,
      isConnected: false,
      isConnecting: false,
      signer: null,
      signingClient: null,
      error: null,
      
      // Connect to Keplr
      connect: async () => {
        set({ isConnecting: true, error: null });
        
        try {
          // Vérifier que Keplr est installé
          if (!window.keplr) {
            throw new Error('Please install Keplr extension');
          }
          
          // Suggérer la chaîne SkillChain à Keplr
          await window.keplr.experimentalSuggestChain({
            chainId: CHAIN_CONFIG.chainId,
            chainName: CHAIN_CONFIG.chainName,
            rpc: CHAIN_CONFIG.rpc,
            rest: CHAIN_CONFIG.rest,
            bip44: {
              coinType: 118,
            },
            bech32Config: {
              bech32PrefixAccAddr: CHAIN_CONFIG.bech32Prefix,
              bech32PrefixAccPub: `${CHAIN_CONFIG.bech32Prefix}pub`,
              bech32PrefixValAddr: `${CHAIN_CONFIG.bech32Prefix}valoper`,
              bech32PrefixValPub: `${CHAIN_CONFIG.bech32Prefix}valoperpub`,
              bech32PrefixConsAddr: `${CHAIN_CONFIG.bech32Prefix}valcons`,
              bech32PrefixConsPub: `${CHAIN_CONFIG.bech32Prefix}valconspub`,
            },
            currencies: CHAIN_CONFIG.currencies,
            feeCurrencies: CHAIN_CONFIG.feeCurrencies,
            stakeCurrency: CHAIN_CONFIG.stakeCurrency,
          });
          
          // Activer la chaîne
          await window.keplr.enable(CHAIN_CONFIG.chainId);
          
          // Obtenir le signer
          const signer = window.keplr.getOfflineSigner(CHAIN_CONFIG.chainId);
          const accounts = await signer.getAccounts();
          
          if (accounts.length === 0) {
            throw new Error('No accounts found');
          }
          
          // Créer le signing client
          const signingClient = await getSigningClient(signer);
          
          set({
            address: accounts[0].address,
            isConnected: true,
            isConnecting: false,
            signer,
            signingClient,
          });
          
        } catch (error: any) {
          set({
            isConnecting: false,
            error: error.message || 'Failed to connect wallet',
          });
        }
      },
      
      // Disconnect
      disconnect: () => {
        set({
          address: null,
          isConnected: false,
          signer: null,
          signingClient: null,
        });
      },
      
      clearError: () => set({ error: null }),
    }),
    {
      name: 'wallet-storage',
      // Ne pas persister le signer et signingClient
      partialize: (state) => ({
        address: state.address,
        isConnected: state.isConnected,
      }),
    }
  )
);

// Type augmentation pour window.keplr
declare global {
  interface Window {
    keplr?: {
      experimentalSuggestChain: (chainInfo: any) => Promise<void>;
      enable: (chainId: string) => Promise<void>;
      getOfflineSigner: (chainId: string) => OfflineSigner;
    };
  }
}
```

---

## 7.9 Hooks personnalisés

**src/hooks/useQueries.ts :**
```typescript
import { useQuery, useQueryClient } from '@tanstack/react-query';
import * as api from '@/services/api';
import { useWalletStore } from '@/stores/walletStore';

// Hook pour les paramètres du module
export function useParams() {
  return useQuery({
    queryKey: ['marketplace', 'params'],
    queryFn: api.getParams,
    staleTime: 60000, // 1 minute
  });
}

// Hook pour un profil spécifique
export function useProfile(address: string | null) {
  return useQuery({
    queryKey: ['marketplace', 'profile', address],
    queryFn: () => address ? api.getProfile(address) : null,
    enabled: !!address,
  });
}

// Hook pour le profil de l'utilisateur connecté
export function useMyProfile() {
  const { address } = useWalletStore();
  return useProfile(address);
}

// Hook pour tous les gigs
export function useGigs() {
  return useQuery({
    queryKey: ['marketplace', 'gigs'],
    queryFn: api.getAllGigs,
    staleTime: 10000, // 10 secondes
  });
}

// Hook pour les gigs ouverts
export function useOpenGigs() {
  return useQuery({
    queryKey: ['marketplace', 'gigs', 'open'],
    queryFn: api.getOpenGigs,
    staleTime: 10000,
  });
}

// Hook pour un gig spécifique
export function useGig(id: string | null) {
  return useQuery({
    queryKey: ['marketplace', 'gig', id],
    queryFn: () => id ? api.getGig(id) : null,
    enabled: !!id,
  });
}

// Hook pour les applications d'un gig
export function useApplicationsByGig(gigId: string | null) {
  return useQuery({
    queryKey: ['marketplace', 'applications', 'gig', gigId],
    queryFn: () => gigId ? api.getApplicationsByGig(gigId) : [],
    enabled: !!gigId,
  });
}

// Hook pour les applications de l'utilisateur connecté
export function useMyApplications() {
  const { address } = useWalletStore();
  return useQuery({
    queryKey: ['marketplace', 'applications', 'freelancer', address],
    queryFn: () => address ? api.getApplicationsByFreelancer(address) : [],
    enabled: !!address,
  });
}

// Hook pour les contracts de l'utilisateur
export function useMyContracts() {
  const { address } = useWalletStore();
  return useQuery({
    queryKey: ['marketplace', 'contracts', 'user', address],
    queryFn: () => address ? api.getContractsByUser(address) : [],
    enabled: !!address,
  });
}

// Hook pour le solde de l'utilisateur
export function useBalance(address: string | null, denom: string = 'uskill') {
  return useQuery({
    queryKey: ['bank', 'balance', address, denom],
    queryFn: () => address ? api.getBalanceByDenom(address, denom) : '0',
    enabled: !!address,
    staleTime: 5000, // 5 secondes
  });
}

// Hook pour le solde de l'utilisateur connecté
export function useMyBalance(denom: string = 'uskill') {
  const { address } = useWalletStore();
  return useBalance(address, denom);
}

// Hook pour le solde escrow
export function useEscrowBalance() {
  return useQuery({
    queryKey: ['marketplace', 'escrow'],
    queryFn: api.getEscrowBalance,
    staleTime: 10000,
  });
}

// Hook pour invalider les caches après une transaction
export function useInvalidateQueries() {
  const queryClient = useQueryClient();
  
  return {
    invalidateAll: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace'] });
      queryClient.invalidateQueries({ queryKey: ['bank'] });
    },
    invalidateGigs: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace', 'gigs'] });
    },
    invalidateContracts: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace', 'contracts'] });
    },
    invalidateProfile: (address: string) => {
      queryClient.invalidateQueries({ queryKey: ['marketplace', 'profile', address] });
    },
  };
}
```

---

## 7.10 Composant Navbar

**src/components/Navbar.tsx :**
```typescript
import { useWalletStore } from '@/stores/walletStore';
import { useMyBalance } from '@/hooks/useQueries';
import { formatAmount, shortenAddress } from '@/services/cosmjs';

export function Navbar() {
  const { address, isConnected, isConnecting, connect, disconnect, error } = useWalletStore();
  const { data: balance } = useMyBalance();
  
  return (
    <nav className="bg-gray-800 text-white p-4">
      <div className="container mx-auto flex justify-between items-center">
        {/* Logo */}
        <div className="flex items-center space-x-2">
          <span className="text-2xl font-bold">⚡ SkillChain</span>
        </div>
        
        {/* Navigation Links */}
        <div className="flex items-center space-x-6">
          <a href="/" className="hover:text-blue-400">Home</a>
          <a href="/gigs" className="hover:text-blue-400">Gigs</a>
          {isConnected && (
            <>
              <a href="/my-contracts" className="hover:text-blue-400">My Contracts</a>
              <a href="/profile" className="hover:text-blue-400">Profile</a>
            </>
          )}
        </div>
        
        {/* Wallet Connection */}
        <div className="flex items-center space-x-4">
          {isConnected ? (
            <>
              {/* Balance */}
              <div className="bg-gray-700 px-3 py-1 rounded">
                <span className="text-green-400">
                  {formatAmount(balance || '0')} SKILL
                </span>
              </div>
              
              {/* Address */}
              <div className="bg-gray-700 px-3 py-1 rounded">
                {shortenAddress(address || '')}
              </div>
              
              {/* Disconnect Button */}
              <button
                onClick={disconnect}
                className="bg-red-600 hover:bg-red-700 px-4 py-2 rounded"
              >
                Disconnect
              </button>
            </>
          ) : (
            <button
              onClick={connect}
              disabled={isConnecting}
              className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded disabled:opacity-50"
            >
              {isConnecting ? 'Connecting...' : 'Connect Wallet'}
            </button>
          )}
        </div>
      </div>
      
      {/* Error Message */}
      {error && (
        <div className="bg-red-600 text-white p-2 text-center mt-2">
          {error}
        </div>
      )}
    </nav>
  );
}
```

---

## 7.11 Composant GigCard

**src/components/GigCard.tsx :**
```typescript
import type { Gig } from '@/types/skillchain';
import { formatAmount, shortenAddress } from '@/services/cosmjs';

interface GigCardProps {
  gig: Gig;
  onApply?: (gigId: string) => void;
  showApplyButton?: boolean;
}

export function GigCard({ gig, onApply, showApplyButton = true }: GigCardProps) {
  const statusColors: Record<string, string> = {
    open: 'bg-green-500',
    in_progress: 'bg-yellow-500',
    completed: 'bg-blue-500',
    cancelled: 'bg-red-500',
    disputed: 'bg-purple-500',
  };
  
  const createdDate = new Date(parseInt(gig.createdAt) * 1000).toLocaleDateString();
  
  return (
    <div className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
      {/* Header */}
      <div className="flex justify-between items-start mb-4">
        <h3 className="text-xl font-semibold text-gray-800">{gig.title}</h3>
        <span className={`${statusColors[gig.status]} text-white text-xs px-2 py-1 rounded`}>
          {gig.status.toUpperCase()}
        </span>
      </div>
      
      {/* Description */}
      <p className="text-gray-600 mb-4 line-clamp-3">{gig.description}</p>
      
      {/* Details */}
      <div className="grid grid-cols-2 gap-4 mb-4 text-sm">
        <div>
          <span className="text-gray-500">Category:</span>
          <span className="ml-2 font-medium">{gig.category}</span>
        </div>
        <div>
          <span className="text-gray-500">Delivery:</span>
          <span className="ml-2 font-medium">{gig.deliveryDays} days</span>
        </div>
        <div>
          <span className="text-gray-500">Posted by:</span>
          <span className="ml-2 font-medium">{shortenAddress(gig.owner)}</span>
        </div>
        <div>
          <span className="text-gray-500">Posted:</span>
          <span className="ml-2 font-medium">{createdDate}</span>
        </div>
      </div>
      
      {/* Price and Action */}
      <div className="flex justify-between items-center pt-4 border-t">
        <div className="text-2xl font-bold text-green-600">
          {formatAmount(gig.price)} SKILL
        </div>
        
        {showApplyButton && gig.status === 'open' && onApply && (
          <button
            onClick={() => onApply(gig.id)}
            className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg"
          >
            Apply Now
          </button>
        )}
      </div>
    </div>
  );
}
```

---

## 7.12 Page Home

**src/pages/Home.tsx :**
```typescript
import { useOpenGigs, useParams, useEscrowBalance } from '@/hooks/useQueries';
import { GigCard } from '@/components/GigCard';
import { formatAmount } from '@/services/cosmjs';

export function HomePage() {
  const { data: gigs, isLoading: gigsLoading } = useOpenGigs();
  const { data: params } = useParams();
  const { data: escrowBalance } = useEscrowBalance();
  
  return (
    <div className="container mx-auto px-4 py-8">
      {/* Hero Section */}
      <section className="text-center mb-12">
        <h1 className="text-4xl font-bold mb-4">
          Welcome to SkillChain
        </h1>
        <p className="text-xl text-gray-600 mb-8">
          The decentralized marketplace for skills and services
        </p>
        
        {/* Stats */}
        <div className="grid grid-cols-3 gap-8 max-w-2xl mx-auto">
          <div className="bg-blue-100 p-4 rounded-lg">
            <div className="text-3xl font-bold text-blue-600">
              {gigs?.length || 0}
            </div>
            <div className="text-gray-600">Open Gigs</div>
          </div>
          <div className="bg-green-100 p-4 rounded-lg">
            <div className="text-3xl font-bold text-green-600">
              {params?.platformFeePercent || 5}%
            </div>
            <div className="text-gray-600">Platform Fee</div>
          </div>
          <div className="bg-purple-100 p-4 rounded-lg">
            <div className="text-3xl font-bold text-purple-600">
              {formatAmount(escrowBalance || '0')}
            </div>
            <div className="text-gray-600">In Escrow</div>
          </div>
        </div>
      </section>
      
      {/* Recent Gigs */}
      <section>
        <h2 className="text-2xl font-bold mb-6">Recent Open Gigs</h2>
        
        {gigsLoading ? (
          <div className="text-center py-8">Loading...</div>
        ) : gigs && gigs.length > 0 ? (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {gigs.slice(0, 6).map((gig) => (
              <GigCard
                key={gig.id}
                gig={gig}
                onApply={(id) => window.location.href = `/gigs/${id}`}
              />
            ))}
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            No open gigs available. Be the first to create one!
          </div>
        )}
        
        {gigs && gigs.length > 6 && (
          <div className="text-center mt-8">
            <a
              href="/gigs"
              className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-lg inline-block"
            >
              View All Gigs
            </a>
          </div>
        )}
      </section>
    </div>
  );
}
```

---

## 7.13 Configuration App.tsx

**src/App.tsx :**
```typescript
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Navbar } from '@/components/Navbar';
import { HomePage } from '@/pages/Home';

// Créer le client React Query
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="min-h-screen bg-gray-100">
          <Navbar />
          <Routes>
            <Route path="/" element={<HomePage />} />
            {/* Autres routes à ajouter */}
          </Routes>
        </div>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
```

---

## 7.14 Lancer et tester

```bash
# Terminal 1: Lancer la blockchain
cd skillchain
ignite chain serve

# Terminal 2: Lancer le frontend
cd skillchain/frontend
npm run dev

# Ouvrir http://localhost:3000
```

**Test avec données :**
```bash
# Créer des données de test via CLI
skillchaind tx marketplace create-gig \
  "Build a DeFi Dashboard" \
  "Need a React dashboard to display DeFi metrics" \
  500000 "development" 14 \
  --from bob --yes

skillchaind tx marketplace create-gig \
  "Smart Contract Audit" \
  "Security audit for a CosmWasm contract" \
  1000000 "security" 7 \
  --from charlie --yes

# Rafraîchir la page frontend pour voir les gigs
```

---

## Questions de révision

1. **Quelle est la différence entre `StargateClient` et `SigningStargateClient` ?**

2. **Pourquoi utilise-t-on Zustand avec `persist` pour le store wallet ?**

3. **Quel endpoint REST permet de récupérer tous les gigs ?**

4. **Comment React Query gère-t-il le cache des données blockchain ?**

5. **Pourquoi les montants sont-ils stockés en `string` plutôt qu'en `number` ?**

6. **Quelle méthode de Keplr permet d'ajouter une chaîne custom ?**

---

## Récapitulatif des commandes

```bash
# Créer le projet frontend
npm create vite@latest frontend -- --template react-ts

# Installer CosmJS
npm install @cosmjs/stargate @cosmjs/proto-signing

# Lancer le frontend
npm run dev

# Endpoints API REST
GET /skillchain/marketplace/gig           # Tous les gigs
GET /skillchain/marketplace/profile/{addr} # Un profil
GET /cosmos/bank/v1beta1/balances/{addr}  # Soldes
```

---

**Prochaine leçon** : Nous allons implémenter les transactions signées avec Keplr pour créer des profils, des gigs et postuler à des missions.
