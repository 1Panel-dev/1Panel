<template>
    <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <el-tab-pane :label="t('aiTools.agents.weixin')" name="weixin">
            <WeixinTab ref="weixinRef" />
        </el-tab-pane>
        <el-tab-pane label="QQ" name="qqbot">
            <QQBotTab ref="qqbotRef" />
        </el-tab-pane>
        <el-tab-pane :label="t('aiTools.agents.wecom')" name="wecom">
            <WecomTab ref="wecomRef" />
        </el-tab-pane>
        <el-tab-pane :label="t('aiTools.agents.dingtalk')" name="dingtalk">
            <DingTalkTab ref="dingtalkRef" />
        </el-tab-pane>
        <el-tab-pane :label="t('aiTools.agents.feishu')" name="feishu">
            <FeishuTab ref="feishuRef" />
        </el-tab-pane>
        <el-tab-pane label="Telegram" name="telegram">
            <TelegramTab ref="telegramRef" />
        </el-tab-pane>
        <el-tab-pane label="Discord" name="discord">
            <DiscordTab ref="discordRef" />
        </el-tab-pane>
    </el-tabs>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import FeishuTab from './channels/feishu.vue';
import TelegramTab from './channels/telegram.vue';
import DiscordTab from './channels/discord.vue';
import QQBotTab from './channels/qq.vue';
import WeixinTab from './channels/weixin.vue';
import WecomTab from './channels/wecom.vue';
import DingTalkTab from './channels/dingtalk.vue';

const { t } = useI18n();
const activeTab = ref('weixin');
const agentId = ref(0);
const feishuRef = ref();
const telegramRef = ref();
const discordRef = ref();
const qqbotRef = ref();
const weixinRef = ref();
const wecomRef = ref();
const dingtalkRef = ref();

const loadCurrentTab = async () => {
    if (agentId.value <= 0) {
        return;
    }
    await nextTick();
    if (activeTab.value === 'discord') {
        await discordRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'telegram') {
        await telegramRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'wecom') {
        await wecomRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'weixin') {
        await weixinRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'dingtalk') {
        await dingtalkRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'qqbot') {
        await qqbotRef.value?.load(agentId.value);
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
