<template>
    <div>
        <div class="source-font">CPU: {{ stats.cpuPercent.toFixed(2) }}%</div>
        <div class="float-left source-font">{{ $t('monitor.memory') }}: {{ stats.memoryPercent.toFixed(2) }}%</div>
        <el-popover placement="right" width="500px" class="float-right">
            <template #reference>
                <SvgIcon iconName="p-xiangqing" className="resource-detail-icon" />
            </template>
            <template #default>
                <el-descriptions direction="vertical" border :column="3" size="small">
                    <el-descriptions-item :label="$t('container.cpuUsage')">
                        {{ computeCPU(stats.cpuTotalUsage) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.cpuTotal')">
                        {{ computeCPU(stats.systemUsage) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.core')">{{ stats.percpuUsage }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('container.memUsage')">
                        {{ computeSizeForDocker(stats.memoryUsage) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.memCache')">
                        {{ computeSizeForDocker(stats.memoryCache) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.memTotal')">
                        {{ computeSizeForDocker(stats.memoryLimit) }}
                    </el-descriptions-item>
                    <slot />
                </el-descriptions>
                <slot name="footer" />
            </template>
        </el-popover>
    </div>
</template>

<script setup lang="ts">
import type { Container } from '@/api/interface/container';
import { computeCPU, computeSizeForDocker } from '@/utils/size';
import SvgIcon from '@/components/svg-icon/svg-icon.vue';
defineProps<{ stats: Container.ContainerListStats }>();
</script>

<style scoped lang="scss">
.source-font {
    font-size: 12px;
}
:deep(.resource-detail-icon) {
    margin-top: -3px;
    font-size: 6px;
    cursor: pointer;
}
</style>
