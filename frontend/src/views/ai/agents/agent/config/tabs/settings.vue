<template>
    <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <el-tab-pane v-if="agentType === 'openclaw'" :label="t('aiTools.agents.securityTab')" name="security">
            <SecurityTab ref="securityRef" />
        </el-tab-pane>
        <el-tab-pane :label="t('aiTools.agents.otherTab')" name="other">
            <OtherTab ref="otherRef" />
        </el-tab-pane>
    </el-tabs>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import SecurityTab from './settings/security.vue';
import OtherTab from './settings/other.vue';

const { t } = useI18n();
const activeTab = ref('security');
const agentId = ref(0);
const agentType = ref<AI.AgentItem['agentType']>('openclaw');
const securityRef = ref();
const otherRef = ref();

const loadCurrentTab = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    if (activeTab.value === 'security' && agentType.value === 'openclaw') {
        await securityRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'other') {
        await otherRef.value?.load(agentId.value);
    }
};

const handleTabClick = async () => {
    await loadCurrentTab();
};

const load = async (agent: AI.AgentItem) => {
    agentId.value = agent.id;
    agentType.value = agent.agentType;
    activeTab.value = agent.agentType === 'openclaw' ? 'security' : 'other';
    await loadCurrentTab();
};

defineExpose({
    load,
});
</script>
