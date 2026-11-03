<template>
    <div>
        <RouterButton :buttons="buttons" />
        <EnvironmentCheck :key="currentNode" v-slot="{ vmUnavailable, vmEnvironmentLoaded }">
            <RouterViewCache
                :route-props="{
                    vmUnavailable,
                    vmEnvironmentLoaded,
                }"
            />
        </EnvironmentCheck>
    </div>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';
import EnvironmentCheck from '@/views/vm/status/index.vue';

const { currentNode } = useGlobalStore();

const buttons = [
    {
        label: i18n.global.t('xpack.vm.overview'),
        path: '/vm/overview',
    },
    {
        label: i18n.global.t('xpack.vm.vm'),
        path: '/vm/list',
    },
    {
        label: i18n.global.t('xpack.vm.iso'),
        path: '/vm/isos',
    },
    {
        label: i18n.global.t('xpack.vm.template'),
        path: '/vm/templates',
    },
    {
        label: i18n.global.t('xpack.vm.network'),
        path: '/vm/networks',
    },
    {
        label: i18n.global.t('xpack.vm.storagePool'),
        path: '/vm/storage-pools',
    },
];
</script>
