<template>
    <DialogPro
        v-model="open"
        :show-close="showClose"
        :close-on-click-modal="true"
        @closed="handleClosed"
        :width="width"
    >
        <template #content>
            <div v-if="open">
                <LogFile :config="config" :showTail="showTail"></LogFile>
            </div>
        </template>
        <template #footer>
            <slot name="task-footer"></slot>
        </template>
    </DialogPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import bus from '@/global/bus';
import LogFile from '@/components/log/file/index.vue';

defineProps({
    showClose: {
        type: Boolean,
        default: true,
    },
    width: {
        type: String,
        default: '50%',
    },
    tail: {
        type: Boolean,
        default: true,
    },
});

const config = reactive({
    id: 0,
    name: '',
    taskID: '',
    type: 'task',
    taskOperate: '',
    resourceID: 0,
    taskType: '',
    tail: true,
    colorMode: 'task',

    operateNode: '',
});
const open = ref(false);
const showTail = ref(true);

const openWithTaskID = (id: string, tail?: boolean, operateNode?: string) => {
    config.taskID = id;
    config.type = 'task';
    config.taskType = '';
    config.taskOperate = '';
    config.resourceID = 0;
    if (tail === undefined) {
        config.tail = true;
    } else {
        config.tail = tail;
    }
    config.operateNode = operateNode || '';
    open.value = true;
    bus.emit('refreshTask', true);
};

const openWithResourceID = (taskType: string, taskOperate: string, resourceID: number, operateNode?: string) => {
    config.taskID = '';
    config.type = 'task';
    config.taskType = taskType;
    config.resourceID = resourceID;
    config.taskOperate = taskOperate;
    config.operateNode = operateNode || '';
    open.value = true;
};

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
};

const handleClosed = () => {
    em('close', true);
    bus.emit('refreshTask', true);
    bus.emit('refreshApp', true);
};

defineExpose({ openWithResourceID, openWithTaskID, handleClose });
</script>
