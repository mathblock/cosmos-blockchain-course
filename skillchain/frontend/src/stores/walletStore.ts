/* eslint-disable @typescript-eslint/no-explicit-any */
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
    (set) => ({
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
          // Keplr must be installed
          if (!window.keplr) {
            throw new Error('Please install Keplr extension');
          }
          
          // Suggest the chain to Keplr
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
          
          // Activate the chain
          await window.keplr.enable(CHAIN_CONFIG.chainId);
          
          // Get the signer
          const signer = window.keplr.getOfflineSigner(CHAIN_CONFIG.chainId);
          const accounts = await signer.getAccounts();
          
          if (accounts.length === 0) {
            throw new Error('No accounts found');
          }
          
          // Get the signing client
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
      // Do not persist sensitive data
      partialize: (state) => ({
        address: state.address,
        isConnected: state.isConnected,
      }),
    }
  )
);

// Type augmentation for window.keplr
declare global {
  interface Window {
    keplr?: {
      experimentalSuggestChain: (chainInfo: any) => Promise<void>;
      enable: (chainId: string) => Promise<void>;
      getOfflineSigner: (chainId: string) => OfflineSigner;
    };
  }
}