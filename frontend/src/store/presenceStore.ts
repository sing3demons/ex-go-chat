import { create } from 'zustand';

interface PresenceState {
  onlineUsers: Set<string>;
  lastSeen: Record<string, string>; // userId -> timestamp
  typingUsers: Record<string, Set<string>>; // roomId -> Set of userIds
  
  setUserOnline: (userId: string) => void;
  setUserOffline: (userId: string, lastSeen: string) => void;
  isUserOnline: (userId: string) => boolean;
  getLastSeen: (userId: string) => string | undefined;
  setTypingUsers: (roomId: string, userId: string, isTyping: boolean) => void;
  getTypingUsers: (roomId: string) => string[];
}

export const usePresenceStore = create<PresenceState>((set, get) => ({
  onlineUsers: new Set(),
  lastSeen: {},
  typingUsers: {},

  setUserOnline: (userId: string) => {
    set((state) => ({
      onlineUsers: new Set(state.onlineUsers).add(userId),
    }));
  },

  setUserOffline: (userId: string, lastSeen: string) => {
    set((state) => {
      const newOnlineUsers = new Set(state.onlineUsers);
      newOnlineUsers.delete(userId);
      return {
        onlineUsers: newOnlineUsers,
        lastSeen: { ...state.lastSeen, [userId]: lastSeen },
      };
    });
  },

  isUserOnline: (userId: string) => {
    return get().onlineUsers.has(userId);
  },

  getLastSeen: (userId: string) => {
    return get().lastSeen[userId];
  },

  setTypingUsers: (roomId: string, userId: string, isTyping: boolean) => {
    set((state) => {
      const newTypingUsers = { ...state.typingUsers };
      if (!newTypingUsers[roomId]) {
        newTypingUsers[roomId] = new Set();
      }
      const roomTyping = new Set(newTypingUsers[roomId]);
      if (isTyping) {
        roomTyping.add(userId);
      } else {
        roomTyping.delete(userId);
      }
      newTypingUsers[roomId] = roomTyping;
      return { typingUsers: newTypingUsers };
    });
  },

  getTypingUsers: (roomId: string) => {
    const typingSet = get().typingUsers[roomId];
    return typingSet ? Array.from(typingSet) : [];
  },
}));
