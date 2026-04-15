<template>
    <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <el-tab-pane :label="t('aiTools.agents.securityTab')" name="security">
            <SecurityTab ref="securityRef" />
        </el-tab-pane>
        <el-tab-pane :label="t('aiTools.agents.otherTab')" name="other">
            <OtherTab ref="otherRef" />
        </el-tab-pane>
        <el-tab-pane :label="t('website.source')" name="configFile">
            <ConfigFileTab ref="configFileRef" />
        </el-tab-pane>
    </el-tabs>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import SecurityTab from './settings/security.vue';
import OtherTab from './settings/other.vue';
import ConfigFileTab from './settings/config-file.vue';

const { t } = useI18n();
const activeTab = ref('security');
const agentId = ref(0);
const appVersion = ref('');
const securityRef = ref();
const otherRef = ref();
const configFileRef = ref();

const loadCurrentTab = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    if (activeTab.value === 'security') {
        await securityRef.value?.load(agentId.value, appVersion.value);
        return;
    }
    if (activeTab.value === 'other') {
        await otherRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'configFile') {
        await configFileRef.value?.load(agentId.value);
    }
};

const handleTabClick = async () => {
    await loadCurrentTab();
};

const load = async (params: { agentId: number; appVersion: string }) => {
    agentId.value = params.agentId;
    appVersion.value = params.appVersion;
    activeTab.value = 'security';
    await loadCurrentTab();
};

defineExpose({
    load,
});
</script>
