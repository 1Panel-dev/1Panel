<template>
    <el-dialog
        v-model="open"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :show-close="showClose"
        :before-close="handleClose"
        :width="width"
    >
        <div>
            <LogFile :config="config" :showTail="showTail"></LogFile>
        </div>
    </el-dialog>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import bus from '@/global/bus';

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
    taskID: '',
    type: 'task',
    taskOperate: '',
    resourceID: 0,
    taskType: '',
    tail: true,
    colorMode: 'task',
});
const open = ref(false);
const showTail = ref(true);

const openWithTaskID = (id: string, tail: boolean) => {
    config.taskID = id;
    if (tail === undefined) {
        config.tail = true;
    } else {
        config.tail = tail;
    }
    open.value = true;
    bus.emit('refreshTask', true);
};

const openWithResourceID = (taskType: string, taskOperate: string, resourceID: number) => {
    config.taskType = taskType;
    config.resourceID = resourceID;
    config.taskOperate = taskOperate;
    open.value = true;
};

const handleClose = () => {
    open.value = false;
    bus.emit('refreshTask', true);
};

defineExpose({ openWithResourceID, openWithTaskID });
</script>
