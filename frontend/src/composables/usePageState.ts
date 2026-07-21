import { reactive, type UnwrapNestedRefs } from 'vue';
import { useRoute } from 'vue-router';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { getPageState } from '@/utils/page-state-cache';

export const usePageState = <T extends object>(factory: () => T): UnwrapNestedRefs<T> => {
    const route = useRoute();
    const { currentNode } = useGlobalStore();
    const key = `${currentNode.value}:${String(route.name || route.path)}`;
    return reactive(getPageState(key, factory));
};
