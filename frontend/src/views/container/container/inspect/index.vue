<template>
    <DrawerPro v-model="visible" :header="$t('commons.button.view')" size="large" @close="handleClose">
        <el-tabs v-model="activeTab" type="border-card">
            <el-tab-pane :label="$t('menu.container')" name="overview">
                <el-descriptions :column="1" border :title="$t('home.baseInfo')">
                    <el-descriptions-item :label="$t('commons.table.name')">
                        <el-text>{{ inspectData?.Name?.substring(1) || '-' }}</el-text>
                        <CopyButton :content="inspectData?.Name || ''" />
                    </el-descriptions-item>
                    <el-descriptions-item label="ID">
                        <el-text class="text-xs">{{ inspectData?.Id?.substring(0, 12) }}</el-text>
                        <CopyButton :content="inspectData?.Id || ''" />
                    </el-descriptions-item>
                    <el-descriptions-item label="PID">
                        {{ inspectData?.State?.Pid }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.image')">
                        {{ inspectData?.Config?.Image }}
                        <CopyButton :content="inspectData?.Config?.Image || ''" />
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.createdAt')">
                        {{ formatDate(inspectData?.Created) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('process.startTime')">
                        {{ formatDate(inspectData?.State?.StartedAt) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.finishTime')">
                        {{ formatDate(inspectData?.State?.FinishedAt) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.restartPolicy')">
                        {{ getRestartPolicy() }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.status')">
                        <el-tag :type="getStatusType(inspectData?.State?.Status)">
                            {{ inspectData?.State?.Status }}
                            {{
                                inspectData?.State?.Health?.Status ? '(' + inspectData?.State?.Health?.Status + ')' : ''
                            }}
                        </el-tag>
                    </el-descriptions-item>
                </el-descriptions>

                <el-descriptions class="mt-4" :column="1" border :title="$t('container.command')">
                    <el-descriptions-item :label="$t('container.command')">
                        <div v-if="inspectData?.Config?.Cmd?.length">
                            <el-tag type="info" v-for="(entry, index) in inspectData?.Config?.Cmd" :key="index">
                                {{ entry }}
                            </el-tag>
                        </div>
                        <span v-else>-</span>
                    </el-descriptions-item>
                    <el-descriptions-item label="ENTRYPONT">
                        <div v-if="inspectData?.Config?.Entrypoint?.length">
                            <el-tag type="info" v-for="(entry, index) in inspectData?.Config?.Entrypoint" :key="index">
                                {{ entry }}
                            </el-tag>
                        </div>
                        <span v-else>-</span>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.workingDir')">
                        {{ inspectData?.Config?.WorkingDir || '-' }}
                    </el-descriptions-item>
                </el-descriptions>

                <span class="envTitle">{{ $t('container.env') }}</span>
                <el-collapse accordion :title="$t('container.env')">
                    <el-collapse-item v-for="(env, index) in inspectData?.Config?.Env" :key="index" :name="index">
                        <template #title>
                            <el-text class="text-sm">{{ getEnvKey(env) }}</el-text>
                        </template>
                        <el-text class="text-xs break-all">{{ getEnvValue(env) }}</el-text>
                        <CopyButton :content="getEnvValue(env)" />
                    </el-collapse-item>
                </el-collapse>
                <el-empty v-if="!inspectData?.Config?.Env?.length" :description="$t('commons.msg.noData')" />
            </el-tab-pane>

            <el-tab-pane :label="$t('container.network')" name="network">
                <el-descriptions :column="1" border :title="$t('home.baseInfo')">
                    <el-descriptions-item :label="$t('container.networkName')">
                        {{ inspectData?.HostConfig?.NetworkMode }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.hostname')">
                        {{ inspectData?.Config?.Hostname }}
                    </el-descriptions-item>
                    <el-descriptions-item label="Domain">
                        {{ inspectData?.Config?.Domainname || '-' }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('commons.table.port')">
                        <div v-for="item of getPortBindings()" :key="item">
                            <span>{{ item.hostIp }}:{{ item.hostPort }}</span>
                            <span class="mx-2">→</span>
                            <span>{{ item.containerPort }}</span>
                        </div>
                    </el-descriptions-item>
                </el-descriptions>

                <div v-for="(network, name) in inspectData?.NetworkSettings?.Networks" :key="name" class="mb-4 mt-4">
                    <el-descriptions :column="2" border :title="name">
                        <el-descriptions-item label="Network ID">
                            <el-text class="text-xs">{{ network?.NetworkID?.substring(0, 12) }}</el-text>
                        </el-descriptions-item>
                        <el-descriptions-item label="Endpoint ID">
                            <el-text class="text-xs">{{ network?.EndpointID?.substring(0, 12) }}</el-text>
                        </el-descriptions-item>
                        <el-descriptions-item label="IPv4">
                            {{ network?.IPAddress || '-' }}
                        </el-descriptions-item>
                        <el-descriptions-item label="IPv4 Gateway">
                            {{ network?.Gateway || '-' }}
                        </el-descriptions-item>
                        <el-descriptions-item label="MAC">
                            {{ network?.MacAddress || '-' }}
                        </el-descriptions-item>
                        <el-descriptions-item label="IPv6 Gateway">
                            {{ network?.IPv6Gateway || '-' }}
                        </el-descriptions-item>
                    </el-descriptions>
                </div>
            </el-tab-pane>

            <el-tab-pane :label="$t('container.volume')" name="storage">
                <el-table :data="inspectData?.Mounts" border>
                    <el-table-column :label="$t('commons.table.type')" width="100">
                        <template #default="{ row }">
                            <el-tag size="small">
                                {{ row.Type === 'bind' ? $t('container.volumeOption') : $t('container.hostOption') }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.hostOption')" show-overflow-tooltip>
                        <template #default="{ row }">
                            <el-text
                                v-if="row.Source"
                                type="primary"
                                class="cursor-pointer"
                                @click="handleJumpToFile(row.Source)"
                            >
                                {{ row.Source }}
                            </el-text>
                            <span v-else>-</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.volumeOption')" prop="Destination" show-overflow-tooltip />
                    <el-table-column :label="$t('container.tag')" width="80" align="center">
                        <template #default="{ row }">
                            {{ row.RW ? $t('container.modeRW') : $t('container.modeR') }}
                        </template>
                    </el-table-column>
                </el-table>
            </el-tab-pane>

            <el-tab-pane :label="$t('commons.button.view')" name="view">
                <CodemirrorPro v-model="rawJson" :height-diff="240" :disabled="true" mode="json" />
            </el-tab-pane>
        </el-tabs>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import { routerToFileWithPath } from '@/utils/router';
import i18n from '@/lang';

const visible = ref(false);
const activeTab = ref('overview');
const inspectData = ref<any>(null);
const rawJson = ref('');
const showRawJson = ref(false);

interface DialogProps {
    data: any;
}

const acceptParams = (props: DialogProps): void => {
    visible.value = true;
    activeTab.value = 'overview';
    showRawJson.value = false;

    try {
        if (typeof props.data === 'string') {
            inspectData.value = JSON.parse(props.data);
        } else {
            inspectData.value = props.data;
        }
        rawJson.value = JSON.stringify(inspectData.value, null, 2);
    } catch (e) {
        console.error('Failed to parse inspect data:', e);
    }
};

const handleClose = () => {
    visible.value = false;
    inspectData.value = null;
    rawJson.value = '';
};

const getStatusType = (status: string): string => {
    const statusMap: Record<string, string> = {
        running: 'success',
        paused: 'warning',
        restarting: 'warning',
        exited: 'info',
        dead: 'danger',
    };
    return statusMap[status?.toLowerCase()] || 'info';
};

const formatDate = (dateStr: string): string => {
    if (!dateStr || dateStr === '0001-01-01T00:00:00Z') {
        return '-';
    }
    try {
        const date = new Date(dateStr);
        const y = date.getFullYear();
        const m = String(date.getMonth() + 1).padStart(2, '0');
        const d = String(date.getDate()).padStart(2, '0');
        const h = String(date.getHours()).padStart(2, '0');
        const minute = String(date.getMinutes()).padStart(2, '0');
        const second = String(date.getSeconds()).padStart(2, '0');
        return `${y}-${m}-${d} ${h}:${minute}:${second}`;
    } catch {
        return dateStr;
    }
};

const getPortBindings = (): any[] => {
    const ports: any[] = [];
    const portBindings = inspectData.value?.HostConfig?.PortBindings || {};

    for (const [containerPort, bindings] of Object.entries(portBindings)) {
        if (Array.isArray(bindings)) {
            for (const binding of bindings) {
                ports.push({
                    containerPort,
                    hostIp: (binding as any).HostIp || '0.0.0.0',
                    hostPort: (binding as any).HostPort,
                });
            }
        }
    }
    return ports;
};

const getEnvKey = (env: string): string => {
    const index = env.indexOf('=');
    return index > 0 ? env.substring(0, index) : env;
};

const getEnvValue = (env: string): string => {
    const index = env.indexOf('=');
    return index > 0 ? env.substring(index + 1) : '';
};

const getRestartPolicy = () => {
    switch (inspectData.value?.HostConfig?.RestartPolicy?.Name) {
        case 'no':
            return i18n.global.t('container.no');
        case 'always':
            return i18n.global.t('container.always');
        case 'on-failure':
            return i18n.global.t('container.onFailure');
        case 'unless-stopped':
            return i18n.global.t('container.unlessStopped');
        default:
            return '-';
    }
};

const handleJumpToFile = (path: string) => {
    if (path) {
        routerToFileWithPath(path);
    }
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.break-all {
    word-break: break-all;
}

:deep(.el-descriptions__label) {
    width: 180px;
    background-color: transparent !important;
}

.envTitle {
    font-size: 16px;
    color: var(--el-text-color-primary);
    margin-top: 20px;
    margin-bottom: 16px;
    display: block;
}
</style>
