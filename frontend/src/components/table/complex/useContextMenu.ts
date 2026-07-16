import { computed, onBeforeUnmount, ref, toValue, type MaybeRefOrGetter } from 'vue';
import { isOperationDisabled, isOperationVisible, type FuTableOperationButton } from '@/components/table/shared';

export const useContextMenu = <Row>(buttons: MaybeRefOrGetter<FuTableOperationButton<Row>[] | undefined>) => {
    const contextMenu = ref({
        visible: false,
        left: 0,
        top: 0,
        currentRow: null as Row | null,
    });

    const close = () => {
        contextMenu.value.visible = false;
        document.removeEventListener('click', close);
    };

    const open = (row: Row, event: MouseEvent) => {
        if (!toValue(buttons)?.length) {
            return;
        }
        event.preventDefault();
        contextMenu.value = {
            visible: true,
            left: event.clientX + 5,
            top: event.clientY,
            currentRow: row,
        };
        document.addEventListener('click', close);
    };

    const visibleButtons = computed(() =>
        (toValue(buttons) || []).filter((button) => isOperationVisible(button, contextMenu.value.currentRow as Row)),
    );
    const isDisabled = (button: FuTableOperationButton<Row>) =>
        isOperationDisabled(button, contextMenu.value.currentRow as Row);
    const click = (button: FuTableOperationButton<Row>) => {
        close();
        button.click?.(contextMenu.value.currentRow as Row);
    };

    onBeforeUnmount(close);

    return { contextMenu, open, close, visibleButtons, isDisabled, click };
};
