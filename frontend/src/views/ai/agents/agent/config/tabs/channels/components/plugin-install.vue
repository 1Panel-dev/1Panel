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
        <el-button type="primary" :loading="installing" :disabled="!hasManagePermission" @click="handleInstall">
            {{ t('commons.button.install') }}
        </el-button>
    </el-form-item>
    <div v-else class="plugin-install-status">
        <div class="plugin-install-status__row">
            <div class="plugin-install-status__text">
                <span class="plugin-install-status__label">{{ t('app.version') }}</span>
                <span class="plugin-install-status__value">{{ currentVersion || '-' }}</span>
            </div>
            <el-button
                type="danger"
                plain
                size="small"
                :loading="uninstalling"
                :disabled="!hasManagePermission"
                @click="handleUninstall"
            >
                {{ t('commons.button.uninstall') }}
            </el-button>
        </div>
        <div v-if="upgradable" class="plugin-install-status__row mt-4">
            <div class="plugin-install-status__text">
                <span class="plugin-install-status__label">{{ t('app.newVersion') }}</span>
                <span class="plugin-install-status__value">{{ latestVersion || '-' }}</span>
            </div>
            <el-button
                type="primary"
                size="small"
                :loading="upgrading"
                :disabled="!hasManagePermission"
                @click="handleUpgrade"
            >
                {{ t('commons.button.upgrade') }}
            </el-button>
        </div>
    </div>
    <TaskLog ref="taskLogRef" @close="handleTaskClose" />
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useMenuManagePermission } from '@/composables/useMenuManagePermission';
import { useI18n } from 'vue-i18n';
import TaskLog from '@/components/log/task/index.vue';

const props = defineProps<{
    installed: boolean;
    installing: boolean;
    upgrading: boolean;
    uninstalling: boolean;
    currentVersion: string;
    latestVersion: string;
    upgradable: boolean;
    installAction: () => Promise<string>;
    upgradeAction: () => Promise<string>;
    uninstallAction: () => Promise<string>;
    onTaskClose?: () => void | Promise<void>;
}>();

const { t } = useI18n();
const { hasManagePermission } = useMenuManagePermission();
const taskLogRef = ref();

const openTaskLog = (taskID: string) => {
    if (taskID) {
        taskLogRef.value?.openWithTaskID(taskID);
    }
};

const handleInstall = async () => {
    openTaskLog(await props.installAction());
};

const handleUpgrade = async () => {
    openTaskLog(await props.upgradeAction());
};

const handleUninstall = async () => {
    openTaskLog(await props.uninstallAction());
};

const handleTaskClose = async () => {
    await props.onTaskClose?.();
};
</script>

<style scoped lang="scss">
.plugin-install-alert {
    :deep(.el-alert__title),
    :deep(.el-alert__description) {
        font-size: var(--el-font-size-base);
        line-height: 1.5;
    }
}

.plugin-install-status__row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    width: 100%;
    max-width: 300px;
    padding: 8px 12px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    background-color: var(--el-fill-color-blank);
    color: var(--el-text-color-regular);
}

.plugin-install-status__text {
    display: flex;
    align-items: baseline;
    gap: 8px;
}

.plugin-install-status__label {
    color: var(--el-text-color-secondary);
}

.plugin-install-status__value {
    font-weight: 500;
    color: var(--el-text-color-primary);
}
</style>
