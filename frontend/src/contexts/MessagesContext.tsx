import React, { createContext, useContext, useReducer, useEffect } from 'react';
import type { Message } from '../types';
import { useChatStore } from '../store/chatStore';

interface MessagesState {
  messagesByRoom: Record<string, Message[]>;
  updateCount: number;
}

type MessagesAction = 
  | { type: 'SET_MESSAGES'; roomId: string; messages: Message[] }
  | { type: 'ADD_MESSAGE'; message: Message }
  | { type: 'FORCE_UPDATE' };

const messagesReducer = (state: MessagesState, action: MessagesAction): MessagesState => {
  switch (action.type) {
    case 'SET_MESSAGES':
      return {
        ...state,
        messagesByRoom: {
          ...state.messagesByRoom,
          [action.roomId]: [...action.messages]
        },
        updateCount: state.updateCount + 1
      };
    case 'ADD_MESSAGE':
      const existingMessages = state.messagesByRoom[action.message.roomId] || [];
      const messageExists = existingMessages.some(m => m.id === action.message.id);
      
      if (messageExists) {
        return state;
      }
      
      return {
        ...state,
        messagesByRoom: {
          ...state.messagesByRoom,
          [action.message.roomId]: [...existingMessages, action.message]
        },
        updateCount: state.updateCount + 1
      };
    case 'FORCE_UPDATE':
      return {
        ...state,
        updateCount: state.updateCount + 1
      };
    default:
      return state;
  }
};

interface MessagesContextType {
  state: MessagesState;
  dispatch: React.Dispatch<MessagesAction>;
  getMessages: (roomId: string) => Message[];
}

const MessagesContext = createContext<MessagesContextType | undefined>(undefined);

export const MessagesProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, dispatch] = useReducer(messagesReducer, {
    messagesByRoom: {},
    updateCount: 0
  });

  // Sync with Zustand store
  useEffect(() => {
    const unsubscribe = useChatStore.subscribe((zustandState) => {
      console.log('🔄 Zustand store changed, syncing to React Context');
      
      // Sync all rooms
      Object.keys(zustandState.messages).forEach(roomId => {
        const messages = zustandState.messages[roomId] || [];
        dispatch({ type: 'SET_MESSAGES', roomId, messages });
      });
    });

    // Initial sync
    const initialState = useChatStore.getState();
    Object.keys(initialState.messages).forEach(roomId => {
      const messages = initialState.messages[roomId] || [];
      dispatch({ type: 'SET_MESSAGES', roomId, messages });
    });

    return unsubscribe;
  }, []);

  // Force update every second as backup
  useEffect(() => {
    const interval = setInterval(() => {
      const zustandState = useChatStore.getState();
      let hasChanges = false;
      
      Object.keys(zustandState.messages).forEach(roomId => {
        const zustandMessages = zustandState.messages[roomId] || [];
        const contextMessages = state.messagesByRoom[roomId] || [];
        
        if (zustandMessages.length !== contextMessages.length) {
          hasChanges = true;
          dispatch({ type: 'SET_MESSAGES', roomId, messages: zustandMessages });
        }
      });
      
      if (hasChanges) {
        console.log('🔄 Polling detected changes, forcing update');
        dispatch({ type: 'FORCE_UPDATE' });
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [state.messagesByRoom]);

  const getMessages = (roomId: string): Message[] => {
    return state.messagesByRoom[roomId] || [];
  };

  return (
    <MessagesContext.Provider value={{ state, dispatch, getMessages }}>
      {children}
    </MessagesContext.Provider>
  );
};

export const useMessages = () => {
  const context = useContext(MessagesContext);
  if (context === undefined) {
    throw new Error('useMessages must be used within a MessagesProvider');
  }
  return context;
};