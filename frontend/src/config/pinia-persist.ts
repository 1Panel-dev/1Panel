import { PersistenceOptions } from 'pinia-plugin-persistedstate';

const piniaPersistConfig = (key: string) => {
    const persist: PersistenceOptions = {
        key,
        storage: window.localStorage,
    };
    return persist;
};

export default piniaPersistConfig;
