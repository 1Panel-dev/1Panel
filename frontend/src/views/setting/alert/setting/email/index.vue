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
                    <el-form-item :label="$t('commons.login.username')" prop="userName">
                        <el-input v-model.trim="form.config.userName" />
                        <span class="input-help">
                            {{ $t('xpack.alert.userNameHelper') }}
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
                                v-permission
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
                                v-permission
                                v-model="form.config.encryption"
                                :active-value="'TLS'"
                                :inactive-value="form.config.encryption === 'TLS' ? 'NONE' : form.config.encryption"
                            />
                        </div>
                        <span class="input-help">
                            {{ $t('xpack.alert.tlsHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('xpack.alert.recipient')" prop="recipients">
                        <div class="w-full">
                            <div
                                v-for="(item, index) in form.config.recipients"
                                :key="index"
                                class="flex items-center mb-2 gap-2"
                            >
                                <el-input
                                    v-model="form.config.recipients[index]"
                                    :placeholder="$t('xpack.alert.recipientPlaceholder')"
                                />
                                <el-button
                                    v-permission
                                    type="danger"
                                    plain
                                    circle
                                    @click="removeRecipient(index)"
                                    :disabled="form.config.recipients.length <= 1"
                                >
                                    <el-icon><Minus /></el-icon>
                                </el-button>
                            </div>
                            <el-button v-permission type="primary" plain @click="addRecipient">
                                <el-icon><Plus /></el-icon>
                                {{ $t('xpack.alert.addRecipient') }}
                            </el-button>
                        </div>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <div class="flex items-center justify-between">
                <el-button v-permission @click="onTest(formRef)" plain type="primary">
                    {{ $t('xpack.alert.test') }}
                </el-button>
                <div>
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button v-permission :disabled="loading || !isOK" type="primary" @click="onSave(formRef)">
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
import { Plus, Minus } from '@element-plus/icons-vue';

const emit = defineEmits<{ (e: 'search'): void }>();

const rules = {
    displayName: [Rules.requiredInput],
    sender: [Rules.requiredInput],
    host: [Rules.requiredInput],
    port: [Rules.requiredInput],
    recipients: [
        {
            validator: (rule: any, value: string[], callback: any) => {
                if (!value || value.length === 0) {
                    callback(new Error(i18n.global.t('commons.rule.requiredInput')));
                } else {
                    const hasEmpty = value.some((item) => !item || !item.trim());
                    if (hasEmpty) {
                        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
                    } else {
                        callback();
                    }
                }
            },
            trigger: 'blur',
        },
    ],
};
interface Config {
    status: string;
    displayName: string;
    sender: string;
    userName: string;
    password: string;
    host: string;
    port: number;
    encryption: string;
    recipient?: string;
    recipients?: string[];
}
interface DialogProps {
    id: number;
    config: Config;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    id: undefined as number | undefined,
    config: {
        displayName: '',
        sender: '',
        password: '',
        userName: '',
        host: '',
        port: 465,
        encryption: 'NONE',
        status: 'Enable',
        recipients: [''] as string[],
    },
});
const isOK = ref(false);
const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.id = params.id;
    form.config.displayName = params.config.displayName || '';
    form.config.sender = params.config.sender || '';
    form.config.password = params.config.password || '';
    form.config.userName = params.config.userName || '';
    form.config.host = params.config.host || '';
    form.config.port = params.config.port || 465;
    form.config.encryption = params.config.encryption || 'NONE';
    form.config.status = params.config.status || 'Enable';

    if (params.config.recipients && params.config.recipients.length > 0) {
        form.config.recipients = [...params.config.recipients];
    } else if (params.config.recipient) {
        form.config.recipients = params.config.recipient
            .split(',')
            .map((r) => r.trim())
            .filter((r) => r);
    } else {
        form.config.recipients = [''];
    }

    drawerVisible.value = true;
};

const addRecipient = () => {
    form.config.recipients.push('');
};

const removeRecipient = (index: number) => {
    if (form.config.recipients.length > 1) {
        form.config.recipients.splice(index, 1);
    }
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            form.config.status = 'Enable';
            const configInfo = {
                ...form.config,
                recipient: form.config.recipients.filter((r) => r && r.trim()).join(','),
            };
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
            const testConfig = {
                ...form.config,
                recipient: form.config.recipients.filter((r) => r && r.trim()).join(','),
            };
            await TestAlertConfig(testConfig)
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
