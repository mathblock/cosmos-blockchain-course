import { StargateClient, SigningStargateClient, GasPrice } from '@cosmjs/stargate';
import { type OfflineSigner } from '@cosmjs/proto-signing';

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
      coinMinimalDenom: 'skill',
      coinDecimals: 6,
    },
    {
      coinDenom: 'SKILL',
      coinMinimalDenom: 'skill',
      coinDecimals: 6,
    },
  ],
  feeCurrencies: [
    {
      coinDenom: 'SKILL',
      coinMinimalDenom: 'skill',
      coinDecimals: 6,
      gasPriceStep: {
        low: 0.01,
        average: 0.025,
        high: 0.04,
      },
    },
  ],
  stakeCurrency: {
    coinDenom: 'SKILL',
    coinMinimalDenom: 'skill',
    coinDecimals: 6,
  },
};

// Client in read-only mode (no wallet needed)
let queryClient: StargateClient | null = null;

export async function getQueryClient(): Promise<StargateClient> {
  if (!queryClient) {
    queryClient = await StargateClient.connect(CHAIN_CONFIG.rpc);
  }
  return queryClient;
}

// Client with signing capabilities (wallet needed)
export async function getSigningClient(signer: OfflineSigner): Promise<SigningStargateClient> {
  return SigningStargateClient.connectWithSigner(
    CHAIN_CONFIG.rpc,
    signer,
    {
      gasPrice: GasPrice.fromString('0.025uskill'),
    }
  );
}

// Util for formatting amounts
export function formatAmount(amount: string, decimals: number = 6): string {
  const value = parseInt(amount) / Math.pow(10, decimals);
  return value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: decimals,
  });
}

// Util for converting to micro units
export function toMicroUnits(amount: number, decimals: number = 6): string {
  return Math.floor(amount * Math.pow(10, decimals)).toString();
}

// Util for shortening addresses
export function shortenAddress(address: string, chars: number = 8): string {
  if (!address) return '';
  return `${address.slice(0, chars + 5)}...${address.slice(-chars)}`;
}