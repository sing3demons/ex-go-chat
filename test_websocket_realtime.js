const WebSocket = require('ws');

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY3Yjk1ZGRhMjhiOWZiYmEwYTE4NzUiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzY4NzAwNzg2LCJuYmYiOjE3Njg2MTQzODYsImlhdCI6MTc2ODYxNDM4Nn0.9Cg5W38Cb87gkOEfXLgGoBe-Hmwcv5G7zySuMiDWLog";

const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.on('open', function open() {
  console.log('WebSocket connected');
  
  // Test sending a message to a room
  const testMessage = {
    type: 'message',
    roomId: '696aea6abacf94d10c061470', // Use new room ID
    payload: {
      content: 'Test realtime message from WebSocket',
      tempId: `temp-${Date.now()}`
    }
  };
  
  console.log('Sending message:', testMessage);
  ws.send(JSON.stringify(testMessage));
});

ws.on('message', function message(data) {
  console.log('Received:', data.toString());
});

ws.on('error', function error(err) {
  console.error('WebSocket error:', err);
});

ws.on('close', function close() {
  console.log('WebSocket disconnected');
});

// Keep connection alive for 10 seconds
setTimeout(() => {
  ws.close();
}, 10000);