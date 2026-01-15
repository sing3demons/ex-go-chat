// User types
export interface User {
  id: string;
  username: string;
  email: string;
  createdAt: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

// Room types
export interface Room {
  id: string;
  type: 'direct' | 'group';
  name?: string;
  members: string[];
  createdAt: string;
  updatedAt: string;
}

// Message types
export interface MessageStatus {
  delivered: string;
  read: string;
}

export interface Message {
  id: string;
  roomId: string;
  senderId: string;
  content: string;
  status: Record<string, MessageStatus>;
  edited: boolean;
  deleted: boolean;
  createdAt: string;
  updatedAt: string;
  // Optimistic update fields
  pending?: boolean;
  failed?: boolean;
  tempId?: string;
}

// Notification types
export interface Notification {
  id: string;
  userId: string;
  roomId: string;
  messageId: string;
  type: 'message' | 'mention' | 'group_invite';
  read: boolean;
  createdAt: string;
}

// WebSocket message types
export type WSMessageType = 
  | 'message'
  | 'typing'
  | 'read'
  | 'delivered'
  | 'presence'
  | 'edit'
  | 'delete'
  | 'heartbeat'
  | 'join_room'
  | 'error';

export interface WSMessage {
  type: WSMessageType;
  roomId?: string;
  payload: any;
}

export interface ChatMessagePayload {
  messageId: string;
  content: string;
  senderId: string;
  timestamp: string;
  tempId?: string; // For optimistic updates
}

export interface TypingPayload {
  userId: string;
  username: string;
  isTyping: boolean;
}

export interface StatusPayload {
  messageId: string;
  userId: string;
  status: 'delivered' | 'read';
  timestamp: string;
}

export interface PresencePayload {
  userId: string;
  online: boolean;
  lastSeen?: string;
}

export interface EditPayload {
  messageId: string;
  content: string;
  editedAt: string;
}

export interface DeletePayload {
  messageId: string;
}

export interface ErrorPayload {
  code: string;
  message: string;
}

// API Response types
export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
}
