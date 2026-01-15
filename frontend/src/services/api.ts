import axios from 'axios';
import type { AxiosInstance, AxiosError } from 'axios';
import type { APIResponse, AuthResponse, Room, Message, Notification } from '../types';
import { retryWithExponentialBackoff, isRetryableError } from '../utils/retry';

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class APIService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<APIResponse>) => {
        // Handle different error types
        if (error.response) {
          // Server responded with error status
          const status = error.response.status;
          const message = error.response.data?.error || 'An error occurred';

          if (status === 401) {
            // Token expired or invalid - redirect to login
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            if (window.location.pathname !== '/login') {
              window.location.href = '/login';
            }
          } else if (status === 403) {
            // Forbidden - user doesn't have permission
            console.error('Permission denied:', message);
          } else if (status === 404) {
            // Not found
            console.error('Resource not found:', message);
          } else if (status >= 500) {
            // Server error
            console.error('Server error:', message);
          }

          // Attach user-friendly message
          error.message = message;
        } else if (error.request) {
          // Request made but no response received - network error
          error.message = 'Network error. Please check your connection.';
          console.error('Network error:', error);
        } else {
          // Something else happened
          error.message = 'An unexpected error occurred.';
          console.error('Error:', error);
        }

        return Promise.reject(error);
      }
    );
  }

  // Auth endpoints
  async register(username: string, email: string, password: string): Promise<AuthResponse> {
    const response = await this.client.post<APIResponse<AuthResponse>>('/api/auth/register', {
      username,
      email,
      password,
    });
    return response.data.data!;
  }

  async login(identifier: string, password: string): Promise<AuthResponse> {
    const response = await this.client.post<APIResponse<AuthResponse>>('/api/auth/login', {
      identifier,
      password,
    });
    return response.data.data!;
  }

  // Room endpoints
  async getRooms(): Promise<Room[]> {
    // Retry room loading on network errors (critical for app initialization)
    return retryWithExponentialBackoff(
      async () => {
        const response = await this.client.get<APIResponse<Room[]>>('/api/rooms');
        return response.data.data || [];
      },
      {
        maxRetries: 3,
        initialDelay: 1000,
        shouldRetry: (error) => isRetryableError(error),
      }
    );
  }

  async createGroupRoom(name: string, members: string[]): Promise<Room> {
    const response = await this.client.post<APIResponse<Room>>('/api/rooms', {
      name,
      memberIds: members,
    });
    return response.data.data!;
  }

  async createDirectRoom(memberId: string): Promise<Room> {
    const response = await this.client.post<APIResponse<Room>>('/api/rooms', {
      type: 'direct',
      members: [memberId],
    });
    return response.data.data!;
  }

  async addMembers(roomId: string, members: string[]): Promise<void> {
    await this.client.post(`/api/rooms/${roomId}/members`, { members });
  }

  async removeMembers(roomId: string, members: string[]): Promise<void> {
    await this.client.delete(`/api/rooms/${roomId}/members`, { data: { members } });
  }

  // Message endpoints
  async getMessages(roomId: string, limit = 50, offset = 0): Promise<Message[]> {
    // Retry message loading on network errors
    return retryWithExponentialBackoff(
      async () => {
        const response = await this.client.get<APIResponse<Message[]>>(
          `/api/messages/room/${roomId}`,
          { params: { limit, offset } }
        );
        return response.data.data || [];
      },
      {
        maxRetries: 3,
        initialDelay: 1000,
        shouldRetry: (error) => isRetryableError(error),
      }
    );
  }

  // Notification endpoints
  async getNotifications(limit = 20, offset = 0): Promise<Notification[]> {
    const response = await this.client.get<APIResponse<Notification[]>>('/api/notifications', {
      params: { limit, offset },
    });
    return response.data.data || [];
  }

  async getPendingNotifications(): Promise<Notification[]> {
    // Retry pending notifications on network errors
    return retryWithExponentialBackoff(
      async () => {
        const response = await this.client.get<APIResponse<Notification[]>>(
          '/api/notifications/pending'
        );
        return response.data.data || [];
      },
      {
        maxRetries: 2,
        initialDelay: 1000,
        shouldRetry: (error) => isRetryableError(error),
      }
    );
  }

  async getUnreadCount(): Promise<number> {
    const response = await this.client.get<APIResponse<{ count: number }>>(
      '/api/notifications/unread-count'
    );
    return response.data.data?.count || 0;
  }

  async markNotificationAsRead(notificationId: string): Promise<void> {
    await this.client.post('/api/notifications/mark-read', { notificationId });
  }

  async markAllNotificationsAsRead(): Promise<void> {
    await this.client.post('/api/notifications/mark-all-read');
  }

  // User endpoints
  async searchUsers(query: string): Promise<{ id: string; username: string }[]> {
    const response = await this.client.get<APIResponse<{ id: string; username: string }[]>>(
      '/api/users/search',
      { params: { q: query } }
    );
    return response.data.data || [];
  }

  async createDirectChatByUsername(username: string): Promise<Room> {
    const response = await this.client.post<APIResponse<Room>>('/api/users/chat', {
      username,
    });
    return response.data.data!;
  }
}

export const api = new APIService();
