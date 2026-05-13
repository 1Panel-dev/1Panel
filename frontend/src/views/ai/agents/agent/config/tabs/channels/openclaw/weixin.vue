<template>
    <VersionSupport v-if="!supported" :min-version="openclawMinSupportedVersion" />
    <el-form v-else v-loading="loading" label-position="top">
        <PluginInstall
            :installed="installed"
            :installing="installing"
            :upgrading="upgrading"
            :uninstalling="uninstalling"
            :current-version="currentVersion"
            :latest-version="latestVersion"
            :upgradable="upgradable"
            :install-action="installPlugin"
            :upgrade-action="upgradePlugin"
            :uninstall-action="uninstallPlugin"
            :on-task-close="reload"
        />
        <template v-if="installed">
            <el-form-item class="mt-4">
                <el-button
                    v-permission
                    type="primary"
                    :loading="loggingIn"
                    :disabled="!installed"
                    @click="loginChannel"
                >
                    {{ t('aiTools.agents.scanConnect') }}
                </el-button>
            </el-form-item>
            <el-alert type="info" :closable="false" :title="t('aiTools.agents.scanConnectHelper')" />
        </template>
    </el-form>
    <TaskLog ref="loginTaskLogRef" @close="reload" />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { loginAgentWeixinChannel } from '@/api/modules/ai';
import { newUUID } from '@/utils/id';
import TaskLog from '@/components/log/task/index.vue';
import { isOpenclawCurrentHTTPVersion } from '@/utils/agent';
import PluginInstall from '../components/plugin-install.vue';
import VersionSupport from '../../components/version-support.vue';
import { useAgentPluginChannel } from './useAgentPluginChannel';

const openclawMinSupportedVersion = '2026.3.23';
const props = defineProps<{
    appVersion: string;
}>();

const { t } = useI18n();
const loggingIn = ref(false);
const loginTaskLogRef = ref();
const {
    agentId,
    loading,
    installed,
    installing,
    upgrading,
    uninstalling,
    currentVersion,
    latestVersion,
    upgradable,
    loadPlugin,
    installPlugin,
    upgradePlugin,
    uninstallPlugin,
} = useAgentPluginChannel('weixin');
const supported = computed(() => isOpenclawCurrentHTTPVersion(props.appVersion));

const load = async (id: number) => {
    if (!supported.value) {
        return;
    }
    await loadPlugin(id);
};

const reload = async () => {
    if (!agentId.value) {
        return;
    }
    await load(agentId.value);
};

const loginChannel = async () => {
    if (!supported.value || !agentId.value) {
        return;
    }
    const taskID = newUUID();
    loggingIn.value = true;
    try {
        await loginAgentWeixinChannel({
            agentId: agentId.value,
            taskID,
        });
        loginTaskLogRef.value?.openWithTaskID(taskID);
    } finally {
        loggingIn.value = false;
    }
};

defineExpose({
    load,
});
</script>
