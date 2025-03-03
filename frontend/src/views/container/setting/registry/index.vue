<template>
    <div>
        <DrawerPro v-model="drawerVisible" :header="$t('container.registries')" @close="handleClose" size="small">
            <el-form
                ref="formRef"
                label-position="top"
                :model="form"
                :rules="rules"
                @submit.prevent
                v-loading="loading"
            >
                <el-form-item :label="$t('container.registries')" prop="registries">
                    <el-input
                        type="textarea"
                        :placeholder="$t('container.registrieHelper')"
                        :rows="5"
                        v-model="form.registries"
                    />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSave">
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
    registries: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    registries: '',
});
const formRef = ref<FormInstance>();
const rules = reactive({
    registries: [{ validator: checkRegistries, trigger: 'blur' }],
});

function checkRegistries(rule: any, value: any, callback: any) {
    if (form.registries !== '') {
        const reg = /^[a-zA-Z0-9]{1}[a-z:A-Z0-9_/.-]{0,150}$/;
        let regis = form.registries.split('\n');
        for (const item of regis) {
            if (item === '') {
                continue;
            }
            if (!reg.test(item)) {
                return callback(new Error(i18n.global.t('commons.rule.imageName')));
            }
        }
    }
    callback();
}

const acceptParams = (params: DialogProps): void => {
    form.registries = params.registries || params.registries.replaceAll(',', '\n');
    drawerVisible.value = true;
};

const onSave = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const onSubmit = async () => {
    loading.value = true;
    await updateDaemonJson('Registries', emptyLineFilter(form.registries, '\n').replaceAll('\n', ','))
        .then(() => {
            loading.value = false;
            handleClose();
            emit('search');
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
