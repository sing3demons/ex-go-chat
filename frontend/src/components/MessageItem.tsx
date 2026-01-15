import { useState } from 'react';
import type { Message } from '../types';
import { useAuthStore } from '../store/authStore';
import { usePresence } from '../hooks/usePresence';
import { useChatStore } from '../store/chatStore';

interface MessageItemProps {
  message: Message;
  'data-message-id'?: string;
  'data-sender-id'?: string;
}

export const MessageItem = ({ message, ...props }: MessageItemProps) => {
  const { user } = useAuthStore();
  const { isUserOnline } = usePresence();
  const editMessage = useChatStore((state) => state.editMessage);
  const removeMessage = useChatStore((state) => state.removeMessage);
  const retryMessage = useChatStore((state) => state.retryMessage);
  
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(message.content);
  const [showActions, setShowActions] = useState(false);
  
  const isOwn = message.senderId === user?.id;
  const senderOnline = isUserOnline(message.senderId);

  // Handle edit submit
  const handleEditSubmit = () => {
    if (editContent.trim() && editContent !== message.content) {
      editMessage(message.roomId, message.id, editContent.trim());
    }
    setIsEditing(false);
  };

  // Handle delete
  const handleDelete = () => {
    if (window.confirm('Are you sure you want to delete this message?')) {
      removeMessage(message.roomId, message.id);
    }
  };

  // Handle retry
  const handleRetry = () => {
    if (message.tempId) {
      retryMessage(message.tempId);
    }
  };

  // Format timestamp
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  // Get delivery/read status icon
  const getStatusIcon = () => {
    if (!isOwn) return null;
    
    // Show pending indicator
    if (message.pending) {
      return (
        <div className="flex items-center gap-1" title="Sending...">
          <svg className="w-4 h-4 text-gray-300 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>
      );
    }

    // Show failed indicator
    if (message.failed) {
      return (
        <div className="flex items-center gap-1" title="Failed to send">
          <svg className="w-4 h-4 text-red-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd"/>
          </svg>
        </div>
      );
    }
    
    const statusValues = Object.values(message.status || {});
    
    if (statusValues.length === 0) {
      // No status yet - message just sent
      return (
        <div className="flex items-center gap-1" title="Sent">
          <svg className="w-4 h-4 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
            <path d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
          </svg>
        </div>
      );
    }

    const allRead = statusValues.every(s => s.read);
    const allDelivered = statusValues.every(s => s.delivered);
    const someRead = statusValues.some(s => s.read);
    const someDelivered = statusValues.some(s => s.delivered);
    
    if (allRead) {
      // All recipients have read
      return (
        <div className="flex items-center gap-1" title={`Read by ${statusValues.length} ${statusValues.length === 1 ? 'person' : 'people'}`}>
          <svg className="w-4 h-4 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
            <path d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
            <path d="M12.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-1-1a1 1 0 011.414-1.414l.293.293 7.293-7.293a1 1 0 011.414 0z"/>
          </svg>
        </div>
      );
    } else if (someRead) {
      // Some recipients have read
      const readCount = statusValues.filter(s => s.read).length;
      return (
        <div className="flex items-center gap-1" title={`Read by ${readCount} of ${statusValues.length}`}>
          <svg className="w-4 h-4 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
            <path d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
            <path d="M12.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-1-1a1 1 0 011.414-1.414l.293.293 7.293-7.293a1 1 0 011.414 0z"/>
          </svg>
        </div>
      );
    } else if (allDelivered) {
      // All recipients have received but not read
      return (
        <div className="flex items-center gap-1" title={`Delivered to ${statusValues.length} ${statusValues.length === 1 ? 'person' : 'people'}`}>
          <svg className="w-4 h-4 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
            <path d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
            <path d="M12.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-1-1a1 1 0 011.414-1.414l.293.293 7.293-7.293a1 1 0 011.414 0z"/>
          </svg>
        </div>
      );
    } else if (someDelivered) {
      // Some recipients have received
      const deliveredCount = statusValues.filter(s => s.delivered).length;
      return (
        <div className="flex items-center gap-1" title={`Delivered to ${deliveredCount} of ${statusValues.length}`}>
          <svg className="w-4 h-4 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
            <path d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
            <path d="M12.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-1-1a1 1 0 011.414-1.414l.293.293 7.293-7.293a1 1 0 011.414 0z"/>
          </svg>
        </div>
      );
    }
    
    // Default: sent but not delivered yet
    return (
      <div className="flex items-center gap-1" title="Sent">
        <svg className="w-4 h-4 text-gray-300" fill="currentColor" viewBox="0 0 20 20">
          <path d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
        </svg>
      </div>
    );
  };

  if (message.deleted) {
    return (
      <div className={`flex ${isOwn ? 'justify-end' : 'justify-start'} mb-4`}>
        <div className="bg-gray-200 text-gray-500 italic px-4 py-2 rounded-lg max-w-md">
          <div className="flex items-center gap-2">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            <span>Message deleted</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div 
      className={`flex ${isOwn ? 'justify-end' : 'justify-start'} mb-4 group animate-fadeIn`}
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
      {...props}
    >
      <div className={`flex items-end gap-2 max-w-md ${isOwn ? 'flex-row-reverse' : 'flex-row'} ${isOwn ? 'animate-slideInRight' : 'animate-slideInLeft'}`}>
        {/* Avatar */}
        {!isOwn && (
          <div className="relative">
            <div className="w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center text-sm font-semibold">
              {message.senderId.slice(0, 2).toUpperCase()}
            </div>
            {senderOnline && (
              <div className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-white rounded-full" />
            )}
          </div>
        )}

        {/* Message Content */}
        <div className="flex flex-col">
          {isEditing ? (
            <div className="bg-white border-2 border-blue-500 rounded-lg p-2">
              <input
                type="text"
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleEditSubmit();
                  if (e.key === 'Escape') setIsEditing(false);
                }}
                className="w-full px-2 py-1 border-none focus:outline-none"
                autoFocus
              />
              <div className="flex gap-2 mt-2">
                <button
                  onClick={handleEditSubmit}
                  className="text-xs px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600"
                >
                  Save
                </button>
                <button
                  onClick={() => setIsEditing(false)}
                  className="text-xs px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <div>
              <div
                className={`px-4 py-2 rounded-lg ${
                  isOwn 
                    ? message.failed
                      ? 'bg-red-100 text-red-900 border border-red-300 rounded-br-none'
                      : message.pending
                      ? 'bg-blue-400 text-white rounded-br-none opacity-70'
                      : 'bg-blue-500 text-white rounded-br-none'
                    : 'bg-white text-gray-800 border border-gray-200 rounded-bl-none'
                }`}
              >
                <p className="break-words whitespace-pre-wrap">{message.content}</p>
                
                <div className={`flex items-center gap-2 mt-1 text-xs ${
                  isOwn 
                    ? message.failed 
                      ? 'text-red-700' 
                      : message.pending
                      ? 'text-blue-200'
                      : 'text-blue-100'
                    : 'text-gray-500'
                }`}>
                  <span>{formatTime(message.createdAt)}</span>
                  {message.edited && <span>• edited</span>}
                  {message.pending && <span>• sending...</span>}
                  {message.failed && <span>• failed</span>}
                  {getStatusIcon()}
                </div>
              </div>

              {/* Retry button for failed messages */}
              {message.failed && isOwn && (
                <button
                  onClick={handleRetry}
                  className="mt-1 text-xs px-3 py-1 bg-red-500 text-white rounded hover:bg-red-600 flex items-center gap-1"
                >
                  <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Retry
                </button>
              )}
            </div>
          )}

          {/* Action Buttons */}
          {isOwn && showActions && !isEditing && !message.deleted && (
            <div className="flex gap-1 mt-1 justify-end">
              <button
                onClick={() => setIsEditing(true)}
                className="p-1 hover:bg-gray-200 rounded text-gray-600"
                title="Edit message"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button
                onClick={handleDelete}
                className="p-1 hover:bg-red-100 rounded text-red-600"
                title="Delete message"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
