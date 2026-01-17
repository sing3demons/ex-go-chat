const WebSocket = require('ws');

// Test Redis caching by sending multiple messages
async function testRedisCaching() {
  console.log('🚀 Testing Redis caching with multiple messages...');
  
  const token2 = await getUser2Token();
  const roomId = '696b02347d48e8e6de7d5752';
  
  console.log('👤 Connecting User 2 to send multiple messages...');
  const ws2 = new WebSocket(`ws://localhost:8080/ws?token=${token2}`);
  
  ws2.on('open', function() {
    console.log('✅ User 2 connected');
    
    // Send 5 messages to test caching
    let messageCount = 0;
    const sendMessage = () => {
      if (messageCount < 5) {
        messageCount++;
        const message = {
          type: 'message',
          roomId: roomId,
          payload: {
            content: `Redis cache test message #${messageCount} at ${new Date().toLocaleTimeString()}`,
            tempId: `temp-cache-${Date.now()}-${messageCount}`
          }
        };
        
        console.log(`📤 Sending message ${messageCount}:`, message.payload.content);
        ws2.send(JSON.stringify(message));
        
        // Send next message after 1 second
        setTimeout(sendMessage, 1000);
      } else {
        console.log('✅ All 5 messages sent! These should be cached in Redis.');
        console.log('📱 Check frontend - messages should load faster from cache on next visit');
        
        setTimeout(() => {
          ws2.close();
          console.log('🔚 Test complete');
        }, 1000);
      }
    };
    
    // Start sending messages after 500ms
    setTimeout(sendMessage, 500);
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

testRedisCaching().catch(console.error);