import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useEffect } from 'react';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { ChatPage } from './pages/ChatPage';
import { useAuthStore } from './store/authStore';
import { websocket } from './services/websocket';
import { ErrorBoundary } from './components';
import { MessagesProvider } from './contexts/MessagesContext';

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" />;
}

function App() {
  const { token, isAuthenticated } = useAuthStore();

  // Reconnect WebSocket on app load if user is authenticated
  useEffect(() => {
    let mounted = true;
    
    const reconnectWebSocket = async () => {
      // Only reconnect if user is authenticated, has token, and WebSocket is not connected
      if (isAuthenticated && token && !websocket.isConnected() && mounted) {
        try {
          console.log('App: Reconnecting WebSocket on app load...');
          await websocket.connect(token);
          console.log('App: WebSocket reconnected successfully');
        } catch (error) {
          console.error('App: Failed to reconnect WebSocket:', error);
          // Don't show error to user on app load - they can still use the app
        }
      }
    };

    // Add a small delay to ensure auth state is properly loaded
    const timer = setTimeout(() => {
      if (mounted) {
        reconnectWebSocket();
      }
    }, 500);
    
    return () => {
      mounted = false;
      clearTimeout(timer);
    };
  }, [isAuthenticated, token]);

  return (
    <ErrorBoundary>
      <MessagesProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route
              path="/chat"
              element={
                <PrivateRoute>
                  <ChatPage />
                </PrivateRoute>
              }
            />
            <Route path="/" element={<Navigate to="/chat" />} />
          </Routes>
        </BrowserRouter>
      </MessagesProvider>
    </ErrorBoundary>
  );
}

export default App;
