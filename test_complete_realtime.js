const WebSocket = require('ws');

// Complete realtime test - one user listens, another sends
async function testCompleteRealtime() {
  console.log('🚀 Starting complete realtime test...');
  
  // Tokens
  const token1 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY3Yjk1ZGRhMjhiOWZiYmEwYTE4NzUiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzY4NzAwNzg2LCJuYmYiOjE3Njg2MTQzODYsImlhdCI6MTc2ODYxNDM4Nn0.9Cg5W38Cb87gkOEfXLgGoBe-Hmwcv5G7zySuMiDWLog";
  const token2 = await getUser2Token();
  const roomId = '696b02347d48e8e6de7d5752';
  
  // Connect User 1 (receiver)
  console.log('👤 Connecting User 1 (testuser) as receiver...');
  const ws1 = new WebSocket(`ws://localhost:8080/ws?token=${token1}`);
  
  let user1Ready = false;
  
  ws1.on('open', function() {
    console.log('✅ User 1 connected and listening for messages');
    user1Ready = true;
    startTest();
  });
  
  ws1.on('message', function(data) {
    console.log('📨 User 1 received raw data:', data.toString());
    
    const messages = data.toString().trim().split('\n');
    for (const messageData of messages) {
      if (messageData.trim()) {
        try {
          const msg = JSON.parse(messageData);
          if (msg.type === 'message') {
            console.log('🎉 SUCCESS! User 1 received realtime message:');
            console.log('   Content:', msg.payload.content);
            console.log('   From:', msg.payload.senderId);
            console.log('   Time:', msg.payload.timestamp);
            console.log('   Room:', msg.roomId);
          }
        } catch (error) {
          console.error('Failed to parse message:', error);
        }
      }
    }
  });
  
  ws1.on('error', function(err) {
    console.error('❌ User 1 error:', err);
  });
  
  function startTest() {
    if (!user1Ready) return;
    
    console.log('⏳ Waiting 2 seconds before sending message...');
    setTimeout(() => {
      console.log('👤 Connecting User 2 (testuser2) to send message...');
      
      const ws2 = new WebSocket(`ws://localhost:8080/ws?token=${token2}`);
      
      ws2.on('open', function() {
        console.log('✅ User 2 connected');
        
        setTimeout(() => {
          const message = {
            type: 'message',
            roomId: roomId,
            payload: {
              content: `REALTIME TEST MESSAGE at ${new Date().toLocaleTimeString()}`,
              tempId: `temp-realtime-${Date.now()}`
            }
          };
          
          console.log('📤 User 2 sending message:', message.payload.content);
          ws2.send(JSON.stringify(message));
          
          // Close User 2 after sending
          setTimeout(() => {
            ws2.close();
            console.log('🔚 User 2 disconnected after sending');
          }, 1000);
        }, 500);
      });
      
      ws2.on('error', function(err) {
        console.error('❌ User 2 error:', err);
      });
    }, 2000);
  }
  
  // Close User 1 after 10 seconds
  setTimeout(() => {
    console.log('🔚 Test complete, closing User 1 connection');
    ws1.close();
  }, 10000);
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

testCompleteRealtime().catch(console.error);