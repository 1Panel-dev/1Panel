<template>
    <DrawerPro v-model="open" :header="header" size="large" @close="handleClose">
        <template #content>
            <el-tabs v-model="activeTab" tab-position="left" class="config-tabs" @tab-click="handleTabClick">
                <el-tab-pane
                    v-if="agentType === 'openclaw' || agentType === 'hermes-agent'"
                    :label="t('aiTools.agents.channelsTab')"
                    name="channels"
                >
                    <ChannelsTab ref="channelsRef" :app-version="appVersion" :agent-type="agentType" />
                </el-tab-pane>
                <el-tab-pane :label="t('aiTools.model.model')" name="model">
                    <ModelTab ref="modelRef" @updated="handleModelUpdated" />
                </el-tab-pane>
                <el-tab-pane v-if="agentType === 'openclaw'" :label="t('aiTools.agents.agentRoleTab')" name="agent">
                    <AgentTab ref="agentRef" />
                </el-tab-pane>
                <el-tab-pane
                    v-if="agentType === 'openclaw' || agentType === 'hermes-agent'"
                    :label="t('aiTools.agents.skillsTab')"
                    name="skills"
                >
                    <SkillsTab ref="skillsRef" :app-version="appVersion" :agent-type="agentType" />
                </el-tab-pane>
                <el-tab-pane v-if="agentType === 'openclaw'" :label="t('aiTools.agents.pluginsTab')" name="plugins">
                    <PluginsTab ref="pluginsRef" />
                </el-tab-pane>
                <el-tab-pane
                    v-if="agentType === 'openclaw' || agentType === 'hermes-agent'"
                    :label="t('file.setting')"
                    name="settings"
                >
                    <SettingsTab ref="settingsRef" />
                </el-tab-pane>
            </el-tabs>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import type { TabsPaneContext } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import ChannelsTab from './tabs/channels.vue';
import ModelTab from './tabs/model.vue';
import AgentTab from './tabs/agents/index.vue';
import SkillsTab from './tabs/skills.vue';
import PluginsTab from './tabs/plugins.vue';
import SettingsTab from './tabs/settings.vue';

const { t } = useI18n();
const emit = defineEmits(['updated']);
const open = ref(false);
const activeTab = ref('channels');
const header = ref('');
const agentId = ref(0);
const accountId = ref(0);
const model = ref('');
const appVersion = ref('');
const configPath = ref('');
const agentType = ref<AI.AgentType>('openclaw');
const channelsRef = ref();
const modelRef = ref();
const agentRef = ref();
const skillsRef = ref();
const pluginsRef = ref();
const settingsRef = ref();

const loadSettings = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await settingsRef.value?.load({
        agentId: agentId.value,
        appVersion: appVersion.value,
        agentType: agentType.value,
    });
};

const loadModel = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await modelRef.value?.load({
        agentId: agentId.value,
        agentType: agentType.value,
    });
};

const loadChannels = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await channelsRef.value?.load(agentId.value);
};

const loadAgent = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await agentRef.value?.load({
        agentId: agentId.value,
        agentType: agentType.value,
        accountId: accountId.value,
        model: model.value,
        configPath: configPath.value,
    });
};

const loadSkills = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await skillsRef.value?.load(agentId.value);
};

const loadPlugins = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await pluginsRef.value?.load(agentId.value);
};

const handleClose = () => {
    activeTab.value = 'channels';
};

const handleTabClick = async (pane: TabsPaneContext) => {
    if (pane.paneName === 'settings' && agentId.value > 0) {
        await loadSettings();
    }
    if (pane.paneName === 'model') {
        await loadModel();
    }
    if (pane.paneName === 'skills') {
        await loadSkills();
    }
    if (pane.paneName === 'plugins') {
        await loadPlugins();
    }
    if (pane.paneName === 'agent') {
        await loadAgent();
    }
    if (pane.paneName === 'channels' && agentId.value > 0) {
        await loadChannels();
    }
};

const handleModelUpdated = () => {
    emit('updated');
};

const openDrawer = async (agent: AI.AgentItem) => {
    agentId.value = agent.id;
    accountId.value = agent.accountId;
    model.value = agent.model;
    appVersion.value = agent.appVersion;
    configPath.value = agent.configPath;
    agentType.value = agent.agentType;
    header.value = `${agent.name} - ${t('menu.config')}`;
    activeTab.value = agent.agentType === 'openclaw' || agent.agentType === 'hermes-agent' ? 'channels' : 'model';
    open.value = true;
    if (agent.agentType === 'openclaw' || agent.agentType === 'hermes-agent') {
        await loadChannels();
        return;
    }
    await loadModel();
};

defineExpose({
    open: openDrawer,
});
</script>

<style scoped lang="scss">
.config-tabs {
    min-height: 440px;
}
</style>
