<template>
    <el-alert
        v-if="!installed"
        class="plugin-install-alert"
        type="warning"
        :closable="false"
        :title="t('aiTools.agents.pluginNotInstalled')"
        :description="t('aiTools.agents.pluginInstallNPMRegistryHelper')"
    />
    <el-form-item v-if="!installed" class="mt-4">
        <el-button type="primary" :loading="installing" @click="emit('install')">
            {{ t('commons.button.install') }}
        </el-button>
    </el-form-item>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

defineProps<{
    installed: boolean;
    installing: boolean;
}>();

const emit = defineEmits<{
    install: [];
}>();

const { t } = useI18n();
</script>

<style scoped lang="scss">
.plugin-install-alert {
    :deep(.el-alert__title),
    :deep(.el-alert__description) {
        font-size: var(--el-font-size-base);
        line-height: 1.5;
    }
}
</style>
