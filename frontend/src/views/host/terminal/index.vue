<template>
    <div>
        <el-card class="router_card">
            <el-radio-group v-model="activeNames" @change="handleChange">
                <el-radio-button class="router_card_button" size="large" label="terminal">
                    {{ $t('menu.terminal') }}
                </el-radio-button>
                <el-radio-button class="router_card_button" size="large" label="host">
                    {{ $t('menu.host') }}
                </el-radio-button>
                <el-radio-button class="router_card_button" size="large" label="command">
                    {{ $t('terminal.quickCommand') }}
                </el-radio-button>
            </el-radio-group>
        </el-card>

        <div v-show="activeNames === 'terminal'">
            <TerminalTab ref="terminalTabRef" />
        </div>
        <div v-if="activeNames === 'host'">
            <HostTab ref="hostTabRef" />
        </div>
        <div v-if="activeNames === 'command'">
            <CommandTab ref="commandTabRef" />
        </div>
    </div>
</template>

<script setup lang="ts">
import HostTab from '@/views/host/terminal/host/index.vue';
import CommandTab from '@/views/host/terminal/command/index.vue';
import TerminalTab from '@/views/host/terminal/terminal/index.vue';
import { onMounted, onUnmounted, ref } from 'vue';

const activeNames = ref<string>('terminal');
const hostTabRef = ref();
const commandTabRef = ref();
const terminalTabRef = ref();

const handleChange = (tab: any) => {
    if (tab === 'host') {
        hostTabRef.value!.acceptParams();
    }
    if (tab === 'command') {
        commandTabRef.value!.acceptParams();
    }
    if (tab === 'terminal') {
        terminalTabRef.value!.acceptParams();
    }
};

onMounted(() => {
    handleChange('terminal');
});
onUnmounted(() => {
    terminalTabRef.value?.cleanTimer();
});
</script>

<style lang="scss">
.router_card {
    --el-card-border-radius: 8px;
    --el-card-padding: 0;
    padding: 0px;
    padding-bottom: 2px;
    padding-top: 2px;
}
.router_card_button {
    margin-left: 2px;
    .el-radio-button__inner {
        min-width: 100px;
        height: 100%;
        border: 0 !important;
    }

    .el-radio-button__original-radio:checked + .el-radio-button__inner {
        border-radius: 3px;
        color: $primary-color;
        background-color: var(--panel-button-active);
        box-shadow: 0 0 0 2px $primary-color !important;
    }

    .el-radio-button:first-child .el-radio-button__inner {
        border-radius: 3px;
        color: $primary-color;
        background-color: var(--panel-button-active);
        box-shadow: 0 0 0 2px $primary-color !important;
    }
}
</style>
