<template>
    <div class="network-page">
        <RouterButton class="shrink-0" :buttons="buttons" />
        <Topology />
    </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import i18n from '@/lang';
import Topology from './topology.vue';

const buttons = computed<RouterButton[]>(() => [
    { label: i18n.global.t('networkManage.title'), name: 'NetworkManage' },
]);
</script>

<style scoped>
@media (min-width: 768px) {
    .network-page {
        display: flex;
        flex-direction: column;
        height: 100%;
    }
}
</style>
