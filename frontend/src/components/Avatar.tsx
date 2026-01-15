interface AvatarProps {
  userId: string;
  username?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  online?: boolean;
  className?: string;
}

export const Avatar = ({ 
  userId, 
  username, 
  size = 'md', 
  online = false,
  className = '' 
}: AvatarProps) => {
  // Size mappings
  const sizeClasses = {
    sm: 'w-8 h-8 text-xs',
    md: 'w-10 h-10 text-sm',
    lg: 'w-12 h-12 text-base',
    xl: 'w-16 h-16 text-lg',
  };

  const onlineIndicatorSizes = {
    sm: 'w-2 h-2',
    md: 'w-2.5 h-2.5',
    lg: 'w-3 h-3',
    xl: 'w-4 h-4',
  };

  // Generate initials from username or userId
  const getInitials = () => {
    const name = username || userId;
    if (!name) return '??';
    
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
  };

  // Generate consistent color from userId
  const getBackgroundColor = () => {
    const colors = [
      'bg-blue-500',
      'bg-green-500',
      'bg-yellow-500',
      'bg-red-500',
      'bg-purple-500',
      'bg-pink-500',
      'bg-indigo-500',
      'bg-teal-500',
    ];
    
    // Simple hash function
    let hash = 0;
    for (let i = 0; i < userId.length; i++) {
      hash = userId.charCodeAt(i) + ((hash << 5) - hash);
    }
    
    return colors[Math.abs(hash) % colors.length];
  };

  return (
    <div className={`relative inline-block ${className}`}>
      <div
        className={`
          ${sizeClasses[size]}
          ${getBackgroundColor()}
          rounded-full
          flex items-center justify-center
          text-white font-semibold
          select-none
        `}
        title={username || userId}
      >
        {getInitials()}
      </div>
      
      {online && (
        <div
          className={`
            absolute bottom-0 right-0
            ${onlineIndicatorSizes[size]}
            bg-green-500
            border-2 border-white
            rounded-full
          `}
          title="Online"
        />
      )}
    </div>
  );
};
