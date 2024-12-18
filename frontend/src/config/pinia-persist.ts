import { PersistedStateOptions } from 'pinia-plugin-persistedstate';

const piniaPersistConfig = (key: string) => {
    const persist: PersistedStateOptions = {
        key,
        storage: window.localStorage,
    };
    return persist;
};

export default piniaPersistConfig;
