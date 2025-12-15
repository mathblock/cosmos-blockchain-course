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