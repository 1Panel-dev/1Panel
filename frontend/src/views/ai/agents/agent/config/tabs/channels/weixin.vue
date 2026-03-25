<template>
    <VersionSupport v-if="!supported" :min-version="openclawMinSupportedVersion" />
    <el-form v-else label-position="top">
        <PluginInstall :installed="installed" :installing="installing" @install="installPlugin" />
        <el-alert type="info" :closable="false" :title="t('aiTools.agents.scanConnectHelper')" />
        <el-form-item class="mt-4">
            <el-button type="primary" :loading="loggingIn" :disabled="!installed" @click="loginChannel">
                {{ t('aiTools.agents.scanConnect') }}
            </el-button>
        </el-form-item>
    </el-form>
    <TaskLog ref="taskLogRef" @close="reload" />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { loginAgentWeixinChannel } from '@/api/modules/ai';
import { newUUID } from '@/utils/util';
import TaskLog from '@/components/log/task/index.vue';
import { isOpenclawCurrentHTTPVersion } from '@/utils/agent';
import PluginInstall from './components/plugin-install.vue';
import VersionSupport from '../components/version-support.vue';
import { useAgentPluginChannel } from './useAgentPluginChannel';

const openclawMinSupportedVersion = '2026.3.23';
const props = defineProps<{
    appVersion: string;
}>();

const { t } = useI18n();
const loggingIn = ref(false);
const { agentId, installed, installing, taskLogRef, loadPlugin, installPlugin } = useAgentPluginChannel('weixin');
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
        taskLogRef.value?.openWithTaskID(taskID);
    } finally {
        loggingIn.value = false;
    }
};

defineExpose({
    load,
});
</script>
