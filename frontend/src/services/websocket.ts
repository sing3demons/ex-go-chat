import type { WSMessage, WSMessageType } from '../types';

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';

type MessageHandler = (message: WSMessage) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private token: string | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 3; // ลดจาก 5 เป็น 3
  private reconnectDelay = 2000; // เพิ่มจาก 1000 เป็น 2000
  private handlers: Map<WSMessageType, MessageHandler[]> = new Map();
  private heartbeatInterval: number | null = null;
  private connectionErrorCallback: ((error: string) => void) | null = null;
  private isReconnecting = false; // เพิ่ม flag เพื่อป้องกัน multiple reconnection
  private shouldReconnect = true; // เพิ่ม flag เพื่อควบคุม reconnection

  setConnectionErrorCallback(callback: (error: string) => void) {
    this.connectionErrorCallback = callback;
  }

  connect(token: string): Promise<void> {
    return new Promise((resolve, reject) => {
      // If already connected or connecting, don't create a new connection
      if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
        console.log('WebSocket already connected or connecting');
        resolve();
        return;
      }

      // If already reconnecting, don't start another connection
      if (this.isReconnecting) {
        console.log('WebSocket reconnection already in progress');
        resolve();
        return;
      }

      this.token = token;
      this.shouldReconnect = true; // Reset reconnection flag
      const url = `${WS_URL}/ws?token=${token}`;
      console.log('🔌 Connecting to WebSocket:', url);

      try {
        this.ws = new WebSocket(url);

        this.ws.onopen = () => {
          console.log('✅ WebSocket connected successfully');
          this.reconnectAttempts = 0;
          this.isReconnecting = false;
          this.startHeartbeat();
          resolve();
        };

        this.ws.onmessage = (event) => {
          console.log('📨 WebSocket message received:', event.data);
          try {
            // Handle multiple JSON messages separated by newlines
            const messages = event.data.trim().split('\n');
            for (const messageData of messages) {
              if (messageData.trim()) {
                const message: WSMessage = JSON.parse(messageData);
                console.log('📨 Parsed message:', message);
                this.handleMessage(message);
              }
            }
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
            console.error('Raw message data:', event.data);
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          this.isReconnecting = false;
          const errorMessage = 'WebSocket connection error. Please check your connection.';
          if (this.connectionErrorCallback) {
            this.connectionErrorCallback(errorMessage);
          }
          reject(new Error(errorMessage));
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket disconnected', event.code, event.reason);
          this.stopHeartbeat();
          this.isReconnecting = false;
          
          // Handle different close codes
          if (event.code === 1000) {
            // Normal closure - don't reconnect
            console.log('WebSocket closed normally');
            this.shouldReconnect = false;
          } else if (event.code === 4001) {
            // Custom: Authentication failed - don't reconnect
            console.log('WebSocket authentication failed');
            this.shouldReconnect = false;
            const errorMessage = 'Authentication failed. Please login again.';
            if (this.connectionErrorCallback) {
              this.connectionErrorCallback(errorMessage);
            }
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/login';
          } else if (this.shouldReconnect && !this.isReconnecting) {
            // Other errors - try to reconnect only if we should
            console.log('WebSocket closed abnormally, attempting reconnect...');
            this.attemptReconnect();
          }
        };
      } catch (error) {
        this.isReconnecting = false;
        reject(error);
      }
    });
  }

  disconnect(): void {
    console.log('🔌 Disconnecting WebSocket...');
    this.shouldReconnect = false; // ป้องกัน reconnection
    this.isReconnecting = false;
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close(1000, 'User disconnected'); // Normal closure
      this.ws = null;
    }
  }

  send(message: WSMessage): void {
    console.log('📤 Sending WebSocket message:', message);
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify(message));
        console.log('✅ Message sent successfully');
      } catch (error) {
        console.error('Failed to send WebSocket message:', error);
        if (this.connectionErrorCallback) {
          this.connectionErrorCallback('Failed to send message. Please try again.');
        }
      }
    } else {
      console.error('❌ WebSocket is not connected, readyState:', this.ws?.readyState);
      if (this.connectionErrorCallback) {
        this.connectionErrorCallback('Not connected. Please check your connection.');
      }
    }
  }

  on(type: WSMessageType, handler: MessageHandler): () => void {
    console.log(`📝 Registering handler for message type: ${type}`);
    if (!this.handlers.has(type)) {
      this.handlers.set(type, []);
    }
    this.handlers.get(type)!.push(handler);
    console.log(`✅ Handler registered. Total handlers for ${type}: ${this.handlers.get(type)!.length}`);

    // Return unsubscribe function
    return () => {
      console.log(`🗑️ Unregistering handler for message type: ${type}`);
      const handlers = this.handlers.get(type);
      if (handlers) {
        const index = handlers.indexOf(handler);
        if (index > -1) {
          handlers.splice(index, 1);
          console.log(`✅ Handler unregistered. Remaining handlers for ${type}: ${handlers.length}`);
        }
      }
    };
  }

  private handleMessage(message: WSMessage): void {
    console.log('🎯 Handling WebSocket message:', message);
    const handlers = this.handlers.get(message.type);
    if (handlers) {
      console.log(`📡 Found ${handlers.length} handlers for message type: ${message.type}`);
      handlers.forEach((handler) => handler(message));
    } else {
      console.warn(`⚠️ No handlers found for message type: ${message.type}`);
    }
  }

  private startHeartbeat(): void {
    this.heartbeatInterval = window.setInterval(() => {
      this.send({ type: 'heartbeat', payload: {} });
    }, 30000); // 30 seconds
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private attemptReconnect(): void {
    // Don't reconnect if we shouldn't or if already reconnecting
    if (!this.shouldReconnect || this.isReconnecting) {
      console.log('Skipping reconnect - shouldReconnect:', this.shouldReconnect, 'isReconnecting:', this.isReconnecting);
      return;
    }

    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached');
      this.shouldReconnect = false;
      if (this.connectionErrorCallback) {
        this.connectionErrorCallback('Connection lost. Please refresh the page.');
      }
      return;
    }

    // Don't reconnect if we don't have a token
    if (!this.token) {
      console.log('No token available for reconnection');
      return;
    }

    this.isReconnecting = true;
    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

    setTimeout(() => {
      if (this.shouldReconnect && this.token && (!this.ws || this.ws.readyState === WebSocket.CLOSED)) {
        this.connect(this.token).catch((error) => {
          console.error('Reconnect failed:', error);
          this.isReconnecting = false;
        });
      } else {
        this.isReconnecting = false;
      }
    }, delay);
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

export const websocket = new WebSocketService();
