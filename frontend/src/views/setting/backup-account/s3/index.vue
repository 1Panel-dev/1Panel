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
                <DrawerHeader :header="title + $t('setting.backupAccount').toLowerCase()" :back="handleClose" />
            </template>
            <el-form @submit.prevent ref="formRef" v-loading="loading" label-position="top" :model="s3Data.rowData">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                            <el-tag>{{ $t('setting.' + s3Data.rowData!.type) }}</el-tag>
                        </el-form-item>
                        <el-form-item :label="$t('setting.mode')" prop="varsJson.mode" :rules="Rules.requiredSelect">
                            <el-radio-group v-model="s3Data.rowData!.varsJson['mode']">
                                <el-radio value="virtual hosted">Virtual Hosted</el-radio>
                                <el-radio value="path">Path</el-radio>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item label="Access key ID" prop="accessKey" :rules="Rules.requiredInput">
                            <el-input v-model.trim="s3Data.rowData!.accessKey" />
                        </el-form-item>
                        <el-form-item label="Secret Key" prop="credential" :rules="Rules.requiredInput">
                            <el-input show-password clearable v-model.trim="s3Data.rowData!.credential" />
                        </el-form-item>
                        <el-form-item label="Region" prop="varsJson.region" :rules="Rules.requiredInput">
                            <el-input pro v-model.trim="s3Data.rowData!.varsJson['region']" />
                        </el-form-item>
                        <el-form-item label="Endpoint" prop="varsJson.endpointItem" :rules="Rules.requiredInput">
                            <el-input v-model="s3Data.rowData!.varsJson['endpointItem']">
                                <template #prepend>
                                    <el-select v-model="endpointProto" style="width: 120px">
                                        <el-option label="http" value="http" />
                                        <el-option label="https" value="https" />
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item label="Bucket" prop="bucket" :rules="Rules.requiredInput">
                            <el-checkbox v-model="s3Data.rowData!.bucketInput" :label="$t('container.input')" />
                            <el-input clearable v-if="s3Data.rowData!.bucketInput" v-model="s3Data.rowData!.bucket" />
                            <div v-else class="w-full">
                                <el-select style="width: 80%" v-model="s3Data.rowData!.bucket">
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
                            <el-select v-model="s3Data.rowData!.varsJson['scType']">
                                <el-option value="STANDARD" :label="$t('setting.scStandard')" />
                                <el-option value="STANDARD_IA" :label="$t('setting.scStandard_IA')" />
                                <el-option value="GLACIER" :label="$t('setting.scArchive')" />
                                <el-option value="DEEP_ARCHIVE" :label="$t('setting.scDeep_Archive')" />
                            </el-select>
                            <el-alert
                                v-if="s3Data.rowData!.varsJson['scType'] === 'Archive' || s3Data.rowData!.varsJson['scType'] === 'ColdArchive'"
                                style="margin-top: 10px"
                                :closable="false"
                                type="warning"
                                :title="$t('setting.archiveHelper')"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('setting.backupDir')" prop="backupPath">
                            <el-input clearable v-model.trim="s3Data.rowData!.backupPath" placeholder="/1panel" />
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

const endpointProto = ref('http');
const emit = defineEmits(['search']);

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const s3Data = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    buckets.value = [];
    s3Data.value = params;
    if (!s3Data.value.rowData.varsJson['scType']) {
        s3Data.value.rowData.varsJson['scType'] = 'STANDARD';
    }
    if (!s3Data.value.rowData.varsJson['mode']) {
        s3Data.value.rowData.varsJson['mode'] = 'virtual hosted';
    }
    if (s3Data.value.title === 'edit') {
        let httpItem = splitHttp(s3Data.value.rowData!.varsJson['endpoint']);
        s3Data.value.rowData!.varsJson['endpointItem'] = httpItem.url;
        endpointProto.value = httpItem.proto;
    }
    title.value = i18n.global.t('commons.button.' + s3Data.value.title);
    drawerVisible.value = true;
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const getBuckets = async () => {
    loading.value = true;
    let item = deepCopy(s3Data.value.rowData!.varsJson);
    item['endpoint'] = spliceHttp(endpointProto.value, s3Data.value.rowData!.varsJson['endpointItem']);
    listBucket({
        type: s3Data.value.rowData!.type,
        vars: JSON.stringify(item),
        accessKey: s3Data.value.rowData!.accessKey,
        credential: s3Data.value.rowData!.credential,
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
        if (!s3Data.value.rowData) return;
        s3Data.value.rowData!.varsJson['endpoint'] = spliceHttp(
            endpointProto.value,
            s3Data.value.rowData!.varsJson['endpointItem'],
        );
        s3Data.value.rowData.vars = JSON.stringify(s3Data.value.rowData!.varsJson);
        loading.value = true;
        if (s3Data.value.title === 'create') {
            await addBackup(s3Data.value.rowData)
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
        await editBackup(s3Data.value.rowData)
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
