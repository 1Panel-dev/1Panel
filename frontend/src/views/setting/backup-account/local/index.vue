<template>
    <div v-loading="loading">
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
            <el-form @submit.prevent ref="formRef" v-loading="loading" label-position="top" :model="dialogData.rowData">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                            <el-tag>{{ $t('setting.' + dialogData.rowData!.type) }}</el-tag>
                        </el-form-item>
                        <el-form-item
                            :label="$t('setting.backupDir')"
                            prop="varsJson['dir']"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model="dialogData.rowData!.varsJson['dir']">
                                <template #prepend>
                                    <FileList @choose="loadDir" :dir="true"></FileList>
                                </template>
                            </el-input>
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
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Backup } from '@/api/interface/backup';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { addBackup, editBackup } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
}
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

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const loadDir = async (path: string) => {
    dialogData.value.rowData!.varsJson['dir'] = path;
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!dialogData.value.rowData) return;
        dialogData.value.rowData.vars = JSON.stringify(dialogData.value.rowData!.varsJson);
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await addBackup(dialogData.value.rowData)
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
        await editBackup(dialogData.value.rowData)
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

<style lang="scss" scoped>
.append-button {
    width: 80px;
    background-color: var(--el-fill-color-light);
    color: var(--el-color-info);
}
</style>
