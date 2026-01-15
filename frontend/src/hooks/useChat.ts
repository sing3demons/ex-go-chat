import { useState, useCallback, useRef, useEffect } from 'react';
import { useChatStore } from '../store/chatStore';
import { websocket } from '../services/websocket';

export const useChat = (roomId: string | null) => {
  const [isTyping, setIsTyping] = useState(false);
  const typingTimeoutRef = useRef<number | null>(null);
  
  const {
    messages,
    isLoading,
    sendMessage: sendMessageStore,
    editMessage: editMessageStore,
    removeMessage: removeMessageStore,
    markAsRead: markAsReadStore,
    loadMessages,
  } = useChatStore();

  const roomMessages = roomId ? messages[roomId] || [] : [];

  // Send message
  const sendMessage = useCallback((content: string) => {
    if (!roomId || !content.trim()) return;
    sendMessageStore(roomId, content);
  }, [roomId, sendMessageStore]);

  // Edit message
  const editMessage = useCallback((messageId: string, content: string) => {
    if (!roomId || !content.trim()) return;
    editMessageStore(roomId, messageId, content);
  }, [roomId, editMessageStore]);

  // Delete message
  const deleteMessage = useCallback((messageId: string) => {
    if (!roomId) return;
    removeMessageStore(roomId, messageId);
  }, [roomId, removeMessageStore]);

  // Mark message as read
  const markAsRead = useCallback((messageId: string) => {
    if (!roomId) return;
    markAsReadStore(roomId, messageId);
  }, [roomId, markAsReadStore]);

  // Start typing indicator
  const startTyping = useCallback(() => {
    if (!roomId) return;
    
    // Send typing start event
    if (!isTyping) {
      websocket.send({
        type: 'typing',
        roomId,
        payload: { isTyping: true },
      });
      setIsTyping(true);
    }

    // Clear existing timeout
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    // Set timeout to stop typing after 3 seconds of inactivity
    typingTimeoutRef.current = setTimeout(() => {
      stopTyping();
    }, 3000);
  }, [roomId, isTyping]);

  // Stop typing indicator
  const stopTyping = useCallback(() => {
    if (!roomId || !isTyping) return;
    
    websocket.send({
      type: 'typing',
      roomId,
      payload: { isTyping: false },
    });
    setIsTyping(false);

    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
      typingTimeoutRef.current = null;
    }
  }, [roomId, isTyping]);

  // Load more messages (pagination)
  const loadMoreMessages = useCallback(async (offset: number) => {
    if (!roomId) return;
    await loadMessages(roomId, 50, offset);
  }, [roomId, loadMessages]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
      // Stop typing when component unmounts
      if (isTyping && roomId) {
        websocket.send({
          type: 'typing',
          roomId,
          payload: { isTyping: false },
        });
      }
    };
  }, [isTyping, roomId]);

  return {
    messages: roomMessages,
    isLoading,
    sendMessage,
    editMessage,
    deleteMessage,
    markAsRead,
    startTyping,
    stopTyping,
    loadMoreMessages,
  };
};
