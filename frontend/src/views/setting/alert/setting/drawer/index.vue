<template>
    <DrawerPro v-model="drawerVisible" :header="drawerHeader" @close="handleClose" size="736">
        <el-form
            ref="formRef"
            :rules="currentRules"
            label-position="top"
            :model="form"
            @submit.prevent
            v-loading="loading"
        >
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item v-if="!isEdit" :label="$t('commons.table.type')" prop="type">
                        <el-select v-model="form.type" class="w-full" @change="onTypeChange">
                            <el-option
                                v-for="item in typeOptions"
                                :key="item.value"
                                :label="item.label"
                                :value="item.value"
                            />
                        </el-select>
                    </el-form-item>

                    <template v-if="form.type === 'email'">
                        <el-form-item :label="$t('xpack.alert.displayName')" prop="config.displayName">
                            <el-input v-model="form.config.displayName" />
                            <span class="input-help">{{ $t('xpack.alert.displayNameHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('xpack.alert.sender')" prop="config.sender">
                            <el-input v-model.trim="form.config.sender" />
                            <span class="input-help">{{ $t('xpack.alert.senderHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.username')" prop="config.userName">
                            <el-input v-model.trim="form.config.userName" />
                            <span class="input-help">{{ $t('xpack.alert.userNameHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('xpack.alert.password')" prop="config.password">
                            <el-input v-model.trim="form.config.password" type="password" show-password />
                            <span class="input-help">{{ $t('xpack.alert.passwordHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('xpack.alert.host')" prop="config.host">
                            <el-input v-model.trim="form.config.host" placeholder="smtp.qq.com" />
                            <span class="input-help">{{ $t('xpack.alert.hostHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('xpack.alert.port')" prop="config.port">
                            <el-input v-model.number="form.config.port" :min="1" :max="65535" />
                        </el-form-item>
                        <el-form-item :label="$t('xpack.alert.encryption')" prop="config.encryption">
                            <div class="flex items-center gap-2">
                                <span class="el-form-item__label">SSL</span>
                                <el-switch
                                    v-permission
                                    v-model="form.config.encryption"
                                    :active-value="'SSL'"
                                    :inactive-value="form.config.encryption === 'SSL' ? 'NONE' : form.config.encryption"
                                />
                            </div>
                            <span class="input-help">{{ $t('xpack.alert.sslHelper') }}</span>
                            <div class="flex items-center gap-2">
                                <span class="el-form-item__label">TLS</span>
                                <el-switch
                                    v-permission
                                    v-model="form.config.encryption"
                                    :active-value="'TLS'"
                                    :inactive-value="form.config.encryption === 'TLS' ? 'NONE' : form.config.encryption"
                                />
                            </div>
                            <span class="input-help">{{ $t('xpack.alert.tlsHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('xpack.alert.recipient')" prop="recipient">
                            <el-input
                                v-model.trim="form.recipient"
                                :placeholder="$t('xpack.alert.recipientPlaceholder')"
                            />
                        </el-form-item>
                    </template>

                    <template v-else-if="form.type === 'sms'">
                        <el-form-item :label="$t('xpack.alert.displayName')" prop="smsDisplayName">
                            <el-input v-model.trim="form.smsDisplayName" />
                            <span class="input-help">{{ $t('xpack.alert.displayNameHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('xpack.alert.phone')" prop="smsPhone">
                            <el-input clearable v-model.trim="form.smsPhone" />
                            <span class="input-help">{{ $t('xpack.alert.phoneHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('xpack.alert.dailyAlertNum')" prop="smsDailyAlertNum">
                            <el-input clearable v-model.number="form.smsDailyAlertNum" min="20" max="100" />
                            <span class="input-help">{{ $t('xpack.alert.dailyAlertNumHelper') }}</span>
                        </el-form-item>
                    </template>

                    <template v-else>
                        <el-form-item :label="$t('xpack.alert.webhookName')" prop="webhookName">
                            <el-input v-model="form.webhookName" />
                        </el-form-item>
                        <el-form-item :label="$t('xpack.alert.webhookUrl')" prop="webhookUrl">
                            <el-input v-model.trim="form.webhookUrl" :rows="2" type="password" show-password />
                        </el-form-item>
                    </template>

                    <el-form-item v-if="isEdit && isEE" :label="$t('commons.table.updater')">
                        <el-input :model-value="form.updateUser || '-'" readonly />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <div v-if="form.type === 'email'" class="flex items-center justify-between">
                <el-button v-permission @click="onTest(formRef)" plain type="primary">
                    {{ $t('xpack.alert.test') }}
                </el-button>
                <div>
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button
                        v-permission
                        :disabled="loading || (form.type === 'email' && !isOK)"
                        type="primary"
                        @click="onSave(formRef)"
                    >
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </div>
            </div>
            <div v-else class="flex justify-end gap-2">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-permission :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </div>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { ListAlertConfigs, TestAlertConfig, UpdateAlertConfig } from '@/api/modules/alert';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { Alert } from '@/api/interface/alert';

const emit = defineEmits<{ (e: 'search'): void }>();

const { isProductPro, isIntl, isEE } = useGlobalStore();

const emailRules = {
    'config.displayName': [Rules.requiredInput, { validator: checkDisplayNameDuplicate, trigger: 'blur' }],
    'config.sender': [Rules.requiredInput],
    'config.host': [Rules.requiredInput],
    'config.port': [Rules.requiredInput],
    recipient: [Rules.requiredInput],
};

const smsRules = {
    smsDisplayName: [Rules.requiredInput, { validator: checkSmsDisplayNameDuplicate, trigger: 'blur' }],
    smsPhone: [Rules.requiredInput, Rules.phone, { validator: checkPhoneDuplicate, trigger: 'blur' }],
    smsDailyAlertNum: [Rules.integerNumber, checkNumberRange(20, 100)],
};

const webhookRules = {
    webhookName: [Rules.requiredInput, { validator: checkDisplayNameDuplicate, trigger: 'blur' }],
    webhookUrl: [Rules.requiredInput],
};

const currentRules = computed(() => {
    if (form.type === 'email') return emailRules;
    if (form.type === 'sms') return smsRules;
    return webhookRules;
});

const typeOptions = computed(() => {
    const options: { value: string; label: string }[] = [{ value: 'email', label: i18n.global.t('xpack.alert.mail') }];
    if (isProductPro.value && !isIntl.value) {
        options.push({ value: 'weCom', label: i18n.global.t('xpack.alert.weCom') });
        options.push({ value: 'dingTalk', label: i18n.global.t('xpack.alert.dingTalk') });
        options.push({ value: 'feiShu', label: i18n.global.t('xpack.alert.feiShu') });
    }
    options.push({ value: 'bark', label: i18n.global.t('xpack.alert.bark') });
    if (isProductPro.value && !isEE.value && !isIntl.value) {
        options.push({ value: 'sms', label: i18n.global.t('xpack.alert.sms') });
    }
    return options;
});

const defaultEmailForm = {
    displayName: '',
    sender: '',
    userName: '',
    password: '',
    host: '',
    port: 465,
    encryption: 'NONE',
    status: 'Enable',
    recipient: '',
};

const drawerVisible = ref(false);
const loading = ref(false);
const isEdit = ref(false);
const isOK = ref(false);
const formRef = ref<FormInstance>();
const alertConfigs = ref<Alert.AlertConfigInfo[]>([]);

const loadAlertConfigs = async () => {
    loading.value = true;
    try {
        const res = await ListAlertConfigs();
        alertConfigs.value = res.data?.filter((item: Alert.AlertConfigInfo) => item.type !== 'common') || [];
    } catch {
        alertConfigs.value = [];
    } finally {
        loading.value = false;
    }
};

const form = reactive({
    id: undefined as number | undefined,
    type: 'email',
    title: '',
    status: 'Enable',
    updateUser: '',
    config: { ...defaultEmailForm } as Record<string, any>,
    recipient: '',
    webhookName: '',
    webhookUrl: '',
    smsDisplayName: '',
    smsPhone: '',
    smsDailyAlertNum: 50,
});

const drawerHeader = computed(() => {
    if (isEdit.value) {
        return i18n.global.t('xpack.alert.' + form.type);
    }
    return i18n.global.t('commons.button.create');
});

const normalizeDisplayName = (value?: string) => value?.trim() || '';

function checkDisplayNameDuplicate(_rule: unknown, value: string, callback: (error?: Error) => void) {
    const currentValue = normalizeDisplayName(value);
    if (!currentValue) {
        callback();
        return;
    }

    const duplicated = alertConfigs.value.some((item) => {
        if (item.type !== form.type) {
            return false;
        }
        if (form.id && item.id === form.id) {
            return false;
        }
        try {
            const config = JSON.parse(item.config || '{}') as { displayName?: string };
            return normalizeDisplayName(config.displayName) === currentValue;
        } catch {
            return false;
        }
    });

    if (duplicated) {
        callback(new Error(i18n.global.t('commons.rule.duplicate')));
        return;
    }

    callback();
}

function checkSmsDisplayNameDuplicate(_rule: unknown, value: string, callback: (error?: Error) => void) {
    const currentValue = normalizeDisplayName(value);
    if (!currentValue) {
        callback();
        return;
    }

    const duplicated = alertConfigs.value.some((item) => {
        if (item.type !== 'sms') {
            return false;
        }
        if (form.id && item.id === form.id) {
            return false;
        }
        try {
            const config = JSON.parse(item.config || '{}') as { displayName?: string };
            return normalizeDisplayName(config.displayName) === currentValue;
        } catch {
            return false;
        }
    });

    if (duplicated) {
        callback(new Error(i18n.global.t('commons.rule.duplicate')));
        return;
    }

    callback();
}

function checkPhoneDuplicate(_rule: unknown, value: string, callback: (error?: Error) => void) {
    const currentValue = normalizeDisplayName(value);
    if (!currentValue) {
        callback();
        return;
    }

    const duplicated = alertConfigs.value.some((item) => {
        if (item.type !== 'sms') {
            return false;
        }
        if (form.id && item.id === form.id) {
            return false;
        }
        try {
            const config = JSON.parse(item.config || '{}') as { phone?: string };
            return normalizeDisplayName(config.phone) === currentValue;
        } catch {
            return false;
        }
    });

    if (duplicated) {
        callback(new Error(i18n.global.t('commons.rule.duplicate')));
        return;
    }

    callback();
}

const titleMap: Record<string, string> = {
    email: 'xpack.alert.emailConfig',
    weCom: 'xpack.alert.weCom',
    dingTalk: 'xpack.alert.dingTalk',
    feiShu: 'xpack.alert.feiShu',
    bark: 'xpack.alert.bark',
    sms: 'xpack.alert.smsConfig',
};

interface DrawerProps {
    id?: number;
    type?: string;
    config?: Record<string, any>;
    status?: string;
    updateUser?: string;
}

const acceptParams = (params: DrawerProps): void => {
    if (params.id && params.id > 0) {
        isEdit.value = true;
        form.id = params.id;
        form.type = params.type || 'email';
        form.title = titleMap[form.type] || '';
        form.status = params.status || 'Enable';
        form.updateUser = params.updateUser || '';

        if (form.type === 'email') {
            form.config = { ...defaultEmailForm, ...(params.config || {}) };
            form.recipient = params.config?.recipient || '';
        } else if (form.type === 'sms') {
            form.smsDisplayName = params.config?.displayName || '';
            form.smsPhone = params.config?.phone || '';
            form.smsDailyAlertNum = params.config?.alertDailyNum || 50;
        } else {
            form.webhookName = params.config?.displayName || '';
            form.webhookUrl = params.config?.url || '';
        }
    } else {
        isEdit.value = false;
        form.id = undefined;
        form.type = params.type || 'email';
        form.title = titleMap[form.type] || '';
        form.status = 'Enable';
        form.updateUser = '';
        form.config = { ...defaultEmailForm };
        form.recipient = '';
        form.smsDisplayName = '';
        form.webhookName = '';
        form.webhookUrl = '';
        form.smsPhone = '';
        form.smsDailyAlertNum = 50;
    }

    isOK.value = false;
    drawerVisible.value = true;
    void loadAlertConfigs();
};

const onTypeChange = (type: string) => {
    form.title = titleMap[type] || '';
    if (type === 'email') {
        form.config = { ...defaultEmailForm };
        form.recipient = '';
    } else if (type === 'sms') {
        form.smsDisplayName = '';
        form.smsPhone = '';
        form.smsDailyAlertNum = 50;
    } else {
        form.config = {};
        form.webhookName = '';
        form.webhookUrl = '';
    }
    isOK.value = false;
    formRef.value?.clearValidate();
};

const buildSavePayload = () => {
    if (form.type === 'email') {
        const configInfo = {
            ...form.config,
            recipient: form.recipient,
        };
        return {
            id: form.id,
            type: 'email',
            title: titleMap['email'],
            status: form.status,
            config: JSON.stringify(configInfo),
        };
    }
    if (form.type === 'sms') {
        const configInfo = {
            displayName: form.smsDisplayName,
            phone: form.smsPhone,
            alertDailyNum: form.smsDailyAlertNum,
        };
        return {
            id: form.id,
            type: 'sms',
            title: titleMap['sms'],
            status: form.status,
            config: JSON.stringify(configInfo),
        };
    }
    const configInfo = {
        displayName: form.webhookName,
        url: form.webhookUrl,
    };
    return {
        id: form.id,
        type: form.type,
        title: form.title,
        status: form.status,
        config: JSON.stringify(configInfo),
    };
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            await UpdateAlertConfig(buildSavePayload());
            loading.value = false;
            void loadAlertConfigs();
            handleClose();
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        } catch (error) {
            loading.value = false;
        }
    });
};

const onTest = async (formEl: FormInstance | undefined) => {
    if (!formEl || form.type !== 'email') return;
    await formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            const testConfig = {
                ...form.config,
                recipient: form.recipient,
            } as Alert.AlertConfigTest;
            const res = await TestAlertConfig(testConfig);
            loading.value = false;
            if (res.data) {
                isOK.value = true;
                MsgSuccess(i18n.global.t('xpack.alert.alertTestOk'));
            } else {
                MsgError(i18n.global.t('xpack.alert.alertTestFailed'));
            }
        } catch {
            loading.value = false;
            MsgError(i18n.global.t('xpack.alert.alertTestFailed'));
        }
    });
};

watch(
    () => form.config,
    () => {
        if (form.type === 'email') {
            isOK.value = false;
        }
    },
    { deep: true },
);

watch(
    () => form.recipient,
    () => {
        if (form.type === 'email') {
            isOK.value = false;
        }
    },
);

watch(
    () => [form.smsDisplayName, form.smsPhone, form.smsDailyAlertNum],
    () => {
        if (form.type === 'sms') {
            formRef.value?.clearValidate(['smsDisplayName', 'smsPhone']);
        }
    },
);

const handleClose = () => {
    isOK.value = false;
    drawerVisible.value = false;
};

onMounted(() => {
    void loadAlertConfigs();
});

defineExpose({
    acceptParams,
});
</script>
