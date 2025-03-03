<template>
    <div>
        <DrawerPro v-model="drawerVisible" :header="$t('container.mirrors')" @close="handleClose" size="small">
            <el-form
                ref="formRef"
                label-position="top"
                :model="form"
                @submit.prevent
                :rules="rules"
                v-loading="loading"
            >
                <el-form-item :label="$t('container.mirrors')" prop="mirrors">
                    <el-input
                        type="textarea"
                        :placeholder="$t('container.mirrorHelper')"
                        :rows="5"
                        v-model="form.mirrors"
                    />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </template>
        </DrawerPro>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit" />
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { updateDaemonJson } from '@/api/modules/container';
import { FormInstance } from 'element-plus';
import { emptyLineFilter } from '@/utils/util';

const emit = defineEmits<{ (e: 'search'): void }>();

const confirmDialogRef = ref();

interface DialogProps {
    mirrors: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    mirrors: '',
});
const formRef = ref<FormInstance>();
const rules = reactive({
    mirrors: [{ validator: checkMirrors, trigger: 'blur' }],
});

function checkMirrors(rule: any, value: any, callback: any) {
    if (form.mirrors !== '') {
        const reg = /^https?:\/\/[a-zA-Z0-9.-]+(:[0-9]{1,5})?(\/[a-zA-Z0-9./-]*)?$/;
        let mirrors = form.mirrors.split('\n');
        for (const item of mirrors) {
            if (item === '') {
                continue;
            }
            if (!reg.test(item)) {
                return callback(new Error(i18n.global.t('commons.rule.mirror')));
            }
        }
    }
    callback();
}

const acceptParams = (params: DialogProps): void => {
    form.mirrors = params.mirrors || params.mirrors.replaceAll(',', '\n');
    drawerVisible.value = true;
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

const onSubmit = async () => {
    loading.value = true;
    await updateDaemonJson('Mirrors', emptyLineFilter(form.mirrors, '\n').replaceAll('\n', ','))
        .then(() => {
            loading.value = false;
            emit('search');
            handleClose();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
