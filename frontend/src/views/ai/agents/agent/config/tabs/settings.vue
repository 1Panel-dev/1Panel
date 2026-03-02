<template>
    <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <el-tab-pane :label="t('aiTools.agents.browserTab')" name="browser">
            <BrowserTab ref="browserRef" />
        </el-tab-pane>
        <el-tab-pane :label="t('aiTools.agents.otherTab')" name="other">
            <OtherTab ref="otherRef" />
        </el-tab-pane>
    </el-tabs>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import BrowserTab from './settings/browser.vue';
import OtherTab from './settings/other.vue';

const { t } = useI18n();
const activeTab = ref('browser');
const agentId = ref(0);
const browserRef = ref();
const otherRef = ref();

const loadCurrentTab = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    if (activeTab.value === 'browser') {
        await browserRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'other') {
        await otherRef.value?.load(agentId.value);
    }
};

const handleTabClick = async () => {
    await loadCurrentTab();
};

const load = async (id: number) => {
    agentId.value = id;
    activeTab.value = 'browser';
    await loadCurrentTab();
};

defineExpose({
    load,
});
</script>
