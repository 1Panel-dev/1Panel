<template>
    <el-tag :type="getType(status)" round effect="light">
        <span class="flx-align-center">
            {{ $t('commons.status.' + status) }}
            <el-icon v-if="loadingIcon(status)" class="is-loading">
                <Loading />
            </el-icon>
        </span>
    </el-tag>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';

const props = defineProps({
    status: {
        type: String,
        default: 'running',
    },
});
let status = ref('running');

const getType = (status: string) => {
    if (status.includes('error') || status.includes('err')) {
        return 'danger';
    }
    switch (status) {
        case 'running':
            return 'success';
        case 'stopped':
            return 'danger';
        case 'unhealthy':
        case 'paused':
        case 'exited':
        case 'dead':
        case 'removing':
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
];

const loadingIcon = (status: string): boolean => {
    return loadingStatus.indexOf(status) > -1;
};

onMounted(() => {
    status.value = props.status.toLocaleLowerCase();
});
</script>
