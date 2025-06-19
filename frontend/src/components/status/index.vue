<template>
    <el-tooltip v-if="msg && msg != ''" effect="dark" placement="bottom">
        <template #content>
            <div class="content">{{ msg }}</div>
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
    <span v-else>
        <el-tag size="small" :type="getType(statusItem)" round effect="light" v-if="!operate">
            <span class="flx-align-center">
                {{ $t('commons.status.' + statusItem) }}
                <el-icon v-if="loadingIcon(statusItem)" class="is-loading">
                    <Loading />
                </el-icon>
            </span>
        </el-tag>
        <el-button size="small" v-else :type="getType(statusItem)" plain round>
            {{ $t('commons.status.' + statusItem) }}
            <el-icon v-if="loadingIcon(statusItem)" class="is-loading">
                <Loading />
            </el-icon>
            <el-icon size="15" v-else class="status-icon">
                <svg-icon iconName="p-stop" v-if="statusItem == 'stopped' || statusItem == 'exited'"></svg-icon>
                <svg-icon iconName="p-start" class="svg-icon" v-if="statusItem == 'running'"></svg-icon>
            </el-icon>
        </el-button>
    </span>
</template>

<script lang="ts" setup>
import { computed } from 'vue';

const props = defineProps({
    status: String,
    msg: String,
    hasIcon: Boolean,
    operate: {
        type: Boolean,
        default: false,
        required: false,
    },
});

const statusItem = computed(() => {
    return props.status?.toLowerCase() || '';
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
        case 'unused':
        case 'executing':
            return 'success';
        case 'stopped':
        case 'exceptional':
        case 'disable':
        case 'unhealthy':
        case 'failed':
        case 'lost':
        case 'exited':
            return 'danger';
        case 'paused':
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
    'executing',
];

const loadingIcon = (status: string): boolean => {
    return loadingStatus.indexOf(status) > -1;
};
</script>

<style lang="scss" scoped>
.content {
    width: 300px;
    word-break: break-all;
}

.status-icon {
    width: 1em;
    height: 1em;
}
</style>
