const WebSocket = require('ws');

// Test complete flow: frontend sends, backend receives, other user gets it
async function testCompleteFlow() {
  console.log('🚀 Testing complete message flow...');
  
  const token1 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY3Yjk1ZGRhMjhiOWZiYmEwYTE4NzUiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzY4NzAwNzg2LCJuYmYiOjE3Njg2MTQzODYsImlhdCI6MTc2ODYxNDM4Nn0.9Cg5W38Cb87gkOEfXLgGoBe-Hmwcv5G7zySuMiDWLog";
  const token2 = await getUser2Token();
  const roomId = '696b02347d48e8e6de7d5752';
  
  // Connect User 1 as listener
  console.log('👤 Connecting User 1 (testuser) as listener...');
  const ws1 = new WebSocket(`ws://localhost:8080/ws?token=${token1}`);
  
  ws1.on('open', function() {
    console.log('✅ User 1 connected and listening');
  });
  
  ws1.on('message', function(data) {
    const messages = data.toString().trim().split('\n');
    for (const messageData of messages) {
      if (messageData.trim()) {
        try {
          const msg = JSON.parse(messageData);
          if (msg.type === 'message' && msg.payload.content.includes('Complete flow test')) {
            console.log('🎉 SUCCESS! Message received by User 1:');
            console.log('   Content:', msg.payload.content);
            console.log('   ✅ Complete flow working: Send → Backend → Receive');
          }
        } catch (error) {
          // Ignore parse errors
        }
      }
    }
  });
  
  // Wait a bit then send message from User 2
  setTimeout(() => {
    console.log('👤 Connecting User 2 to send message...');
    const ws2 = new WebSocket(`ws://localhost:8080/ws?token=${token2}`);
    
    ws2.on('open', function() {
      console.log('✅ User 2 connected');
      
      setTimeout(() => {
        const message = {
          type: 'message',
          roomId: roomId,
          payload: {
            content: `Complete flow test message at ${new Date().toLocaleTimeString()}`,
            tempId: `temp-flow-${Date.now()}`
          }
        };
        
        console.log('📤 User 2 sending message:', message.payload.content);
        console.log('📱 This message should appear in frontend within 3 seconds due to polling');
        ws2.send(JSON.stringify(message));
        
        setTimeout(() => {
          ws2.close();
          console.log('🔚 User 2 disconnected');
        }, 1000);
      }, 500);
    });
  }, 2000);
  
  // Close test after 8 seconds
  setTimeout(() => {
    console.log('🔚 Test complete');
    ws1.close();
  }, 8000);
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

testCompleteFlow().catch(console.error);