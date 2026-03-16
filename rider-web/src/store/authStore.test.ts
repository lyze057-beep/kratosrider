import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useAuthStore } from './authStore';

vi.stubGlobal('localStorage', {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
});

describe('authStore', () => {
  beforeEach(() => {
    useAuthStore.setState({
      isLoggedIn: false,
      token: null,
      refreshToken: null,
      userInfo: null,
    });
    vi.clearAllMocks();
  });

  it('should initialize with default state', () => {
    const state = useAuthStore.getState();
    expect(state.isLoggedIn).toBe(false);
    expect(state.userInfo).toBeNull();
    expect(state.token).toBeNull();
    expect(state.refreshToken).toBeNull();
  });

  it('should set authentication state', () => {
    const mockUser = {
      id: 1,
      phone: '13800138000',
      nickname: 'Test User',
      avatar: '',
    };
    const mockToken = 'test-token';
    const mockRefreshToken = 'test-refresh-token';

    useAuthStore.getState().setAuth(mockToken, mockRefreshToken, mockUser);

    const state = useAuthStore.getState();
    expect(state.isLoggedIn).toBe(true);
    expect(state.userInfo).toEqual(mockUser);
    expect(state.token).toEqual(mockToken);
    expect(state.refreshToken).toEqual(mockRefreshToken);
    expect(localStorage.setItem).toHaveBeenCalledWith('token', mockToken);
    expect(localStorage.setItem).toHaveBeenCalledWith('refreshToken', mockRefreshToken);
    expect(localStorage.setItem).toHaveBeenCalledWith('userInfo', JSON.stringify(mockUser));
  });

  it('should update user info', () => {
    const mockUser = {
      id: 1,
      phone: '13800138000',
      nickname: 'Test User',
      avatar: '',
    };
    const mockToken = 'test-token';
    const mockRefreshToken = 'test-refresh-token';
    useAuthStore.getState().setAuth(mockToken, mockRefreshToken, mockUser);

    const updatedUser = {
      ...mockUser,
      nickname: 'Updated User',
    };
    useAuthStore.getState().setUserInfo(updatedUser);

    const state = useAuthStore.getState();
    expect(state.userInfo).toEqual(updatedUser);
    expect(localStorage.setItem).toHaveBeenCalledWith('userInfo', JSON.stringify(updatedUser));
  });

  it('should clear authentication state', () => {
    const mockUser = {
      id: 1,
      phone: '13800138000',
      nickname: 'Test User',
      avatar: '',
    };
    const mockToken = 'test-token';
    const mockRefreshToken = 'test-refresh-token';
    useAuthStore.getState().setAuth(mockToken, mockRefreshToken, mockUser);

    useAuthStore.getState().logout();

    const state = useAuthStore.getState();
    expect(state.isLoggedIn).toBe(false);
    expect(state.userInfo).toBeNull();
    expect(state.token).toBeNull();
    expect(state.refreshToken).toBeNull();
    expect(localStorage.removeItem).toHaveBeenCalledWith('token');
    expect(localStorage.removeItem).toHaveBeenCalledWith('refreshToken');
    expect(localStorage.removeItem).toHaveBeenCalledWith('userInfo');
  });
});
