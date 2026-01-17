import { useEffect, useRef, useState, useMemo, useCallback } from 'react';
import { MessageItem } from './MessageItem';
import { useChatStore } from '../store/chatStore';
import { useAuthStore } from '../store/authStore';
import { shallow } from 'zustand/shallow';

// Empty array constant to avoid creating new references
const EMPTY_MESSAGES: never[] = [];

interface MessageListProps {
  roomId: string;
}

export const MessageList = ({ roomId }: MessageListProps) => {
  // Use shallow comparison to ensure proper reactivity
  const messages = useChatStore(
    (state) => state.messages[roomId] || EMPTY_MESSAGES,
    shallow
  );
  const loadMessages = useChatStore((state) => state.loadMessages);
  const markAsRead = useChatStore((state) => state.markAsRead);
  const isLoading = useChatStore((state) => state.isLoading);
  const user = useAuthStore((state) => state.user);
  
  // Create a key that changes when messages change to force re-render
  const messagesKey = `${roomId}-${messages.length}-${messages[messages.length - 1]?.id || 'empty'}`;
  
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesContainerRef = useRef<HTMLDivElement>(null);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const observedMessagesRef = useRef<Set<string>>(new Set());

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
  };

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Load initial messages
  useEffect(() => {
    if (roomId) {
      loadMessages(roomId);
    }
  }, [roomId, loadMessages]);

  // Setup Intersection Observer for read receipts
  useEffect(() => {
    if (!user) return;

    // Create observer to detect when messages are visible
    observerRef.current = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const messageId = entry.target.getAttribute('data-message-id');
            const senderId = entry.target.getAttribute('data-sender-id');
            
            if (messageId && senderId && senderId !== user.id) {
              // Only mark as read if it's not our own message and we haven't already marked it
              if (!observedMessagesRef.current.has(messageId)) {
                observedMessagesRef.current.add(messageId);
                markAsRead(roomId, messageId);
              }
            }
          }
        });
      },
      {
        root: messagesContainerRef.current,
        threshold: 0.5, // Message must be 50% visible
      }
    );

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, [roomId, user, markAsRead]);

  // Observe messages for read receipts
  useEffect(() => {
    if (!observerRef.current) return;

    // Observe all message elements
    const messageElements = messagesContainerRef.current?.querySelectorAll('[data-message-id]');
    messageElements?.forEach((element) => {
      observerRef.current?.observe(element);
    });

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, [messages]);

  // Clear observed messages when room changes
  useEffect(() => {
    observedMessagesRef.current.clear();
  }, [roomId]);

  // Handle scroll for infinite loading
  const handleScroll = async () => {
    const container = messagesContainerRef.current;
    if (!container || isLoadingMore) return;

    // Check if scrolled to top
    if (container.scrollTop === 0 && messages.length > 0) {
      setIsLoadingMore(true);
      const oldScrollHeight = container.scrollHeight;
      
      // Load more messages
      await loadMessages(roomId, 50, messages.length);
      
      // Maintain scroll position
      setTimeout(() => {
        if (container) {
          container.scrollTop = container.scrollHeight - oldScrollHeight;
        }
        setIsLoadingMore(false);
      }, 100);
    }
  };

  // Group messages by date
  const groupMessagesByDate = () => {
    const groups: { date: string; messages: typeof messages }[] = [];
    let currentDate = '';
    let currentGroup: typeof messages = [];

    messages.forEach((message) => {
      const messageDate = new Date(message.createdAt).toLocaleDateString();
      
      if (messageDate !== currentDate) {
        if (currentGroup.length > 0) {
          groups.push({ date: currentDate, messages: currentGroup });
        }
        currentDate = messageDate;
        currentGroup = [message];
      } else {
        currentGroup.push(message);
      }
    });

    if (currentGroup.length > 0) {
      groups.push({ date: currentDate, messages: currentGroup });
    }

    return groups;
  };

  const messageGroups = groupMessagesByDate();

  // Format date for separator
  const formatDateSeparator = (dateStr: string) => {
    const date = new Date(dateStr);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    if (date.toLocaleDateString() === today.toLocaleDateString()) {
      return 'Today';
    } else if (date.toLocaleDateString() === yesterday.toLocaleDateString()) {
      return 'Yesterday';
    } else {
      return date.toLocaleDateString('en-US', { 
        weekday: 'long', 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric' 
      });
    }
  };

  if (isLoading && messages.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-gray-500">Loading messages...</div>
      </div>
    );
  }

  return (
    <div 
      key={messagesKey}
      ref={messagesContainerRef}
      onScroll={handleScroll}
      className="flex-1 overflow-y-auto p-2 sm:p-3 bg-gray-50 scroll-smooth min-h-0"
    >
      {isLoadingMore && (
        <div className="text-center text-gray-500 text-xs sm:text-sm py-2">
          Loading more messages...
        </div>
      )}
      
      {messages.length === 0 ? (
        <div className="flex items-center justify-center h-full text-gray-500 p-4">
          <div className="text-center">
            <svg className="w-12 h-12 sm:w-16 sm:h-16 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            <p className="text-sm sm:text-base font-medium">No messages yet</p>
            <p className="text-xs sm:text-sm mt-1">Start the conversation!</p>
          </div>
        </div>
      ) : (
        <>
          {messageGroups.map((group, groupIndex) => (
            <div key={`group-${groupIndex}-${group.date}`}>
              {/* Date Separator */}
              <div className="flex items-center justify-center my-3 sm:my-4">
                <div className="bg-gray-200 text-gray-600 text-xs px-2 sm:px-3 py-1 rounded-full">
                  {formatDateSeparator(group.date)}
                </div>
              </div>
              
              {/* Messages */}
              {group.messages.map((message, messageIndex) => (
                <MessageItem 
                  key={`${groupIndex}-${message.id}-${messageIndex}`} 
                  message={message}
                  data-message-id={message.id}
                  data-sender-id={message.senderId}
                />
              ))}
            </div>
          ))}
          <div ref={messagesEndRef} />
        </>
      )}
    </div>
  );
};
