<template>
    <DrawerPro v-model="open" :header="header" size="large" @close="handleClose">
        <el-tabs v-model="activeTab" tab-position="left" class="config-tabs" @tab-click="handleTabClick">
            <el-tab-pane :label="t('aiTools.model.model')" name="model">
                <ModelTab ref="modelRef" @updated="handleModelUpdated" />
            </el-tab-pane>
            <el-tab-pane :label="t('aiTools.agents.channelsTab')" name="channels">
                <ChannelsTab ref="channelsRef" />
            </el-tab-pane>
        </el-tabs>
    </DrawerPro>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import type { TabsPaneContext } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import ChannelsTab from './tabs/channels.vue';
import ModelTab from './tabs/model.vue';

const { t } = useI18n();
const emit = defineEmits(['updated']);
const open = ref(false);
const activeTab = ref('model');
const header = ref('');
const agentId = ref(0);
const currentAgent = ref<AI.AgentItem>();
const channelsRef = ref();
const modelRef = ref();

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
    activeTab.value = 'model';
};

const handleTabClick = async (pane: TabsPaneContext) => {
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
    activeTab.value = 'model';
    open.value = true;
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
