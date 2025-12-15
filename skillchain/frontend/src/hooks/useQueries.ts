import { useQuery, useQueryClient } from '@tanstack/react-query';
import * as api from '@/services/api';
import { useWalletStore } from '@/stores/walletStore';

export function useParams() {
  return useQuery({
    queryKey: ['marketplace', 'params'],
    queryFn: api.getParams,
    staleTime: 60000, // 1 minute
  });
}

export function useProfile(address: string | null) {
  return useQuery({
    queryKey: ['marketplace', 'profile', address],
    queryFn: () => address ? api.getProfile(address) : null,
    enabled: !!address,
  });
}

export function useMyProfile() {
  const { address } = useWalletStore();
  return useProfile(address);
}

export function useGigs() {
  return useQuery({
    queryKey: ['marketplace', 'gigs'],
    queryFn: api.getAllGigs,
    staleTime: 10000, // 10 secondes
  });
}


export function useOpenGigs() {
  return useQuery({
    queryKey: ['marketplace', 'gigs', 'open'],
    queryFn: api.getOpenGigs,
    staleTime: 10000,
  });
}

export function useGig(id: string | null) {
  return useQuery({
    queryKey: ['marketplace', 'gig', id],
    queryFn: () => id ? api.getGig(id) : null,
    enabled: !!id,
  });
}

export function useApplicationsByGig(gigId: string | null) {
  return useQuery({
    queryKey: ['marketplace', 'applications', 'gig', gigId],
    queryFn: () => gigId ? api.getApplicationsByGig(gigId) : [],
    enabled: !!gigId,
  });
}

export function useMyApplications() {
  const { address } = useWalletStore();
  return useQuery({
    queryKey: ['marketplace', 'applications', 'freelancer', address],
    queryFn: () => address ? api.getApplicationsByFreelancer(address) : [],
    enabled: !!address,
  });
}

export function useMyContracts() {
  const { address } = useWalletStore();
  return useQuery({
    queryKey: ['marketplace', 'contracts', 'user', address],
    queryFn: () => address ? api.getContractsByUser(address) : [],
    enabled: !!address,
  });
}

export function useBalance(address: string | null, denom: string = 'skill') {
  return useQuery({
    queryKey: ['bank', 'balance', address, denom],
    queryFn: () => address ? api.getBalanceByDenom(address, denom) : '0',
    enabled: !!address,
    staleTime: 5000, // 5 secondes
  });
}

export function useMyBalance(denom: string = 'skill') {
  const { address } = useWalletStore();
  return useBalance(address, denom);
}

export function useEscrowBalance() {
  return useQuery({
    queryKey: ['marketplace', 'escrow'],
    queryFn: api.getEscrowBalance,
    staleTime: 10000,
  });
}

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