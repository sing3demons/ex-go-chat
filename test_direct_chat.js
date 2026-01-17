const WebSocket = require('ws');

// Test tokens (you'll need to get these from actual login)
const user1Token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY4YTY5NmEyYjY1OWIwMmUzODJkYjMiLCJ1c2VybmFtZSI6InRlc3QxMSIsImV4cCI6MTc2ODU3NzgyOSwibmJmIjoxNzY4NDkxNDI5LCJpYXQiOjE3Njg0OTE0Mjl9.CQSrxNV1r4y5fPiN84uMg8YAA8cD03vc8q27a_6ite8';
const user2Token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY4YTY5NmEyYjY1OWIwMmUzODJkYjQiLCJ1c2VybmFtZSI6InRlc3QyMiIsImV4cCI6MTc2ODU3NzgyOSwibmJmIjoxNzY4NDkxNDI5LCJpYXQiOjE3Njg0OTE0Mjl9.example';

// Connect two users
const ws1 = new WebSocket(`ws://localhost:8080/ws?token=${user1Token}`);
const ws2 = new WebSocket(`ws://localhost:8080/ws?token=${user2Token}`);

let roomId = null;

ws1.on('open', () => {
    console.log('User 1 connected');
});

ws2.on('open', () => {
    console.log('User 2 connected');
});

ws1.on('message', (data) => {
    const msg = JSON.parse(data);
    console.log('User 1 received:', msg.type, msg);
    
    if (msg.type === 'room_created') {
        roomId = msg.payload.roomId;
        console.log('Room created, sending message...');
        
        // Send a message to the room
        setTimeout(() => {
            ws1.send(JSON.stringify({
                type: 'message',
                roomId: roomId,
                payload: {
                    content: 'Hello from User 1!'
                }
            }));
        }, 1000);
    }
});

ws2.on('message', (data) => {
    const msg = JSON.parse(data);
    console.log('User 2 received:', msg.type, msg);
});

// Create direct chat via API
setTimeout(() => {
    console.log('Creating direct chat...');
    fetch('http://localhost:8080/api/users/chat', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${user1Token}`
        },
        body: JSON.stringify({
            username: 'test22'
        })
    })
    .then(res => res.json())
    .then(data => {
        console.log('Direct chat created:', data);
    })
    .catch(err => {
        console.error('Error creating chat:', err);
    });
}, 2000);

// Close connections after 10 seconds
setTimeout(() => {
    ws1.close();
    ws2.close();
    process.exit(0);
}, 10000);