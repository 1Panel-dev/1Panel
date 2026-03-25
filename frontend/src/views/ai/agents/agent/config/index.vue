<template>
    <DrawerPro v-model="open" :header="header" size="large" @close="handleClose">
        <template #content>
            <el-tabs v-model="activeTab" tab-position="left" class="config-tabs" @tab-click="handleTabClick">
                <el-tab-pane :label="t('aiTools.agents.channelsTab')" name="channels">
                    <ChannelsTab ref="channelsRef" />
                </el-tab-pane>
                <el-tab-pane :label="t('aiTools.model.model')" name="model">
                    <ModelTab ref="modelRef" @updated="handleModelUpdated" />
                </el-tab-pane>
                <el-tab-pane :label="t('aiTools.agents.skillsTab')" name="skills">
                    <SkillsTab ref="skillsRef" />
                </el-tab-pane>
                <el-tab-pane :label="t('file.setting')" name="settings">
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
import SkillsTab from './tabs/skills.vue';
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
const channelsRef = ref();
const modelRef = ref();
const skillsRef = ref();
const settingsRef = ref();

const loadSettings = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await settingsRef.value?.load({
        agentId: agentId.value,
        appVersion: appVersion.value,
    });
};

const loadModel = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await modelRef.value?.load({
        agentId: agentId.value,
        accountId: accountId.value,
        model: model.value,
    });
};

const loadChannels = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await channelsRef.value?.load(agentId.value);
};

const loadSkills = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await skillsRef.value?.load(agentId.value);
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
    header.value = `${agent.name} - ${t('menu.config')}`;
    activeTab.value = 'channels';
    open.value = true;
    await loadChannels();
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
