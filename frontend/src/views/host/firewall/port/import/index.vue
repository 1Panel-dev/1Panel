<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="large">
        <div>
            <el-alert :closable="false" show-icon type="info">
                <template #default>
                    <div>{{ $t('firewall.importHelper') }}</div>
                </template>
            </el-alert>
            <el-upload
                action="#"
                :auto-upload="false"
                ref="uploadRef"
                class="float-left mt-2"
                :show-file-list="false"
                :limit="1"
                accept=".json"
                :on-change="fileOnChange"
                :on-exceed="handleExceed"
                v-model:file-list="uploaderFiles"
            >
                <el-button class="float-left" type="primary">{{ $t('commons.button.upload') }}</el-button>
            </el-upload>

            <el-card class="mt-2 w-full" v-loading="loading">
                <div v-if="compareResult.new.length > 0 || compareResult.conflict.length > 0">
                    <el-alert
                        :closable="false"
                        show-icon
                        type="success"
                        class="mb-2"
                        v-if="compareResult.new.length > 0"
                    >
                        <template #default>
                            <span>
                                {{ $t('firewall.importNew') }}:
                                <strong>{{ compareResult.new.length }}</strong>
                            </span>
                        </template>
                    </el-alert>
                    <el-alert
                        :closable="false"
                        show-icon
                        type="warning"
                        class="mb-2"
                        v-if="compareResult.conflict.length > 0"
                    >
                        <template #default>
                            <span>
                                {{ $t('firewall.importConflict') }}:
                                <strong>{{ compareResult.conflict.length }}</strong>
                            </span>
                        </template>
                    </el-alert>
                    <el-alert
                        :closable="false"
                        show-icon
                        type="info"
                        class="mb-2"
                        v-if="compareResult.duplicate.length > 0"
                    >
                        <template #default>
                            <span>
                                {{ $t('firewall.importDuplicate') }}:
                                <strong>{{ compareResult.duplicate.length }}</strong>
                            </span>
                        </template>
                    </el-alert>
                </div>

                <el-table :data="displayData" @selection-change="handleSelectionChange">
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('commons.table.status')" :min-width="80">
                        <template #default="{ row }">
                            <el-tag v-if="row.status === 'new'" type="success">{{ $t('firewall.new') }}</el-tag>
                            <el-tag v-else-if="row.status === 'conflict'" type="warning">
                                {{ $t('firewall.conflict') }}
                            </el-tag>
                            <el-tag v-else type="info">{{ $t('firewall.duplicate') }}</el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.protocol')" :min-width="70" prop="protocol" />
                    <el-table-column :label="$t('commons.table.port')" :min-width="70" prop="port" />
                    <el-table-column :label="$t('firewall.address')" :min-width="80">
                        <template #default="{ row }">
                            <span v-if="row.address && row.address !== 'Anywhere'">{{ row.address }}</span>
                            <span v-else>{{ $t('firewall.allIP') }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('firewall.strategy')" :min-width="80" prop="strategy">
                        <template #default="{ row }">
                            <el-tag v-if="row.strategy === 'accept'" type="success">{{ $t('firewall.accept') }}</el-tag>
                            <el-tag v-else type="danger">{{ $t('firewall.drop') }}</el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.description')"
                        :min-width="120"
                        prop="description"
                        show-overflow-tooltip
                    />
                </el-table>
            </el-card>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="visible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" :disabled="selects.length === 0" @click="onImport">
                    {{ $t('commons.button.import') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { genFileId, UploadFile, UploadFiles, UploadProps, UploadRawFile } from 'element-plus';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { operatePortRule, searchFireRule } from '@/api/modules/host';
import { Host } from '@/api/interface/host';

const emit = defineEmits<{ (e: 'search'): void }>();

const visible = ref(false);
const loading = ref(false);
const selects = ref<any>([]);
const displayData = ref<any>([]);
const currentRules = ref<Host.RuleInfo[]>([]);

const uploadRef = ref();
const uploaderFiles = ref();

interface CompareResult {
    new: any[];
    conflict: any[];
    duplicate: any[];
}

const compareResult = ref<CompareResult>({
    new: [],
    conflict: [],
    duplicate: [],
});

const acceptParams = async (): Promise<void> => {
    visible.value = true;
    displayData.value = [];
    selects.value = [];
    compareResult.value = { new: [], conflict: [], duplicate: [] };

    loading.value = true;
    try {
        const res = await searchFireRule({
            type: 'port',
            status: '',
            strategy: '',
            info: '',
            page: 1,
            pageSize: 10000,
        });
        currentRules.value = res.data.items || [];
    } catch (error) {
        MsgError(i18n.global.t('commons.msg.searchFailed'));
    } finally {
        loading.value = false;
    }
};

const handleSelectionChange = (val: any) => {
    selects.value = val;
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    loading.value = true;
    displayData.value = [];
    compareResult.value = { new: [], conflict: [], duplicate: [] };
    uploaderFiles.value = uploadFiles;

    const reader = new FileReader();
    reader.onload = (e) => {
        try {
            const content = e.target.result as string;
            const parsed = JSON.parse(content);

            if (!Array.isArray(parsed)) {
                MsgError(i18n.global.t('firewall.errImportFormat'));
                loading.value = false;
                return;
            }

            for (const item of parsed) {
                if (!checkDataFormat(item)) {
                    MsgError(i18n.global.t('firewall.errImportFormat'));
                    loading.value = false;
                    return;
                }
            }

            compareRules(parsed);
            loading.value = false;
        } catch (error) {
            MsgError(i18n.global.t('firewall.errImport') + error.message);
            loading.value = false;
        }
    };
    reader.readAsText(_uploadFile.raw);
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value!.handleStart(file);
};

const checkDataFormat = (item: any): boolean => {
    if (!item.port || !item.protocol || !item.strategy) {
        return false;
    }
    if (!['tcp', 'udp', 'tcp/udp'].includes(item.protocol)) {
        return false;
    }
    if (!['accept', 'drop'].includes(item.strategy)) {
        return false;
    }
    return true;
};

const compareRules = (importedRules: any[]) => {
    const newRules: any[] = [];
    const conflictRules: any[] = [];
    const duplicateRules: any[] = [];

    for (const importedRule of importedRules) {
        const key = `${importedRule.address || 'Anywhere'}:${importedRule.port}:${importedRule.protocol}`;

        const existingRule = currentRules.value.find((rule) => {
            const existingKey = `${rule.address || 'Anywhere'}:${rule.port}:${rule.protocol}`;
            return existingKey === key;
        });

        if (!existingRule) {
            newRules.push({ ...importedRule, status: 'new' });
        } else if (existingRule.strategy !== importedRule.strategy) {
            conflictRules.push({
                ...importedRule,
                status: 'conflict',
                existingStrategy: existingRule.strategy,
            });
        } else {
            duplicateRules.push({ ...importedRule, status: 'duplicate' });
        }
    }

    compareResult.value = {
        new: newRules,
        conflict: conflictRules,
        duplicate: duplicateRules,
    };

    displayData.value = [...newRules, ...conflictRules, ...duplicateRules];
};

const onImport = async () => {
    if (selects.value.length === 0) {
        MsgError(i18n.global.t('firewall.selectImportRules'));
        return;
    }

    loading.value = true;
    let successCount = 0;
    let errorCount = 0;

    for (const rule of selects.value) {
        try {
            const params: Host.RulePort = {
                operation: 'add',
                address: rule.address || 'Anywhere',
                port: rule.port,
                source: '',
                protocol: rule.protocol,
                strategy: rule.strategy,
                description: rule.description || '',
            };

            await operatePortRule(params);
            successCount++;
        } catch (error) {
            errorCount++;
            console.error('Failed to import rule:', rule, error);
        }
    }

    loading.value = false;

    if (errorCount === 0) {
        MsgSuccess(i18n.global.t('firewall.importSuccess', [successCount]));
        visible.value = false;
        emit('search');
    } else {
        MsgError(i18n.global.t('firewall.importPartialSuccess', [successCount, errorCount]));
        emit('search');
    }
};

defineExpose({
    acceptParams,
});
</script>
