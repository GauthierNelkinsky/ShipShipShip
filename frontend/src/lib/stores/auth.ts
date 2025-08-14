import { writable } from 'svelte/store';
import { api } from '$lib/api';

export interface AuthState {
    isAuthenticated: boolean;
    loading: boolean;
    initialized: boolean;
    user?: { username: string };
}

// Create the auth store with initial state
function createAuthStore() {
    const { subscribe, set, update } = writable<AuthState>({
        isAuthenticated: false,
        loading: true,
        initialized: false,
    });

    return {
        subscribe,

        // Initialize authentication state
        async init() {
            update(state => ({ ...state, loading: true }));

            try {
                if (!api.isAuthenticated()) {
                    set({
                        isAuthenticated: false,
                        loading: false,
                        initialized: true,
                    });
                    return false;
                }

                const result = await api.validateToken();
                set({
                    isAuthenticated: true,
                    loading: false,
                    initialized: true,
                    user: { username: result.username },
                });
                return true;
            } catch (error) {
                api.clearToken();
                set({
                    isAuthenticated: false,
                    loading: false,
                    initialized: true,
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
                user,
            });
        },

        // Clear authentication state
        logout() {
            api.clearToken();
            set({
                isAuthenticated: false,
                loading: false,
                initialized: true,
            });
        },

        // Reset to initial state
        reset() {
            set({
                isAuthenticated: false,
                loading: true,
                initialized: false,
            });
        }
    };
}

export const authStore = createAuthStore();
