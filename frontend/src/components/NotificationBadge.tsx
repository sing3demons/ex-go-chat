import { useEffect } from 'react';
import { useNotifications } from '../hooks/useNotifications';

interface NotificationBadgeProps {
  className?: string;
}

export const NotificationBadge = ({ className = '' }: NotificationBadgeProps) => {
  const { unreadCount, fetchUnreadCount } = useNotifications();

  useEffect(() => {
    // Fetch unread count on mount
    fetchUnreadCount();
    
    // Poll for updates every 30 seconds
    const interval = setInterval(() => {
      fetchUnreadCount();
    }, 30000);

    return () => clearInterval(interval);
  }, [fetchUnreadCount]);

  if (unreadCount === 0) {
    return null;
  }

  return (
    <div
      className={`
        absolute -top-1 -right-1
        min-w-[20px] h-5
        px-1.5
        bg-red-500 text-white
        text-xs font-bold
        rounded-full
        flex items-center justify-center
        ${className}
      `}
    >
      {unreadCount > 99 ? '99+' : unreadCount}
    </div>
  );
};
