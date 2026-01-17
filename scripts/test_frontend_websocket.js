const WebSocket = require('ws');

// Test frontend WebSocket behavior
async function testFrontendWebSocket() {
  console.log('🚀 Testing frontend WebSocket behavior...');
  
  // User 1 token (testuser)
  const token1 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY3Yjk1ZGRhMjhiOWZiYmEwYTE4NzUiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzY4NzAwNzg2LCJuYmYiOjE3Njg2MTQzODYsImlhdCI6MTc2ODYxNDM4Nn0.9Cg5W38Cb87gkOEfXLgGoBe-Hmwcv5G7zySuMiDWLog";
  
  const roomId = '696b02347d48e8e6de7d5752';
  
  // Connect as User 1
  const ws1 = new WebSocket(`ws://localhost:8080/ws?token=${token1}`);
  
  ws1.on('open', function open() {
    console.log('✅ User 1 connected');
    
    // Send a message after connection
    setTimeout(() => {
      const message = {
        type: 'message',
        roomId: roomId,
        payload: {
          content: `Frontend test message at ${new Date().toISOString()}`,
          tempId: `temp-frontend-${Date.now()}`
        }
      };
      console.log('📤 Sending message:', message.payload.content);
      ws1.send(JSON.stringify(message));
    }, 1000);
  });

  ws1.on('message', function message(data) {
    console.log('📨 Raw WebSocket data received:', data.toString());
    
    // Parse like frontend does
    const messages = data.toString().trim().split('\n');
    for (const messageData of messages) {
      if (messageData.trim()) {
        try {
          const msg = JSON.parse(messageData);
          console.log('📨 Parsed message:', JSON.stringify(msg, null, 2));
          
          if (msg.type === 'message') {
            console.log('✅ Message received - Type:', msg.type, 'Content:', msg.payload.content);
            
            // Simulate frontend store update
            console.log('🔄 Would update store with message:', {
              id: msg.payload.messageId,
              roomId: msg.roomId,
              senderId: msg.payload.senderId,
              content: msg.payload.content,
              createdAt: msg.payload.timestamp
            });
          }
        } catch (error) {
          console.error('❌ Failed to parse message:', error);
        }
      }
    }
  });

  ws1.on('error', function error(err) {
    console.error('❌ WebSocket error:', err);
  });

  ws1.on('close', function close() {
    console.log('🔚 WebSocket closed');
  });

  // Close after 5 seconds
  setTimeout(() => {
    console.log('🔚 Closing connection...');
    ws1.close();
  }, 5000);
}

testFrontendWebSocket().catch(console.error);