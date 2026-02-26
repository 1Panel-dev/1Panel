<template>
    <DrawerPro v-model="open" :header="header" size="large" @close="handleClose">
        <template #content>
            <el-tabs v-model="activeTab" tab-position="left" class="config-tabs" @tab-click="handleTabClick">
                <el-tab-pane :label="t('aiTools.agents.settingsTab')" name="settings">
                    <SettingsTab ref="settingsRef" />
                </el-tab-pane>
                <el-tab-pane :label="t('aiTools.model.model')" name="model">
                    <ModelTab ref="modelRef" @updated="handleModelUpdated" />
                </el-tab-pane>
                <el-tab-pane :label="t('aiTools.agents.channelsTab')" name="channels">
                    <ChannelsTab ref="channelsRef" />
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
import SettingsTab from './tabs/settings.vue';

const { t } = useI18n();
const emit = defineEmits(['updated']);
const open = ref(false);
const activeTab = ref('settings');
const header = ref('');
const agentId = ref(0);
const currentAgent = ref<AI.AgentItem>();
const channelsRef = ref();
const modelRef = ref();
const settingsRef = ref();

const loadSettings = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await settingsRef.value?.load(agentId.value);
};

const loadModel = async () => {
    if (!currentAgent.value) {
        return;
    }
    await nextTick();
    await modelRef.value?.load(currentAgent.value);
};

const loadChannels = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await channelsRef.value?.load(agentId.value);
};

const handleClose = () => {
    activeTab.value = 'settings';
};

const handleTabClick = async (pane: TabsPaneContext) => {
    if (pane.paneName === 'settings' && agentId.value > 0) {
        await loadSettings();
    }
    if (pane.paneName === 'model' && currentAgent.value) {
        await loadModel();
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
    currentAgent.value = agent;
    header.value = `${agent.name} - ${t('menu.config')}`;
    activeTab.value = 'settings';
    open.value = true;
    await loadSettings();
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
