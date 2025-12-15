// Types for the SkillChain frontend application (protobuf def)

export interface Profile {
  owner: string;
  name: string;
  bio: string;
  skills: string[];
  hourlyRate: string;  // uint64 to string to avoid precision issues
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

export interface Coin {
  denom: string;
  amount: string;
}

export interface Balance {
  balances: Coin[];
}