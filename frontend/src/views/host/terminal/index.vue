<template>
    <div>
        <el-card class="topCard">
            <el-radio-group @change="handleChange" v-model="activeNames">
                <el-radio-button class="topButton" size="large" label="terminal">
                    {{ $t('menu.terminal') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="host">
                    {{ $t('menu.host') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="command">
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

<style>
.topCard {
    --el-card-border-color: var(--el-border-color-light);
    --el-card-border-radius: 4px;
    --el-card-padding: 0px;
    --el-card-bg-color: var(--el-fill-color-blank);
}
.topButton .el-radio-button__inner {
    display: inline-block;
    line-height: 1;
    white-space: nowrap;
    vertical-align: middle;
    background: var(--el-button-bg-color, var(--el-fill-color-blank));
    border: 0;
    font-weight: 350;
    border-left: 0;
    color: var(--el-button-text-color, var(--el-text-color-regular));
    text-align: center;
    box-sizing: border-box;
    outline: 0;
    margin: 0;
    position: relative;
    cursor: pointer;
    transition: var(--el-transition-all);
    -webkit-user-select: none;
    user-select: none;
    padding: 8px 15px;
    font-size: var(--el-font-size-base);
    border-radius: 0;
}
</style>
