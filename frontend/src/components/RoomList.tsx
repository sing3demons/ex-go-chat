import { useEffect } from 'react';
import { useRoomStore } from '../store/roomStore';
import { useChatStore } from '../store/chatStore';
import { usePresence } from '../hooks/usePresence';
import { LoadingSpinner } from './LoadingSpinner';
import type { Room } from '../types';

interface RoomItemProps {
  room: Room;
  isSelected: boolean;
  onSelect: () => void;
}

const RoomItem = ({ room, isSelected, onSelect }: RoomItemProps) => {
  const messages = useChatStore((state) => state.messages[room.id] || []);
  const lastMessage = messages[messages.length - 1];
  const { areUsersOnline } = usePresence();
  
  // For direct rooms, check if the other user is online
  const isOnline = room.type === 'direct' && areUsersOnline(room.members);
  
  // Calculate unread count (simplified - would need proper tracking)
  const unreadCount = 0; // TODO: Implement proper unread tracking

  return (
    <button
      onClick={onSelect}
      className={`w-full text-left p-3 sm:p-4 rounded-lg mb-2 hover:bg-gray-200 active:bg-gray-300 transition-all duration-200 ease-in-out touch-manipulation transform hover:scale-[1.02] ${
        isSelected ? 'bg-blue-100 border-2 border-blue-500 shadow-md' : 'bg-white hover:shadow-sm'
      }`}
    >
      <div className="flex items-center justify-between mb-1">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <div className="font-semibold truncate text-sm sm:text-base">
            {room.name || `Room ${room.id.slice(0, 8)}`}
          </div>
          {isOnline && (
            <div className="w-2 h-2 bg-green-500 rounded-full flex-shrink-0" title="Online" />
          )}
        </div>
        {unreadCount > 0 && (
          <span className="bg-blue-500 text-white text-xs px-2 py-1 rounded-full flex-shrink-0 ml-2">
            {unreadCount}
          </span>
        )}
      </div>
      
      {lastMessage && (
        <div className="text-xs sm:text-sm text-gray-500 truncate">
          {lastMessage.deleted ? (
            <span className="italic">Message deleted</span>
          ) : (
            lastMessage.content
          )}
        </div>
      )}
      
      <div className="text-xs text-gray-400 mt-1">
        {room.type === 'direct' ? 'Direct' : `Group • ${room.members.length} members`}
      </div>
    </button>
  );
};

export const RoomList = () => {
  const { rooms, loadRooms, selectRoom, selectedRoomId, isLoading } = useRoomStore();
  const setCurrentRoom = useChatStore((state) => state.setCurrentRoom);
  const loadMessages = useChatStore((state) => state.loadMessages);

  useEffect(() => {
    loadRooms();
  }, [loadRooms]);

  const handleSelectRoom = async (roomId: string) => {
    selectRoom(roomId);
    setCurrentRoom(roomId);
    // Load messages for the selected room
    await loadMessages(roomId);
  };

  return (
    <div className="w-full lg:w-80 xl:w-96 bg-gray-100 border-r overflow-y-auto flex flex-col h-full">
      <div className="p-3 sm:p-4 border-b bg-white shadow-sm">
        <h2 className="text-lg sm:text-xl font-bold">Chats</h2>
        <p className="text-xs sm:text-sm text-gray-500">{rooms.length} conversation{rooms.length !== 1 ? 's' : ''}</p>
      </div>
      
      <div className="flex-1 overflow-y-auto p-2">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <LoadingSpinner size="md" color="blue" text="Loading rooms..." />
          </div>
        ) : rooms.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center px-4">
            <div className="text-gray-400 mb-2">
              <svg className="w-12 h-12 sm:w-16 sm:h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </div>
            <p className="text-sm sm:text-base text-gray-500 font-medium">No conversations yet</p>
            <p className="text-xs sm:text-sm text-gray-400 mt-1">Start a new chat to begin messaging</p>
          </div>
        ) : (
          rooms.map((room) => (
            <RoomItem
              key={room.id}
              room={room}
              isSelected={selectedRoomId === room.id}
              onSelect={() => handleSelectRoom(room.id)}
            />
          ))
        )}
      </div>
    </div>
  );
};
