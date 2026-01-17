const WebSocket = require('ws');

// User 1 (testuser)
const token1 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY3Yjk1ZGRhMjhiOWZiYmEwYTE4NzUiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzY4NzAwNzg2LCJuYmYiOjE3Njg2MTQzODYsImlhdCI6MTc2ODYxNDM4Nn0.9Cg5W38Cb87gkOEfXLgGoBe-Hmwcv5G7zySuMiDWLog";

// User 2 (testuser2) - need to get token
async function getUser2Token() {
  const response = await fetch('http://localhost:8080/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identifier: 'testuser2', password: 'Password123' })
  });
  const data = await response.json();
  return data.data.token;
}

async function testRealtimeChat() {
  console.log('🚀 Starting realtime chat test...');
  
  // Get token for user 2
  const token2 = await getUser2Token();
  console.log('✅ Got token for user 2');

  const roomId = '696b02347d48e8e6de7d5752'; // Test room

  // Connect User 1
  const ws1 = new WebSocket(`ws://localhost:8080/ws?token=${token1}`);
  
  // Connect User 2
  const ws2 = new WebSocket(`ws://localhost:8080/ws?token=${token2}`);

  let user1Connected = false;
  let user2Connected = false;

  ws1.on('open', function open() {
    console.log('👤 User 1 (testuser) connected');
    user1Connected = true;
    checkBothConnected();
  });

  ws2.on('open', function open() {
    console.log('👤 User 2 (testuser2) connected');
    user2Connected = true;
    checkBothConnected();
  });

  function checkBothConnected() {
    if (user1Connected && user2Connected) {
      console.log('✅ Both users connected, starting test...');
      
      // User 1 sends a message
      setTimeout(() => {
        const message = {
          type: 'message',
          roomId: roomId,
          payload: {
            content: `Test message from User 1 at ${new Date().toISOString()}`,
            tempId: `temp-user1-${Date.now()}`
          }
        };
        console.log('📤 User 1 sending message:', message.payload.content);
        ws1.send(JSON.stringify(message));
      }, 1000);

      // User 2 sends a message
      setTimeout(() => {
        const message = {
          type: 'message',
          roomId: roomId,
          payload: {
            content: `Test message from User 2 at ${new Date().toISOString()}`,
            tempId: `temp-user2-${Date.now()}`
          }
        };
        console.log('📤 User 2 sending message:', message.payload.content);
        ws2.send(JSON.stringify(message));
      }, 2000);
    }
  }

  ws1.on('message', function message(data) {
    const messages = data.toString().trim().split('\n');
    for (const messageData of messages) {
      if (messageData.trim()) {
        try {
          const msg = JSON.parse(messageData);
          if (msg.type === 'message') {
            console.log('📨 User 1 received message:', msg.payload.content, 'from:', msg.payload.senderId);
          }
        } catch (error) {
          console.error('Failed to parse message for User 1:', error);
        }
      }
    }
  });

  ws2.on('message', function message(data) {
    const messages = data.toString().trim().split('\n');
    for (const messageData of messages) {
      if (messageData.trim()) {
        try {
          const msg = JSON.parse(messageData);
          if (msg.type === 'message') {
            console.log('📨 User 2 received message:', msg.payload.content, 'from:', msg.payload.senderId);
          }
        } catch (error) {
          console.error('Failed to parse message for User 2:', error);
        }
      }
    }
  });

  ws1.on('error', function error(err) {
    console.error('❌ User 1 WebSocket error:', err);
  });

  ws2.on('error', function error(err) {
    console.error('❌ User 2 WebSocket error:', err);
  });

  // Close connections after 10 seconds
  setTimeout(() => {
    console.log('🔚 Closing connections...');
    ws1.close();
    ws2.close();
  }, 10000);
}

testRealtimeChat().catch(console.error);