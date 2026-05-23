<template>
    <div>
        <el-card class="router_card">
            <el-radio-group v-model="activeNames" @change="handleChange">
                <el-radio-button class="router_card_button" size="large" value="terminal">
                    {{ $t('menu.terminal', 2) }}
                </el-radio-button>
                <el-radio-button v-if="isAdmin" class="router_card_button" size="large" value="host">
                    {{ $t('terminal.host', 2) }}
                </el-radio-button>
                <el-radio-button class="router_card_button" size="large" value="command">
                    {{ $t('terminal.quickCommand', 2) }}
                </el-radio-button>
                <el-radio-button v-if="isAdmin" class="router_card_button" size="large" value="setting">
                    {{ $t('container.setting') }}
                </el-radio-button>
            </el-radio-group>
        </el-card>

        <div v-show="activeNames === 'terminal'">
            <TerminalTab ref="terminalTabRef" />
        </div>
        <div v-if="isAdmin && activeNames === 'host'">
            <HostTab ref="hostTabRef" />
        </div>
        <div v-if="activeNames === 'command'">
            <CommandTab ref="commandTabRef" />
        </div>
        <div v-if="isAdmin && activeNames === 'setting'">
            <SettingTab ref="settingTabRef" />
        </div>
    </div>
</template>

<script setup lang="ts">
import HostTab from '@/views/terminal/host/index.vue';
import CommandTab from '@/views/terminal/command/index.vue';
import TerminalTab from '@/views/terminal/terminal/index.vue';
import SettingTab from '@/views/terminal/setting/index.vue';
import { onMounted, onUnmounted, ref } from 'vue';
import { getTerminalInfo } from '@/api/modules/setting';
import { TerminalStore } from '@/store';
import { useGlobalStore } from '@/composables/useGlobalStore';

const terminalStore = TerminalStore();
const { isAdmin } = useGlobalStore();
const activeNames = ref<string>('terminal');
const hostTabRef = ref();
const commandTabRef = ref();
const terminalTabRef = ref();
const settingTabRef = ref();

const handleChange = (tab: any) => {
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
    handleChange('terminal');
});
onUnmounted(() => {
    terminalTabRef.value?.cleanTimer();
});
</script>

<style lang="scss" scoped>
.router_card {
    --el-card-padding: 0;
}

.router_card_button {
    :deep(.el-radio-button__inner) {
        min-width: 100px;
        height: 100%;
        background-color: var(--panel-button-active) !important;
        box-shadow: none !important;
        outline: none !important;
        border: 2px solid transparent !important;
    }

    :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
        color: $primary-color;
        border-color: $primary-color !important;
        border-radius: 4px;
    }
}
</style>
