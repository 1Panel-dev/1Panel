<template>
    <div v-loading="loading">
        <LogPro v-model="content" :heightDiff="320"></LogPro>
    </div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { getSupervisorLog } from '@/api/modules/host-tool';
import LogPro from '@/components/log-pro/index.vue';

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
