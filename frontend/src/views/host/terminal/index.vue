<template>
    <div>
        <el-card class="router_card">
            <el-radio-group v-model="activeNames" @change="handleChange">
                <el-radio-button class="router_card_button" size="large" value="terminal">
                    {{ $t('menu.terminal', 2) }}
                </el-radio-button>
                <el-radio-button class="router_card_button" size="large" value="host">
                    {{ $t('menu.host', 2) }}
                </el-radio-button>
                <el-radio-button class="router_card_button" size="large" value="command">
                    {{ $t('terminal.quickCommand', 2) }}
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
    --el-card-padding: 0;
}

.router_card_button {
    .el-radio-button__inner {
        min-width: 100px;
        height: 100%;
        background-color: var(--panel-button-active) !important;
        box-shadow: none !important;
        border: 2px solid transparent !important;
    }

    .el-radio-button__original-radio:checked + .el-radio-button__inner {
        color: $primary-color;
        border-color: $primary-color !important;
        border-radius: 4px;
    }
}
</style>
