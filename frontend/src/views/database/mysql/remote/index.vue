<template>
    <el-drawer v-model="dialogVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <DrawerHeader :header="$t('database.remoteAccess')" :back="handleClose" />
        </template>
        <el-form @submit.prevent v-loading="loading" ref="formRef" :model="form" label-position="top">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('database.remoteAccess')" :rules="Rules.requiredInput" prop="privilege">
                        <el-switch v-model="form.privilege" />
                        <span class="input-help">{{ $t('database.remoteConnHelper') }}</span>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit"></ConfirmDialog>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { updateMysqlAccess } from '@/api/modules/database';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const dialogVisiable = ref(false);
const form = reactive({
    privilege: false,
});

const confirmDialogRef = ref();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

interface DialogProps {
    privilege: boolean;
}

const acceptParams = (prop: DialogProps): void => {
    form.privilege = prop.privilege;
    dialogVisiable.value = true;
};

const handleClose = () => {
    dialogVisiable.value = false;
};

const onSubmit = async () => {
    let param = {
        id: 0,
        value: form.privilege ? '%' : 'localhost',
    };
    loading.value = true;
    await updateMysqlAccess(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            dialogVisiable.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};

defineExpose({
    acceptParams,
});
</script>
