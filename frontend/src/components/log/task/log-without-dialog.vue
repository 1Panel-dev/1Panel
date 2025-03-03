<template>
    <div>
        <LogFile :config="config" :showTail="false" :showDownload="false" :heightDiff="heightDiff"></LogFile>
    </div>
</template>
<script lang="ts" setup>
import { reactive } from 'vue';

const props = defineProps({
    taskID: String,
    heightDiff: Number,
});

const config = reactive({
    taskID: props.taskID,
    type: 'task',
    taskOperate: '',
    resourceID: 0,
    taskType: '',
    tail: false,
});
</script>
