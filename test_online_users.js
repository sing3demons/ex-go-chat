const WebSocket = require('ws');

// Test online users tracking
async function testOnlineUsers() {
  console.log('🚀 Testing online users tracking...');
  
  const token1 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY3Yjk1ZGRhMjhiOWZiYmEwYTE4NzUiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzY4NzAwNzg2LCJuYmYiOjE3Njg2MTQzODYsImlhdCI6MTc2ODYxNDM4Nn0.9Cg5W38Cb87gkOEfXLgGoBe-Hmwcv5G7zySuMiDWLog";
  const token2 = await getUser2Token();
  
  console.log('👤 Connecting User 1 (testuser)...');
  const ws1 = new WebSocket(`ws://localhost:8080/ws?token=${token1}`);
  
  ws1.on('open', function() {
    console.log('✅ User 1 connected');
    
    // Wait a bit then connect User 2
    setTimeout(() => {
      console.log('👤 Connecting User 2 (testuser2)...');
      const ws2 = new WebSocket(`ws://localhost:8080/ws?token=${token2}`);
      
      ws2.on('open', function() {
        console.log('✅ User 2 connected');
        console.log('📊 Both users should now be online');
        
        // Wait then disconnect User 1
        setTimeout(() => {
          console.log('🔌 Disconnecting User 1...');
          ws1.close();
        }, 2000);
        
        // Wait then disconnect User 2
        setTimeout(() => {
          console.log('🔌 Disconnecting User 2...');
          ws2.close();
          console.log('🔚 Test complete');
        }, 4000);
      });
      
      ws2.on('message', function(data) {
        const messages = data.toString().trim().split('\n');
        for (const messageData of messages) {
          if (messageData.trim()) {
            try {
              const msg = JSON.parse(messageData);
              if (msg.type === 'presence') {
                console.log('📡 User 2 received presence update:', {
                  userId: msg.payload.userId,
                  online: msg.payload.online,
                  lastSeen: msg.payload.lastSeen
                });
              }
            } catch (error) {
              // Ignore parse errors
            }
          }
        }
      });
    }, 1000);
  });
  
  ws1.on('message', function(data) {
    const messages = data.toString().trim().split('\n');
    for (const messageData of messages) {
      if (messageData.trim()) {
        try {
          const msg = JSON.parse(messageData);
          if (msg.type === 'presence') {
            console.log('📡 User 1 received presence update:', {
              userId: msg.payload.userId,
              online: msg.payload.online,
              lastSeen: msg.payload.lastSeen
            });
          }
        } catch (error) {
          // Ignore parse errors
        }
      }
    }
  });
  
  ws1.on('close', function() {
    console.log('🔌 User 1 disconnected');
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

testOnlineUsers().catch(console.error);