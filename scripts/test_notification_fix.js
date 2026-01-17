// Test script to verify notification and message fixes
console.log('Testing notification and message fixes...');

// Run tests
async function runTests() {
  try {
    console.log('Testing server connectivity...\n');
    
    // Test server health
    const healthCheck = await fetch('http://localhost:8080/health');
    if (healthCheck.ok) {
      console.log('✓ Server is running and accessible');
    } else {
      console.log('✗ Server health check failed');
    }
    
    console.log('\n=== Manual Testing Instructions ===');
    console.log('To test the fixes:');
    console.log('1. Open the frontend in two different browsers/tabs');
    console.log('2. Login with two different users (e.g., test11 and test22)');
    console.log('3. In browser 1: Create a direct chat with the other user');
    console.log('4. In browser 1: Send a message');
    console.log('5. Check browser 2: Message should appear immediately');
    console.log('6. Close browser 2 (simulate offline)');
    console.log('7. Send another message from browser 1');
    console.log('8. Reopen browser 2: Should see notification');
    
    console.log('\n=== Expected Results ===');
    console.log('✓ Messages should not show "failed" status');
    console.log('✓ Messages should appear immediately for online users');
    console.log('✓ Notifications should be created for offline users');
    console.log('✓ Room subscription should work for new direct chats');
    
  } catch (error) {
    console.error('Test failed:', error.message);
  }
}

runTests();