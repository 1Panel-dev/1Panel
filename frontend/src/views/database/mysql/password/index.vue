<template>
    <el-dialog v-model="dialogVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
        <template #header>
            <div class="card-header">
                <span>{{ $t('database.rootPassword') }}</span>
            </div>
        </template>
        <el-form v-loading="loading" ref="formRef" :model="form" label-width="80px">
            <el-form-item :label="$t('database.rootPassword')" :rules="Rules.requiredInput" prop="password">
                <el-input type="password" show-password clearable v-model="form.password" />
            </el-form-item>
        </el-form>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit"></ConfirmDialog>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="dialogVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { updateMysqlPassword } from '@/api/modules/database';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { GetAppPassword } from '@/api/modules/app';

const loading = ref(false);

const dialogVisiable = ref(false);
const form = reactive({
    password: '',
});

const confirmDialogRef = ref();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const acceptParams = (): void => {
    form.password = '';
    loadPassword();
    dialogVisiable.value = true;
};

const loadPassword = async () => {
    const res = await GetAppPassword('mysql');
    form.password = res.data;
};

const onSubmit = async () => {
    let param = {
        id: 0,
        value: form.password,
    };
    loading.value = true;
    await updateMysqlPassword(param)
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
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
