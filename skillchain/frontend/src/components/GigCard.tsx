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