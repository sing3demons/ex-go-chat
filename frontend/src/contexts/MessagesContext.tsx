import React, { createContext, useContext, useReducer, useEffect, useRef } from 'react';
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

  const prevMessagesRef = useRef<Record<string, Message[]>>({});

  // Sync with Zustand store - Subscribe to changes and sync immediately
  useEffect(() => {
    // Subscribe to entire store state
    const unsubscribe = useChatStore.subscribe(
      (state) => state,
      (newState) => {
        // Compare messages and sync only if changed
        const currentMessages = newState.messages;
        let hasChanges = false;

        // Check for changes and dispatch updates
        Object.keys(currentMessages).forEach(roomId => {
          const newMessages = currentMessages[roomId] || [];
          const prevMessages = prevMessagesRef.current[roomId] || [];

          // Deep comparison: check if messages array actually changed
          if (JSON.stringify(newMessages) !== JSON.stringify(prevMessages)) {
            hasChanges = true;
            dispatch({ type: 'SET_MESSAGES', roomId, messages: newMessages });
          }
        });

        // Check for removed rooms
        Object.keys(prevMessagesRef.current).forEach(roomId => {
          if (!currentMessages[roomId]) {
            hasChanges = true;
            dispatch({ type: 'SET_MESSAGES', roomId, messages: [] });
          }
        });

        if (hasChanges) {
          prevMessagesRef.current = JSON.parse(JSON.stringify(currentMessages));
        }
      }
    );

    // Initial sync
    const initialState = useChatStore.getState();
    prevMessagesRef.current = JSON.parse(JSON.stringify(initialState.messages));
    Object.keys(initialState.messages).forEach(roomId => {
      const messages = initialState.messages[roomId] || [];
      dispatch({ type: 'SET_MESSAGES', roomId, messages });
    });

    return unsubscribe;
  }, []);

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