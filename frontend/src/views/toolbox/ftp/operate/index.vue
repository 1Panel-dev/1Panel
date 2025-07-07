<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="title"
        :resource="dialogData.title === 'add' ? '' : dialogData.rowData?.user"
        @close="handleClose"
        size="large"
    >
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('commons.login.username')" prop="user">
                <el-input :disabled="dialogData.title === 'edit'" clearable v-model.trim="dialogData.rowData!.user" />
            </el-form-item>
            <el-form-item :label="$t('commons.login.password')" prop="password">
                <el-input type="password" clearable v-model="dialogData.rowData!.password" show-password>
                    <template #append>
                        <el-button @click="random">{{ $t('commons.button.random') }}</el-button>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('file.root')" prop="path">
                <el-input v-model="dialogData.rowData!.path">
                    <template #prepend>
                        <FileList @choose="loadDir" :dir="true"></FileList>
                    </template>
                </el-input>
                <span class="input-help">{{ $t('toolbox.ftp.dirHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input type="textarea" :rows="3" clearable v-model="dialogData.rowData!.description" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import { Toolbox } from '@/api/interface/toolbox';
import { createFtp, updateFtp } from '@/api/modules/toolbox';
import { getRandomStr, isSensitiveLinuxPath } from '@/utils/util';

interface DialogProps {
    title: string;
    rowData?: Toolbox.FtpInfo;
    getTableList?: () => Promise<any>;
}
const loading = ref();
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const random = async () => {
    dialogData.value.rowData.password = getRandomStr(16);
};

const handleClose = () => {
    drawerVisible.value = false;
};

const verifyPath = (rule: any, value: any, callback: any) => {
    if (isSensitiveLinuxPath(dialogData.value.rowData.path)) {
        callback(new Error(i18n.global.t('toolbox.ftp.dirSystem')));
        return;
    }
    callback();
};

const rules = reactive({
    user: [Rules.simpleName],
    password: [Rules.simplePassword],
    path: [Rules.requiredInput, Rules.noSpace, { validator: verifyPath, trigger: 'change', required: true }],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const loadDir = async (path: string) => {
    dialogData.value.rowData!.path = path;
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('toolbox.ftp.dirMsg', [dialogData.value.rowData.path]), 'FTP', {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            loading.value = true;
            if (dialogData.value.title === 'edit') {
                await updateFtp(dialogData.value.rowData)
                    .then(() => {
                        loading.value = false;
                        drawerVisible.value = false;
                        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                        emit('search');
                    })
                    .catch(() => {
                        loading.value = false;
                    });

                return;
            }

            await createFtp(dialogData.value.rowData)
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
    });
};

defineExpose({
    acceptParams,
});
</script>
