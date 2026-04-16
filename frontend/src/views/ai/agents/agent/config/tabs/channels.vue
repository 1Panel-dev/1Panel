<template>
    <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <template v-if="agentType === 'hermes-agent'">
            <el-tab-pane :label="t('aiTools.agents.weixin')" name="weixin">
                <HermesWeixinTab ref="hermesWeixinRef" :agent-id="agentId" />
            </el-tab-pane>
            <el-tab-pane label="QQ" name="qqbot">
                <HermesQQBotTab ref="hermesQQBotRef" />
            </el-tab-pane>
            <el-tab-pane :label="t('aiTools.agents.wecom')" name="wecom">
                <HermesWecomTab ref="hermesWecomRef" />
            </el-tab-pane>
            <el-tab-pane :label="t('aiTools.agents.dingtalk')" name="dingtalk">
                <HermesDingTalkTab ref="hermesDingTalkRef" />
            </el-tab-pane>
            <el-tab-pane :label="t('aiTools.agents.feishu')" name="feishu">
                <HermesFeishuTab ref="hermesFeishuRef" />
            </el-tab-pane>
            <el-tab-pane label="Telegram" name="telegram">
                <HermesTelegramTab ref="hermesTelegramRef" />
            </el-tab-pane>
            <el-tab-pane label="Discord" name="discord">
                <HermesDiscordTab ref="hermesDiscordRef" />
            </el-tab-pane>
        </template>
        <template v-else>
            <el-tab-pane :label="t('aiTools.agents.weixin')" name="weixin">
                <WeixinTab ref="weixinRef" :app-version="appVersion" />
            </el-tab-pane>
            <el-tab-pane label="QQ" name="qqbot">
                <QQBotTab ref="qqbotRef" />
            </el-tab-pane>
            <el-tab-pane :label="t('aiTools.agents.wecom')" name="wecom">
                <WecomTab ref="wecomRef" />
            </el-tab-pane>
            <el-tab-pane :label="t('aiTools.agents.dingtalk')" name="dingtalk">
                <DingTalkTab ref="dingtalkRef" :app-version="appVersion" />
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
        </template>
    </el-tabs>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import FeishuTab from './channels/openclaw/feishu.vue';
import HermesTelegramTab from './channels/hermes/telegram.vue';
import HermesDiscordTab from './channels/hermes/discord.vue';
import HermesWecomTab from './channels/hermes/wecom.vue';
import HermesWeixinTab from './channels/hermes/weixin.vue';
import HermesQQBotTab from './channels/hermes/qq.vue';
import HermesDingTalkTab from './channels/hermes/dingtalk.vue';
import HermesFeishuTab from './channels/hermes/feishu.vue';
import TelegramTab from './channels/openclaw/telegram.vue';
import DiscordTab from './channels/openclaw/discord.vue';
import QQBotTab from './channels/openclaw/qq.vue';
import WeixinTab from './channels/openclaw/weixin.vue';
import WecomTab from './channels/openclaw/wecom.vue';
import DingTalkTab from './channels/openclaw/dingtalk.vue';

const props = defineProps<{
    appVersion: string;
    agentType: AI.AgentType;
}>();
const { t } = useI18n();
const activeTab = ref('weixin');
const agentId = ref(0);
const feishuRef = ref();
const hermesWecomRef = ref();
const hermesWeixinRef = ref();
const hermesQQBotRef = ref();
const hermesDingTalkRef = ref();
const hermesFeishuRef = ref();
const hermesTelegramRef = ref();
const hermesDiscordRef = ref();
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
    if (activeTab.value === 'wecom' && hermesWecomRef.value) {
        await hermesWecomRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'weixin' && hermesWeixinRef.value) {
        await hermesWeixinRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'qqbot' && hermesQQBotRef.value) {
        await hermesQQBotRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'dingtalk' && hermesDingTalkRef.value) {
        await hermesDingTalkRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'feishu' && hermesFeishuRef.value) {
        await hermesFeishuRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'discord' && hermesDiscordRef.value) {
        await hermesDiscordRef.value?.load(agentId.value);
        return;
    }
    if (activeTab.value === 'telegram' && hermesTelegramRef.value) {
        await hermesTelegramRef.value?.load(agentId.value);
        return;
    }
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
    if (
        props.agentType === 'hermes-agent' &&
        activeTab.value !== 'wecom' &&
        activeTab.value !== 'weixin' &&
        activeTab.value !== 'qqbot' &&
        activeTab.value !== 'dingtalk' &&
        activeTab.value !== 'feishu' &&
        activeTab.value !== 'telegram' &&
        activeTab.value !== 'discord'
    ) {
        activeTab.value = 'weixin';
    }
    await loadCurrentTab();
};

defineExpose({
    load,
});
</script>
