<template>
    <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <el-tab-pane :label="t('aiTools.agents.feishu')" name="feishu">
            <FeishuTab ref="feishuRef" />
        </el-tab-pane>
        <el-tab-pane label="Telegram" name="telegram">
            <TelegramTab ref="telegramRef" />
        </el-tab-pane>
    </el-tabs>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import FeishuTab from './channels/feishu.vue';
import TelegramTab from './channels/telegram.vue';

const { t } = useI18n();
const activeTab = ref('feishu');
const agentId = ref(0);
const feishuRef = ref();
const telegramRef = ref();

const loadCurrentTab = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    if (activeTab.value === 'telegram') {
        await telegramRef.value?.load(agentId.value);
        return;
    }
    await feishuRef.value?.load(agentId.value);
};

const handleTabClick = async () => {
    await loadCurrentTab();
};

const load = async (id: number) => {
    agentId.value = id;
    await loadCurrentTab();
};

defineExpose({
    load,
});
</script>
