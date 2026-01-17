const WebSocket = require('ws');

// Test to send a message that the frontend should receive
async function testFrontendReceive() {
  console.log('🚀 Testing frontend message reception...');
  
  const token2 = await getUser2Token();
  const roomId = '696b02347d48e8e6de7d5752';
  
  console.log('👤 Connecting User 2 to send message to frontend...');
  const ws2 = new WebSocket(`ws://localhost:8080/ws?token=${token2}`);
  
  ws2.on('open', function() {
    console.log('✅ User 2 connected');
    
    setTimeout(() => {
      const message = {
        type: 'message',
        roomId: roomId,
        payload: {
          content: `Frontend test message at ${new Date().toLocaleTimeString()}`,
          tempId: `temp-frontend-${Date.now()}`
        }
      };
      
      console.log('📤 Sending message to frontend:', message.payload.content);
      console.log('📱 Check your frontend at http://localhost:5173 to see if this message appears within 3 seconds');
      ws2.send(JSON.stringify(message));
      
      // Close after sending
      setTimeout(() => {
        ws2.close();
        console.log('🔚 Message sent, connection closed');
      }, 1000);
    }, 500);
  });
  
  ws2.on('error', function(err) {
    console.error('❌ User 2 error:', err);
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

testFrontendReceive().catch(console.error);