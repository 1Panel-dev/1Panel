<template>
    <div v-loading="loading">
        <HighlightLog v-model="content" :heightDiff="320" />
    </div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { getSupervisorLog } from '@/api/modules/host-tool';
import HighlightLog from '@/components/log/hightlight-log/index.vue';

let content = ref('');
let loading = ref(false);

const getConfig = async () => {
    const res = await getSupervisorLog();
    content.value = res.data;
};

onMounted(() => {
    getConfig();
});
</script>
