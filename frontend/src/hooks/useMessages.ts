import { useContext } from 'react';
import { MessagesContext } from '../contexts/MessagesContext';

export const useMessages = () => {
  const context = useContext(MessagesContext);
  if (context === undefined) {
    throw new Error('useMessages must be used within a MessagesProvider');
  }
  return context;
};
