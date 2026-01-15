import { useEffect, useCallback } from 'react';
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
  TypingPayload
} from '../types';

export const useWebSocket = () => {
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

  // Subscribe to room events when rooms change
  const subscribeToRooms = useCallback(() => {
    if (rooms.length > 0) {
      console.log(`Subscribing to ${rooms.length} rooms...`);
      rooms.forEach(room => {
        websocket.send({
          type: 'join_room',
          roomId: room.id,
          payload: {}
        });
      });
      console.log('Room subscriptions sent');
    }
  }, [rooms]);

  useEffect(() => {
    // Subscribe to rooms when WebSocket is connected and rooms are loaded
    if (websocket.isConnected() && rooms.length > 0) {
      subscribeToRooms();
    }
  }, [subscribeToRooms, rooms]);

  useEffect(() => {
    // Handle incoming messages
    const unsubMessage = websocket.on('message', (msg) => {
      const payload = msg.payload as ChatMessagePayload & { tempId?: string };
      
      // Check if this is a confirmation of an optimistic message
      if (payload.tempId) {
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

      // Send delivery acknowledgment automatically
      websocket.send({
        type: 'delivered',
        roomId: msg.roomId,
        payload: { messageId: payload.messageId }
      });
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

    return () => {
      unsubMessage();
      unsubEdit();
      unsubDelete();
      unsubDelivered();
      unsubRead();
      unsubTyping();
      unsubPresence();
      unsubError();
    };
  }, [addMessage, confirmMessage, failMessage, updateMessage, deleteMessage, updateMessageStatus, setUserOnline, setUserOffline, setTypingUsers]);

  return {
    isConnected: websocket.isConnected(),
    connect: (token: string) => websocket.connect(token),
    disconnect: () => websocket.disconnect(),
    subscribeToRooms,
  };
};
