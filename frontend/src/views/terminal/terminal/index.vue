<template>
    <div class="terminal-page" :class="{ 'is-mobile': isMobile }">
        <el-tabs
            type="card"
            class="terminal-tabs"
            style="background-color: var(--panel-terminal-tag-bg-color)"
            v-model="terminalValue"
            addable
            @tab-add="showConnections = !showConnections"
            @tab-change="quickCmd = ''"
            @edit="handleTabsRemove"
        >
            <el-tab-pane
                :key="item.key"
                v-for="item in store.entries"
                :closable="true"
                :label="item.title"
                :name="item.key"
            >
                <template #label>
                    <el-tooltip
                        :content="`${item.title} · ${
                            item.status === 'online' ? `${item.latency} ms` : $t('commons.button.reconnect')
                        }`"
                        placement="top-start"
                        :show-after="300"
                    >
                        <span class="terminal-tab-label">
                            <span v-if="item.status === 'online'" class="terminal-tab-status" aria-hidden="true">
                                <span class="terminal-status-dot"></span>
                            </span>
                            <el-button
                                v-else
                                icon="Refresh"
                                class="terminal-tab-reconnect"
                                :aria-label="$t('commons.button.reconnect')"
                                link
                                @click.stop="onReconnect(item)"
                            />
                            <span class="terminal-tab-title">{{ item.title }}</span>
                            <span
                                v-if="item.key === terminalValue && item.status === 'online'"
                                class="terminal-tab-latency"
                            >
                                {{ item.latency }} ms
                            </span>
                        </span>
                    </el-tooltip>
                </template>
                <div
                    class="terminal-slot"
                    :ref="(el: any) => onSlot(item.key, el)"
                    :style="{
                        height: `calc(100vh - ${loadHeight()})`,
                        'background-color': `var(--panel-logs-bg-color)`,
                    }"
                ></div>

                <div class="flex items-center gap-2 w-full py-2 flex-wrap">
                    <AiSetting v-if="!isMobile" class="shrink-0" />
                    <el-cascader
                        v-model="quickCmd"
                        :options="commandTree"
                        :props="quickCommandProps"
                        :show-all-levels="false"
                        filterable
                        clearable
                        class="quick-command-cascader min-w-[180px] max-w-[260px] shrink-0"
                        :placeholder="$t('terminal.quickCommand')"
                        @change="handleQuickCommandChange"
                    >
                        <template #default="{ data }">
                            <el-tooltip
                                v-if="!data.children?.length"
                                placement="right"
                                popper-class="command-detail-tooltip"
                            >
                                <template #content>
                                    <div class="command-detail-content">{{ data.value }}</div>
                                </template>
                                <div class="cascader-option">
                                    <span class="cascader-option-label">{{ data.label }}</span>
                                </div>
                            </el-tooltip>
                            <div v-else class="cascader-option">
                                <span class="cascader-option-label">{{ data.label }}</span>
                            </div>
                        </template>
                    </el-cascader>
                    <el-input
                        v-model="batchVal"
                        @keydown.enter.exact.prevent="batchInput"
                        type="textarea"
                        :autosize="{ minRows: 1, maxRows: 3 }"
                        class="flex-1 basis-[300px] min-w-[200px]"
                        placeholder=">"
                    ></el-input>
                    <el-checkbox
                        :label="$t('terminal.batchInput')"
                        v-model="isBatch"
                        class="shrink-0 whitespace-nowrap"
                    />
                </div>
            </el-tab-pane>
            <template #add-icon>
                <ConnectionMenu ref="connectionMenuRef" v-model="showConnections" :open-session="openConnection" />
            </template>
            <div v-if="store.entries.length === 0">
                <el-empty
                    :style="{ height: `calc(100vh - ${loadEmptyHeight()})`, 'background-color': '#000' }"
                    :description="$t('terminal.emptyTerminal')"
                ></el-empty>
            </div>
        </el-tabs>
        <div v-if="!isMobile" class="terminal-actions">
            <el-tooltip :content="loadTooltip()" placement="top">
                <el-button
                    class="terminal-action"
                    icon="FullScreen"
                    text
                    :aria-label="loadTooltip()"
                    @click="toggleFullscreen"
                />
            </el-tooltip>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onBeforeUnmount, onActivated, onDeactivated } from 'vue';
import screenfull from 'screenfull';
import i18n from '@/lang';
import { testByID, testLocalConn } from '@/api/modules/terminal';
import { useGlobalStore } from '@/composables/useGlobalStore';
import router from '@/routers';
import { getCommandTree } from '@/api/modules/command';
import { getAgentSettingInfo } from '@/api/modules/setting';
import AiSetting from '@/views/terminal/setting/ai/index.vue';
import { TerminalSessionStore } from '@/store';
import ConnectionMenu from '@/components/terminal/connection-menu/index.vue';
import type { TerminalConnectionOptions } from '@/components/terminal/connection-menu/types';

const { isFullScreen, isMobile, isNodeAdmin, openMenuTabs } = useGlobalStore();
const store = TerminalSessionStore();

const connectionMenuRef = ref<InstanceType<typeof ConnectionMenu>>();

const toggleFullscreen = () => {
    if (screenfull.isEnabled) {
        screenfull.toggle();
    }
};
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (isFullScreen.value ? 'quitFullscreen' : 'fullscreen'));
};

let timer: ReturnType<typeof setInterval> | null = null;
const terminalValue = ref();

const commandTree = ref();
const quickCommandProps = {
    expandTrigger: 'hover' as const,
};
let quickCmd = ref();
let batchVal = ref();
let isBatch = ref<boolean>(false);

const showConnections = ref(false);
const initCmd = ref('');

const acceptParams = async () => {
    isFullScreen.value = false;
    loadCommandTree();
    if (store.entries.length === 0) {
        await openDefaultLocalConn();
    } else {
        if (!store.find(terminalValue.value)) {
            terminalValue.value = store.entries[0].key;
        }
        await claim();
        store.sync();
    }
    timer = setInterval(store.sync, 1000 * 5);
    if (!isMobile.value) {
        screenfull.on('change', () => {
            isFullScreen.value = screenfull.isFullscreen;
        });
    }
};

const openDefaultLocalConn = async () => {
    await nextTick();
    if (isNodeAdmin.value) {
        await connectionMenuRef.value?.connectLocal();
        return;
    }
    await getAgentSettingInfo().then(async (res) => {
        if (res.data?.localSSHConnShow === 'Enable') {
            await connectionMenuRef.value?.connectLocal();
        }
    });
};

const cleanTimer = () => {
    clearInterval(Number(timer));
    timer = null;
};

const slotEls: Record<string, HTMLElement> = {};
const onSlot = (key: string, el: HTMLElement | null) => {
    if (el) slotEls[key] = el;
    else delete slotEls[key];
};
let pageVisible = true;
const claim = async () => {
    for (const item of store.entries) {
        if (pageVisible) {
            store.setSlot(item.key, slotEls[item.key] || null);
        } else if (store.slots[item.key] && store.slots[item.key] === slotEls[item.key]) {
            store.setSlot(item.key, null);
        }
    }
    if (!pageVisible) return;
    await nextTick();
    for (const item of store.entries) {
        store.instances[item.key]?.refit();
    }
};
watch(
    () => store.entries.length,
    () => nextTick(claim),
);
onActivated(() => {
    pageVisible = true;
    claim();
});
onDeactivated(() => {
    pageVisible = false;
    claim();
});

const loadHeight = () => {
    return openMenuTabs.value ? '250px' : '210px';
};
const loadEmptyHeight = () => {
    return openMenuTabs.value ? '201px' : '156px';
};

const handleTabsRemove = async (targetName: string, action: 'remove' | 'add') => {
    if (action !== 'remove') {
        return;
    }
    if (!store.find(targetName)) {
        return;
    }
    const tabs = store.entries;
    let activeName = terminalValue.value;
    if (activeName === targetName) {
        tabs.forEach((tab, index) => {
            if (tab.key === targetName) {
                const nextTab = tabs[index + 1] || tabs[index - 1];
                if (nextTab) {
                    activeName = nextTab.key;
                }
            }
        });
    }
    terminalValue.value = activeName;
    store.remove(targetName);
};

const loadCommandTree = async () => {
    const res = await getCommandTree('command');
    commandTree.value = res.data || [];
    for (const item of commandTree.value) {
        if (item.label === 'Default') {
            item.label = i18n.global.t('commons.table.default');
        }
    }
};

const sendToTerminals = (command: string, all: boolean) => {
    const keys = all ? store.entries.map((e) => e.key) : [terminalValue.value];
    for (const key of keys) {
        store.instances[key]?.sendMsg(command);
    }
};

const handleQuickCommandChange = (val: Array<string>) => {
    if (!val?.length) {
        return;
    }
    sendToTerminals(val[val.length - 1] + '\n', isBatch.value);
    quickCmd.value = '';
};

function batchInput() {
    if (batchVal.value === '') {
        return;
    }
    sendToTerminals(batchVal.value + '\n', isBatch.value);
    batchVal.value = '';
}

const connectionError = 'Failed to set up the connection. Please check the host information';

const openConnection = async (options: TerminalConnectionOptions) => {
    const cmd = initCmd.value;
    initCmd.value = '';
    terminalValue.value = await store.open({ ...options, initCmd: cmd });
};

const onReconnect = async (item: any) => {
    const nodeName = new URLSearchParams(item.args).get('operateNode') || undefined;
    const res = item.wsID === 0 ? await testLocalConn(nodeName) : await testByID(item.wsID);
    const cmd = initCmd.value;
    initCmd.value = '';
    await store.reconnect(item.key, res.data ? '' : connectionError, cmd);
    store.sync();
};

const changeFullScreen = () => {
    isFullScreen.value = screenfull.isFullscreen;
};

defineExpose({
    acceptParams,
});

onBeforeUnmount(() => {
    document.removeEventListener('fullscreenchange', changeFullScreen);
    cleanTimer();
    pageVisible = false;
    claim();
});

onMounted(() => {
    if (router.currentRoute.value.query.path) {
        const path = String(router.currentRoute.value.query.path);
        initCmd.value = `cd "${path}" \n`;
    }
    document.addEventListener('fullscreenchange', changeFullScreen);
});
</script>

<style lang="scss" scoped>
.terminal-page {
    --terminal-actions-width: 40px;
    position: relative;
    min-width: 0;
    padding-top: 7px;

    &.is-mobile {
        --terminal-actions-width: 0px;
    }
}

.terminal-actions {
    position: absolute;
    top: 7px;
    right: 0;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    width: var(--terminal-actions-width);
    height: var(--el-tabs-header-height, 40px);
}

.terminal-action {
    width: 32px;
    height: 32px;
    margin: 0;
    padding: 0;
    border-radius: 6px;
    color: var(--el-text-color-regular);

    &:hover {
        color: var(--el-color-primary);
    }
}

.terminal-tabs {
    :deep(.el-tabs__header) {
        justify-content: flex-start;
        padding: 0 var(--terminal-actions-width) 0 0;
        min-height: var(--el-tabs-header-height);
        position: relative;
        margin: 0 0 3px 0;
    }
    :deep(.el-tabs__nav) {
        white-space: nowrap;
        position: relative;
        transition: transform var(--el-transition-duration);
        float: left;
        z-index: calc(var(--el-index-normal) + 1);
    }
    :deep(.el-tabs__item) {
        padding: 0;
    }
    :deep(.el-tabs__nav-wrap) {
        flex: 0 1 auto;
        min-width: 0;
    }

    :deep(.el-tabs__new-tab) {
        width: auto;
        height: auto;
        margin: 0;
        border: 0;
    }
    :deep(.el-tabs__item.is-active) {
        color: var(--panel-terminal-tag-active-text-color);
        background-color: var(--panel-terminal-tag-active-bg-color);
    }
    :deep(.el-tabs__item:hover) {
        color: var(--panel-terminal-tag-hover-text-color);
    }
    :deep(.el-tabs__item.is-active:hover) {
        color: var(--panel-terminal-tag-active-text-color);
    }
    :deep(.el-tabs__header .el-tabs__item.is-closable) {
        padding: 0 12px;

        .is-icon-close {
            width: 14px;
            margin-left: 6px;
            right: 0;
            opacity: 0;
            pointer-events: none;
            transition: opacity var(--el-transition-duration);
        }

        &.is-active,
        &:hover,
        &:focus-within {
            .is-icon-close {
                opacity: 1;
                pointer-events: auto;
            }
        }

        @media (hover: none), (pointer: coarse) {
            .is-icon-close {
                opacity: 1;
                pointer-events: auto;
            }
        }
    }
}

.terminal-tab-label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    max-width: 220px;
}

.terminal-tab-status,
.terminal-tab-reconnect {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 14px;
    height: 14px;
    padding: 0;
}

.terminal-status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background-color: var(--el-color-success);
}

.terminal-tab-reconnect {
    color: inherit;
}

.terminal-tab-title {
    min-width: 0;
    max-width: 140px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.terminal-tab-latency {
    flex-shrink: 0;
    font-size: 12px;
    font-weight: 400;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
    opacity: 0.7;
}

.terminal-slot {
    width: 100%;
}

.vertical-tabs > .el-tabs__content {
    padding: 32px;
    color: #6b778c;
    font-size: 32px;
    font-weight: 600;
}
.el-input__wrapper {
    border-radius: 50px;
}

:deep(.el-textarea__inner) {
    border-radius: 4px;
    resize: none;
    min-height: 32px;
    transition: height 0.2s ease;
}

.quick-command-cascader {
    :deep(.el-input__wrapper) {
        border-radius: 6px;
    }
}

.cascader-option {
    width: 100%;
    display: flex;
    align-items: center;
    min-width: 0;
}

.cascader-option-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

:deep(.command-detail-tooltip) {
    max-width: 420px;
}

.command-detail-content {
    white-space: pre-wrap;
    word-break: break-all;
    line-height: 1.5;
}
</style>
