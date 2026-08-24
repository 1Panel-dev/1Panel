<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="large">
        <el-alert class="mb-3" type="info" :closable="false" :title="$t('firewall.importBackendHelper', [provider])" />
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
import { checkFirewallRules, createFirewallRules } from '@/api/modules/firewall';
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
    const wildcard =
        rule.scope.family === 'ipv6' ? '::/0' : rule.scope.family === 'inet' ? '0.0.0.0/0, ::/0' : '0.0.0.0/0';
    if (address && address !== wildcard) return address;
    return `${wildcard}（${i18n.global.t('firewall.anyWhere')}）`;
};

const isRule = (value: unknown): value is Firewall.Rule => {
    if (!value || typeof value !== 'object') return false;
    const rule = value as Partial<Firewall.Rule>;
    return (
        Boolean(rule.scope) &&
        ['iptables', 'nftables', 'firewalld', 'ufw'].includes(String(rule.scope?.provider)) &&
        typeof rule.protocol === 'string' &&
        ['accept', 'drop', 'reject'].includes(String(rule.action))
    );
};

const targetScope = (family: Firewall.Family): Firewall.Scope => {
    if (provider.value === 'iptables' || provider.value === 'nftables') {
        return { provider: provider.value, family, table: 'filter', chain: '1PANEL_BASIC', direction: 'input' };
    }
    if (provider.value === 'firewalld') {
        return { provider: provider.value, family, zone: 'public', direction: 'input' };
    }
    return { provider: 'ufw', family, chain: 'incoming', direction: 'input' };
};

const normalizeImportedRule = (rule: Firewall.Rule): Firewall.Rule[] => {
    const splitInet = rule.scope.family === 'inet' && provider.value !== 'firewalld';
    const addresses = [rule.sourceAddress, rule.destinationAddress].filter((value): value is string => Boolean(value));
    const families: Firewall.Family[] = !splitInet
        ? [rule.scope.family]
        : addresses.length === 0
          ? ['ipv4', 'ipv6']
          : addresses.some((address) => address.includes(':'))
            ? ['ipv6']
            : ['ipv4'];
    return families.map((family) => ({
        ...rule,
        uuid: undefined,
        nativeKind: undefined,
        priority: undefined,
        orderIndex: undefined,
        orderBucket: undefined,
        scope: targetScope(family),
    }));
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
            rules.value = parsed.flatMap(normalizeImportedRule);
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

const importedCreateRequest = (plan: Firewall.RuleCheckResult): Firewall.CreateItem | undefined => {
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
    return {
        checkFlag: plan.checkFlag,
        action: resolution,
        adoptInstanceKey: resolution === 'adopt' ? plan.candidates?.[0]?.instanceKey : undefined,
        rule: plan.requestedRule,
        sourceKind: 'imported',
    };
};

const onImport = async () => {
    loading.value = true;
    let success = 0;
    let failed = 0;
    try {
        const plans: Firewall.RuleCheckResult[] = [];
        for (let offset = 0; offset < selects.value.length; offset += 256) {
            const batch = selects.value.slice(offset, offset + 256);
            plans.push(...(await checkFirewallRules({ items: batch.map((rule) => ({ rule })) })).data.items);
        }
        const items: Firewall.CreateItem[] = [];
        for (const plan of plans) {
            try {
                const item = importedCreateRequest(plan);
                if (item) {
                    items.push(item);
                } else {
                    success++;
                }
            } catch {
                failed++;
            }
        }
        items.sort((left, right) => JSON.stringify(left.rule.scope).localeCompare(JSON.stringify(right.rule.scope)));
        for (let offset = 0; offset < items.length; offset += 256) {
            const batch = items.slice(offset, offset + 256);
            const result = (await createFirewallRules({ items: batch })).data;
            success += result.succeeded;
            failed += result.failed + result.skipped;
        }
    } catch {
        failed += selects.value.length - success - failed;
    } finally {
        loading.value = false;
    }
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
