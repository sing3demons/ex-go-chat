import { create } from 'zustand';

interface PresenceState {
  onlineUsers: Record<string, boolean>; // userId -> true if online
  lastSeen: Record<string, string>; // userId -> timestamp
  typingUsers: Record<string, Record<string, boolean>>; // roomId -> userId -> true
  
  setUserOnline: (userId: string) => void;
  setUserOffline: (userId: string, lastSeen: string) => void;
  isUserOnline: (userId: string) => boolean;
  getLastSeen: (userId: string) => string | undefined;
  setTypingUsers: (roomId: string, userId: string, isTyping: boolean) => void;
  getTypingUsers: (roomId: string) => string[];
}

export const usePresenceStore = create<PresenceState>((set, get) => ({
  onlineUsers: {},
  lastSeen: {},
  typingUsers: {},

  setUserOnline: (userId: string) => {
    set((state) => ({
      onlineUsers: { ...state.onlineUsers, [userId]: true },
    }));
  },

  setUserOffline: (userId: string, lastSeen: string) => {
    set((state) => {
      const newOnlineUsers = { ...state.onlineUsers };
      delete newOnlineUsers[userId];
      return {
        onlineUsers: newOnlineUsers,
        lastSeen: { ...state.lastSeen, [userId]: lastSeen },
      };
    });
  },

  isUserOnline: (userId: string) => {
    return !!get().onlineUsers[userId];
  },

  getLastSeen: (userId: string) => {
    return get().lastSeen[userId];
  },

  setTypingUsers: (roomId: string, userId: string, isTyping: boolean) => {
    set((state) => {
      const newTypingUsers = { ...state.typingUsers };
      if (!newTypingUsers[roomId]) {
        newTypingUsers[roomId] = {};
      }
      const roomTyping = { ...newTypingUsers[roomId] };
      if (isTyping) {
        roomTyping[userId] = true;
      } else {
        delete roomTyping[userId];
      }
      newTypingUsers[roomId] = roomTyping;
      return { typingUsers: newTypingUsers };
    });
  },

  getTypingUsers: (roomId: string) => {
    const typingObj = get().typingUsers[roomId];
    return typingObj ? Object.keys(typingObj) : [];
  },
}));
