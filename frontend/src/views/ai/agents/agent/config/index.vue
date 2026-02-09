<template>
    <DrawerPro v-model="open" :header="header" size="large" @close="handleClose">
        <el-tabs v-model="activeTab" tab-position="left" class="config-tabs" @tab-click="handleTabClick">
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

const { t } = useI18n();
const open = ref(false);
const activeTab = ref('channels');
const header = ref('');
const agentId = ref(0);
const channelsRef = ref();

const loadChannels = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    await channelsRef.value?.load(agentId.value);
};

const handleClose = () => {
    activeTab.value = 'channels';
};

const handleTabClick = async (pane: TabsPaneContext) => {
    if (pane.paneName === 'channels' && agentId.value > 0) {
        await loadChannels();
    }
};

const openDrawer = async (agent: AI.AgentItem) => {
    agentId.value = agent.id;
    header.value = `${agent.name} - ${t('aiTools.agents.configTitle')}`;
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
