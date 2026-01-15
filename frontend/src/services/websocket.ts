import type { WSMessage, WSMessageType } from '../types';

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';

type MessageHandler = (message: WSMessage) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private token: string | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private handlers: Map<WSMessageType, MessageHandler[]> = new Map();
  private heartbeatInterval: number | null = null;
  private connectionErrorCallback: ((error: string) => void) | null = null;

  setConnectionErrorCallback(callback: (error: string) => void) {
    this.connectionErrorCallback = callback;
  }

  connect(token: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.token = token;
      const url = `${WS_URL}/ws?token=${token}`;

      try {
        this.ws = new WebSocket(url);

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          this.startHeartbeat();
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WSMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          const errorMessage = 'WebSocket connection error. Please check your connection.';
          if (this.connectionErrorCallback) {
            this.connectionErrorCallback(errorMessage);
          }
          reject(new Error(errorMessage));
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket disconnected', event.code, event.reason);
          this.stopHeartbeat();
          
          // Handle different close codes
          if (event.code === 1000) {
            // Normal closure
            console.log('WebSocket closed normally');
          } else if (event.code === 1006) {
            // Abnormal closure - try to reconnect
            console.log('WebSocket closed abnormally, attempting reconnect...');
            this.attemptReconnect();
          } else if (event.code === 4001) {
            // Custom: Authentication failed
            const errorMessage = 'Authentication failed. Please login again.';
            if (this.connectionErrorCallback) {
              this.connectionErrorCallback(errorMessage);
            }
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/login';
          } else {
            // Other errors - try to reconnect
            this.attemptReconnect();
          }
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  disconnect(): void {
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  send(message: WSMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify(message));
      } catch (error) {
        console.error('Failed to send WebSocket message:', error);
        if (this.connectionErrorCallback) {
          this.connectionErrorCallback('Failed to send message. Please try again.');
        }
      }
    } else {
      console.error('WebSocket is not connected');
      if (this.connectionErrorCallback) {
        this.connectionErrorCallback('Not connected. Please check your connection.');
      }
    }
  }

  on(type: WSMessageType, handler: MessageHandler): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, []);
    }
    this.handlers.get(type)!.push(handler);

    // Return unsubscribe function
    return () => {
      const handlers = this.handlers.get(type);
      if (handlers) {
        const index = handlers.indexOf(handler);
        if (index > -1) {
          handlers.splice(index, 1);
        }
      }
    };
  }

  private handleMessage(message: WSMessage): void {
    const handlers = this.handlers.get(message.type);
    if (handlers) {
      handlers.forEach((handler) => handler(message));
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
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached');
      if (this.connectionErrorCallback) {
        this.connectionErrorCallback('Connection lost. Please refresh the page.');
      }
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

    setTimeout(() => {
      if (this.token) {
        this.connect(this.token).catch((error) => {
          console.error('Reconnect failed:', error);
        });
      }
    }, delay);
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

export const websocket = new WebSocketService();
