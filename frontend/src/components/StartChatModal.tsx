import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { useRoomStore } from '../store/roomStore';
import { useChatStore } from '../store/chatStore';

interface StartChatModalProps {
  isOpen: boolean;
  onClose: () => void;
}

interface SearchUser {
  id: string;
  username: string;
}

export const StartChatModal = ({ isOpen, onClose }: StartChatModalProps) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<SearchUser[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const { selectRoom, addRoom } = useRoomStore();
  const setCurrentRoom = useChatStore((state) => state.setCurrentRoom);
  const loadMessages = useChatStore((state) => state.loadMessages);

  // Debounced search
  const searchUsers = useCallback(async (query: string) => {
    if (query.length < 2) {
      setSearchResults([]);
      return;
    }

    setIsSearching(true);
    setError(null);
    
    try {
      const results = await api.searchUsers(query);
      setSearchResults(results);
    } catch (err: any) {
      setError('Failed to search users');
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  }, []);

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      searchUsers(searchQuery);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery, searchUsers]);

  // Reset state when modal closes
  useEffect(() => {
    if (!isOpen) {
      setSearchQuery('');
      setSearchResults([]);
      setError(null);
    }
  }, [isOpen]);

  const handleStartChat = async (user: SearchUser) => {
    setIsCreating(true);
    setError(null);

    try {
      const room = await api.createDirectChatByUsername(user.username);
      
      // Add room to store
      addRoom(room);
      
      // Select the room
      selectRoom(room.id);
      setCurrentRoom(room.id);
      
      // Load messages
      await loadMessages(room.id);
      
      // Close modal
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to create chat');
    } finally {
      setIsCreating(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <h2 className="text-lg font-semibold">เริ่มแชทใหม่</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Search Input */}
        <div className="p-4">
          <div className="relative">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="ค้นหาชื่อผู้ใช้..."
              className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              autoFocus
            />
            {isSearching && (
              <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                <svg className="animate-spin h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
              </div>
            )}
          </div>
          
          {error && (
            <p className="text-red-500 text-sm mt-2">{error}</p>
          )}
        </div>

        {/* Search Results */}
        <div className="max-h-64 overflow-y-auto">
          {searchQuery.length < 2 ? (
            <div className="p-4 text-center text-gray-500 text-sm">
              พิมพ์อย่างน้อย 2 ตัวอักษรเพื่อค้นหา
            </div>
          ) : searchResults.length === 0 && !isSearching ? (
            <div className="p-4 text-center text-gray-500 text-sm">
              ไม่พบผู้ใช้ที่ตรงกับ "{searchQuery}"
            </div>
          ) : (
            searchResults.map((user) => (
              <button
                key={user.id}
                onClick={() => handleStartChat(user)}
                disabled={isCreating}
                className="w-full flex items-center gap-3 p-4 hover:bg-gray-50 transition-colors border-b last:border-b-0 disabled:opacity-50"
              >
                <div className="w-10 h-10 bg-blue-500 rounded-full flex items-center justify-center text-white font-semibold">
                  {user.username.charAt(0).toUpperCase()}
                </div>
                <div className="flex-1 text-left">
                  <p className="font-medium">{user.username}</p>
                </div>
                <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </button>
            ))
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t bg-gray-50 rounded-b-lg">
          <p className="text-xs text-gray-500 text-center">
            ค้นหาผู้ใช้ด้วยชื่อผู้ใช้เพื่อเริ่มแชท
          </p>
        </div>
      </div>
    </div>
  );
};
