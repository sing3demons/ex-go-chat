import { create } from 'zustand';
import type { Message } from '../types';
import { api } from '../services/api';
import { websocket } from '../services/websocket';

interface ChatState {
  messages: Record<string, Message[]>; // roomId -> messages
  currentRoomId: string | null;
  isLoading: boolean;
  
  setCurrentRoom: (roomId: string) => void;
  loadMessages: (roomId: string, limit?: number, offset?: number) => Promise<void>;
  addMessage: (message: Message) => void;
  addOptimisticMessage: (roomId: string, content: string, senderId: string) => string;
  confirmMessage: (tempId: string, serverMessage: Message) => void;
  failMessage: (tempId: string) => void;
  retryMessage: (tempId: string) => void;
  updateMessage: (messageId: string, content: string) => void;
  updateMessageStatus: (messageId: string, userId: string, statusType: 'delivered' | 'read', timestamp: string) => void;
  deleteMessage: (messageId: string) => void;
  sendMessage: (roomId: string, content: string) => void;
  editMessage: (roomId: string, messageId: string, content: string) => void;
  removeMessage: (roomId: string, messageId: string) => void;
  markAsRead: (roomId: string, messageId: string) => void;
  clearMessages: (roomId: string) => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  messages: {},
  currentRoomId: null,
  isLoading: false,

  setCurrentRoom: (roomId: string) => {
    set({ currentRoomId: roomId });
  },

  loadMessages: async (roomId: string, limit = 50, offset = 0) => {
    set({ isLoading: true });
    try {
      const messages = await api.getMessages(roomId, limit, offset);
      set((state) => ({
        messages: {
          ...state.messages,
          [roomId]: offset === 0 ? messages : [...(state.messages[roomId] || []), ...messages],
        },
        isLoading: false,
      }));
    } catch (error) {
      console.error('Failed to load messages:', error);
      set({ isLoading: false });
    }
  },

  addMessage: (message: Message) => {
    set((state) => ({
      messages: {
        ...state.messages,
        [message.roomId]: [...(state.messages[message.roomId] || []), message],
      },
    }));
  },

  // Optimistic update: Add message immediately with pending status
  addOptimisticMessage: (roomId: string, content: string, senderId: string) => {
    const tempId = `temp-${Date.now()}-${Math.random()}`;
    const optimisticMessage: Message = {
      id: tempId,
      tempId,
      roomId,
      senderId,
      content,
      status: {},
      edited: false,
      deleted: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      pending: true,
      failed: false,
    };

    set((state) => ({
      messages: {
        ...state.messages,
        [roomId]: [...(state.messages[roomId] || []), optimisticMessage],
      },
    }));

    // Set timeout to mark as failed if no confirmation after 10 seconds
    setTimeout(() => {
      const currentState = get();
      const roomMessages = currentState.messages[roomId] || [];
      const msg = roomMessages.find(m => m.tempId === tempId);
      
      // If message is still pending after timeout, mark as failed
      if (msg && msg.pending) {
        get().failMessage(tempId);
      }
    }, 10000); // 10 second timeout

    return tempId;
  },

  // Confirm message: Replace optimistic message with server response
  confirmMessage: (tempId: string, serverMessage: Message) => {
    set((state) => {
      const newMessages = { ...state.messages };
      Object.keys(newMessages).forEach((roomId) => {
        newMessages[roomId] = newMessages[roomId].map((msg) =>
          msg.tempId === tempId ? { ...serverMessage, pending: false, failed: false } : msg
        );
      });
      return { messages: newMessages };
    });
  },

  // Mark message as failed
  failMessage: (tempId: string) => {
    set((state) => {
      const newMessages = { ...state.messages };
      Object.keys(newMessages).forEach((roomId) => {
        newMessages[roomId] = newMessages[roomId].map((msg) =>
          msg.tempId === tempId ? { ...msg, pending: false, failed: true } : msg
        );
      });
      return { messages: newMessages };
    });
  },

  // Retry failed message
  retryMessage: (tempId: string) => {
    const state = get();
    let messageToRetry: Message | undefined;
    let targetRoomId: string | undefined;

    // Find the failed message
    for (const roomId of Object.keys(state.messages)) {
      const msg = state.messages[roomId].find((m) => m.tempId === tempId);
      if (msg) {
        messageToRetry = msg;
        targetRoomId = roomId;
        break;
      }
    }

    if (!messageToRetry || !targetRoomId) {
      return;
    }

    // Mark as pending again
    set((state) => {
      const newMessages = { ...state.messages };
      if (targetRoomId) {
        newMessages[targetRoomId] = newMessages[targetRoomId].map((msg) =>
          msg.tempId === tempId ? { ...msg, pending: true, failed: false } : msg
        );
      }
      return { messages: newMessages };
    });

    // Resend the message
    websocket.send({
      type: 'message',
      roomId: targetRoomId,
      payload: { content: messageToRetry.content, tempId },
    });
  },

  updateMessage: (messageId: string, content: string) => {
    set((state) => {
      const newMessages = { ...state.messages };
      Object.keys(newMessages).forEach((roomId) => {
        newMessages[roomId] = newMessages[roomId].map((msg) =>
          msg.id === messageId ? { ...msg, content, edited: true } : msg
        );
      });
      return { messages: newMessages };
    });
  },

  updateMessageStatus: (messageId: string, userId: string, statusType: 'delivered' | 'read', timestamp: string) => {
    set((state) => {
      const newMessages = { ...state.messages };
      Object.keys(newMessages).forEach((roomId) => {
        newMessages[roomId] = newMessages[roomId].map((msg) => {
          if (msg.id === messageId) {
            const newStatus = { ...msg.status };
            if (!newStatus[userId]) {
              newStatus[userId] = { delivered: '', read: '' };
            }
            if (statusType === 'delivered') {
              newStatus[userId].delivered = timestamp;
            } else if (statusType === 'read') {
              newStatus[userId].read = timestamp;
            }
            return { ...msg, status: newStatus };
          }
          return msg;
        });
      });
      return { messages: newMessages };
    });
  },

  deleteMessage: (messageId: string) => {
    set((state) => {
      const newMessages = { ...state.messages };
      Object.keys(newMessages).forEach((roomId) => {
        newMessages[roomId] = newMessages[roomId].map((msg) =>
          msg.id === messageId ? { ...msg, deleted: true } : msg
        );
      });
      return { messages: newMessages };
    });
  },

  sendMessage: (roomId: string, content: string) => {
    // Get current user ID from auth store
    const authStore = JSON.parse(localStorage.getItem('auth-storage') || '{}');
    const senderId = authStore.state?.user?.id;

    if (!senderId) {
      console.error('Cannot send message: User not authenticated');
      return;
    }

    // Add optimistic message
    const tempId = get().addOptimisticMessage(roomId, content, senderId);

    // Send via WebSocket with tempId for tracking
    websocket.send({
      type: 'message',
      roomId,
      payload: { content, tempId },
    });
  },

  editMessage: (roomId: string, messageId: string, content: string) => {
    websocket.send({
      type: 'edit',
      roomId,
      payload: { messageId, content },
    });
  },

  removeMessage: (roomId: string, messageId: string) => {
    websocket.send({
      type: 'delete',
      roomId,
      payload: { messageId },
    });
  },

  markAsRead: (roomId: string, messageId: string) => {
    websocket.send({
      type: 'read',
      roomId,
      payload: { messageId },
    });
  },

  clearMessages: (roomId: string) => {
    set((state) => {
      const newMessages = { ...state.messages };
      delete newMessages[roomId];
      return { messages: newMessages };
    });
  },
}));
