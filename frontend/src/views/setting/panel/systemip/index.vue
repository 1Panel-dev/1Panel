<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.systemIP')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('setting.systemIP')" prop="systemIP">
                <el-input clearable v-model="form.systemIP" />
                <span class="input-help">{{ $t('commons.rule.hostHelper') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="loading" type="primary" @click="onSaveSystemIP(formRef)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { updateAgentSetting } from '@/api/modules/setting';
import { FormInstance } from 'element-plus';
import { checkDomain, checkIpV4V6 } from '@/utils/util';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    systemIP: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    systemIP: '',
});

const formRef = ref<FormInstance>();
const rules = reactive({
    systemIP: [{ validator: checkSystemIP, trigger: 'blur' }],
});

function checkSystemIP(rule: any, value: any, callback: any) {
    if (form.systemIP !== '') {
        if (checkIpV4V6(form.systemIP) && checkDomain(form.systemIP)) {
            return callback(new Error(i18n.global.t('commons.rule.host')));
        }
    }
    callback();
}

const acceptParams = (params: DialogProps): void => {
    form.systemIP = params.systemIP;
    drawerVisible.value = true;
};

const onSaveSystemIP = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await updateAgentSetting({ key: 'SystemIP', value: form.systemIP })
            .then(async () => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loading.value = false;
                drawerVisible.value = false;
                emit('search');
                return;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
