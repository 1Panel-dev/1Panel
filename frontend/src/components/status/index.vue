<template>
    <span
        class="status-wrapper"
        :class="{ 'is-disabled': isDisabled }"
        :aria-disabled="isDisabled"
        @click.capture="handlePermissionClick"
    >
        <el-tooltip v-if="msg && msg != ''" effect="dark" placement="bottom">
            <template #content>
                <div class="content">{{ msg }}</div>
            </template>
            <el-tag size="small" :type="getType(statusItem)" round effect="light">
                <span class="flx-align-center">
                    <span v-if="statusItem != ''">{{ $t('commons.status.' + statusItem) }}</span>
                    <el-icon v-if="loadingIcon(statusItem)" class="is-loading">
                        <Loading />
                    </el-icon>
                </span>
            </el-tag>
        </el-tooltip>
        <template v-else>
            <el-tag size="small" :type="getType(statusItem)" round effect="light" v-if="!operate">
                <span class="flx-align-center">
                    <span v-if="statusItem != ''">{{ $t('commons.status.' + statusItem) }}</span>
                    <el-icon v-if="loadingIcon(statusItem)" class="is-loading">
                        <Loading />
                    </el-icon>
                </span>
            </el-tag>
            <el-button size="small" v-else :type="getType(statusItem)" plain round :disabled="isDisabled">
                <span v-if="statusItem != ''">{{ $t('commons.status.' + statusItem) }}</span>
                <el-icon v-if="loadingIcon(statusItem)" class="is-loading">
                    <Loading />
                </el-icon>
                <svg-icon iconName="p-stop" className="status-icon" v-if="stopIcon(statusItem)"></svg-icon>
                <svg-icon iconName="p-start" className="status-icon" v-if="runningIcon(statusItem)"></svg-icon>
            </el-button>
        </template>
    </span>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';

const props = defineProps({
    status: String,
    msg: String,
    hasIcon: Boolean,
    disabled: Boolean,
    operate: {
        type: Boolean,
        default: false,
        required: false,
    },
});

const permissionDisabled = ref(false);

defineExpose({
    setPermissionDisabled: (disabled: boolean) => {
        permissionDisabled.value = disabled;
    },
});

const statusItem = computed(() => {
    return props.status?.toLowerCase() || '';
});

const isDisabled = computed(() => {
    return !!props.disabled || permissionDisabled.value;
});

const handlePermissionClick = (event: MouseEvent) => {
    if (!isDisabled.value) {
        return;
    }
    event.preventDefault();
    event.stopImmediatePropagation();
};

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
        case 'new':
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
        case 'conflict':
        case 'partial':
            return 'warning';
        case 'duplicate':
        case 'unexecuted':
        case 'canceled':
            return 'info';
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
    'loading',
];

const stopStatus = ['stopped', 'exited', 'disable'];
const runningStatus = ['running', 'enable'];

const loadingIcon = (status: string): boolean => {
    return loadingStatus.indexOf(status) > -1;
};
const stopIcon = (status: string): boolean => {
    return stopStatus.indexOf(status.toLocaleLowerCase()) > -1;
};
const runningIcon = (status: string): boolean => {
    return runningStatus.indexOf(status.toLocaleLowerCase()) > -1;
};
</script>

<style lang="scss" scoped>
.status-wrapper {
    display: inline-flex;
}

.status-wrapper.is-disabled {
    cursor: not-allowed;
}

.status-wrapper.is-disabled :deep(.el-tag) {
    opacity: 0.55;
    filter: grayscale(1);
}

.content {
    width: 300px;
    word-break: break-all;
}

.status-icon {
    width: 1em;
    height: 1em;
}
</style>
