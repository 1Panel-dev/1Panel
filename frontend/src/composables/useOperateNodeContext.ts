import { onUnmounted, watch, type ComputedRef, type Ref } from 'vue';
import { clearOperateNodeOverride, setOperateNodeOverride } from '@/utils/operate-node';

export const useOperateNodeContext = (node: Ref<string> | ComputedRef<string>) => {
    let activeNode = '';
    const stop = watch(
        node,
        (value) => {
            activeNode = value || '';
            setOperateNodeOverride(activeNode);
        },
        { immediate: true },
    );

    onUnmounted(() => {
        stop();
        clearOperateNodeOverride(activeNode);
    });
};
