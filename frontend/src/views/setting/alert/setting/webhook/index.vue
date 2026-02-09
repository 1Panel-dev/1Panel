<template>
    <DrawerPro v-model="drawerVisible" :header="$t('xpack.alert.' + form.type)" @close="handleClose" size="736">
        <el-form
            ref="formRef"
            :rules="rules"
            label-position="top"
            :model="form.config"
            @submit.prevent
            v-loading="loading"
        >
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('xpack.alert.webhookName')" prop="displayName">
                        <el-input v-model="form.config.displayName" />
                    </el-form-item>
                    <el-form-item :label="$t('xpack.alert.webhookUrl')" prop="url">
                        <el-input v-model.trim="form.config.url" :rows="2" type="password" show-password />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { UpdateAlertConfig } from '@/api/modules/alert';
import { Rules } from '@/global/form-rules';

const emit = defineEmits<{ (e: 'search'): void }>();

const rules = {
    displayName: [Rules.requiredInput],
    url: [Rules.requiredInput],
};
interface Config {
    displayName: string;
    url: string;
}

interface DialogProps {
    id: number;
    type: string;
    title: string;
    config: Config;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    id: undefined,
    type: '',
    title: '',
    config: {
        displayName: '',
        url: '',
    },
});
const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.id = params.id;
    form.type = params.type;
    form.title = params.title;
    form.config = params.config;
    form.config = { ...params.config };
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            const configInfo = form.config;
            await UpdateAlertConfig({
                id: form.id,
                type: form.type,
                title: form.title,
                status: 'Enable',
                config: JSON.stringify(configInfo),
            });

            loading.value = false;
            handleClose();
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        } catch (error) {
            loading.value = false;
        }
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
