<template>
    <DrawerPro v-model="drawerVisible" :header="$t('xpack.alert.emailConfig')" @close="handleClose" size="736">
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
                    <el-form-item :label="$t('xpack.alert.displayName')" prop="displayName">
                        <el-input v-model="form.config.displayName" />
                        <span class="input-help">
                            {{ $t('xpack.alert.displayNameHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('xpack.alert.sender')" prop="sender">
                        <el-input v-model.trim="form.config.sender" />
                        <span class="input-help">
                            {{ $t('xpack.alert.senderHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('xpack.alert.password')" prop="password">
                        <el-input v-model.trim="form.config.password" type="password" show-password />
                        <span class="input-help">
                            {{ $t('xpack.alert.passwordHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('xpack.alert.host')" prop="host">
                        <el-input v-model.trim="form.config.host" placeholder="smtp.qq.com" />
                        <span class="input-help">
                            {{ $t('xpack.alert.hostHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('xpack.alert.port')" prop="port">
                        <el-input v-model.number="form.config.port" :min="1" :max="65535" />
                    </el-form-item>
                    <el-form-item :label="$t('xpack.alert.encryption')" prop="encryption">
                        <div class="flex items-center gap-2">
                            <span class="el-form-item__label">SSL</span>
                            <el-switch
                                v-model="form.config.encryption"
                                :active-value="'SSL'"
                                :inactive-value="form.config.encryption === 'SSL' ? 'NONE' : form.config.encryption"
                            />
                        </div>
                        <span class="input-help">
                            {{ $t('xpack.alert.sslHelper') }}
                        </span>
                        <div class="flex items-center gap-2">
                            <span class="el-form-item__label">TLS</span>
                            <el-switch
                                v-model="form.config.encryption"
                                :active-value="'TLS'"
                                :inactive-value="form.config.encryption === 'TLS' ? 'NONE' : form.config.encryption"
                            />
                        </div>
                        <span class="input-help">
                            {{ $t('xpack.alert.tlsHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('xpack.alert.recipient')" prop="recipient">
                        <el-input v-model="form.config.recipient" />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <div class="flex items-center justify-between">
                <el-button @click="onTest(formRef)" plain type="primary">
                    {{ $t('xpack.alert.test') }}
                </el-button>
                <div>
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading || !isOK" type="primary" @click="onSave(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </div>
            </div>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { TestAlertConfig, UpdateAlertConfig } from '@/api/modules/alert';
import { Rules } from '@/global/form-rules';

const emit = defineEmits<{ (e: 'search'): void }>();

const rules = {
    displayName: [Rules.requiredInput],
    sender: [Rules.requiredInput],
    host: [Rules.requiredInput],
    port: [Rules.requiredInput],
    recipient: [Rules.requiredInput],
};
interface Config {
    status: string;
    displayName: string;
    sender: string;
    password: string;
    host: string;
    port: number;
    encryption: string;
    recipient: string;
}
interface DialogProps {
    id: number;
    config: Config;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    id: undefined,
    config: {
        displayName: '',
        sender: '',
        password: '',
        host: '',
        port: 465,
        encryption: 'NONE',
        status: 'Enable',
        recipient: '',
    },
});
const isOK = ref(false);
const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.id = params.id;
    form.config = params.config;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            form.config.status = 'Enable';
            const configInfo = form.config;
            await UpdateAlertConfig({
                id: form.id,
                type: 'email',
                title: 'xpack.alert.emailConfig',
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

const onTest = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            await TestAlertConfig(form.config)
                .then((res) => {
                    loading.value = false;
                    if (res.data) {
                        isOK.value = true;
                        MsgSuccess(i18n.global.t('xpack.alert.alertTestOk'));
                    } else {
                        MsgError(i18n.global.t('xpack.alert.alertTestFailed'));
                    }
                })
                .catch(() => {
                    loading.value = false;
                    MsgError(i18n.global.t('xpack.alert.alertTestFailed'));
                });
        } finally {
            loading.value = false;
        }
    });
};

const handleClose = () => {
    isOK.value = false;
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
