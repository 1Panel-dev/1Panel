<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="large">
        <div>
            <el-alert :closable="false" show-icon type="info">
                <template #default>
                    <div>{{ $t('commons.msg.importHelper') }}</div>
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
                <el-table :data="displayData" @selection-change="handleSelectionChange">
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('commons.table.status')" :min-width="80">
                        <template #default="{ row }">
                            <Status :status="row.status" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.protocol')" :min-width="70" prop="protocol" />
                    <el-table-column :label="$t('firewall.sourcePort')" :min-width="70" prop="port" />
                    <el-table-column :label="$t('firewall.targetIP')" :min-width="100" prop="targetIP" />
                    <el-table-column :label="$t('firewall.targetPort')" :min-width="70" prop="targetPort" />
                    <el-table-column
                        v-if="currentFireName === 'ufw'"
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
import { operateForwardRule, searchFireRule, getNetworkOptions } from '@/api/modules/host';
import { Host } from '@/api/interface/host';

const emit = defineEmits<{ (e: 'search'): void }>();

const visible = ref(false);
const loading = ref(false);
const selects = ref<any>([]);
const displayData = ref<any>([]);
const currentRules = ref<Host.RuleInfo[]>([]);
const currentFireName = ref('');
const availableInterfaces = ref<string[]>([]);

const uploadRef = ref();
const uploaderFiles = ref();

const acceptParams = async (fireName: string): Promise<void> => {
    visible.value = true;
    displayData.value = [];
    selects.value = [];
    currentFireName.value = fireName;
    loadCurrentData(fireName);
};

const loadCurrentData = async (fireName: string) => {
    const res = await searchFireRule({
        type: 'forward',
        status: '',
        strategy: '',
        info: '',
        page: 1,
        pageSize: 10000,
    });
    currentRules.value = res.data.items || [];
    if (fireName === 'ufw') {
        const networkRes = await getNetworkOptions();
        availableInterfaces.value = networkRes.data || [];
    }
};

const handleSelectionChange = (val: any) => {
    selects.value = val;
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    loading.value = true;
    displayData.value = [];
    uploaderFiles.value = uploadFiles;

    const reader = new FileReader();
    reader.onload = (e) => {
        try {
            const content = e.target.result as string;
            const parsed = JSON.parse(content);

            if (!Array.isArray(parsed)) {
                MsgError(i18n.global.t('commons.msg.errImportFormat'));
                loading.value = false;
                return;
            }

            for (const item of parsed) {
                if (!checkDataFormat(item)) {
                    MsgError(i18n.global.t('commons.msg.errImportFormat'));
                    loading.value = false;
                    return;
                }
            }

            compareRules(parsed);
            loading.value = false;
        } catch (error) {
            MsgError(i18n.global.t('commons.msg.errImport') + error.message);
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
    if (!item.protocol || !item.port || !item.targetIP || !item.targetPort) {
        return false;
    }
    if (!['tcp', 'udp', 'tcp/udp'].includes(item.protocol)) {
        return false;
    }

    if (currentFireName.value === 'ufw' && item.interface !== undefined && item.interface !== null) {
        const interfaceValue = item.interface;
        if (interfaceValue !== '' && interfaceValue !== 'all') {
            if (!availableInterfaces.value.includes(interfaceValue)) {
                return false;
            }
        }
    }

    return true;
};

const compareRules = (importedRules: any[]) => {
    const newRules: any[] = [];
    const conflictRules: any[] = [];
    const duplicateRules: any[] = [];

    for (const importedRule of importedRules) {
        const key = `${importedRule.protocol}:${importedRule.port}:${importedRule.targetIP}:${importedRule.targetPort}`;

        const existingRule = currentRules.value.find((rule) => {
            const existingKey = `${rule.protocol}:${rule.port}:${rule.targetIP}:${rule.targetPort}`;
            return existingKey === key;
        });

        if (!existingRule) {
            newRules.push({ ...importedRule, status: 'new' });
        } else {
            duplicateRules.push({ ...importedRule, status: 'duplicate' });
        }
    }

    displayData.value = [...newRules, ...conflictRules, ...duplicateRules];
};

const onImport = async () => {
    loading.value = true;
    const rules: Host.RuleForward[] = [];
    for (const rule of selects.value) {
        rules.push({
            operation: 'add',
            protocol: rule.protocol,
            port: rule.port,
            targetIP: rule.targetIP,
            targetPort: rule.targetPort,
            interface: rule.interface || '',
        });
    }

    await operateForwardRule({ rules })
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
            visible.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
