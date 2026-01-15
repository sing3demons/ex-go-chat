import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
  RoomList, 
  ChatWindow, 
  CreateRoomModal, 
  RoomSettings,
  NotificationBadge,
  NotificationList,
  Avatar,
  PageLoader,
  ErrorMessage
} from '../components';
import { StartChatModal } from '../components/StartChatModal';
import { useAuthStore } from '../store/authStore';
import { useRoomStore } from '../store/roomStore';
import { useWebSocket } from '../hooks/useWebSocket';
import { websocket } from '../services/websocket';

export const ChatPage = () => {
  const { user, token, logout } = useAuthStore();
  const { selectedRoomId, rooms } = useRoomStore();
  const navigate = useNavigate();
  const { isConnected, connect } = useWebSocket();
  
  const [isCreateRoomModalOpen, setIsCreateRoomModalOpen] = useState(false);
  const [isStartChatModalOpen, setIsStartChatModalOpen] = useState(false);
  const [isRoomSettingsOpen, setIsRoomSettingsOpen] = useState(false);
  const [isNotificationsOpen, setIsNotificationsOpen] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isInitializing, setIsInitializing] = useState(true);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const [isRetrying, setIsRetrying] = useState(false);

  const selectedRoom = rooms.find(r => r.id === selectedRoomId);

  // Manual retry connection
  const handleRetryConnection = async () => {
    if (!token || isRetrying) return;
    
    setIsRetrying(true);
    setConnectionError(null);
    
    try {
      await connect(token);
      console.log('Reconnected successfully');
    } catch (error: any) {
      console.error('Retry failed:', error);
      setConnectionError(error.message || 'Failed to reconnect');
    } finally {
      setIsRetrying(false);
    }
  };

  // Redirect if not authenticated
  useEffect(() => {
    if (!user || !token) {
      navigate('/login');
    }
  }, [user, token, navigate]);

  // Connect WebSocket on mount if not connected and load rooms
  useEffect(() => {
    let mounted = true;

    const initializeChat = async () => {
      if (!token || !mounted) return;
      
      try {
        // Setup error callback
        websocket.setConnectionErrorCallback((error) => {
          if (mounted) {
            setConnectionError(error);
            setTimeout(() => {
              if (mounted) setConnectionError(null);
            }, 5000);
          }
        });

        // Connect WebSocket if not already connected
        if (!isConnected) {
          await connect(token);
          console.log('WebSocket connected successfully');
        }
        
        // Load rooms after ensuring WebSocket connection
        await useRoomStore.getState().loadRooms();
        console.log('Rooms loaded successfully');
      } catch (error: any) {
        console.error('Failed to initialize chat:', error);
        if (mounted) {
          setConnectionError(error.message || 'Failed to initialize chat');
        }
      } finally {
        if (mounted) {
          setIsInitializing(false);
        }
      }
    };

    initializeChat();

    return () => {
      mounted = false;
    };
  }, []); // Only run once on mount

  const handleLogout = () => {
    websocket.disconnect();
    logout();
    navigate('/login');
  };

  if (!user) {
    return null;
  }

  // Show loading screen while initializing
  if (isInitializing) {
    return <PageLoader text="Loading chat..." />;
  }

  return (
    <div className="h-screen flex flex-col bg-gray-50">
      {/* Connection Error Banner */}
      {connectionError && (
        <div className="fixed top-0 left-0 right-0 z-50 p-4">
          <div className="max-w-4xl mx-auto">
            <ErrorMessage 
              message={connectionError} 
              onDismiss={() => setConnectionError(null)}
              type="error"
            />
            {!isConnected && (
              <div className="mt-2 flex justify-center">
                <button
                  onClick={handleRetryConnection}
                  disabled={isRetrying}
                  className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
                >
                  {isRetrying ? (
                    <>
                      <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      <span>Retrying...</span>
                    </>
                  ) : (
                    <>
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                      </svg>
                      <span>Retry Connection</span>
                    </>
                  )}
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Header */}
      <header className="bg-white border-b shadow-sm z-10">
        <div className="px-2 sm:px-4 py-2 sm:py-3 flex justify-between items-center">
          {/* Left: Logo & Title */}
          <div className="flex items-center gap-2 sm:gap-3 min-w-0">
            <button
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              className="lg:hidden p-2 hover:bg-gray-100 active:bg-gray-200 rounded-lg touch-manipulation transition-all duration-200 ease-in-out transform hover:scale-110"
            >
              <svg className="w-5 h-5 sm:w-6 sm:h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            
            <div className="flex items-center gap-1 sm:gap-2 min-w-0">
              <div className="w-7 h-7 sm:w-8 sm:h-8 bg-blue-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg className="w-4 h-4 sm:w-5 sm:h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </div>
              <h1 className="text-base sm:text-xl font-bold text-gray-800 hidden sm:block truncate">Chat App</h1>
            </div>
          </div>

          {/* Right: Actions & User */}
          <div className="flex items-center gap-1 sm:gap-2">
            {/* Connection Status */}
            <div className="hidden md:flex items-center gap-2 px-2 sm:px-3 py-1 rounded-full bg-gray-100">
              <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
              <span className="text-xs text-gray-600">
                {isConnected ? 'Connected' : 'Disconnected'}
              </span>
            </div>

            {/* Start Chat Button */}
            <button
              onClick={() => setIsStartChatModalOpen(true)}
              className="p-2 hover:bg-gray-100 active:bg-gray-200 rounded-lg transition-all duration-200 ease-in-out touch-manipulation transform hover:scale-110"
              title="เริ่มแชทใหม่"
            >
              <svg className="w-5 h-5 sm:w-6 sm:h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </button>

            {/* Create Room Button */}
            <button
              onClick={() => setIsCreateRoomModalOpen(true)}
              className="p-2 hover:bg-gray-100 active:bg-gray-200 rounded-lg transition-all duration-200 ease-in-out touch-manipulation transform hover:scale-110 hover:rotate-90"
              title="Create new room"
            >
              <svg className="w-5 h-5 sm:w-6 sm:h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            </button>

            {/* Notifications */}
            <button
              onClick={() => setIsNotificationsOpen(true)}
              className="relative p-2 hover:bg-gray-100 active:bg-gray-200 rounded-lg transition-all duration-200 ease-in-out touch-manipulation transform hover:scale-110"
              title="Notifications"
            >
              <svg className="w-5 h-5 sm:w-6 sm:h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
              <NotificationBadge />
            </button>

            {/* Room Settings (if room selected) */}
            {selectedRoom && (
              <button
                onClick={() => setIsRoomSettingsOpen(true)}
                className="p-2 hover:bg-gray-100 active:bg-gray-200 rounded-lg transition-colors hidden md:block touch-manipulation"
                title="Room settings"
              >
                <svg className="w-5 h-5 sm:w-6 sm:h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </button>
            )}

            {/* User Menu */}
            <div className="flex items-center gap-1 sm:gap-2 pl-1 sm:pl-2 border-l">
              <Avatar userId={user.id} username={user.username} size="sm" />
              <span className="text-xs sm:text-sm font-medium text-gray-700 hidden lg:block truncate max-w-[100px]">
                {user.username}
              </span>
              <button
                onClick={handleLogout}
                className="p-2 hover:bg-red-50 active:bg-red-100 rounded-lg transition-colors text-red-600 touch-manipulation"
                title="Logout"
              >
                <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar - Room List */}
        <aside className={`
          ${isMobileMenuOpen ? 'block' : 'hidden'} lg:block
          fixed lg:relative
          inset-0 lg:inset-auto
          z-20
          lg:z-0
          bg-gray-100 lg:bg-transparent
          ${isMobileMenuOpen ? 'animate-slideInLeft' : ''}
        `}>
          <div className="h-full lg:h-auto" onClick={() => setIsMobileMenuOpen(false)}>
            <div onClick={(e) => e.stopPropagation()}>
              <RoomList />
            </div>
          </div>
        </aside>

        {/* Mobile Overlay */}
        {isMobileMenuOpen && (
          <div
            className="fixed inset-0 bg-black bg-opacity-50 z-10 lg:hidden animate-fadeIn"
            onClick={() => setIsMobileMenuOpen(false)}
          />
        )}

        {/* Chat Area */}
        <main className="flex-1 flex flex-col min-w-0">
          <ChatWindow />
        </main>
      </div>

      {/* Modals */}
      <StartChatModal
        isOpen={isStartChatModalOpen}
        onClose={() => setIsStartChatModalOpen(false)}
      />

      <CreateRoomModal
        isOpen={isCreateRoomModalOpen}
        onClose={() => setIsCreateRoomModalOpen(false)}
      />

      {selectedRoom && (
        <RoomSettings
          room={selectedRoom}
          isOpen={isRoomSettingsOpen}
          onClose={() => setIsRoomSettingsOpen(false)}
        />
      )}

      <NotificationList
        isOpen={isNotificationsOpen}
        onClose={() => setIsNotificationsOpen(false)}
      />
    </div>
  );
};
