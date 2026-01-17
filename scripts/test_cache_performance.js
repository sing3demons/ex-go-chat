// Test cache performance by fetching messages multiple times
async function testCachePerformance() {
  console.log('🚀 Testing cache performance...');
  
  const token1 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2OTY3Yjk1ZGRhMjhiOWZiYmEwYTE4NzUiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzY4NzAwNzg2LCJuYmYiOjE3Njg2MTQzODYsImlhdCI6MTc2ODYxNDM4Nn0.9Cg5W38Cb87gkOEfXLgGoBe-Hmwcv5G7zySuMiDWLog";
  const roomId = '696b02347d48e8e6de7d5752';
  
  // First fetch (should hit database and populate cache)
  console.log('📊 First fetch (database + cache population)...');
  const start1 = Date.now();
  const response1 = await fetch(`http://localhost:8080/api/messages/room/${roomId}?limit=50&offset=0`, {
    headers: { 'Authorization': `Bearer ${token1}` }
  });
  const data1 = await response1.json();
  const time1 = Date.now() - start1;
  console.log(`   ⏱️  Time: ${time1}ms, Messages: ${data1.data.length}`);
  
  // Second fetch (should hit cache)
  console.log('📊 Second fetch (should use cache)...');
  const start2 = Date.now();
  const response2 = await fetch(`http://localhost:8080/api/messages/room/${roomId}?limit=50&offset=0`, {
    headers: { 'Authorization': `Bearer ${token1}` }
  });
  const data2 = await response2.json();
  const time2 = Date.now() - start2;
  console.log(`   ⏱️  Time: ${time2}ms, Messages: ${data2.data.length}`);
  
  // Third fetch (should still use cache)
  console.log('📊 Third fetch (should still use cache)...');
  const start3 = Date.now();
  const response3 = await fetch(`http://localhost:8080/api/messages/room/${roomId}?limit=50&offset=0`, {
    headers: { 'Authorization': `Bearer ${token1}` }
  });
  const data3 = await response3.json();
  const time3 = Date.now() - start3;
  console.log(`   ⏱️  Time: ${time3}ms, Messages: ${data3.data.length}`);
  
  // Performance comparison
  console.log('\n📈 Performance Analysis:');
  console.log(`   First fetch:  ${time1}ms`);
  console.log(`   Second fetch: ${time2}ms`);
  console.log(`   Third fetch:  ${time3}ms`);
  
  if (time2 < time1 && time3 < time1) {
    console.log('✅ Cache is working! Subsequent fetches are faster.');
  } else {
    console.log('⚠️  Cache might not be working as expected.');
  }
  
  // Show some recent messages
  console.log('\n📝 Recent messages:');
  data3.data.slice(0, 3).forEach((msg, i) => {
    console.log(`   ${i + 1}. ${msg.content.substring(0, 50)}...`);
  });
}

testCachePerformance().catch(console.error);