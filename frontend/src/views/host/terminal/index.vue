<template>
    <div>
        <el-card class="topRouterCard">
            <el-radio-group @change="handleChange" v-model="activeNames">
                <el-radio-button class="topRouterButton" size="default" label="terminal">
                    {{ $t('menu.terminal') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="host">
                    {{ $t('menu.host') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="command">
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
