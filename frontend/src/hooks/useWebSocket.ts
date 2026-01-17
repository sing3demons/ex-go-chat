import { useEffect, useCallback, useState } from 'react';
import { websocket } from '../services/websocket';
import { useChatStore } from '../store/chatStore';
import { usePresenceStore } from '../store/presenceStore';
import { useRoomStore } from '../store/roomStore';
import type { 
  ChatMessagePayload, 
  PresencePayload,
  EditPayload,
  DeletePayload,
  StatusPayload,
  TypingPayload,
  RoomCreatedPayload
} from '../types';

export const useWebSocket = () => {
  const [isConnected, setIsConnected] = useState(false);
  
  const addMessage = useChatStore((state) => state.addMessage);
  const confirmMessage = useChatStore((state) => state.confirmMessage);
  const failMessage = useChatStore((state) => state.failMessage);
  const updateMessage = useChatStore((state) => state.updateMessage);
  const deleteMessage = useChatStore((state) => state.deleteMessage);
  const updateMessageStatus = useChatStore((state) => state.updateMessageStatus);
  const setUserOnline = usePresenceStore((state) => state.setUserOnline);
  const setUserOffline = usePresenceStore((state) => state.setUserOffline);
  const setTypingUsers = usePresenceStore((state) => state.setTypingUsers);
  const rooms = useRoomStore((state) => state.rooms);
  const addRoom = useRoomStore((state) => state.addRoom);

  // Update connection status
  useEffect(() => {
    const updateConnectionStatus = () => {
      setIsConnected(websocket.isConnected());
    };

    // Check initial status
    updateConnectionStatus();

    // Set up interval to check connection status
    const interval = setInterval(updateConnectionStatus, 1000);

    return () => clearInterval(interval);
  }, []);

  // Subscribe to room events when rooms change
  const subscribeToRooms = useCallback(() => {
    // The backend already subscribes users to their rooms on connection
    // This is only needed for newly created rooms
    if (rooms.length > 0 && websocket.isConnected()) {
      console.log(`Rooms available: ${rooms.length}`);
    }
  }, [rooms]);

  useEffect(() => {
    // No need to subscribe - backend handles this on connection
    subscribeToRooms();
  }, [subscribeToRooms]);

  useEffect(() => {
    // Handle incoming messages
    const unsubMessage = websocket.on('message', (msg) => {
      console.log('🎯 Received message in useWebSocket:', msg);
      const payload = msg.payload as ChatMessagePayload & { tempId?: string };
      
      // Check if this is a confirmation of an optimistic message
      if (payload.tempId) {
        console.log('✅ Confirming optimistic message:', payload.tempId);
        // This is a server confirmation - replace the optimistic message
        confirmMessage(payload.tempId, {
          id: payload.messageId,
          roomId: msg.roomId!,
          senderId: payload.senderId,
          content: payload.content,
          status: {},
          edited: false,
          deleted: false,
          createdAt: payload.timestamp,
          updatedAt: payload.timestamp,
        });
      } else {
        console.log('📨 Adding new message from another user:', payload);
        // This is a new message from another user
        addMessage({
          id: payload.messageId,
          roomId: msg.roomId!,
          senderId: payload.senderId,
          content: payload.content,
          status: {},
          edited: false,
          deleted: false,
          createdAt: payload.timestamp,
          updatedAt: payload.timestamp,
        });
      }

      // Send delivery acknowledgment automatically with a small delay
      setTimeout(() => {
        websocket.send({
          type: 'delivered',
          roomId: msg.roomId,
          payload: { messageId: payload.messageId }
        });
      }, 100); // 100ms delay to ensure message is saved to database
    });

    // Handle message edits
    const unsubEdit = websocket.on('edit', (msg) => {
      const payload = msg.payload as EditPayload;
      updateMessage(payload.messageId, payload.content);
    });

    // Handle message deletes
    const unsubDelete = websocket.on('delete', (msg) => {
      const payload = msg.payload as DeletePayload;
      deleteMessage(payload.messageId);
    });

    // Handle delivery status updates
    const unsubDelivered = websocket.on('delivered', (msg) => {
      const payload = msg.payload as StatusPayload;
      updateMessageStatus(payload.messageId, payload.userId, 'delivered', payload.timestamp);
    });

    // Handle read status updates
    const unsubRead = websocket.on('read', (msg) => {
      const payload = msg.payload as StatusPayload;
      updateMessageStatus(payload.messageId, payload.userId, 'read', payload.timestamp);
    });

    // Handle typing indicators
    const unsubTyping = websocket.on('typing', (msg) => {
      const payload = msg.payload as TypingPayload;
      if (msg.roomId) {
        setTypingUsers(msg.roomId, payload.userId, payload.isTyping);
      }
    });

    // Handle presence updates
    const unsubPresence = websocket.on('presence', (msg) => {
      const payload = msg.payload as PresencePayload;
      if (payload.online) {
        setUserOnline(payload.userId);
      } else {
        setUserOffline(payload.userId, payload.lastSeen || '');
      }
    });

    // Handle errors
    const unsubError = websocket.on('error', (msg) => {
      console.error('WebSocket error:', msg.payload);
      
      // If error contains tempId, mark that message as failed
      if (msg.payload.tempId) {
        failMessage(msg.payload.tempId);
      }
    });

    // Handle room creation notifications
    const unsubRoomCreated = websocket.on('room_created', (msg) => {
      const payload = msg.payload as RoomCreatedPayload;
      
      // Add the new room to the room store
      const newRoom = {
        id: payload.roomId,
        type: payload.roomType,
        name: payload.name || '',
        members: payload.members,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      
      addRoom(newRoom);
      console.log('New room created and added:', payload.roomId);
    });

    return () => {
      unsubMessage();
      unsubEdit();
      unsubDelete();
      unsubDelivered();
      unsubRead();
      unsubTyping();
      unsubPresence();
      unsubError();
      unsubRoomCreated();
    };
  }, [addMessage, confirmMessage, failMessage, updateMessage, deleteMessage, updateMessageStatus, setUserOnline, setUserOffline, setTypingUsers, addRoom]);

  return {
    isConnected,
    connect: async (token: string) => {
      await websocket.connect(token);
      setIsConnected(websocket.isConnected());
    },
    disconnect: () => {
      websocket.disconnect();
      setIsConnected(false);
    },
    subscribeToRooms,
  };
};
