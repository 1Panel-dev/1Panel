import { defineStore } from 'pinia';
import { AuthState } from '../interface';
import piniaPersistConfig from '@/config/piniaPersist';

export const AuthStore = defineStore({
    id: 'AuthState',
    state: (): AuthState => ({
        authRouter: [],
    }),
    getters: {
        dynamicRouter: (state) => {
            return state.authRouter;
        },
    },
    actions: {
        async setAuthRouter(dynamicRouter: string[]) {
            this.authRouter = dynamicRouter;
        },
    },
    persist: piniaPersistConfig('AuthState'),
});
