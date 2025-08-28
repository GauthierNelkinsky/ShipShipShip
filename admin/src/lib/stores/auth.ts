import { writable } from "svelte/store";
import { api } from "$lib/api";

let isDemoMode = false;

export interface AuthState {
  isAuthenticated: boolean;
  loading: boolean;
  initialized: boolean;
  isDemoMode: boolean;
  user?: { username: string };
}

// Create the auth store with initial state
function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>({
    isAuthenticated: false,
    loading: true,
    initialized: false,
    isDemoMode: false,
  });

  return {
    subscribe,

    // Initialize authentication state
    async init() {
      update((state) => ({ ...state, loading: true }));

      try {
        // Check if demo mode is enabled
        const demoResponse = await api.checkDemoMode();
        isDemoMode = demoResponse.demo_mode;

        if (isDemoMode) {
          // In demo mode, bypass authentication
          set({
            isAuthenticated: true,
            loading: false,
            initialized: true,
            isDemoMode: true,
            user: { username: "Demo User" },
          });
          return true;
        }

        if (!api.isAuthenticated()) {
          set({
            isAuthenticated: false,
            loading: false,
            initialized: true,
            isDemoMode: false,
          });
          return false;
        }

        const result = await api.validateToken();
        set({
          isAuthenticated: true,
          loading: false,
          initialized: true,
          isDemoMode: false,
          user: { username: result.username },
        });
        return true;
      } catch (error) {
        api.clearToken();
        set({
          isAuthenticated: false,
          loading: false,
          initialized: true,
          isDemoMode: false,
        });
        return false;
      }
    },

    // Set authenticated state after login
    setAuthenticated(user?: { username: string }) {
      set({
        isAuthenticated: true,
        loading: false,
        initialized: true,
        isDemoMode: false,
        user,
      });
    },

    // Clear authentication state
    logout() {
      if (!isDemoMode) {
        api.clearToken();
      }
      set({
        isAuthenticated: false,
        loading: false,
        initialized: true,
        isDemoMode: false,
      });
    },

    // Reset to initial state
    reset() {
      set({
        isAuthenticated: false,
        loading: true,
        initialized: false,
        isDemoMode: false,
      });
    },
  };
}

export const authStore = createAuthStore();
