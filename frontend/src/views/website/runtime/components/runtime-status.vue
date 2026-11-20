<template>
    <el-popover v-if="showMessage" placement="bottom" :width="400" trigger="hover" :content="message">
        <template #reference>
            <Status :key="row.status" :status="row.status" />
        </template>
    </el-popover>

    <div v-else>
        <Status :key="row.status" :status="row.status" />
    </div>
</template>

<script setup>
import { computed } from 'vue';
import i18n from '@/lang';
import Status from '@/components/status/index.vue';

const props = defineProps({
    row: {
        type: Object,
        required: true,
    },
});

const message = computed(() => props.row.message?.trim() || i18n.global.t('runtime.statusMessageUnavailable'));

const showMessage = computed(() => {
    return props.row.status === 'Error' || props.row.status === 'SystemRestart';
});
</script>
