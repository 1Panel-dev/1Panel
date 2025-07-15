import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';

export const useGlobalStore = () => {
    const globalStore = GlobalStore();
    const storeRefs = storeToRefs(globalStore);

    return {
        globalStore,
        ...storeRefs,
    };
};
