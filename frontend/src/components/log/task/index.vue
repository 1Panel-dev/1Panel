<template>
    <el-dialog
        v-model="open"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :show-close="showClose"
        @close="handleClose"
        :width="width"
    >
        <div v-if="open">
            <LogFile :config="config" :showTail="showTail"></LogFile>
        </div>
    </el-dialog>
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

const openWithTaskID = (id: string, tail: boolean, operateNode?: string) => {
    config.taskID = id;
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
    config.taskType = taskType;
    config.resourceID = resourceID;
    config.taskOperate = taskOperate;
    config.operateNode = operateNode || '';
    open.value = true;
};

const em = defineEmits(['close']);
const handleClose = () => {
    em('close', true);
    open.value = false;
    bus.emit('refreshTask', true);
    bus.emit('refreshApp', true);
};

defineExpose({ openWithResourceID, openWithTaskID });
</script>
