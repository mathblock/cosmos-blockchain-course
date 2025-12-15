/* eslint-disable @typescript-eslint/no-explicit-any */
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

const API_BASE = 'http://localhost:1317';

const api = axios.create({
  baseURL: API_BASE,
  timeout: 10000,
});

// ============ QUERIES MARKETPLACE ============

// Module Params
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

// Gigs (missions)
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