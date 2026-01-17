const WebSocket = require('ws');

// Send a test message to the room
async function sendTestMessage() {
  console.log('🚀 Sending test message...');
  
  // User 2 token (testuser2)
  const token2 = await getUser2Token();
  const roomId = '696b02347d48e8e6de7d5752';
  
  const ws = new WebSocket(`ws://localhost:8080/ws?token=${token2}`);
  
  ws.on('open', function open() {
    console.log('✅ Connected as testuser2');
    
    // Send a message after connection
    setTimeout(() => {
      const message = {
        type: 'message',
        roomId: roomId,
        payload: {
          content: `Test message from testuser2 at ${new Date().toLocaleTimeString()}`,
          tempId: `temp-test-${Date.now()}`
        }
      };
      console.log('📤 Sending message:', message.payload.content);
      ws.send(JSON.stringify(message));
      
      // Close after sending
      setTimeout(() => {
        ws.close();
        console.log('🔚 Message sent, connection closed');
      }, 1000);
    }, 1000);
  });

  ws.on('error', function error(err) {
    console.error('❌ WebSocket error:', err);
  });
}

async function getUser2Token() {
  const response = await fetch('http://localhost:8080/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identifier: 'testuser2', password: 'Password123' })
  });
  const data = await response.json();
  return data.data.token;
}

sendTestMessage().catch(console.error);