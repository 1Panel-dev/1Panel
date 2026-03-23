<template>
    <el-alert
        v-if="!installed"
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
