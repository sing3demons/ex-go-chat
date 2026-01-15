import { create } from 'zustand';
import type { Room } from '../types';
import { api } from '../services/api';

interface RoomState {
  rooms: Room[];
  selectedRoomId: string | null;
  isLoading: boolean;
  
  loadRooms: () => Promise<void>;
  selectRoom: (roomId: string) => void;
  createGroupRoom: (name: string, members: string[]) => Promise<Room>;
  createDirectRoom: (memberId: string) => Promise<Room>;
  addRoom: (room: Room) => void;
  updateRoom: (roomId: string, updates: Partial<Room>) => void;
}

export const useRoomStore = create<RoomState>((set) => ({
  rooms: [],
  selectedRoomId: null,
  isLoading: false,

  loadRooms: async () => {
    set({ isLoading: true });
    try {
      const rooms = await api.getRooms();
      set({ rooms, isLoading: false });
    } catch (error) {
      console.error('Failed to load rooms:', error);
      set({ isLoading: false });
    }
  },

  selectRoom: (roomId: string) => {
    set({ selectedRoomId: roomId });
  },

  createGroupRoom: async (name: string, members: string[]) => {
    const room = await api.createGroupRoom(name, members);
    set((state) => ({
      rooms: [...state.rooms, room],
    }));
    return room;
  },

  createDirectRoom: async (memberId: string) => {
    const room = await api.createDirectRoom(memberId);
    set((state) => ({
      rooms: [...state.rooms, room],
    }));
    return room;
  },

  addRoom: (room: Room) => {
    set((state) => ({
      rooms: [...state.rooms, room],
    }));
  },

  updateRoom: (roomId: string, updates: Partial<Room>) => {
    set((state) => ({
      rooms: state.rooms.map((room) =>
        room.id === roomId ? { ...room, ...updates } : room
      ),
    }));
  },
}));
