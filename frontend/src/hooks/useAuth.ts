import { useState } from 'react';
import { useAuthStore } from '../store/authStore';
import { api } from '../services/api';
import type { User } from '../types';

interface RegisterData {
  username: string;
  email: string;
  password: string;
}

interface LoginData {
  identifier: string;
  password: string;
}

export const useAuth = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const { user, token, setAuth, clearAuth } = useAuthStore();

  const register = async (data: RegisterData): Promise<User | null> => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await api.register(data.username, data.email, data.password);
      setAuth(response.user, response.token);
      return response.user;
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || 'Registration failed';
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  const login = async (data: LoginData): Promise<User | null> => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await api.login(data.identifier, data.password);
      setAuth(response.user, response.token);
      return response.user;
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || 'Login failed';
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  const logout = () => {
    clearAuth();
  };

  return {
    user,
    token,
    loading,
    error,
    register,
    login,
    logout,
    isAuthenticated: !!token,
  };
};
