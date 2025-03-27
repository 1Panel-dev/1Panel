<template>
    <DrawerPro v-model="drawerVisible" :header="$t('app.defaultWebDomain')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('app.defaultWebDomain')" prop="defaultDomain">
                <el-input v-model="form.defaultDomain"></el-input>
                <span class="input-help">{{ $t('app.defaultWebDomainHepler') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="handleClose()">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="loading" type="primary" @click="submit()">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { updateAppStoreConfig } from '@/api/modules/app';
import { MsgSuccess } from '@/utils/message';

const emit = defineEmits<{ (e: 'close'): void }>();
const drawerVisible = ref();
const loading = ref();
const form = reactive({
    defaultDomain: '',
});
const rules = reactive({
    defaultDomain: [Rules.requiredInput, Rules.ipV4V6OrDomain],
});
const formRef = ref<FormInstance>();
interface DialogProps {
    domain: string;
}

const acceptParams = (config: DialogProps): void => {
    form.defaultDomain = config.domain;
    drawerVisible.value = true;
};

const handleClose = () => {
    drawerVisible.value = false;
    emit('close');
};

const submit = async () => {
    if (!formRef.value) return;
    await formRef.value.validate(async (valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        try {
            const req = {
                defaultDomain: form.defaultDomain,
            };
            await updateAppStoreConfig(req);
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        } catch (error) {
        } finally {
            loading.value = false;
            handleClose();
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
