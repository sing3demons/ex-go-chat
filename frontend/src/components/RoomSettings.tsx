import { useState } from 'react';
import { api } from '../services/api';
import { useRoomStore } from '../store/roomStore';
import type { Room } from '../types';

interface RoomSettingsProps {
  room: Room;
  isOpen: boolean;
  onClose: () => void;
}

export const RoomSettings = ({ room, isOpen, onClose }: RoomSettingsProps) => {
  const [newMemberIds, setNewMemberIds] = useState('');
  const [removeMemberIds, setRemoveMemberIds] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const loadRooms = useRoomStore((state) => state.loadRooms);

  if (!isOpen) return null;

  const handleAddMembers = async () => {
    if (!newMemberIds.trim()) {
      setError('Please enter member IDs to add');
      return;
    }

    setIsLoading(true);
    setError(null);
    setSuccessMessage(null);

    try {
      const members = newMemberIds
        .split(',')
        .map(id => id.trim())
        .filter(id => id.length > 0);

      await api.addMembers(room.id, members);
      await loadRooms();
      
      setSuccessMessage(`Added ${members.length} member(s) successfully`);
      setNewMemberIds('');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to add members');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRemoveMembers = async () => {
    if (!removeMemberIds.trim()) {
      setError('Please enter member IDs to remove');
      return;
    }

    setIsLoading(true);
    setError(null);
    setSuccessMessage(null);

    try {
      const members = removeMemberIds
        .split(',')
        .map(id => id.trim())
        .filter(id => id.length > 0);

      await api.removeMembers(room.id, members);
      await loadRooms();
      
      setSuccessMessage(`Removed ${members.length} member(s) successfully`);
      setRemoveMemberIds('');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to remove members');
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setNewMemberIds('');
    setRemoveMemberIds('');
    setError(null);
    setSuccessMessage(null);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b sticky top-0 bg-white">
          <h2 className="text-xl font-semibold">Room Settings</h2>
          <button
            onClick={handleClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="p-6">
          {/* Room Info */}
          <div className="mb-6 p-4 bg-gray-50 rounded-lg">
            <h3 className="font-semibold text-lg mb-2">{room.name || 'Unnamed Room'}</h3>
            <div className="text-sm text-gray-600 space-y-1">
              <p>Type: <span className="font-medium">{room.type === 'direct' ? 'Direct' : 'Group'}</span></p>
              <p>Members: <span className="font-medium">{room.members.length}</span></p>
              <p className="text-xs text-gray-500 mt-2">Room ID: {room.id}</p>
            </div>
          </div>

          {/* Messages */}
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">
              {error}
            </div>
          )}
          
          {successMessage && (
            <div className="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded-lg text-sm">
              {successMessage}
            </div>
          )}

          {/* Current Members */}
          <div className="mb-6">
            <h3 className="font-semibold mb-3">Current Members</h3>
            <div className="border rounded-lg divide-y max-h-40 overflow-y-auto">
              {room.members.length === 0 ? (
                <div className="p-4 text-center text-gray-500 text-sm">
                  No members
                </div>
              ) : (
                room.members.map((memberId, index) => (
                  <div key={index} className="p-3 flex items-center justify-between hover:bg-gray-50">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center text-sm font-semibold">
                        {memberId.slice(0, 2).toUpperCase()}
                      </div>
                      <span className="text-sm font-mono">{memberId}</span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Add Members */}
          {room.type === 'group' && (
            <div className="mb-6">
              <h3 className="font-semibold mb-3">Add Members</h3>
              <div className="space-y-3">
                <textarea
                  value={newMemberIds}
                  onChange={(e) => setNewMemberIds(e.target.value)}
                  placeholder="Enter member IDs separated by commas&#10;Example: user1, user2, user3"
                  rows={2}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                  disabled={isLoading}
                />
                <button
                  onClick={handleAddMembers}
                  disabled={isLoading || !newMemberIds.trim()}
                  className="w-full px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
                >
                  {isLoading ? 'Adding...' : 'Add Members'}
                </button>
              </div>
            </div>
          )}

          {/* Remove Members */}
          {room.type === 'group' && (
            <div className="mb-6">
              <h3 className="font-semibold mb-3">Remove Members</h3>
              <div className="space-y-3">
                <textarea
                  value={removeMemberIds}
                  onChange={(e) => setRemoveMemberIds(e.target.value)}
                  placeholder="Enter member IDs to remove, separated by commas&#10;Example: user1, user2"
                  rows={2}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500 resize-none"
                  disabled={isLoading}
                />
                <button
                  onClick={handleRemoveMembers}
                  disabled={isLoading || !removeMemberIds.trim()}
                  className="w-full px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
                >
                  {isLoading ? 'Removing...' : 'Remove Members'}
                </button>
              </div>
            </div>
          )}

          {/* Leave Group */}
          {room.type === 'group' && (
            <div className="pt-6 border-t">
              <button
                onClick={() => {
                  if (window.confirm('Are you sure you want to leave this group?')) {
                    // TODO: Implement leave group functionality
                    alert('Leave group functionality not yet implemented');
                  }
                }}
                className="w-full px-4 py-2 bg-gray-100 text-red-600 rounded-lg hover:bg-red-50 transition-colors font-medium"
              >
                Leave Group
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
