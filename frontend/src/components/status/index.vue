<template>
    <el-tag :type="getType(status)" round effect="light">
        <!-- <span :style="{ color: getColor(status) }"> -->
        {{ $t('commons.status.' + status) }}
        <!-- </span> -->
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

// const getColor = (status: string) => {
//     switch (status) {
//         case 'running':
//             return '#00c957';
//         case 'error':
//             return '#ff0000';
//         default:
//             return '';
//     }
// };

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
