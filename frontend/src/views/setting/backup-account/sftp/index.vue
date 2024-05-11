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
            <el-form @submit.prevent ref="formRef" v-loading="loading" label-position="top" :model="sftpData.rowData">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                            <el-tag>{{ $t('setting.' + sftpData.rowData!.type) }}</el-tag>
                        </el-form-item>
                        <el-form-item :label="$t('setting.address')" prop="varsJson.address" :rules="Rules.host">
                            <el-input v-model.trim="sftpData.rowData!.varsJson['address']" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.port')" prop="varsJson.port" :rules="[Rules.port]">
                            <el-input-number
                                :min="0"
                                :max="65535"
                                v-model.number="sftpData.rowData!.varsJson['port']"
                            />
                        </el-form-item>
                        <el-form-item
                            :label="$t('commons.login.username')"
                            prop="accessKey"
                            :rules="[Rules.requiredInput]"
                        >
                            <el-input v-model.trim="sftpData.rowData!.accessKey" />
                        </el-form-item>
                        <el-form-item
                            :label="$t('commons.login.password')"
                            prop="credential"
                            :rules="[Rules.requiredInput]"
                        >
                            <el-input
                                type="password"
                                clearable
                                show-password
                                v-model.trim="sftpData.rowData!.credential"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('setting.backupDir')" prop="bucket" :rules="[Rules.requiredInput]">
                            <el-input v-model.trim="sftpData.rowData!.bucket" />
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
import { addBackup, editBackup } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const emit = defineEmits(['search']);

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const sftpData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    sftpData.value = params;
    if (sftpData.value.title === 'create') {
        sftpData.value.rowData.varsJson['port'] = 22;
    }
    title.value = i18n.global.t('commons.button.' + sftpData.value.title);
    drawerVisible.value = true;
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!sftpData.value.rowData) return;
        sftpData.value.rowData.vars = JSON.stringify(sftpData.value.rowData!.varsJson);
        loading.value = true;
        if (sftpData.value.title === 'create') {
            await addBackup(sftpData.value.rowData)
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
        await editBackup(sftpData.value.rowData)
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
