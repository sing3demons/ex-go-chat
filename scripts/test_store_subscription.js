// Test if Zustand store subscription works correctly
// This simulates what the React component should do

const { create } = require('zustand');

// Mock the store structure
const useChatStore = create((set, get) => ({
  messages: {},
  updateCounter: 0,
  
  addMessage: (message) => {
    set((state) => {
      const existingMessages = state.messages[message.roomId] || [];
      
      // Check if message already exists
      const messageExists = existingMessages.some(m => m.id === message.id);
      
      if (messageExists) {
        console.log('Message already exists, skipping:', message.id);
        return state;
      }
      
      console.log('Adding new message to store:', message.id, 'Room:', message.roomId);
      const newMessages = [...existingMessages, message];
      
      const newState = {
        ...state,
        messages: {
          ...state.messages,
          [message.roomId]: newMessages,
        },
        updateCounter: state.updateCounter + 1,
      };
      
      console.log('Updated messages count for room', message.roomId, ':', newMessages.length);
      console.log('🔔 Notifying subscribers of store change...');
      return newState;
    });
  }
}));

function testStoreSubscription() {
  console.log('🚀 Testing Zustand store subscription...');
  
  const roomId = 'test-room';
  let subscriptionCallCount = 0;
  
  // Subscribe to store changes (like React component does)
  const unsubscribe = useChatStore.subscribe((state) => {
    subscriptionCallCount++;
    const messages = state.messages[roomId] || [];
    console.log(`📡 Subscription callback #${subscriptionCallCount} - Room: ${roomId}, Messages: ${messages.length}`);
    
    if (messages.length > 0) {
      console.log('   Latest message:', messages[messages.length - 1].content);
    }
  });
  
  // Get initial state
  const initialState = useChatStore.getState();
  console.log('📊 Initial state:', {
    messagesCount: Object.keys(initialState.messages).length,
    updateCounter: initialState.updateCounter
  });
  
  // Add a message (simulating WebSocket message received)
  setTimeout(() => {
    console.log('📤 Adding test message...');
    useChatStore.getState().addMessage({
      id: 'msg-1',
      roomId: roomId,
      content: 'Test message 1',
      senderId: 'user-1',
      createdAt: new Date().toISOString()
    });
  }, 1000);
  
  // Add another message
  setTimeout(() => {
    console.log('📤 Adding second test message...');
    useChatStore.getState().addMessage({
      id: 'msg-2',
      roomId: roomId,
      content: 'Test message 2',
      senderId: 'user-2',
      createdAt: new Date().toISOString()
    });
  }, 2000);
  
  // Check final state
  setTimeout(() => {
    const finalState = useChatStore.getState();
    console.log('📊 Final state:', {
      messagesCount: (finalState.messages[roomId] || []).length,
      updateCounter: finalState.updateCounter,
      subscriptionCalls: subscriptionCallCount
    });
    
    console.log('✅ Test complete');
    unsubscribe();
  }, 3000);
}

testStoreSubscription();