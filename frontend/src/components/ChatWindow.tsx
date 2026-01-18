import { MessageList } from './MessageList';
import { MessageInput } from './MessageInput';
import { TypingIndicator } from './TypingIndicator';
import { useRoomStore } from '../store/roomStore';

export const ChatWindow = () => {
  const { selectedRoomId, rooms } = useRoomStore();
  const selectedRoom = rooms.find(r => r.id === selectedRoomId);

  if (!selectedRoomId || !selectedRoom) {
    return (
      <div className="flex-1 flex items-center justify-center bg-gray-50 p-4">
        <div className="text-center max-w-md">
          <div className="text-gray-400 mb-4">
            <svg className="w-16 h-16 sm:w-20 sm:h-20 md:w-24 md:h-24 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <h3 className="text-lg sm:text-xl font-semibold text-gray-700 mb-2">
            Select a conversation
          </h3>
          <p className="text-sm sm:text-base text-gray-500">
            Choose a chat from the list to start messaging
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col bg-white min-h-0">
      {/* Chat Header - reduced padding */}
      <div className="border-b p-2 sm:p-3 bg-white shadow-sm shrink-0">
        <div className="flex items-center justify-between">
          <div className="min-w-0 flex-1">
            <h2 className="text-base sm:text-lg font-semibold truncate">
              {selectedRoom.name || `Room ${selectedRoom.id.slice(0, 8)}`}
            </h2>
            <p className="text-xs sm:text-sm text-gray-500 truncate">
              {selectedRoom.type === 'direct' 
                ? 'Direct message' 
                : `${selectedRoom.members.length} members`}
            </p>
          </div>
          
          {/* Room actions */}
          <div className="flex gap-1 sm:gap-2 ml-2">
            <button 
              className="p-2 hover:bg-gray-100 active:bg-gray-200 rounded-lg transition-colors touch-manipulation"
              title="Room info"
            >
              <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {/* Messages Area */}
      <MessageList roomId={selectedRoomId} />

      {/* Typing Indicator */}
      <TypingIndicator roomId={selectedRoomId} />

      {/* Message Input */}
      <MessageInput roomId={selectedRoomId} />
    </div>
  );
};
