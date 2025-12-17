<template>
    <div class="icon">
        <el-avatar @click="$emit('openDetail')" shape="square" :size="77" :src="getAppIconUrl(appID, currentNode)" />
    </div>
</template>

<script lang="ts" setup>
import { getAppIconUrl } from '@/api/modules/app';

interface Props {
    currentNode: string;
    appID: number;
}
defineProps<Props>();

defineEmits(['openDetail']);
</script>
