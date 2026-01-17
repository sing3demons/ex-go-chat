import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '../types';
import { api } from '../services/api';
import { websocket } from '../services/websocket';

// Decode JWT token to get user info
function decodeToken(token: string): { userId: string; username: string } | null {
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));
    return {
      userId: decoded.userId,
      username: decoded.username,
    };
  } catch {
    return null;
  }
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  
  login: (identifier: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  setAuth: (user: User, token: string) => void;
  clearAuth: () => void;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (identifier: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          console.log('Login attempt:', identifier);
          const response = await api.login(identifier, password);
          console.log('Login response:', response);
          
          // Store token
          localStorage.setItem('token', response.token);
          
          // Decode user from JWT token
          const decoded = decodeToken(response.token);
          console.log('Decoded token:', decoded);
          
          if (!decoded) {
            throw new Error('Invalid token received');
          }
          
          const user = {
            id: decoded.userId,
            username: decoded.username,
            email: identifier.includes('@') ? identifier : '',
            createdAt: new Date().toISOString(),
          };
          
          console.log('Setting user:', user);
          localStorage.setItem('user', JSON.stringify(user));
          
          // Update auth state first
          set({
            user,
            token: response.token,
            isAuthenticated: true,
            isLoading: false,
          });
          console.log('Auth state updated, isAuthenticated: true');
          
          // Connect WebSocket after successful login
          try {
            console.log('Connecting WebSocket...');
            await websocket.connect(response.token);
            console.log('WebSocket connected successfully');
          } catch (wsError) {
            console.error('WebSocket connection failed:', wsError);
            // Don't fail login if WebSocket fails - user can still use the app
          }
          
        } catch (error: any) {
          console.error('Login error:', error);
          set({
            error: error.response?.data?.error || error.message || 'Login failed',
            isLoading: false,
          });
          throw error;
        }
      },

      register: async (username: string, email: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await api.register(username, email, password);
          localStorage.setItem('token', response.token);
          localStorage.setItem('user', JSON.stringify(response.user));
          
          set({
            user: response.user,
            token: response.token,
            isAuthenticated: true,
            isLoading: false,
          });
          
          // Connect WebSocket (don't block registration if it fails)
          websocket.connect(response.token).catch((err) => {
            console.error('WebSocket connection failed:', err);
          });
        } catch (error: any) {
          set({
            error: error.response?.data?.error || 'Registration failed',
            isLoading: false,
          });
          throw error;
        }
      },

      logout: () => {
        console.log('Logging out...');
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        websocket.disconnect(); // Properly disconnect WebSocket
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          error: null,
        });
      },

      setAuth: (user: User, token: string) => {
        localStorage.setItem('token', token);
        localStorage.setItem('user', JSON.stringify(user));
        set({
          user,
          token,
          isAuthenticated: true,
        });
      },

      clearAuth: () => {
        console.log('Clearing auth...');
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        websocket.disconnect(); // Properly disconnect WebSocket
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          error: null,
        });
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
