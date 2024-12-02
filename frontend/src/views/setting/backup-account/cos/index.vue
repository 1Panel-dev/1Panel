<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader :header="title + $t('setting.backupAccount')" :back="handleClose" />
            </template>
            <el-form @submit.prevent ref="formRef" v-loading="loading" label-position="top" :model="cosData.rowData">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                            <el-tag>{{ $t('setting.' + cosData.rowData!.type) }}</el-tag>
                        </el-form-item>
                        <el-form-item label="Access key ID" prop="accessKey" :rules="Rules.requiredInput">
                            <el-input v-model.trim="cosData.rowData!.accessKey" />
                        </el-form-item>
                        <el-form-item label="Secret key" prop="credential" :rules="Rules.requiredInput">
                            <el-input show-password clearable v-model.trim="cosData.rowData!.credential" />
                        </el-form-item>
                        <el-form-item label="Region" prop="varsJson.region" :rules="Rules.requiredInput">
                            <el-checkbox v-model="regionInput" :label="$t('container.input')" />
                            <el-select
                                v-if="!regionInput"
                                v-model="cosData.rowData!.varsJson['region']"
                                filterable
                                clearable
                            >
                                <el-option
                                    v-for="item in cities"
                                    :key="item.value"
                                    :label="item.label"
                                    :value="item.value"
                                >
                                    <span style="float: left">{{ item.label }}</span>
                                    <span class="option-help">
                                        {{ item.value }}
                                    </span>
                                </el-option>
                            </el-select>
                            <el-input v-else v-model.trim="cosData.rowData!.varsJson['region']" />
                        </el-form-item>
                        <el-form-item label="Endpoint" prop="varsJson.endpointItem" :rules="Rules.requiredInput">
                            <el-input v-model.trim="cosData.rowData!.varsJson['endpointItem']">
                                <template #prepend>
                                    <el-select v-model.trim="endpointProto" style="width: 120px">
                                        <el-option label="http" value="http" />
                                        <el-option label="https" value="https" />
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item label="Bucket" prop="bucket" :rules="Rules.requiredInput">
                            <el-checkbox v-model="cosData.rowData!.bucketInput" :label="$t('container.input')" />
                            <el-input clearable v-if="cosData.rowData!.bucketInput" v-model="cosData.rowData!.bucket" />
                            <div v-else class="w-full">
                                <el-select style="width: 80%" v-model="cosData.rowData!.bucket">
                                    <el-option v-for="item in buckets" :key="item" :value="item" />
                                </el-select>
                                <el-button style="width: 20%" plain @click="getBuckets()">
                                    {{ $t('setting.loadBucket') }}
                                </el-button>
                            </div>
                        </el-form-item>
                        <el-form-item
                            :label="$t('setting.scType')"
                            prop="varsJson.scType"
                            :rules="[Rules.requiredSelect]"
                        >
                            <el-select v-model="cosData.rowData!.varsJson['scType']">
                                <el-option value="Standard" :label="$t('setting.scStandard')" />
                                <el-option value="Standard_IA" :label="$t('setting.scStandard_IA')" />
                                <el-option value="Archive" :label="$t('setting.scArchive')" />
                                <el-option value="Deep_Archive" :label="$t('setting.scDeep_Archive')" />
                            </el-select>
                            <el-alert
                                v-if="cosData.rowData!.varsJson['scType'] === 'Archive' || cosData.rowData!.varsJson['scType'] === 'Deep_Archive'"
                                style="margin-top: 10px"
                                :closable="false"
                                type="warning"
                                :title="$t('setting.archiveHelper')"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('setting.backupDir')" prop="backupPath">
                            <el-input clearable v-model.trim="cosData.rowData!.backupPath" placeholder="/1panel" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="handleClose">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Backup } from '@/api/interface/backup';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { addBackup, editBackup, listBucket } from '@/api/modules/setting';
import { deepCopy, spliceHttp, splitHttp } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const buckets = ref();
const regionInput = ref();

const endpointProto = ref('http');
const emit = defineEmits(['search']);

const cities = [
    { value: 'ap-beijing-1', label: i18n.global.t('setting.ap_beijing_1') },
    { value: 'ap-beijing', label: i18n.global.t('setting.ap_beijing') },
    { value: 'ap-nanjing', label: i18n.global.t('setting.ap_nanjing') },
    { value: 'ap-shanghai', label: i18n.global.t('setting.ap_shanghai') },
    { value: 'ap-guangzhou', label: i18n.global.t('setting.ap_guangzhou') },
    { value: 'ap-chengdu', label: i18n.global.t('setting.ap_chengdu') },
    { value: 'ap-chongqing', label: i18n.global.t('setting.ap_chongqing') },
    { value: 'ap-shenzhen_fsi', label: i18n.global.t('setting.ap_shenzhen_fsi') },
    { value: 'ap-shanghai_fsi', label: i18n.global.t('setting.ap_shanghai_fsi') },
    { value: 'ap-beijing_fsi', label: i18n.global.t('setting.ap_beijing_fsi') },
    { value: 'ap-hongkong', label: i18n.global.t('setting.ap_hongkong') },
    { value: 'ap-singapore', label: i18n.global.t('setting.ap_singapore') },
    { value: 'ap-mumbai', label: i18n.global.t('setting.ap_mumbai') },
    { value: 'ap-jakarta', label: i18n.global.t('setting.ap_jakarta') },
    { value: 'ap-seoul', label: i18n.global.t('setting.ap_seoul') },
    { value: 'ap-bangkok', label: i18n.global.t('setting.ap_bangkok') },
    { value: 'ap-tokyo', label: i18n.global.t('setting.ap_tokyo') },
    { value: 'na-siliconvalley', label: i18n.global.t('setting.na_siliconvalley') },
    { value: 'na-ashburn', label: i18n.global.t('setting.na_ashburn') },
    { value: 'na-toronto', label: i18n.global.t('setting.na_toronto') },
    { value: 'sa-saopaulo', label: i18n.global.t('setting.sa_saopaulo') },
    { value: 'eu-frankfurt', label: i18n.global.t('setting.eu_frankfurt') },
];

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const cosData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    buckets.value = [];
    cosData.value = params;
    if (params.title === 'create' || (params.title === 'edit' && !cosData.value.rowData.varsJson['scType'])) {
        cosData.value.rowData.varsJson['scType'] = 'Standard';
    }
    if (cosData.value.title === 'edit') {
        let httpItem = splitHttp(cosData.value.rowData!.varsJson['endpoint']);
        cosData.value.rowData!.varsJson['endpointItem'] = httpItem.url;
        endpointProto.value = httpItem.proto;
    }
    title.value = i18n.global.t('commons.button.' + cosData.value.title);
    drawerVisible.value = true;
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const getBuckets = async () => {
    loading.value = true;
    let item = deepCopy(cosData.value.rowData!.varsJson);
    item['endpoint'] = spliceHttp(endpointProto.value, cosData.value.rowData!.varsJson['endpointItem']);
    listBucket({
        type: cosData.value.rowData!.type,
        vars: JSON.stringify(item),
        accessKey: cosData.value.rowData!.accessKey,
        credential: cosData.value.rowData!.credential,
    })
        .then((res) => {
            loading.value = false;
            buckets.value = res.data;
        })
        .catch(() => {
            buckets.value = [];
            loading.value = false;
        });
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!cosData.value.rowData) return;
        cosData.value.rowData!.varsJson['endpoint'] = spliceHttp(
            endpointProto.value,
            cosData.value.rowData!.varsJson['endpointItem'],
        );
        cosData.value.rowData.vars = JSON.stringify(cosData.value.rowData!.varsJson);
        loading.value = true;
        if (cosData.value.title === 'create') {
            await addBackup(cosData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisible.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
            return;
        }
        await editBackup(cosData.value.rowData)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisible.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.option-help {
    float: right;
    font-size: 12px;
    word-break: break-all;
    color: #8f959e;
}
</style>
