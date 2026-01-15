import { useEffect, useState } from 'react';
import { websocket } from '../services/websocket';
import type { TypingPayload } from '../types';

interface TypingIndicatorProps {
  roomId: string;
}

export const TypingIndicator = ({ roomId }: TypingIndicatorProps) => {
  const [typingUsers, setTypingUsers] = useState<Set<string>>(new Set());

  useEffect(() => {
    // Subscribe to typing events
    const unsubscribe = websocket.on('typing', (msg) => {
      if (msg.roomId !== roomId) return;
      
      const payload = msg.payload as TypingPayload;
      
      setTypingUsers((prev) => {
        const newSet = new Set(prev);
        if (payload.isTyping) {
          newSet.add(payload.userId);
        } else {
          newSet.delete(payload.userId);
        }
        return newSet;
      });
    });

    return () => {
      unsubscribe();
      setTypingUsers(new Set());
    };
  }, [roomId]);

  if (typingUsers.size === 0) {
    return null;
  }

  const getTypingText = () => {
    const count = typingUsers.size;
    if (count === 1) {
      return 'Someone is typing';
    } else if (count === 2) {
      return '2 people are typing';
    } else {
      return `${count} people are typing`;
    }
  };

  return (
    <div className="px-4 py-2 bg-gray-50 border-t animate-slideDown">
      <div className="flex items-center gap-2 text-sm text-gray-600">
        <div className="flex gap-1">
          <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
          <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
          <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
        </div>
        <span className="italic">{getTypingText()}</span>
      </div>
    </div>
  );
};
