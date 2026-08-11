<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="large">
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
            <el-button type="primary">{{ $t('commons.button.upload') }}</el-button>
        </el-upload>
        <el-card class="mt-3 w-full" v-loading="loading">
            <ComplexTable v-model:selects="selects" :data="rules" :height="420">
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.protocol')" prop="protocol" min-width="90" />
                <el-table-column :label="$t('firewall.sourceIP')" min-width="150">
                    <template #default="{ row }">{{ displayAddress(row, row.sourceAddress) }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.sourcePort')" min-width="110">
                    <template #default="{ row }">{{ row.sourcePort || $t('firewall.allPorts') }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.destIP')" min-width="150">
                    <template #default="{ row }">{{ displayAddress(row, row.destinationAddress) }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.destPort')" min-width="110">
                    <template #default="{ row }">{{ row.destinationPort || $t('firewall.allPorts') }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.action')" prop="action" min-width="90" />
                <el-table-column :label="$t('commons.table.description')" prop="description" min-width="150" />
            </ComplexTable>
        </el-card>
        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :loading="loading" :disabled="selects.length === 0" @click="onImport">
                {{ $t('commons.button.import') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import { checkFirewallRule, createFirewallRule } from '@/api/modules/firewall';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { genFileId, type UploadFile, type UploadFiles, type UploadProps, type UploadRawFile } from 'element-plus';
import { ref } from 'vue';

const emit = defineEmits<{ (event: 'search'): void }>();
const visible = ref(false);
const loading = ref(false);
const provider = ref<Firewall.Provider>('iptables');
const rules = ref<Firewall.Rule[]>([]);
const selects = ref<Firewall.Rule[]>([]);
const uploadRef = ref();
const uploaderFiles = ref<UploadFile[]>([]);

const displayAddress = (rule: Firewall.Rule, address?: string) => {
    if (address) return address;
    if (rule.scope.family === 'ipv6') return '::/0';
    if (rule.scope.family === 'inet') return '0.0.0.0/0, ::/0';
    return '0.0.0.0/0';
};

const isRule = (value: unknown): value is Firewall.Rule => {
    if (!value || typeof value !== 'object') return false;
    const rule = value as Partial<Firewall.Rule>;
    return (
        Boolean(rule.scope) &&
        rule.scope?.provider === provider.value &&
        typeof rule.protocol === 'string' &&
        ['accept', 'drop', 'reject'].includes(String(rule.action))
    );
};

const fileOnChange = (uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    if (!uploadFile.raw) return;
    loading.value = true;
    uploaderFiles.value = uploadFiles;
    const reader = new FileReader();
    reader.onload = (event) => {
        try {
            const parsed: unknown = JSON.parse(String(event.target?.result || ''));
            if (!Array.isArray(parsed) || !parsed.every(isRule)) {
                MsgError(i18n.global.t('commons.msg.errImportFormat'));
                return;
            }
            rules.value = parsed.map((rule) => ({ ...rule, scope: { ...rule.scope }, uuid: undefined }));
            selects.value = [...rules.value];
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

const commitImportedRule = async (rule: Firewall.Rule) => {
    const plan = (await checkFirewallRule({ rule })).data;
    if (plan.decision === 'no_change') return;
    if (plan.decision === 'blocked') throw new Error(plan.reason);
    const allowed = (plan.allowedActions || []).filter(
        (value): value is Firewall.ApplicableCheckAction => value !== 'cancel',
    );
    let resolution = allowed[0];
    if (plan.classification === 'exact_external') {
        if (plan.candidates?.length !== 1 || !allowed.includes('adopt')) throw new Error(plan.reason);
        resolution = 'adopt';
    } else if (plan.classification === 'covered' && allowed.includes('create_anyway')) {
        resolution = 'create_anyway';
    }
    if (!resolution || resolution === 'select_adopt') throw new Error(plan.reason);
    await createFirewallRule({
        checkFlag: plan.checkFlag,
        action: resolution,
        adoptInstanceKey: resolution === 'adopt' ? plan.candidates?.[0]?.instanceKey : undefined,
        rule: plan.requestedRule,
        sourceKind: 'imported',
    });
};

const onImport = async () => {
    loading.value = true;
    let success = 0;
    let failed = 0;
    for (const rule of selects.value) {
        try {
            await commitImportedRule(rule);
            success++;
        } catch {
            failed++;
        }
    }
    loading.value = false;
    if (failed === 0) {
        MsgSuccess(i18n.global.t('firewall.importSuccess', [success]));
        visible.value = false;
    } else {
        MsgError(i18n.global.t('firewall.importPartialSuccess', [success, failed]));
    }
    emit('search');
};

const acceptParams = (value: Firewall.Provider) => {
    provider.value = value;
    rules.value = [];
    selects.value = [];
    uploaderFiles.value = [];
    visible.value = true;
};

defineExpose({ acceptParams });
</script>
