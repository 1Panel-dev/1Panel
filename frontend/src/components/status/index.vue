<template>
    <el-tag :type="getType(status)" round effect="light">
        {{ $t('commons.status.' + status) }}
        <el-icon v-if="status === 'installing'" class="is-loading">
            <Loading />
        </el-icon>
    </el-tag>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';

const props = defineProps({
    status: {
        type: String,
        default: 'runnning',
    },
});
let status = ref('running');

const getType = (status: string) => {
    switch (status) {
        case 'running':
            return 'success';
        case 'error':
            return 'danger';
        case 'stopped':
            return 'warning';
        default:
            return '';
    }
};

onMounted(() => {
    status.value = props.status.toLocaleLowerCase();
});
</script>
