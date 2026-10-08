<template>
    <DialogPro
        v-model="visible"
        :title="initialize ? $t('firewall.initializeFromFile') : $t('commons.button.import')"
        size="w-70"
    >
        <el-alert class="mb-3" type="info" :closable="false" :title="$t('firewall.importFileHelper')" />
        <div class="import-file-bar mt-3">
            <el-upload
                ref="uploadRef"
                v-model:file-list="uploaderFiles"
                action="#"
                :auto-upload="false"
                :show-file-list="false"
                :limit="1"
                accept=".json"
                :on-change="fileOnChange"
                :on-exceed="handleExceed"
            >
                <el-button type="primary" icon="Upload">{{ $t('commons.button.upload') }}</el-button>
            </el-upload>
            <div v-if="uploaderFiles.length" class="import-file-info">
                <el-icon><Document /></el-icon>
                <span class="import-file-name">{{ uploaderFiles[0].name }}</span>
            </div>
            <el-text v-else type="info">.json</el-text>
        </div>
        <el-card class="mt-3 w-full" shadow="never" v-loading="loading">
            <template #header>
                <div class="import-preview-header">
                    <span>{{ $t('commons.button.preview') }}</span>
                    <el-tag v-if="rules.length" type="info" effect="plain">
                        {{ $t('commons.table.total', [rules.length]) }}
                    </el-tag>
                </div>
            </template>
            <ComplexTable :data="rules" :height="300">
                <el-table-column width="48" fixed>
                    <template #header>
                        <el-checkbox
                            :model-value="rules.length > 0 && selects.size === rules.length"
                            :indeterminate="selects.size > 0 && selects.size < rules.length"
                            @change="selects = $event ? new Set(rules) : new Set()"
                        />
                    </template>
                    <template #default="{ row }">
                        <el-checkbox
                            :model-value="selects.has(row)"
                            @change="$event ? selects.add(row) : selects.delete(row)"
                        />
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.protocol')" prop="protocol" min-width="90" />
                <el-table-column :label="$t('firewall.sourceIP')" min-width="150">
                    <template #default="{ row }">{{ displayAddress(row, row.sourceAddress) }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.sourcePort')" min-width="110">
                    <template #default="{ row }">{{ row.sourcePort || '*' }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.destIP')" min-width="150">
                    <template #default="{ row }">{{ displayAddress(row, row.destinationAddress) }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.destPort')" min-width="110">
                    <template #default="{ row }">{{ row.destinationPort || '*' }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.action')" prop="action" min-width="90">
                    <template #default="{ row }">{{ actionLabel(row.action) }}</template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.description')" prop="description" min-width="150" />
                <el-table-column :label="$t('commons.button.view')" prop="raw" min-width="180" show-overflow-tooltip />
            </ComplexTable>
        </el-card>
        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                type="primary"
                :loading="loading"
                :disabled="!fileLoaded || (rules.length > 0 && selects.size === 0)"
                @click="onImport"
            >
                {{ $t('commons.button.import') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { createFirewallRules } from '@/api/modules/firewall';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import { formatHostAddress, inferAddressFamily } from '@/views/host/firewall/utils/validation';
import { Document } from '@element-plus/icons-vue';
import { genFileId, type UploadFile, type UploadFiles, type UploadProps, type UploadRawFile } from 'element-plus';
import { ref } from 'vue';

const emit = defineEmits<{ (event: 'created', taskID: string): void }>();
const visible = ref(false);
const initialize = ref(false);
const provider = ref<Firewall.Provider>('iptables');
const maxFileSize = 64 * 1024 * 1024;
const loading = ref(false);
const fileLoaded = ref(false);
const rules = ref<Firewall.Rule[]>([]);
const selects = ref(new Set<Firewall.Rule>());
const uploadRef = ref();
const uploaderFiles = ref<UploadFile[]>([]);

const displayAddress = (rule: Firewall.Rule, address?: string) => {
    const wildcard =
        rule.scope.family === 'ipv6' ? '::/0' : rule.scope.family === 'inet' ? '0.0.0.0/0, ::/0' : '0.0.0.0/0';
    if (address && address !== wildcard) return formatHostAddress(address, rule.scope.family);
    return `${wildcard}（${i18n.global.t('firewall.anyWhere')}）`;
};

const actionLabel = (action: Firewall.Action) => {
    if (action === 'accept') return i18n.global.t('firewall.accept');
    if (action === 'reject') return i18n.global.t('firewall.reject');
    if (action === 'drop') return i18n.global.t('firewall.drop');
    return i18n.global.t('commons.status.unknown');
};

const isRule = (value: unknown): value is Firewall.Rule => {
    if (!value || typeof value !== 'object') return false;
    const rule = value as Partial<Firewall.Rule>;
    return (
        Boolean(rule.scope) &&
        ['iptables', 'nftables', 'firewalld', 'ufw'].includes(String(rule.scope?.provider)) &&
        typeof rule.protocol === 'string' &&
        (['accept', 'drop', 'reject'].includes(String(rule.action)) ||
            (Boolean(rule.raw) && ['partial', 'opaque'].includes(rule.parseStatus || '')))
    );
};

const normalizeLegacyImportedRule = (value: unknown): Firewall.Rule | undefined => {
    if (!value || typeof value !== 'object') return;
    const rule = value as Record<string, unknown>;
    if (rule.scope !== undefined || !['accept', 'drop'].includes(String(rule.strategy))) return;
    if (typeof rule.address !== 'string') return;
    if (rule.description !== undefined && typeof rule.description !== 'string') return;
    if (rule.port !== undefined && typeof rule.port !== 'string') return;
    if (rule.protocol !== undefined && typeof rule.protocol !== 'string') return;
    if (rule.family && !['ipv4', 'ipv6'].includes(String(rule.family))) return;

    const family = (rule.family as 'ipv4' | 'ipv6' | undefined) || inferAddressFamily(rule.address);
    const port = typeof rule.port === 'string' ? rule.port.trim() : '';
    const protocol = typeof rule.protocol === 'string' ? rule.protocol.trim().toLowerCase() : '';
    if (port && !['tcp', 'udp', 'tcp/udp'].includes(protocol)) return;
    if (!port && protocol && !['all', 'any'].includes(protocol)) return;

    const scope: Firewall.Scope = { provider: provider.value, family, direction: 'input' };
    if (provider.value === 'iptables' || provider.value === 'nftables') {
        scope.table = 'filter';
        scope.chain = '1PANEL_BASIC';
    } else if (provider.value === 'firewalld') {
        scope.zone = 'public';
    } else {
        scope.chain = 'incoming';
    }
    return {
        scope,
        protocol: port ? protocol : 'all',
        sourceAddress: rule.address,
        destinationPort: port || undefined,
        action: rule.strategy as Firewall.Action,
        description: (rule.description as string | undefined) || '',
    };
};

const fileOnChange = (uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    if (!uploadFile.raw) return;
    loading.value = true;

    fileLoaded.value = false;
    rules.value = [];
    selects.value = new Set();
    uploaderFiles.value = uploadFiles;
    if (uploadFile.raw.size > maxFileSize) {
        uploadRef.value?.clearFiles();
        uploaderFiles.value = [];
        loading.value = false;
        MsgError(i18n.global.t('firewall.importFileHelper'));
        return;
    }
    const reader = new FileReader();
    reader.onload = (event) => {
        try {
            const parsed: unknown = JSON.parse(String(event.target?.result || ''));
            if (!Array.isArray(parsed) || (!initialize.value && parsed.length === 0)) {
                MsgError(i18n.global.t('commons.msg.errImportFormat'));
                return;
            }
            const normalized = parsed.map((rule) => (isRule(rule) ? rule : normalizeLegacyImportedRule(rule)));
            if (!normalized.every(isRule)) {
                MsgError(i18n.global.t('commons.msg.errImportFormat'));
                return;
            }
            fileLoaded.value = true;
            rules.value = normalized;
            selects.value = new Set(rules.value);
        } catch (error) {
            MsgError(i18n.global.t('commons.msg.errImport') + String(error));
        } finally {
            loading.value = false;
        }
    };
    reader.readAsText(uploadFile.raw);
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value?.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value?.handleStart(file);
};

const onImport = async () => {
    if (loading.value || !fileLoaded.value || (rules.value.length > 0 && selects.value.size === 0)) return;
    loading.value = true;
    try {
        const result = (
            await createFirewallRules({
                initialize: initialize.value,
                items: rules.value
                    .filter((rule) => selects.value.has(rule))
                    .map((rule) => ({ rule, sourceKind: 'imported', raw: rule.raw, parseStatus: rule.parseStatus })),
            })
        ).data;
        if (!result.taskID || !result.queued) {
            MsgError(i18n.global.t('commons.msg.operationFailed'));
            return;
        }
        visible.value = false;
        emit('created', result.taskID);
    } finally {
        loading.value = false;
    }
};

const acceptParams = (value: Firewall.Provider, withInitialization = false) => {
    provider.value = value;
    initialize.value = withInitialization;
    loading.value = false;

    fileLoaded.value = false;
    rules.value = [];
    selects.value = new Set();
    uploaderFiles.value = [];
    uploadRef.value?.clearFiles();
    visible.value = true;
};

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.import-file-bar {
    display: flex;
    min-height: 32px;
    align-items: center;
    gap: 12px;
}

.import-file-info {
    display: flex;
    min-width: 0;
    align-items: center;
    gap: 6px;
    color: var(--el-text-color-regular);
}

.import-file-name {
    overflow: hidden;
    max-width: 420px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.import-preview-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
}
</style>
