import { usePresence } from '../hooks/usePresence';

interface OnlineStatusProps {
  userId: string;
  showText?: boolean;
  showDot?: boolean;
  className?: string;
}

export const OnlineStatus = ({ 
  userId, 
  showText = true, 
  showDot = true,
  className = '' 
}: OnlineStatusProps) => {
  const { isOnline, statusText, formattedLastSeen } = usePresence(userId);

  if (showDot && !showText) {
    // Just show the dot
    return (
      <div
        className={`
          w-2 h-2 rounded-full
          ${isOnline ? 'bg-green-500' : 'bg-gray-400'}
          ${className}
        `}
        title={statusText}
      />
    );
  }

  if (showText && !showDot) {
    // Just show the text
    return (
      <span className={`text-sm ${isOnline ? 'text-green-600' : 'text-gray-500'} ${className}`}>
        {statusText}
      </span>
    );
  }

  // Show both dot and text
  return (
    <div className={`flex items-center gap-2 ${className}`}>
      <div
        className={`
          w-2 h-2 rounded-full
          ${isOnline ? 'bg-green-500' : 'bg-gray-400'}
        `}
      />
      <span className={`text-sm ${isOnline ? 'text-green-600' : 'text-gray-500'}`}>
        {isOnline ? 'Online' : formattedLastSeen}
      </span>
    </div>
  );
};
