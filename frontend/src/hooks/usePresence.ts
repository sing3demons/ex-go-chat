import { usePresenceStore } from '../store/presenceStore';

export const usePresence = (userId?: string) => {
  const onlineUsers = usePresenceStore((state) => state.onlineUsers);
  const lastSeenMap = usePresenceStore((state) => state.lastSeen);
  const isUserOnline = usePresenceStore((state) => state.isUserOnline);
  const getLastSeen = usePresenceStore((state) => state.getLastSeen);

  // Check if specific user is online
  const isOnline = userId ? !!onlineUsers[userId] : false;
  
  // Get last seen for specific user
  const userLastSeen = userId ? lastSeenMap[userId] : undefined;

  // Format last seen timestamp
  const formatLastSeen = (timestamp?: string): string => {
    if (!timestamp) return 'Never';
    
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    
    return date.toLocaleDateString();
  };

  // Get online status text
  const getStatusText = (): string => {
    if (isOnline) return 'Online';
    if (userLastSeen) return formatLastSeen(userLastSeen);
    return 'Offline';
  };

  // Get all online user IDs as array
  const onlineUserIds = Object.keys(onlineUsers);

  // Check if multiple users are online
  const areUsersOnline = (userIds: string[]): boolean => {
    return userIds.some(id => !!onlineUsers[id]);
  };

  return {
    isOnline,
    lastSeen: userLastSeen,
    statusText: getStatusText(),
    formattedLastSeen: formatLastSeen(userLastSeen),
    onlineUserIds,
    onlineCount: onlineUserIds.length,
    areUsersOnline,
    isUserOnline,
    getLastSeen,
  };
};
