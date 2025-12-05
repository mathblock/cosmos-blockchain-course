# Leçon 8 : Intégration Keplr et Transactions Signées

## Objectifs
- Connecter Keplr wallet à l'application React
- Créer et signer des transactions pour le module marketplace
- Implémenter les messages custom avec CosmJS
- Gérer les erreurs et confirmations de transactions

## Prérequis
- Leçon 7 complétée
- Extension Keplr installée dans le navigateur
- Frontend React fonctionnel

---

## 8.1 Comprendre la signature des transactions

Dans Cosmos SDK, une transaction suit ce flux :

```
┌─────────────────────────────────────────────────────────────────────┐
│                    FLUX D'UNE TRANSACTION                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. Construction         2. Signature           3. Broadcast        │
│  ┌─────────────┐        ┌─────────────┐        ┌─────────────┐     │
│  │  Message    │   ──►  │   Keplr     │   ──►  │   Node      │     │
│  │  + Fee      │        │   signe     │        │   RPC       │     │
│  │  + Memo     │        │   avec clé  │        │             │     │
│  └─────────────┘        └─────────────┘        └──────┬──────┘     │
│                                                        │            │
│                                                        ▼            │
│                                                 ┌─────────────┐     │
│                                                 │  Mempool    │     │
│                                                 │  → Block    │     │
│                                                 │  → Execute  │     │
│                                                 └─────────────┘     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 8.2 Types des messages

**src/types/messages.ts :**
```typescript
// Types des messages pour le module marketplace
// Correspondent aux définitions Protobuf dans proto/skillchain/marketplace/tx.proto

export interface MsgCreateProfile {
  creator: string;
  name: string;
  bio: string;
  skills: string[];
  hourlyRate: string;
}

export interface MsgUpdateProfile {
  creator: string;
  name: string;
  bio: string;
  skills: string[];
  hourlyRate: string;
}

export interface MsgCreateGig {
  creator: string;
  title: string;
  description: string;
  price: string;
  category: string;
  deliveryDays: string;
}

export interface MsgApplyToGig {
  creator: string;
  gigId: string;
  coverLetter: string;
  proposedPrice: string;
  proposedDays: string;
}

export interface MsgAcceptApplication {
  creator: string;
  applicationId: string;
}

export interface MsgRejectApplication {
  creator: string;
  applicationId: string;
}

export interface MsgDeliverContract {
  creator: string;
  contractId: string;
  deliveryNote: string;
}

export interface MsgCompleteContract {
  creator: string;
  contractId: string;
}

export interface MsgOpenDispute {
  creator: string;
  contractId: string;
  reason: string;
  evidence: string;
}

export interface MsgVoteDispute {
  creator: string;
  disputeId: string;
  vote: string;
}

// Type URLs pour l'encodage des messages Amino
export const MSG_TYPE_URLS = {
  CreateProfile: '/skillchain.marketplace.MsgCreateProfile',
  UpdateProfile: '/skillchain.marketplace.MsgUpdateProfile',
  CreateGig: '/skillchain.marketplace.MsgCreateGig',
  ApplyToGig: '/skillchain.marketplace.MsgApplyToGig',
  AcceptApplication: '/skillchain.marketplace.MsgAcceptApplication',
  RejectApplication: '/skillchain.marketplace.MsgRejectApplication',
  DeliverContract: '/skillchain.marketplace.MsgDeliverContract',
  CompleteContract: '/skillchain.marketplace.MsgCompleteContract',
  OpenDispute: '/skillchain.marketplace.MsgOpenDispute',
  VoteDispute: '/skillchain.marketplace.MsgVoteDispute',
} as const;
```

---

## 8.3 Service de transactions

**src/services/transactions.ts :**
```typescript
import { DeliverTxResponse, StdFee } from '@cosmjs/stargate';
import { toMicroUnits } from './cosmjs';
import { MSG_TYPE_URLS } from '@/types/messages';
import { useWalletStore } from '@/stores/walletStore';

// Fee par défaut
const DEFAULT_FEE: StdFee = {
  amount: [{ denom: 'uskill', amount: '5000' }],
  gas: '200000',
};

// Fee pour transactions complexes (escrow, etc.)
const HIGH_GAS_FEE: StdFee = {
  amount: [{ denom: 'uskill', amount: '10000' }],
  gas: '400000',
};

// Type pour le résultat de transaction
export interface TxResult {
  success: boolean;
  txHash?: string;
  error?: string;
  rawLog?: string;
}

// Helper pour exécuter une transaction
async function executeTx(
  typeUrl: string,
  value: Record<string, any>,
  fee: StdFee = DEFAULT_FEE,
  memo: string = ''
): Promise<TxResult> {
  const { signingClient, address } = useWalletStore.getState();
  
  if (!signingClient || !address) {
    return { success: false, error: 'Wallet not connected' };
  }
  
  try {
    const msg = {
      typeUrl,
      value: { ...value, creator: address },
    };
    
    const result: DeliverTxResponse = await signingClient.signAndBroadcast(
      address,
      [msg],
      fee,
      memo
    );
    
    if (result.code !== 0) {
      return {
        success: false,
        txHash: result.transactionHash,
        error: `Transaction failed with code ${result.code}`,
        rawLog: result.rawLog,
      };
    }
    
    return {
      success: true,
      txHash: result.transactionHash,
    };
  } catch (error: any) {
    return {
      success: false,
      error: error.message || 'Transaction failed',
    };
  }
}

// ============ PROFILE TRANSACTIONS ============

export async function createProfile(
  name: string,
  bio: string,
  skills: string[],
  hourlyRate: number
): Promise<TxResult> {
  return executeTx(MSG_TYPE_URLS.CreateProfile, {
    name,
    bio,
    skills,
    hourlyRate: toMicroUnits(hourlyRate),
  });
}

export async function updateProfile(
  name: string,
  bio: string,
  skills: string[],
  hourlyRate: number
): Promise<TxResult> {
  return executeTx(MSG_TYPE_URLS.UpdateProfile, {
    name,
    bio,
    skills,
    hourlyRate: toMicroUnits(hourlyRate),
  });
}

// ============ GIG TRANSACTIONS ============

export async function createGig(
  title: string,
  description: string,
  price: number,
  category: string,
  deliveryDays: number
): Promise<TxResult> {
  return executeTx(MSG_TYPE_URLS.CreateGig, {
    title,
    description,
    price: toMicroUnits(price),
    category,
    deliveryDays: deliveryDays.toString(),
  });
}

// ============ APPLICATION TRANSACTIONS ============

export async function applyToGig(
  gigId: string,
  coverLetter: string,
  proposedPrice: number,
  proposedDays: number
): Promise<TxResult> {
  return executeTx(MSG_TYPE_URLS.ApplyToGig, {
    gigId,
    coverLetter,
    proposedPrice: toMicroUnits(proposedPrice),
    proposedDays: proposedDays.toString(),
  });
}

export async function acceptApplication(applicationId: string): Promise<TxResult> {
  // Accepter une application verrouille les fonds en escrow
  // Utiliser plus de gas
  return executeTx(
    MSG_TYPE_URLS.AcceptApplication,
    { applicationId },
    HIGH_GAS_FEE
  );
}

export async function rejectApplication(applicationId: string): Promise<TxResult> {
  return executeTx(MSG_TYPE_URLS.RejectApplication, { applicationId });
}

// ============ CONTRACT TRANSACTIONS ============

export async function deliverContract(
  contractId: string,
  deliveryNote: string
): Promise<TxResult> {
  return executeTx(MSG_TYPE_URLS.DeliverContract, {
    contractId,
    deliveryNote,
  });
}

export async function completeContract(contractId: string): Promise<TxResult> {
  // Compléter un contrat libère les fonds de l'escrow
  return executeTx(
    MSG_TYPE_URLS.CompleteContract,
    { contractId },
    HIGH_GAS_FEE
  );
}

// ============ DISPUTE TRANSACTIONS ============

export async function openDispute(
  contractId: string,
  reason: string,
  evidence: string
): Promise<TxResult> {
  return executeTx(MSG_TYPE_URLS.OpenDispute, {
    contractId,
    reason,
    evidence,
  });
}

export async function voteDispute(
  disputeId: string,
  vote: 'client' | 'freelancer'
): Promise<TxResult> {
  return executeTx(MSG_TYPE_URLS.VoteDispute, {
    disputeId,
    vote,
  });
}
```

---

## 8.4 Hook useTransaction

**src/hooks/useTransaction.ts :**
```typescript
import { useState, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { TxResult } from '@/services/transactions';

interface UseTransactionOptions {
  onSuccess?: (result: TxResult) => void;
  onError?: (error: string) => void;
  invalidateKeys?: string[][];
}

interface UseTransactionReturn<T extends (...args: any[]) => Promise<TxResult>> {
  execute: (...args: Parameters<T>) => Promise<TxResult>;
  isLoading: boolean;
  error: string | null;
  txHash: string | null;
  reset: () => void;
}

export function useTransaction<T extends (...args: any[]) => Promise<TxResult>>(
  txFn: T,
  options: UseTransactionOptions = {}
): UseTransactionReturn<T> {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [txHash, setTxHash] = useState<string | null>(null);
  
  const queryClient = useQueryClient();
  
  const execute = useCallback(
    async (...args: Parameters<T>): Promise<TxResult> => {
      setIsLoading(true);
      setError(null);
      setTxHash(null);
      
      try {
        const result = await txFn(...args);
        
        if (result.success) {
          setTxHash(result.txHash || null);
          
          // Invalider les caches spécifiés
          if (options.invalidateKeys) {
            for (const key of options.invalidateKeys) {
              queryClient.invalidateQueries({ queryKey: key });
            }
          }
          
          options.onSuccess?.(result);
        } else {
          setError(result.error || 'Transaction failed');
          options.onError?.(result.error || 'Transaction failed');
        }
        
        return result;
      } catch (err: any) {
        const errorMsg = err.message || 'Unknown error';
        setError(errorMsg);
        options.onError?.(errorMsg);
        return { success: false, error: errorMsg };
      } finally {
        setIsLoading(false);
      }
    },
    [txFn, options, queryClient]
  );
  
  const reset = useCallback(() => {
    setError(null);
    setTxHash(null);
  }, []);
  
  return { execute, isLoading, error, txHash, reset };
}
```

---

## 8.5 Composant CreateProfileForm

**src/components/CreateProfileForm.tsx :**
```typescript
import { useState } from 'react';
import { useTransaction } from '@/hooks/useTransaction';
import { createProfile } from '@/services/transactions';
import { useWalletStore } from '@/stores/walletStore';

interface CreateProfileFormProps {
  onSuccess?: () => void;
}

export function CreateProfileForm({ onSuccess }: CreateProfileFormProps) {
  const { isConnected } = useWalletStore();
  
  // Form state
  const [name, setName] = useState('');
  const [bio, setBio] = useState('');
  const [skillsInput, setSkillsInput] = useState('');
  const [hourlyRate, setHourlyRate] = useState('');
  
  // Transaction hook
  const { execute, isLoading, error, txHash } = useTransaction(createProfile, {
    invalidateKeys: [['marketplace', 'profile']],
    onSuccess: () => {
      // Reset form
      setName('');
      setBio('');
      setSkillsInput('');
      setHourlyRate('');
      onSuccess?.();
    },
  });
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Parser les skills (séparés par des virgules)
    const skills = skillsInput
      .split(',')
      .map(s => s.trim())
      .filter(s => s.length > 0);
    
    // Convertir le taux horaire en nombre
    const rate = parseFloat(hourlyRate);
    
    if (skills.length === 0) {
      alert('Please enter at least one skill');
      return;
    }
    
    if (isNaN(rate) || rate <= 0) {
      alert('Please enter a valid hourly rate');
      return;
    }
    
    await execute(name, bio, skills, rate);
  };
  
  if (!isConnected) {
    return (
      <div className="bg-yellow-100 border border-yellow-400 text-yellow-700 p-4 rounded">
        Please connect your wallet to create a profile.
      </div>
    );
  }
  
  return (
    <form onSubmit={handleSubmit} className="space-y-6 max-w-lg">
      <h2 className="text-2xl font-bold">Create Your Profile</h2>
      
      {/* Name */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Display Name
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
          minLength={3}
          maxLength={50}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="John Developer"
        />
      </div>
      
      {/* Bio */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Bio
        </label>
        <textarea
          value={bio}
          onChange={(e) => setBio(e.target.value)}
          required
          rows={4}
          maxLength={500}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Tell clients about your experience and expertise..."
        />
      </div>
      
      {/* Skills */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Skills (comma-separated)
        </label>
        <input
          type="text"
          value={skillsInput}
          onChange={(e) => setSkillsInput(e.target.value)}
          required
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="React, TypeScript, Cosmos SDK, Go"
        />
        <p className="text-sm text-gray-500 mt-1">
          Enter your skills separated by commas
        </p>
      </div>
      
      {/* Hourly Rate */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Hourly Rate (SKILL)
        </label>
        <input
          type="number"
          value={hourlyRate}
          onChange={(e) => setHourlyRate(e.target.value)}
          required
          min="0.001"
          step="0.001"
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="50"
        />
        <p className="text-sm text-gray-500 mt-1">
          Minimum: 0.001 SKILL (1000 uskill)
        </p>
      </div>
      
      {/* Error Message */}
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 p-3 rounded">
          {error}
        </div>
      )}
      
      {/* Success Message */}
      {txHash && (
        <div className="bg-green-100 border border-green-400 text-green-700 p-3 rounded">
          Profile created successfully!
          <br />
          <a
            href={`https://explorer.skillchain.local/tx/${txHash}`}
            target="_blank"
            rel="noopener noreferrer"
            className="underline"
          >
            View transaction
          </a>
        </div>
      )}
      
      {/* Submit Button */}
      <button
        type="submit"
        disabled={isLoading}
        className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-medium py-2 px-4 rounded-md transition-colors"
      >
        {isLoading ? 'Creating Profile...' : 'Create Profile'}
      </button>
    </form>
  );
}
```

---

## 8.6 Composant CreateGigForm

**src/components/CreateGigForm.tsx :**
```typescript
import { useState } from 'react';
import { useTransaction } from '@/hooks/useTransaction';
import { createGig } from '@/services/transactions';
import { useWalletStore } from '@/stores/walletStore';

const CATEGORIES = [
  'development',
  'design',
  'writing',
  'marketing',
  'consulting',
  'security',
  'other',
];

export function CreateGigForm({ onSuccess }: { onSuccess?: () => void }) {
  const { isConnected } = useWalletStore();
  
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [price, setPrice] = useState('');
  const [category, setCategory] = useState(CATEGORIES[0]);
  const [deliveryDays, setDeliveryDays] = useState('');
  
  const { execute, isLoading, error, txHash } = useTransaction(createGig, {
    invalidateKeys: [['marketplace', 'gigs']],
    onSuccess: () => {
      setTitle('');
      setDescription('');
      setPrice('');
      setDeliveryDays('');
      onSuccess?.();
    },
  });
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const priceNum = parseFloat(price);
    const daysNum = parseInt(deliveryDays);
    
    if (isNaN(priceNum) || priceNum < 0.01) {
      alert('Price must be at least 0.01 SKILL');
      return;
    }
    
    if (isNaN(daysNum) || daysNum < 1 || daysNum > 365) {
      alert('Delivery days must be between 1 and 365');
      return;
    }
    
    await execute(title, description, priceNum, category, daysNum);
  };
  
  if (!isConnected) {
    return (
      <div className="bg-yellow-100 border border-yellow-400 text-yellow-700 p-4 rounded">
        Please connect your wallet to create a gig.
      </div>
    );
  }
  
  return (
    <form onSubmit={handleSubmit} className="space-y-6 max-w-lg">
      <h2 className="text-2xl font-bold">Post a New Gig</h2>
      
      {/* Title */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Title
        </label>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
          minLength={10}
          maxLength={100}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Build a React Dashboard for DeFi Analytics"
        />
      </div>
      
      {/* Description */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Description
        </label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          required
          rows={6}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Describe your project requirements in detail..."
        />
      </div>
      
      {/* Category */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Category
        </label>
        <select
          value={category}
          onChange={(e) => setCategory(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {CATEGORIES.map((cat) => (
            <option key={cat} value={cat}>
              {cat.charAt(0).toUpperCase() + cat.slice(1)}
            </option>
          ))}
        </select>
      </div>
      
      {/* Price */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Budget (SKILL)
        </label>
        <input
          type="number"
          value={price}
          onChange={(e) => setPrice(e.target.value)}
          required
          min="0.01"
          step="0.01"
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="500"
        />
      </div>
      
      {/* Delivery Days */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Expected Delivery (days)
        </label>
        <input
          type="number"
          value={deliveryDays}
          onChange={(e) => setDeliveryDays(e.target.value)}
          required
          min="1"
          max="365"
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="14"
        />
      </div>
      
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 p-3 rounded">
          {error}
        </div>
      )}
      
      {txHash && (
        <div className="bg-green-100 border border-green-400 text-green-700 p-3 rounded">
          Gig posted successfully! TX: {txHash.slice(0, 16)}...
        </div>
      )}
      
      <button
        type="submit"
        disabled={isLoading}
        className="w-full bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white font-medium py-2 px-4 rounded-md"
      >
        {isLoading ? 'Posting Gig...' : 'Post Gig'}
      </button>
    </form>
  );
}
```

---

## 8.7 Composant ApplyToGigModal

**src/components/ApplyToGigModal.tsx :**
```typescript
import { useState } from 'react';
import { useTransaction } from '@/hooks/useTransaction';
import { applyToGig } from '@/services/transactions';
import { formatAmount } from '@/services/cosmjs';
import type { Gig } from '@/types/skillchain';

interface Props {
  gig: Gig;
  isOpen: boolean;
  onClose: () => void;
}

export function ApplyToGigModal({ gig, isOpen, onClose }: Props) {
  const [coverLetter, setCoverLetter] = useState('');
  const [proposedPrice, setProposedPrice] = useState(
    (parseInt(gig.price) / 1000000).toString()
  );
  const [proposedDays, setProposedDays] = useState(gig.deliveryDays);
  
  const { execute, isLoading, error, txHash, reset } = useTransaction(applyToGig, {
    invalidateKeys: [
      ['marketplace', 'applications'],
      ['marketplace', 'gig', gig.id],
    ],
    onSuccess: () => {
      setTimeout(() => {
        onClose();
        reset();
      }, 2000);
    },
  });
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await execute(
      gig.id,
      coverLetter,
      parseFloat(proposedPrice),
      parseInt(proposedDays)
    );
  };
  
  if (!isOpen) return null;
  
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-xl font-bold">Apply to Gig</h3>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700">
            ✕
          </button>
        </div>
        
        {/* Gig Info */}
        <div className="bg-gray-100 p-3 rounded mb-4">
          <h4 className="font-semibold">{gig.title}</h4>
          <p className="text-sm text-gray-600">
            Budget: {formatAmount(gig.price)} SKILL • {gig.deliveryDays} days
          </p>
        </div>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Cover Letter */}
          <div>
            <label className="block text-sm font-medium mb-1">
              Cover Letter
            </label>
            <textarea
              value={coverLetter}
              onChange={(e) => setCoverLetter(e.target.value)}
              required
              rows={4}
              className="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
              placeholder="Explain why you're the best fit for this gig..."
            />
          </div>
          
          {/* Proposed Price */}
          <div>
            <label className="block text-sm font-medium mb-1">
              Your Price (SKILL)
            </label>
            <input
              type="number"
              value={proposedPrice}
              onChange={(e) => setProposedPrice(e.target.value)}
              required
              min="0.01"
              step="0.01"
              className="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
            />
          </div>
          
          {/* Proposed Days */}
          <div>
            <label className="block text-sm font-medium mb-1">
              Delivery Time (days)
            </label>
            <input
              type="number"
              value={proposedDays}
              onChange={(e) => setProposedDays(e.target.value)}
              required
              min="1"
              max="365"
              className="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
            />
          </div>
          
          {error && (
            <div className="bg-red-100 text-red-700 p-2 rounded text-sm">
              {error}
            </div>
          )}
          
          {txHash && (
            <div className="bg-green-100 text-green-700 p-2 rounded text-sm">
              Application submitted successfully!
            </div>
          )}
          
          <div className="flex space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 bg-gray-200 hover:bg-gray-300 py-2 rounded-md"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className="flex-1 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white py-2 rounded-md"
            >
              {isLoading ? 'Submitting...' : 'Apply'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
```

---

## 8.8 Page de gestion des contrats

**src/pages/MyContracts.tsx :**
```typescript
import { useState } from 'react';
import { useMyContracts, useGig } from '@/hooks/useQueries';
import { useTransaction } from '@/hooks/useTransaction';
import { deliverContract, completeContract, openDispute } from '@/services/transactions';
import { formatAmount, shortenAddress } from '@/services/cosmjs';
import { useWalletStore } from '@/stores/walletStore';
import type { Contract } from '@/types/skillchain';

function ContractCard({ contract }: { contract: Contract }) {
  const { address } = useWalletStore();
  const { data: gig } = useGig(contract.gigId);
  const [deliveryNote, setDeliveryNote] = useState('');
  const [showDisputeForm, setShowDisputeForm] = useState(false);
  const [disputeReason, setDisputeReason] = useState('');
  const [disputeEvidence, setDisputeEvidence] = useState('');
  
  const isClient = contract.client === address;
  const isFreelancer = contract.freelancer === address;
  
  const deliverTx = useTransaction(deliverContract, {
    invalidateKeys: [['marketplace', 'contracts']],
  });
  
  const completeTx = useTransaction(completeContract, {
    invalidateKeys: [['marketplace', 'contracts'], ['bank']],
  });
  
  const disputeTx = useTransaction(openDispute, {
    invalidateKeys: [['marketplace', 'contracts'], ['marketplace', 'disputes']],
  });
  
  const handleDeliver = () => {
    if (deliveryNote.trim()) {
      deliverTx.execute(contract.id, deliveryNote);
    }
  };
  
  const handleComplete = () => {
    if (confirm('Confirm payment release to freelancer?')) {
      completeTx.execute(contract.id);
    }
  };
  
  const handleDispute = () => {
    if (disputeReason && disputeEvidence) {
      disputeTx.execute(contract.id, disputeReason, disputeEvidence);
    }
  };
  
  const statusColors: Record<string, string> = {
    active: 'bg-blue-500',
    delivered: 'bg-yellow-500',
    completed: 'bg-green-500',
    disputed: 'bg-red-500',
  };
  
  const deadline = new Date(parseInt(contract.deliveryDeadline) * 1000);
  const isOverdue = deadline < new Date() && contract.status === 'active';
  
  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      {/* Header */}
      <div className="flex justify-between items-start mb-4">
        <div>
          <h3 className="text-lg font-semibold">{gig?.title || `Gig #${contract.gigId}`}</h3>
          <p className="text-sm text-gray-500">Contract #{contract.id}</p>
        </div>
        <span className={`${statusColors[contract.status]} text-white text-xs px-2 py-1 rounded`}>
          {contract.status.toUpperCase()}
        </span>
      </div>
      
      {/* Details */}
      <div className="grid grid-cols-2 gap-4 mb-4 text-sm">
        <div>
          <span className="text-gray-500">Client:</span>
          <span className="ml-2">{shortenAddress(contract.client)}</span>
          {isClient && <span className="ml-1 text-blue-600">(You)</span>}
        </div>
        <div>
          <span className="text-gray-500">Freelancer:</span>
          <span className="ml-2">{shortenAddress(contract.freelancer)}</span>
          {isFreelancer && <span className="ml-1 text-blue-600">(You)</span>}
        </div>
        <div>
          <span className="text-gray-500">Price:</span>
          <span className="ml-2 font-medium">{formatAmount(contract.price)} SKILL</span>
        </div>
        <div>
          <span className="text-gray-500">Deadline:</span>
          <span className={`ml-2 ${isOverdue ? 'text-red-600' : ''}`}>
            {deadline.toLocaleDateString()}
          </span>
        </div>
      </div>
      
      {/* Actions based on status and role */}
      <div className="border-t pt-4 space-y-3">
        {/* Freelancer: Deliver */}
        {isFreelancer && contract.status === 'active' && (
          <div className="space-y-2">
            <textarea
              value={deliveryNote}
              onChange={(e) => setDeliveryNote(e.target.value)}
              placeholder="Describe your delivery..."
              className="w-full px-3 py-2 border rounded-md text-sm"
              rows={2}
            />
            <button
              onClick={handleDeliver}
              disabled={deliverTx.isLoading || !deliveryNote.trim()}
              className="w-full bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white py-2 rounded"
            >
              {deliverTx.isLoading ? 'Submitting...' : 'Mark as Delivered'}
            </button>
          </div>
        )}
        
        {/* Client: Complete (after delivery) */}
        {isClient && contract.status === 'delivered' && (
          <button
            onClick={handleComplete}
            disabled={completeTx.isLoading}
            className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white py-2 rounded"
          >
            {completeTx.isLoading ? 'Processing...' : 'Approve & Release Payment'}
          </button>
        )}
        
        {/* Dispute button (both parties, when active or delivered) */}
        {(contract.status === 'active' || contract.status === 'delivered') && (
          <>
            <button
              onClick={() => setShowDisputeForm(!showDisputeForm)}
              className="w-full bg-red-100 hover:bg-red-200 text-red-700 py-2 rounded"
            >
              Open Dispute
            </button>
            
            {showDisputeForm && (
              <div className="space-y-2 p-3 bg-red-50 rounded">
                <input
                  value={disputeReason}
                  onChange={(e) => setDisputeReason(e.target.value)}
                  placeholder="Reason for dispute"
                  className="w-full px-3 py-2 border rounded text-sm"
                />
                <textarea
                  value={disputeEvidence}
                  onChange={(e) => setDisputeEvidence(e.target.value)}
                  placeholder="Evidence (links, descriptions...)"
                  className="w-full px-3 py-2 border rounded text-sm"
                  rows={2}
                />
                <button
                  onClick={handleDispute}
                  disabled={disputeTx.isLoading}
                  className="w-full bg-red-600 hover:bg-red-700 text-white py-2 rounded text-sm"
                >
                  {disputeTx.isLoading ? 'Opening...' : 'Submit Dispute'}
                </button>
              </div>
            )}
          </>
        )}
        
        {/* Completed status */}
        {contract.status === 'completed' && (
          <div className="text-center text-green-600 font-medium">
            ✓ Contract completed successfully
          </div>
        )}
      </div>
      
      {/* Error messages */}
      {(deliverTx.error || completeTx.error || disputeTx.error) && (
        <div className="mt-3 bg-red-100 text-red-700 p-2 rounded text-sm">
          {deliverTx.error || completeTx.error || disputeTx.error}
        </div>
      )}
    </div>
  );
}

export function MyContractsPage() {
  const { isConnected } = useWalletStore();
  const { data: contracts, isLoading } = useMyContracts();
  
  if (!isConnected) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center text-gray-500">
          Please connect your wallet to view your contracts.
        </div>
      </div>
    );
  }
  
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-8">My Contracts</h1>
      
      {isLoading ? (
        <div className="text-center py-8">Loading...</div>
      ) : contracts && contracts.length > 0 ? (
        <div className="grid md:grid-cols-2 gap-6">
          {contracts.map((contract) => (
            <ContractCard key={contract.id} contract={contract} />
          ))}
        </div>
      ) : (
        <div className="text-center py-8 text-gray-500">
          You don't have any contracts yet.
        </div>
      )}
    </div>
  );
}
```

---

## 8.9 Tests complets des transactions

```bash
# Terminal 1: Blockchain
cd skillchain
ignite chain serve --reset-once

# Terminal 2: Frontend
cd skillchain/frontend
npm run dev
```

**Scénario de test manuel :**

1. **Ouvrir http://localhost:3000**
2. **Connecter Keplr** - Cliquer "Connect Wallet", approuver dans Keplr
3. **Créer un profil** - Remplir le formulaire, signer la transaction
4. **Créer un gig** - Remplir le formulaire, signer
5. **Avec un second compte**, postuler au gig
6. **Accepter l'application** - Les fonds sont verrouillés
7. **Marquer comme livré** - Le freelancer soumet
8. **Compléter le contrat** - Le client libère les fonds

---

## 8.10 Gestion des erreurs Keplr

**src/utils/errors.ts :**
```typescript
// Messages d'erreur utilisateur-friendly
export function parseKeplrError(error: any): string {
  const message = error?.message || error?.toString() || 'Unknown error';
  
  // Erreurs Keplr courantes
  if (message.includes('Request rejected')) {
    return 'Transaction cancelled by user';
  }
  
  if (message.includes('insufficient funds')) {
    return 'Insufficient funds to complete this transaction';
  }
  
  if (message.includes('account sequence mismatch')) {
    return 'Transaction conflict. Please try again.';
  }
  
  if (message.includes('out of gas')) {
    return 'Transaction ran out of gas. Please try again.';
  }
  
  // Erreurs métier SkillChain
  if (message.includes('profile already exists')) {
    return 'You already have a profile';
  }
  
  if (message.includes('profile not found')) {
    return 'Profile not found. Create one first.';
  }
  
  if (message.includes('gig is not open')) {
    return 'This gig is no longer accepting applications';
  }
  
  if (message.includes('unauthorized')) {
    return 'You are not authorized to perform this action';
  }
  
  return message;
}
```

---

## Questions de révision

1. **Quelle méthode CosmJS permet de signer et broadcaster une transaction en une seule opération ?**

2. **Pourquoi les montants sont-ils convertis en micro-unités avant l'envoi ?**

3. **Comment invalide-t-on le cache React Query après une transaction réussie ?**

4. **Quelle est la différence entre `DEFAULT_FEE` et `HIGH_GAS_FEE` ?**

5. **Comment Keplr identifie-t-il la chaîne SkillChain lors de la connexion ?**

6. **Pourquoi vérifie-t-on `isConnected` avant d'afficher les formulaires ?**

---

## Récapitulatif du code

```typescript
// Transaction simple
const result = await signingClient.signAndBroadcast(
  address,
  [{ typeUrl: '/skillchain.marketplace.MsgCreateProfile', value: {...} }],
  fee,
  memo
);

// Hook réutilisable
const { execute, isLoading, error } = useTransaction(createProfile, {
  invalidateKeys: [['marketplace', 'profile']],
  onSuccess: () => console.log('Done!'),
});

// Appel
await execute(name, bio, skills, hourlyRate);
```

---

**Prochaine leçon** : Nous allons ajouter les paiements IBC cross-chain pour permettre aux utilisateurs de payer avec des tokens d'autres chaînes Cosmos.
