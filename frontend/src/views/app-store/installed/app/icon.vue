<template>
    <div class="icon">
        <el-avatar
            @click="$emit('openDetail')"
            shape="square"
            :size="77"
            :src="getAppIconUrl(appKey, currentNode, icon)"
        >
            <SvgIcon iconName="p-appstore" />
        </el-avatar>
    </div>
</template>

<script lang="ts" setup>
import { getAppIconUrl } from '@/api/modules/app';

interface Props {
    currentNode: string;
    appKey: string;
    icon?: string;
}
defineProps<Props>();

defineEmits(['openDetail']);
</script>
