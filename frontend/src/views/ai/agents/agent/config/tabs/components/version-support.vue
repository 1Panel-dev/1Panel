<template>
    <div class="version-support">
        <el-empty :description="t('aiTools.agents.versionUnsupportedTitle')">
            <template #image>
                <el-icon class="version-support-icon"><WarningFilled /></el-icon>
            </template>
            <p class="version-support-desc">
                {{ t('aiTools.agents.versionUnsupportedHelper', [minVersion]) }}
            </p>
        </el-empty>
    </div>
</template>

<script setup lang="ts">
import { WarningFilled } from '@element-plus/icons-vue';
import { useI18n } from 'vue-i18n';

defineProps<{
    minVersion: string;
}>();

const { t } = useI18n();
</script>

<style scoped lang="scss">
.version-support {
    min-height: 320px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.version-support-icon {
    font-size: 56px;
    color: var(--el-color-warning);
}

.version-support-desc {
    margin: 8px 0 0;
    color: var(--el-text-color-secondary);
}
</style>
