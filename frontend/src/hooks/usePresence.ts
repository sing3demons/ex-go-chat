import { usePresenceStore } from '../store/presenceStore';

export const usePresence = (userId?: string) => {
  const { onlineUsers, isUserOnline, getLastSeen } = usePresenceStore();

  // Check if specific user is online
  const isOnline = userId ? isUserOnline(userId) : false;
  
  // Get last seen for specific user
  const userLastSeen = userId ? getLastSeen(userId) : undefined;

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
  const onlineUserIds = Array.from(onlineUsers);

  // Check if multiple users are online
  const areUsersOnline = (userIds: string[]): boolean => {
    return userIds.some(id => isUserOnline(id));
  };

  return {
    isOnline,
    lastSeen: userLastSeen,
    statusText: getStatusText(),
    formattedLastSeen: formatLastSeen(userLastSeen),
    onlineUserIds,
    onlineCount: onlineUsers.size,
    areUsersOnline,
    isUserOnline,
    getLastSeen,
  };
};
