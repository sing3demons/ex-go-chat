import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import type { Notification } from '../types';

export const useNotifications = () => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch all notifications
  const fetchNotifications = useCallback(async (limit = 20, offset = 0) => {
    setIsLoading(true);
    setError(null);
    
    try {
      const data = await api.getNotifications(limit, offset);
      setNotifications(prev => offset === 0 ? data : [...prev, ...data]);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch notifications');
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Fetch pending notifications (unread)
  const fetchPendingNotifications = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const data = await api.getPendingNotifications();
      setNotifications(data);
      setUnreadCount(data.length);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch pending notifications');
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Fetch unread count
  const fetchUnreadCount = useCallback(async () => {
    try {
      const count = await api.getUnreadCount();
      setUnreadCount(count);
    } catch (err: any) {
      console.error('Failed to fetch unread count:', err);
    }
  }, []);

  // Mark single notification as read
  const markAsRead = useCallback(async (notificationId: string) => {
    try {
      await api.markNotificationAsRead(notificationId);
      
      // Update local state
      setNotifications(prev =>
        prev.map(notif =>
          notif.id === notificationId ? { ...notif, read: true } : notif
        )
      );
      
      // Update unread count
      setUnreadCount(prev => Math.max(0, prev - 1));
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to mark notification as read');
    }
  }, []);

  // Mark all notifications as read
  const markAllAsRead = useCallback(async () => {
    try {
      await api.markAllNotificationsAsRead();
      
      // Update local state
      setNotifications(prev =>
        prev.map(notif => ({ ...notif, read: true }))
      );
      
      // Reset unread count
      setUnreadCount(0);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to mark all notifications as read');
    }
  }, []);

  // Load more notifications (pagination)
  const loadMore = useCallback(async () => {
    await fetchNotifications(20, notifications.length);
  }, [fetchNotifications, notifications.length]);

  // Auto-fetch unread count on mount
  useEffect(() => {
    fetchUnreadCount();
  }, [fetchUnreadCount]);

  return {
    notifications,
    unreadCount,
    isLoading,
    error,
    fetchNotifications,
    fetchPendingNotifications,
    fetchUnreadCount,
    markAsRead,
    markAllAsRead,
    loadMore,
  };
};
