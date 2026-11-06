<template>
    <CachedTabs
        :items="tabs"
        default-value="terminal"
        cache-key="terminal/main"
        :requested-value="route.path === '/terminal' && route.query.path ? 'terminal' : undefined"
        @activate="handleChange"
    >
        <template #default="{ active }">
            <div v-show="active === 'terminal'">
                <TerminalTab ref="terminalTabRef" />
            </div>
            <HostTab v-if="isAdmin && active === 'host'" ref="hostTabRef" />
            <CommandTab v-if="active === 'command'" ref="commandTabRef" />
            <SettingTab v-if="isAdmin && active === 'setting'" ref="settingTabRef" />
        </template>
    </CachedTabs>
</template>

<script setup lang="ts">
import HostTab from '@/views/terminal/host/index.vue';
import CommandTab from '@/views/terminal/command/index.vue';
import TerminalTab from '@/views/terminal/terminal/index.vue';
import SettingTab from '@/views/terminal/setting/index.vue';
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import CachedTabs from '@/components/cached-tabs/index.vue';
import i18n from '@/lang';
import { getTerminalInfo } from '@/api/modules/setting';
import { TerminalStore } from '@/store';
import { useGlobalStore } from '@/composables/useGlobalStore';

const terminalStore = TerminalStore();
const { isAdmin } = useGlobalStore();
const route = useRoute();
const tabs = computed(() => [
    { value: 'terminal', label: i18n.global.t('menu.terminal', 2) },
    ...(isAdmin.value ? [{ value: 'host', label: i18n.global.t('terminal.host', 2) }] : []),
    { value: 'command', label: i18n.global.t('terminal.quickCommand', 2) },
    ...(isAdmin.value ? [{ value: 'setting', label: i18n.global.t('container.setting') }] : []),
]);
const hostTabRef = ref();
const commandTabRef = ref();
const terminalTabRef = ref();
const settingTabRef = ref();

const handleChange = (tab: string) => {
    if (tab === 'host' && isAdmin.value) {
        hostTabRef.value?.acceptParams();
    }
    if (tab === 'command') {
        commandTabRef.value?.acceptParams();
    }
    if (tab === 'terminal') {
        terminalTabRef.value?.acceptParams();
    }
    if (tab === 'setting' && isAdmin.value) {
        settingTabRef.value?.acceptParams();
    }
};

const loadTerminalSetting = async () => {
    await getTerminalInfo().then((res) => {
        terminalStore.$patch({
            showTerminalButton: res.data.showTerminalButton !== 'Disable',
            lineHeight: Number(res.data.lineHeight),
            letterSpacing: Number(res.data.letterSpacing),
            fontSize: Number(res.data.fontSize),
            fontFamily: res.data.fontFamily || "Monaco, Menlo, Consolas, 'Courier New', monospace",
            backgroundColor: res.data.backgroundColor || '#000000',
            foregroundColor: res.data.foregroundColor || '#f5f5f5',
            cursorBlink: res.data.cursorBlink,
            cursorStyle: res.data.cursorStyle,
            scrollback: Number(res.data.scrollback),
            scrollSensitivity: Number(res.data.scrollSensitivity),
        });
    });
};

onMounted(() => {
    loadTerminalSetting();
});
</script>
