<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="w-70">
        <el-alert
            class="mb-3"
            type="info"
            :closable="false"
            :title="$t('firewall.importLimit', [FIREWALL_BATCH_LIMIT, FIREWALL_IMPORT_MAX_SIZE / 1024])"
        />
        <el-alert v-if="submitError" class="mb-3" type="error" :closable="false" :title="submitError" />
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
                    <el-tag v-if="policies.length" type="info" effect="plain">
                        {{ $t('commons.table.total', [policies.length]) }}
                    </el-tag>
                </div>
            </template>
            <ComplexTable :data="policies" :height="300">
                <el-table-column width="48" fixed>
                    <template #header>
                        <el-checkbox
                            :model-value="policies.length > 0 && selects.size === policies.length"
                            :indeterminate="selects.size > 0 && selects.size < policies.length"
                            @change="selects = $event ? new Set(policies) : new Set()"
                        />
                    </template>
                    <template #default="{ row }">
                        <el-checkbox
                            :model-value="selects.has(row)"
                            @change="$event ? selects.add(row) : selects.delete(row)"
                        />
                    </template>
                </el-table-column>
                <el-table-column label="IP" prop="family" min-width="65">
                    <template #default="{ row }">{{ row.family === 'ipv6' ? 'IPv6' : 'IPv4' }}</template>
                </el-table-column>
                <el-table-column label="IP" prop="hostIP" min-width="150" />
                <el-table-column :label="$t('commons.table.port')" prop="hostPort" min-width="90" />
                <el-table-column :label="$t('commons.table.protocol')" prop="protocol" min-width="90" />
                <el-table-column :label="$t('firewall.protectionMode')" min-width="140">
                    <template #default="{ row }">{{ modeLabel(row.mode) }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.protection')" min-width="180">
                    <template #default="{ row }">{{ displaySources(row) || '-' }}</template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.description')" prop="description" min-width="150" />
            </ComplexTable>
        </el-card>
        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :loading="loading" :disabled="selects.size === 0" @click="onImport">
                {{ $t('commons.button.import') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { upsertDockerPortGuardPolicies } from '@/api/modules/firewall';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import { getErrorMessage } from '@/utils/misc';
import { isAxiosError } from 'axios';
import { genFileId, type UploadFile, type UploadFiles, type UploadProps, type UploadRawFile } from 'element-plus';
import { ref } from 'vue';
import { dockerGuardEndpointKey, normalizeDockerGuardPolicy } from '@/views/host/firewall/docker/model';
import {
    FIREWALL_BATCH_LIMIT,
    FIREWALL_IMPORT_MAX_SIZE,
    formatHostAddressList,
} from '@/views/host/firewall/utils/validation';
import { Document } from '@element-plus/icons-vue';

const emit = defineEmits<{ (event: 'created', taskID: string): void }>();
const visible = ref(false);
const loading = ref(false);
const policies = ref<Firewall.DockerGuardPolicy[]>([]);
const selects = ref(new Set<Firewall.DockerGuardPolicy>());
const uploadRef = ref();
const uploaderFiles = ref<UploadFile[]>([]);
const submitError = ref('');
const displaySources = (policy: Firewall.DockerGuardPolicy) => formatHostAddressList(policy.sources, policy.family);

const fileOnChange = (uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    if (!uploadFile.raw) return;
    loading.value = true;

    policies.value = [];
    selects.value = new Set();
    submitError.value = '';
    uploaderFiles.value = uploadFiles;
    if (uploadFile.raw.size > FIREWALL_IMPORT_MAX_SIZE) {
        uploadRef.value?.clearFiles();
        uploaderFiles.value = [];
        loading.value = false;
        MsgError(i18n.global.t('firewall.importLimit', [FIREWALL_BATCH_LIMIT, FIREWALL_IMPORT_MAX_SIZE / 1024]));
        return;
    }
    const reader = new FileReader();
    reader.onload = (event) => {
        try {
            const parsed: unknown = JSON.parse(String(event.target?.result || ''));
            if (!Array.isArray(parsed)) throw new Error();
            if (parsed.length > FIREWALL_BATCH_LIMIT) {
                MsgError(
                    i18n.global.t('firewall.importLimit', [FIREWALL_BATCH_LIMIT, FIREWALL_IMPORT_MAX_SIZE / 1024]),
                );
                return;
            }
            const normalized = parsed.map(normalizeDockerGuardPolicy);
            if (normalized.some((policy) => !policy)) throw new Error();
            const byEndpoint = new Map<string, Firewall.DockerGuardPolicy>();
            for (const policy of normalized as Firewall.DockerGuardPolicy[]) {
                byEndpoint.set(dockerGuardEndpointKey(policy), policy);
            }
            policies.value = [...byEndpoint.values()];
            selects.value = new Set(policies.value);
        } catch {
            policies.value = [];
            selects.value = new Set();
            MsgError(i18n.global.t('commons.msg.errImportFormat'));
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
    if (loading.value || selects.value.size === 0) return;
    if (selects.value.size > FIREWALL_BATCH_LIMIT) {
        MsgError(i18n.global.t('firewall.importLimit', [FIREWALL_BATCH_LIMIT, FIREWALL_IMPORT_MAX_SIZE / 1024]));
        return;
    }
    loading.value = true;
    submitError.value = '';
    try {
        const result = (
            await upsertDockerPortGuardPolicies({
                policies: policies.value.filter((policy) => selects.value.has(policy)),
            })
        ).data;
        if (!result.taskID || !result.queued) {
            submitError.value = i18n.global.t('commons.msg.operationFailed');
            return;
        }
        visible.value = false;
        emit('created', result.taskID);
    } catch (error) {
        submitError.value =
            (isAxiosError(error) && error.response?.data?.message) ||
            (error && getErrorMessage(error)) ||
            i18n.global.t('commons.res.commonError');
    } finally {
        loading.value = false;
    }
};

const modeLabel = (mode: Firewall.DockerGuardPolicy['mode']) => {
    if (mode === 'deny_sources') return i18n.global.t('firewall.denySources');
    if (mode === 'allow_sources') return i18n.global.t('firewall.allowSources');
    return i18n.global.t('firewall.denyAll');
};

const acceptParams = () => {
    loading.value = false;

    policies.value = [];
    selects.value = new Set();
    uploaderFiles.value = [];
    submitError.value = '';
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
