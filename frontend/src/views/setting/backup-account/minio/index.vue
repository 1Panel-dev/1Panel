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
            <el-form @submit.prevent ref="formRef" v-loading="loading" label-position="top" :model="minioData.rowData">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                            <el-tag>{{ $t('setting.' + minioData.rowData!.type) }}</el-tag>
                        </el-form-item>
                        <el-form-item label="Access key ID" prop="accessKey" :rules="Rules.requiredInput">
                            <el-input v-model.trim="minioData.rowData!.accessKey" />
                        </el-form-item>
                        <el-form-item label="Secret key" prop="credential" :rules="Rules.requiredInput">
                            <el-input show-password clearable v-model.trim="minioData.rowData!.credential" />
                        </el-form-item>
                        <el-form-item label="Endpoint" prop="varsJson.endpointItem" :rules="Rules.requiredInput">
                            <el-input v-model="minioData.rowData!.varsJson['endpointItem']">
                                <template #prepend>
                                    <el-select v-model.trim="endpointProto" style="width: 120px">
                                        <el-option label="http" value="http" />
                                        <el-option label="https" value="https" />
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item label="Bucket" prop="bucket" :rules="Rules.requiredInput">
                            <el-checkbox v-model="minioData.rowData!.bucketInput" :label="$t('container.input')" />
                            <el-input
                                clearable
                                v-if="minioData.rowData!.bucketInput"
                                v-model="minioData.rowData!.bucket"
                            />
                            <div v-else class="w-full">
                                <el-select style="width: 80%" v-model="minioData.rowData!.bucket">
                                    <el-option v-for="item in buckets" :key="item" :value="item" />
                                </el-select>
                                <el-button style="width: 20%" plain @click="getBuckets()">
                                    {{ $t('setting.loadBucket') }}
                                </el-button>
                            </div>
                        </el-form-item>
                        <el-form-item :label="$t('setting.backupDir')" prop="backupPath">
                            <el-input clearable v-model.trim="minioData.rowData!.backupPath" placeholder="/1panel" />
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
import { deepCopy, splitHttp, spliceHttp } from '@/utils/util';
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
const minioData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    buckets.value = [];
    minioData.value = params;
    if (minioData.value.title === 'edit') {
        let httpItem = splitHttp(minioData.value.rowData!.varsJson['endpoint']);
        minioData.value.rowData!.varsJson['endpointItem'] = httpItem.url;
        endpointProto.value = httpItem.proto;
    }
    title.value = i18n.global.t('commons.button.' + minioData.value.title);
    drawerVisible.value = true;
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const getBuckets = async () => {
    loading.value = true;
    let item = deepCopy(minioData.value.rowData!.varsJson);
    item['endpoint'] = spliceHttp(endpointProto.value, minioData.value.rowData!.varsJson['endpointItem']);
    item['endpointItem'] = undefined;
    listBucket({
        type: minioData.value.rowData!.type,
        vars: JSON.stringify(item),
        accessKey: minioData.value.rowData!.accessKey,
        credential: minioData.value.rowData!.credential,
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
        if (!minioData.value.rowData) return;
        minioData.value.rowData!.varsJson['endpoint'] = spliceHttp(
            endpointProto.value,
            minioData.value.rowData!.varsJson['endpointItem'],
        );
        minioData.value.rowData!.varsJson['endpointItem'] = undefined;
        minioData.value.rowData.vars = JSON.stringify(minioData.value.rowData!.varsJson);
        loading.value = true;
        if (minioData.value.title === 'create') {
            await addBackup(minioData.value.rowData)
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
        await editBackup(minioData.value.rowData)
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
