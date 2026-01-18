import { useEffect, useRef } from 'react';
import { MessageItem } from './MessageItem';
import { useChatStore } from '../store/chatStore';
import type { Message } from '../types';

const EMPTY_MESSAGES: Message[] = [];

interface MessageListProps {
  roomId: string;
}

export const MessageList = ({ roomId }: MessageListProps) => {
  const loadMessages = useChatStore((state) => state.loadMessages);
  const isLoading = useChatStore((state) => state.isLoading);
  const updateCount = useChatStore((state) => state.updateCounter);
  const messages = useChatStore((state) => state.messages[roomId] || EMPTY_MESSAGES);
  
  // Debug: Log when component renders
  console.log('🔄 MessageList render - roomId:', roomId, 'messages.length:', messages.length, 'updateCount:', updateCount);
  
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesContainerRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
  };

  // Load initial messages
  useEffect(() => {
    if (!roomId) return;

    const fetchMessages = async () => {
      try {
        await loadMessages(roomId);
      } catch (error) {
        console.error('Failed to load messages:', error);
      }
    };

    fetchMessages();
  }, [roomId, loadMessages]);

  // Auto-scroll when messages change
  useEffect(() => {
    scrollToBottom();
  }, [messages, updateCount]);

  // Group messages by date
  const groupMessagesByDate = () => {
    const groups: { date: string; messages: Message[] }[] = [];
    let currentDate = '';
    let currentGroup: Message[] = [];

    messages.forEach((message: Message) => {
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
      return 'วันนี้';
    } else if (date.toLocaleDateString() === yesterday.toLocaleDateString()) {
      return 'เมื่อวาน';
    } else {
      return date.toLocaleDateString('th-TH', { 
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
        <div className="text-gray-500">กำลังโหลดข้อความ...</div>
      </div>
    );
  }

  return (
    <div 
      ref={messagesContainerRef}
      className="flex-1 overflow-y-auto p-2 sm:p-3 bg-gray-50 scroll-smooth min-h-0"
    >
      {messages.length === 0 ? (
        <div className="flex items-center justify-center h-full text-gray-500 p-4">
          <div className="text-center">
            <svg className="w-12 h-12 sm:w-16 sm:h-16 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            <p className="text-sm sm:text-base font-medium">ยังไม่มีข้อความ</p>
            <p className="text-xs sm:text-sm mt-1">เริ่มการสนทนากันเลย!</p>
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