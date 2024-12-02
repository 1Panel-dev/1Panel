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
            <el-form @submit.prevent ref="formRef" v-loading="loading" label-position="top" :model="kodoData.rowData">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                            <el-tag>{{ $t('setting.' + kodoData.rowData!.type) }}</el-tag>
                        </el-form-item>
                        <el-form-item label="Access key ID" prop="accessKey" :rules="Rules.requiredInput">
                            <el-input v-model.trim="kodoData.rowData!.accessKey" />
                        </el-form-item>
                        <el-form-item label="Secret key" prop="credential" :rules="Rules.requiredInput">
                            <el-input show-password clearable v-model.trim="kodoData.rowData!.credential" />
                        </el-form-item>
                        <el-form-item
                            :label="$t('setting.domain')"
                            prop="varsJson.domainItem"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model="kodoData.rowData!.varsJson['domainItem']">
                                <template #prepend>
                                    <el-select v-model.trim="domainProto" style="width: 120px">
                                        <el-option label="http" value="http" />
                                        <el-option label="https" value="https" />
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item label="Bucket" prop="bucket" :rules="Rules.requiredInput">
                            <el-checkbox v-model="kodoData.rowData!.bucketInput" :label="$t('container.input')" />
                            <el-input
                                clearable
                                v-if="kodoData.rowData!.bucketInput"
                                v-model="kodoData.rowData!.bucket"
                            />
                            <div v-else class="w-full">
                                <el-select style="width: 80%" v-model="kodoData.rowData!.bucket">
                                    <el-option v-for="item in buckets" :key="item" :value="item" />
                                </el-select>
                                <el-button style="width: 20%" plain @click="getBuckets()">
                                    {{ $t('setting.loadBucket') }}
                                </el-button>
                            </div>
                        </el-form-item>

                        <el-form-item :label="$t('cronjob.requestExpirationTime')" prop="varsJson.timeout">
                            <el-input-number
                                style="width: 200px"
                                :min="1"
                                step-strictly
                                :step="1"
                                v-model.number="kodoData.rowData!.varsJson['timeout']"
                            ></el-input-number>
                        </el-form-item>

                        <el-form-item :label="$t('setting.backupDir')" prop="backupPath">
                            <el-input clearable v-model.trim="kodoData.rowData!.backupPath" placeholder="/1panel" />
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

const domainProto = ref('http');
const emit = defineEmits(['search']);

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const kodoData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    buckets.value = [];
    kodoData.value = params;
    if (kodoData.value.title === 'edit') {
        let httpItem = splitHttp(kodoData.value.rowData!.varsJson['domain']);
        kodoData.value.rowData!.varsJson['domainItem'] = httpItem.url;
        domainProto.value = httpItem.proto;
    }
    title.value = i18n.global.t('commons.button.' + kodoData.value.title);
    if (kodoData.value.rowData!.varsJson['timeout'] === undefined) {
        kodoData.value.rowData!.varsJson['timeout'] = 1;
    }
    drawerVisible.value = true;
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const getBuckets = async () => {
    loading.value = true;
    let item = deepCopy(kodoData.value.rowData!.varsJson);
    item['domain'] = spliceHttp(domainProto.value, kodoData.value.rowData!.varsJson['domainItem']);
    listBucket({
        type: kodoData.value.rowData!.type,
        vars: JSON.stringify(item),
        accessKey: kodoData.value.rowData!.accessKey,
        credential: kodoData.value.rowData!.credential,
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
        if (!kodoData.value.rowData) return;
        kodoData.value.rowData!.varsJson['domain'] = spliceHttp(
            domainProto.value,
            kodoData.value.rowData!.varsJson['domainItem'],
        );
        kodoData.value.rowData.vars = JSON.stringify(kodoData.value.rowData!.varsJson);
        loading.value = true;
        if (kodoData.value.title === 'create') {
            await addBackup(kodoData.value.rowData)
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
        await editBackup(kodoData.value.rowData)
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
