<template>
    <el-tooltip placement="top-start">
        <template #content>
            <div class="info-break" :style="{ width: width + 'px' }">{{ info }}</div>
        </template>
        <div class="info-hidden" :style="{ width: width + 'px' }">{{ info }}</div>
    </el-tooltip>
</template>
<script lang="ts" setup>
defineProps({
    width: {
        type: String,
        default: '100',
    },
    info: {
        type: String,
        default: '',
    },
});
defineOptions({ name: 'MsgInfo' });
</script>

<style scoped>
.info-hidden {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.info-break {
    width: 100px;
    word-break: break-all;
    word-wrap: break-word;
}
</style>
