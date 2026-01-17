import React, { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';
import { useChatStore } from '../store/chatStore';

interface RealtimeContextType {
  forceUpdate: () => void;
  updateCount: number;
}

const RealtimeContext = createContext<RealtimeContextType | undefined>(undefined);

interface RealtimeProviderProps {
  children: ReactNode;
}

export const RealtimeProvider: React.FC<RealtimeProviderProps> = ({ children }) => {
  const [updateCount, setUpdateCount] = useState(0);

  const forceUpdate = () => {
    setUpdateCount(prev => prev + 1);
  };

  // Subscribe to chat store changes
  useEffect(() => {
    const unsubscribe = useChatStore.subscribe(() => {
      forceUpdate();
    });

    return unsubscribe;
  }, []);

  return (
    <RealtimeContext.Provider value={{ forceUpdate, updateCount }}>
      {children}
    </RealtimeContext.Provider>
  );
};

export const useRealtime = () => {
  const context = useContext(RealtimeContext);
  if (context === undefined) {
    throw new Error('useRealtime must be used within a RealtimeProvider');
  }
  return context;
};