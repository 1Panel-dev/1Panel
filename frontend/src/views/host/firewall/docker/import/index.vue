<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="large">
        <el-alert class="mb-3" type="info" :closable="false" :title="$t('commons.msg.importHelper')" />
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
            <ComplexTable v-model:selects="selects" :data="policies" :height="420">
                <el-table-column type="selection" fix />
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
                    <template #default="{ row }">{{ row.sources.join(', ') || '-' }}</template>
                </el-table-column>
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
import { upsertDockerPortGuardPolicies } from '@/api/modules/firewall';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { genFileId, type UploadFile, type UploadFiles, type UploadProps, type UploadRawFile } from 'element-plus';
import { ref } from 'vue';
import { dockerGuardEndpointKey, normalizeDockerGuardPolicy } from '@/views/host/firewall/docker/model';

const emit = defineEmits<{ (event: 'search'): void }>();
const visible = ref(false);
const loading = ref(false);
const policies = ref<Firewall.DockerGuardPolicy[]>([]);
const selects = ref<Firewall.DockerGuardPolicy[]>([]);
const uploadRef = ref();
const uploaderFiles = ref<UploadFile[]>([]);

const fileOnChange = (uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    if (!uploadFile.raw) return;
    loading.value = true;
    uploaderFiles.value = uploadFiles;
    const reader = new FileReader();
    reader.onload = (event) => {
        try {
            const parsed: unknown = JSON.parse(String(event.target?.result || ''));
            if (!Array.isArray(parsed)) throw new Error();
            const normalized = parsed.map(normalizeDockerGuardPolicy);
            if (normalized.some((policy) => !policy)) throw new Error();
            const byEndpoint = new Map<string, Firewall.DockerGuardPolicy>();
            for (const policy of normalized as Firewall.DockerGuardPolicy[]) {
                byEndpoint.set(dockerGuardEndpointKey(policy), policy);
            }
            policies.value = [...byEndpoint.values()];
            selects.value = [...policies.value];
        } catch {
            policies.value = [];
            selects.value = [];
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
    loading.value = true;
    let success = 0;
    let failed = 0;
    const groups = new Map<string, Firewall.DockerGuardPolicy[]>();
    for (const policy of selects.value) {
        const key = JSON.stringify([policy.mode, policy.sources, policy.description]);
        groups.set(key, [...(groups.get(key) || []), policy]);
    }
    for (const group of groups.values()) {
        const policy = group[0];
        try {
            await upsertDockerPortGuardPolicies({
                endpoints: group.map(({ family, hostIP, hostPort, protocol }) => ({
                    family,
                    hostIP,
                    hostPort,
                    protocol,
                })),
                mode: policy.mode,
                sources: policy.sources,
                description: policy.description,
            });
            success += group.length;
        } catch {
            failed += group.length;
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

const modeLabel = (mode: Firewall.DockerGuardPolicy['mode']) => {
    if (mode === 'deny_sources') return i18n.global.t('firewall.denySources');
    if (mode === 'allow_sources') return i18n.global.t('firewall.allowSources');
    return i18n.global.t('firewall.denyAll');
};

const acceptParams = () => {
    policies.value = [];
    selects.value = [];
    uploaderFiles.value = [];
    visible.value = true;
};

defineExpose({ acceptParams });
</script>
