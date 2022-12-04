<template>
    <span :style="{ color: getColor(status) }">
        {{ $t('commons.status.' + status) }}
    </span>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';

const props = defineProps({
    status: {
        type: String,
        default: 'runnning',
    },
});
let status = ref('');

const getColor = (status: string) => {
    switch (status) {
        case 'running':
            return '#00c957';
        case 'error':
            return '#ff0000';
        default:
            return '';
    }
};

onMounted(() => {
    status.value = props.status.toLocaleLowerCase();
});
</script>
