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