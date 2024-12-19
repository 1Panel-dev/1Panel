<template>
    <el-tooltip v-if="msg" effect="dark" placement="bottom">
        <template #content>
            <div style="width: 300px; word-break: break-all">{{ msg }}</div>
        </template>
        <el-tag size="small" :type="getType(statusItem)" round effect="light">
            <span class="flx-align-center">
                {{ $t('commons.status.' + statusItem) }}
                <el-icon v-if="loadingIcon(statusItem)" class="is-loading">
                    <Loading />
                </el-icon>
            </span>
        </el-tag>
    </el-tooltip>

    <el-tag size="small" v-else :type="getType(statusItem)" round effect="light">
        <span class="flx-align-center">
            {{ $t('commons.status.' + statusItem) }}
            <el-icon v-if="loadingIcon(statusItem)" class="is-loading">
                <Loading />
            </el-icon>
        </span>
    </el-tag>
</template>

<script lang="ts" setup>
import { computed } from 'vue';

const props = defineProps({
    status: String,
    msg: String,
});

const statusItem = computed(() => {
    return props.status.toLowerCase() || '';
});

const getType = (status: string) => {
    if (status.includes('error') || status.includes('err')) {
        return 'danger';
    }
    switch (status) {
        case 'running':
        case 'free':
        case 'success':
        case 'enable':
        case 'done':
        case 'healthy':
        case 'used':
            return 'success';
        case 'stopped':
        case 'exceptional':
        case 'disable':
        case 'unhealthy':
            return 'danger';
        case 'paused':
        case 'exited':
        case 'dead':
        case 'removing':
        case 'deleted':
            return 'warning';
        default:
            return 'primary';
    }
};

const loadingStatus = [
    'installing',
    'building',
    'restarting',
    'upgrading',
    'rebuilding',
    'recreating',
    'creating',
    'starting',
    'removing',
    'applying',
    'uninstalling',
    'downloading',
    'packing',
    'sending',
    'waiting',
];

const loadingIcon = (status: string): boolean => {
    return loadingStatus.indexOf(status) > -1;
};
</script>
