<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="w-70">
        <el-alert
            class="mb-3"
            type="info"
            :closable="false"
            :title="$t('firewall.importLimit', [FIREWALL_BATCH_LIMIT, FIREWALL_IMPORT_MAX_SIZE / 1024])"
        />
        <div>
            <el-alert v-if="submitError" class="mb-3" type="error" :closable="false" :title="submitError" />
            <div class="import-file-bar mt-3">
                <el-upload
                    ref="uploadRef"
                    v-model:file-list="uploaderFiles"
                    action="#"
                    :auto-upload="false"
                    :disabled="loading"
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

            <el-alert
                v-if="deduplicatedCount"
                class="mt-3"
                type="info"
                :closable="false"
                :title="$t('firewall.importDuplicatesRemoved', [deduplicatedCount])"
            />
            <el-card class="mt-3 w-full" shadow="never" v-loading="loading">
                <template #header>
                    <div class="import-preview-header">
                        <span>{{ $t('commons.button.preview') }}</span>
                        <el-tag v-if="displayData.length" type="info" effect="plain">
                            {{ $t('commons.table.total', [displayData.length]) }}
                        </el-tag>
                    </div>
                </template>
                <ComplexTable v-model:selects="selects" :data="displayData" :height="300">
                    <el-table-column type="selection" fix />
                    <el-table-column label="IP" :min-width="60" prop="family">
                        <template #default="{ row }">
                            {{ row.family === 'ipv6' ? 'IPv6' : 'IPv4' }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.protocol')" :min-width="70" prop="protocol" />
                    <el-table-column :label="$t('firewall.sourcePort')" :min-width="70" prop="port" />
                    <el-table-column :label="$t('firewall.targetIP')" :min-width="100" prop="targetIP" />
                    <el-table-column :label="$t('firewall.targetPort')" :min-width="70" prop="targetPort" />
                    <el-table-column
                        v-if="currentFireName === 'iptables' || currentFireName === 'nftables'"
                        :label="$t('firewall.forwardInboundInterface')"
                        :min-width="100"
                        prop="interface"
                    >
                        <template #default="{ row }">
                            <span>
                                {{
                                    row.interface === '' || row.interface === 'all'
                                        ? $t('commons.table.all')
                                        : row.interface
                                }}
                            </span>
                        </template>
                    </el-table-column>
                </ComplexTable>
            </el-card>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="visible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" :loading="loading" :disabled="selects.length === 0" @click="onImport">
                    {{ $t('commons.button.import') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { genFileId, type UploadFile, type UploadFiles, type UploadProps, type UploadRawFile } from 'element-plus';
import { MsgError } from '@/utils/message';
import i18n from '@/lang';
import { getErrorMessage } from '@/utils/misc';
import { isAxiosError } from 'axios';
import { operateForwardRule, searchForwardRule } from '@/api/modules/firewall';
import { Firewall } from '@/api/interface/firewall';
import { normalizeForwardRuleImport, parseForwardRuleImport } from '../transfer';
import { FIREWALL_BATCH_LIMIT, FIREWALL_IMPORT_MAX_SIZE } from '@/views/host/firewall/utils/validation';
import { Document } from '@element-plus/icons-vue';

const emit = defineEmits<{ (e: 'created', taskID: string): void }>();

const visible = ref(false);
const loading = ref(false);
const submitError = ref('');
const selects = ref<Firewall.RuleForward[]>([]);
const displayData = ref<Firewall.RuleForward[]>([]);
const deduplicatedCount = ref(0);
const currentFireName = ref('');

const uploadRef = ref();
const uploaderFiles = ref<UploadFile[]>([]);
const acceptParams = (fireName: string) => {
    loading.value = false;
    submitError.value = '';
    displayData.value = [];
    selects.value = [];
    deduplicatedCount.value = 0;
    uploaderFiles.value = [];

    uploadRef.value?.clearFiles();
    currentFireName.value = fireName;
    visible.value = true;
};

const fileOnChange = async (uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    if (!uploadFile.raw) return;
    loading.value = true;
    submitError.value = '';
    displayData.value = [];
    deduplicatedCount.value = 0;
    selects.value = [];
    uploaderFiles.value = uploadFiles;

    if (uploadFile.raw.size > FIREWALL_IMPORT_MAX_SIZE) {
        uploadRef.value?.clearFiles();
        uploaderFiles.value = [];
        loading.value = false;
        MsgError(i18n.global.t('firewall.importLimit', [FIREWALL_BATCH_LIMIT, FIREWALL_IMPORT_MAX_SIZE / 1024]));
        return;
    }
    try {
        const parsed = parseForwardRuleImport(await uploadFile.raw.text());
        if (!parsed) throw new Error(i18n.global.t('commons.msg.errImportFormat'));
        const imported = parsed.flatMap((item) => {
            const rule = normalizeForwardRuleImport(item);
            return rule.protocol.split('/').map((protocol) => ({ ...rule, protocol }));
        });
        const response = await searchForwardRule({ strategy: '', info: '', page: 1, pageSize: 10000, all: true });
        if (!visible.value || uploaderFiles.value[0]?.uid !== uploadFile.uid) return;
        const ruleKey = (rule: Firewall.RuleForward) =>
            JSON.stringify([rule.family, rule.protocol, rule.port, rule.targetIP, rule.targetPort, rule.interface]);
        const seen = new Set((response.data.items || []).map((rule) => ruleKey(normalizeForwardRuleImport(rule))));
        const unique: Firewall.RuleForward[] = [];
        let duplicates = 0;
        for (const rule of imported) {
            const key = ruleKey(rule);
            if (seen.has(key)) {
                duplicates++;
                continue;
            }
            seen.add(key);
            unique.push(rule);
        }
        if (unique.length > FIREWALL_BATCH_LIMIT) {
            MsgError(i18n.global.t('firewall.importLimit', [FIREWALL_BATCH_LIMIT, FIREWALL_IMPORT_MAX_SIZE / 1024]));
            return;
        }
        displayData.value = unique;
        deduplicatedCount.value = duplicates;
    } catch (error) {
        if (visible.value && uploaderFiles.value[0]?.uid === uploadFile.uid) {
            submitError.value = getErrorMessage(error) || i18n.global.t('commons.msg.errImportFormat');
        }
    } finally {
        if (uploaderFiles.value[0]?.uid === uploadFile.uid) loading.value = false;
    }
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value!.handleStart(file);
};

const onImport = async () => {
    if (loading.value || selects.value.length === 0) return;
    if (
        selects.value.reduce(
            (total: number, rule: Firewall.RuleForward) => total + rule.protocol.split('/').length,
            0,
        ) > FIREWALL_BATCH_LIMIT
    ) {
        MsgError(i18n.global.t('firewall.importLimit', [FIREWALL_BATCH_LIMIT, FIREWALL_IMPORT_MAX_SIZE / 1024]));
        return;
    }
    loading.value = true;
    submitError.value = '';
    const rules: Firewall.RuleForward[] = [];
    for (const rule of selects.value) {
        rules.push({
            operation: 'add',
            family: rule.family,
            protocol: rule.protocol,
            port: rule.port,
            targetIP: rule.targetIP,
            targetPort: rule.targetPort,
            interface: rule.interface || '',
        });
    }

    try {
        const result = (await operateForwardRule({ rules, import: true })).data;
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

defineExpose({
    acceptParams,
});
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
