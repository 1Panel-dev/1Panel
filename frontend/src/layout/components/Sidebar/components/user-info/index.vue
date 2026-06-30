<template>
    <DrawerPro v-model="open" :header="$t('xpack.user.userInfo')">
        <div v-loading="loading">
            <el-form ref="userRef" label-position="top" :model="form" :rules="userRules">
                <input
                    class="hidden-autofill-field"
                    type="text"
                    name="username"
                    autocomplete="username"
                    tabindex="-1"
                />
                <input
                    class="hidden-autofill-field"
                    type="password"
                    name="password"
                    autocomplete="current-password"
                    tabindex="-1"
                />
                <el-form-item :label="$t('commons.login.username')" prop="name">
                    <el-input type="primary" v-model="form.name" name="profile-name" autocomplete="off" />
                </el-form-item>
                <el-form-item :label="$t('commons.login.password')" prop="password">
                    <el-input
                        type="password"
                        show-password
                        clearable
                        v-model.trim="form.password"
                        name="profile-new-password"
                        autocomplete="new-password"
                    />
                    <span class="input-help">{{ $t('setting.passwordEmptyTip') }}</span>
                </el-form-item>
                <el-form-item v-if="form.password" :label="$t('setting.oldPassword')" prop="oldPassword">
                    <el-input
                        type="password"
                        show-password
                        clearable
                        v-model.trim="form.oldPassword"
                        name="profile-old-password"
                        autocomplete="current-password"
                    />
                </el-form-item>
            </el-form>

            <el-form label-position="top" :model="form" class="setting-section">
                <el-form-item>
                    <template #label>
                        <span class="label-with-help">
                            {{ $t('setting.mfa') }}
                            <el-tooltip placement="top">
                                <template #content>
                                    <div class="tooltip-help-block">{{ $t('setting.mfaAlert') }}</div>
                                    <div>{{ $t('setting.mfaHelper1') }}</div>
                                    <ul class="tooltip-help-list">
                                        <li>Google Authenticator</li>
                                        <li>Microsoft Authenticator</li>
                                        <li>1Password</li>
                                        <li>LastPass</li>
                                        <li>Authenticator</li>
                                    </ul>
                                </template>
                                <el-icon class="help-icon"><QuestionFilled /></el-icon>
                            </el-tooltip>
                        </span>
                    </template>
                    <el-switch
                        @change="handleMFA"
                        v-model="form.mfaStatus"
                        active-value="Enable"
                        inactive-value="Disable"
                    />
                    <span class="input-help">{{ $t('setting.mfaHelper') }}</span>
                </el-form-item>
            </el-form>

            <el-form label-position="top" :model="form" class="setting-section">
                <el-form-item>
                    <template #label>
                        <span class="label-with-help">
                            {{ $t('setting.passkey') }}
                            <el-tooltip placement="top">
                                <template #content>
                                    <div>{{ $t('setting.passkeyRequireSSL') }}</div>
                                </template>
                                <el-icon class="help-icon"><QuestionFilled /></el-icon>
                            </el-tooltip>
                        </span>
                    </template>
                    <el-button @click="openPasskeyDrawer">
                        {{ $t('setting.passkeyManage') }}
                    </el-button>
                    <span class="input-help">{{ $t('setting.passkeyHelper') }}</span>
                </el-form-item>
            </el-form>

            <el-form label-position="top" :model="form" class="setting-section">
                <el-form-item>
                    <template #label>
                        <span class="label-with-help">
                            {{ $t('setting.apiInterface') }}
                            <el-tooltip placement="top">
                                <template #content>
                                    <ul class="tooltip-help-list">
                                        <li>
                                            {{ $t('setting.apiInterfaceAlert1') }}
                                        </li>
                                        <li>
                                            {{ $t('setting.apiInterfaceAlert2') }}
                                        </li>
                                        <li>
                                            <el-link :href="apiURL" target="_blank" class="tooltip-help-link">
                                                {{ $t('setting.apiInterfaceAlert3') }}
                                            </el-link>
                                        </li>
                                        <li v-if="!isFxplay">
                                            <el-link :href="panelURL" target="_blank" class="tooltip-help-link">
                                                {{ $t('setting.apiInterfaceAlert4') }}
                                            </el-link>
                                        </li>
                                    </ul>
                                </template>
                                <el-icon class="help-icon"><QuestionFilled /></el-icon>
                            </el-tooltip>
                        </span>
                    </template>
                    <div class="setting-action-row">
                        <el-switch
                            @change="handleApi"
                            v-model="form.apiInterfaceStatus"
                            active-value="Enable"
                            inactive-value="Disable"
                        />
                        <el-button
                            v-if="form.apiInterfaceStatus === 'Enable'"
                            link
                            type="primary"
                            @click="openApiDetail"
                        >
                            {{ $t('commons.button.view') }}
                        </el-button>
                    </div>
                    <span class="input-help">{{ $t('setting.apiInterfaceHelper') }}</span>
                </el-form-item>
            </el-form>
        </div>
        <template #footer>
            <el-button :disabled="loading" @click="open = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button :disabled="loading" type="primary" @click="onSubmit(userRef)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>

    <DialogPro v-model="mfaDialogOpen" :title="$t('setting.mfa')" size="large" @close="handleMfaDialogClose">
        <el-form
            ref="mfaFormRef"
            :model="mfaForm"
            @submit.prevent
            v-loading="loading"
            label-position="top"
            :rules="mfaRules"
        >
            <el-form-item :label="$t('setting.mfaHelper2')">
                <el-image class="w-32 h-32" :src="qrImage" />
                <span class="input-help flex items-center">
                    <span>{{ $t('setting.secret') }}: {{ mfaForm.secret }}</span>
                    <CopyButton :content="mfaForm.secret" />
                </span>
            </el-form-item>
            <el-form-item :label="$t('commons.table.title')" prop="title">
                <el-input v-model="mfaForm.title">
                    <template #append>
                        <el-button @click="loadMfaCodeBefore(mfaFormRef)">
                            {{ $t('commons.button.save') }}
                        </el-button>
                    </template>
                </el-input>
                <span class="input-help">{{ $t('setting.mfaTitleHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('setting.mfaInterval')" prop="interval">
                <el-input v-model.number="mfaForm.interval">
                    <template #append>
                        <el-button @click="loadMfaCodeBefore(mfaFormRef)">
                            {{ $t('commons.button.save') }}
                        </el-button>
                    </template>
                </el-input>
                <span class="input-help">{{ $t('setting.mfaIntervalHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('setting.mfaCode')" prop="code">
                <el-input v-model="mfaForm.code"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="handleMfaDialogClose">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button :disabled="loading" type="primary" @click="onBindMFA(mfaFormRef)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>

    <DialogPro
        v-model="passkeyPrereqDialogOpen"
        :title="$t('setting.passkey')"
        size="small"
        @close="handlePasskeyPrereqDialogClose"
    >
        <div class="passkey-prereq-dialog">
            <div class="passkey-prereq-title">{{ $t('setting.passkeyPrereqTitle') }}</div>
            <div class="passkey-prereq-list">
                <div class="passkey-prereq-item">
                    <el-icon class="passkey-prereq-status" :class="passkeyPrereqBindDomain ? 'is-ok' : 'is-missing'">
                        <Check v-if="passkeyPrereqBindDomain" />
                        <Close v-else />
                    </el-icon>
                    <span>{{ $t('setting.passkeyPrereqBindDomain') }}</span>
                    <el-button
                        v-if="!passkeyPrereqBindDomain"
                        link
                        type="primary"
                        @click="handlePasskeyConfigureDomain"
                    >
                        {{ $t('setting.passkeyPrereqGoSetup') }}
                    </el-button>
                </div>
                <div class="passkey-prereq-item">
                    <el-icon class="passkey-prereq-status" :class="passkeyPrereqHttps ? 'is-ok' : 'is-missing'">
                        <Check v-if="passkeyPrereqHttps" />
                        <Close v-else />
                    </el-icon>
                    <span>{{ $t('setting.passkeyPrereqHttps') }}</span>
                </div>
                <div class="passkey-prereq-item is-browser">
                    <el-icon class="passkey-prereq-status" :class="passkeyPrereqBrowser ? 'is-ok' : 'is-missing'">
                        <Check v-if="passkeyPrereqBrowser" />
                        <Close v-else />
                    </el-icon>
                    <div class="passkey-prereq-browser-copy">
                        <div>{{ $t('setting.passkeyPrereqBrowser') }}</div>
                        <div v-if="!passkeyPrereqBrowser && passkeyPrereqBrowserDetail" class="passkey-prereq-detail">
                            {{ passkeyPrereqBrowserDetail }}
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <template #footer>
            <el-button @click="handlePasskeyPrereqDialogClose">
                {{ $t('commons.button.cancel') }}
            </el-button>
        </template>
    </DialogPro>

    <DialogPro
        v-model="passkeyDialogOpen"
        :title="$t('setting.passkey')"
        size="large"
        @close="handlePasskeyDialogClose"
    >
        <el-tabs v-model="passkeyActiveTab">
            <el-tab-pane :label="$t('setting.passkeyKeyManagement')" name="keys">
                <el-form label-position="top">
                    <el-form-item :label="$t('setting.passkeyName')">
                        <el-input
                            v-model.trim="passkeyForm.name"
                            :placeholder="$t('setting.passkeyNameHelper')"
                            :disabled="!allPasskeyPrerequisitesMet"
                        />
                    </el-form-item>
                    <el-button type="primary" @click="registerPasskey" :disabled="!canRegisterPasskey">
                        {{ $t('setting.passkeyAdd') }}
                    </el-button>
                    <span class="text-xs text-gray-500 ml-3">{{ passkeyCountText }}</span>
                </el-form>
                <el-table class="mt-4" :data="passkeyList" v-loading="passkeyLoading">
                    <el-table-column prop="name" :label="$t('setting.passkeyName')" min-width="120" />
                    <el-table-column prop="createdAt" :label="$t('setting.passkeyCreatedAt')" min-width="160" />
                    <el-table-column :label="$t('setting.passkeyLastUsedAt')" min-width="160">
                        <template #default="scope">
                            <span>{{ scope.row.lastUsedAt || '-' }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.operate')" width="120">
                        <template #default="scope">
                            <el-button
                                link
                                type="danger"
                                :disabled="!allPasskeyPrerequisitesMet"
                                @click="removePasskey(scope.row.id)"
                            >
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </el-tab-pane>
            <el-tab-pane :label="$t('app.advanced')" name="advanced">
                <el-form label-position="top">
                    <el-form-item :label="$t('setting.passkeyTrustedProxies')">
                        <div class="w-full flex items-start gap-2">
                            <el-input v-model="passkeyTrustedProxies" type="textarea" :rows="3" />
                            <el-button :loading="savePasskeyProxyLoading" @click="onSavePasskeyTrustedProxies">
                                {{ $t('commons.button.save') }}
                            </el-button>
                        </div>
                        <span class="input-help">{{ $t('setting.passkeyTrustedProxiesHelper') }}</span>
                        <span class="input-help">{{ $t('setting.allowIPEgs') }}</span>
                    </el-form-item>
                </el-form>
            </el-tab-pane>
        </el-tabs>
        <template #footer>
            <el-button @click="handlePasskeyDialogClose">
                {{ $t('commons.button.cancel') }}
            </el-button>
        </template>
    </DialogPro>

    <DialogPro v-model="apiDialogOpen" :title="$t('setting.apiInterface')" size="large" @close="handleApiDialogClose">
        <el-form ref="apiRef" :model="form" @submit.prevent v-loading="loading" label-position="top" :rules="apiRules">
            <el-form-item :label="$t('setting.apiKey')" prop="apiKey">
                <el-input v-model="form.apiKey" readonly class="api-key-input" />
                <el-button-group>
                    <CopyButton class="copy_button" :isIcon="false" :content="form.apiKey" />
                    <el-button @click="resetApiKey()">
                        {{ $t('commons.button.reset') }}
                    </el-button>
                </el-button-group>
                <span class="input-help">{{ $t('setting.apiKeyHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('setting.ipWhiteList')" prop="ipWhiteList">
                <el-input
                    type="textarea"
                    :placeholder="$t('setting.ipWhiteListEgs')"
                    :rows="4"
                    v-model="form.ipWhiteList"
                />
                <span class="input-help">{{ $t('setting.ipWhiteListHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('setting.apiKeyValidityTime')" prop="apiKeyValidityTime">
                <el-input :placeholder="$t('setting.apiKeyValidityTimeEgs')" v-model.number="form.apiKeyValidityTime">
                    <template #append>{{ $t('commons.units.minute') }}</template>
                </el-input>
                <span class="input-help">
                    {{ $t('setting.apiKeyValidityTimeHelper') }}
                </span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="handleApiDialogClose">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button :disabled="loading" type="primary" @click="onSaveApi(apiRef)">
                {{ $t('commons.button.save') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessageBox, FormInstance } from 'element-plus';
import { Check, Close, QuestionFilled } from '@element-plus/icons-vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import DialogPro from '@/components/dialog-pro/index.vue';
import { Login } from '@/api/interface/auth';
import {
    bindMFA,
    closeMFA,
    generateApiKey,
    loadMFA,
    passkeyDelete,
    passkeyList as fetchPasskeyList,
    passkeyRegisterBegin,
    passkeyRegisterFinish,
    updateApiConfig,
    updateUserInfo,
} from '@/api/modules/auth';
import { Setting } from '@/api/interface/setting';
import { getSettingBaseInfo, getSettingInfo, updateSetting } from '@/api/modules/setting';
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { base64UrlToBuffer, bufferToBase64Url } from '@/utils/auth';
import { MsgError, MsgSuccess } from '@/utils/message';
import { checkNumberRange, Rules } from '@/global/form-rules';
import { checkCidr, checkCidrV6, checkIpV4V6 } from '@/utils/validate';

const props = defineProps<{ currentUser?: Login.AuthInfo }>();
const emit = defineEmits<{ (e: 'search'): void }>();

const complexityVerification = ref(false);
const { globalStore, docsUrl, entrance, isFxplay } = useGlobalStore();
const router = useRouter();
const open = ref(false);
const loading = ref(false);
const userRef = ref<FormInstance>();
const mfaFormRef = ref<FormInstance>();
const apiRef = ref<FormInstance>();
const mfaDialogOpen = ref(false);
const passkeyPrereqDialogOpen = ref(false);
const passkeyDialogOpen = ref(false);
const apiDialogOpen = ref(false);
const savedMfaStatus = ref('');
const savedApiStatus = ref('');
const qrImage = ref();
const passkeyActiveTab = ref('keys');
const passkeyLoading = ref(false);
const savePasskeyProxyLoading = ref(false);
const passkeyList = ref<Setting.PasskeyInfo[]>([]);
const passkeyTrustedProxies = ref('');
const hasBindDomain = ref(false);
const passkeyPrereqBrowser = ref(false);
const passkeyPrereqBrowserDetailKey = ref('');
const passkeyMaxCount = 5;
const apiURL = `${window.location.protocol}//${window.location.hostname}${
    window.location.port ? `:${window.location.port}` : ''
}/1panel/swagger/index.html`;
const panelURL = `${docsUrl.value}/dev_manual/api_manual/`;
const form = reactive({
    id: 0,
    name: '',
    password: '',
    oldPassword: '',
    mfaStatus: 'Disable',
    mfaInterval: 30,
    apiInterfaceStatus: 'Disable',
    apiKey: '',
    ipWhiteList: '',
    apiKeyValidityTime: 120,
});
const mfaForm = reactive({
    title: '1Panel',
    code: '',
    secret: '',
    interval: 30,
});
const passkeyForm = reactive({
    name: '',
});

const passkeyPrereqBindDomain = computed(() => hasBindDomain.value);
const passkeyPrereqHttps = computed(() => window.isSecureContext);
const passkeyPrereqBrowserDetail = computed(() => {
    if (!passkeyPrereqBrowserDetailKey.value) {
        return '';
    }
    return i18n.global.t(passkeyPrereqBrowserDetailKey.value);
});
const allPasskeyPrerequisitesMet = computed(() => {
    return passkeyPrereqBindDomain.value && passkeyPrereqHttps.value && passkeyPrereqBrowser.value;
});
const passkeyCountText = computed(() => {
    return i18n.global.t('setting.passkeyCount', [passkeyList.value.length, passkeyMaxCount]);
});
const canRegisterPasskey = computed(() => {
    return (
        allPasskeyPrerequisitesMet.value &&
        passkeyList.value.length < passkeyMaxCount &&
        passkeyForm.name.trim().length > 0
    );
});

const validatePassword = (_rule: any, value: string, callback: any) => {
    if (!value) {
        callback();
        return;
    }
    if (value.indexOf(' ') !== -1) {
        callback(new Error(i18n.global.t('setting.noSpace')));
        return;
    }
    if (complexityVerification.value) {
        const reg = /^(?![\d]+$)(?![a-zA-Z]+$)(?![^\da-zA-Z]+$).{8,30}$/;
        if (!reg.test(value)) {
            callback(new Error(i18n.global.t('commons.rule.complexityPassword')));
            return;
        }
    }
    callback();
};

const userRules = reactive({
    name: [Rules.requiredInput, Rules.noSpace],
    oldPassword: [Rules.requiredInput],
    password: [{ validator: validatePassword, trigger: 'blur' }],
});
const mfaRules = reactive({
    code: [Rules.requiredInput],
    title: [Rules.requiredInput],
    interval: [Rules.number, checkNumberRange(15, 60)],
});
const apiRules = reactive({
    ipWhiteList: [Rules.requiredInput, { validator: checkIPs, trigger: 'blur' }],
    apiKey: [Rules.requiredInput],
    apiKeyValidityTime: [Rules.requiredInput, Rules.integerNumberWith0],
});

const getUserFormFields = () => {
    const fields = ['name', 'password'];
    if (form.password) {
        fields.push('oldPassword');
    }
    return fields;
};
const apiFormFields = ['apiKey', 'ipWhiteList', 'apiKeyValidityTime'];

const openDrawer = async () => {
    if (!props.currentUser) {
        return;
    }
    loadComplexitySetting();
    syncCurrentUser(props.currentUser);
    open.value = true;
};

const syncApiConfig = (currentUser: Login.AuthInfo) => {
    form.apiInterfaceStatus = currentUser.apiInterfaceStatus || 'Disable';
    form.apiKey = currentUser.apiKey;
    form.ipWhiteList = currentUser.ipWhiteList;
    form.apiKeyValidityTime = currentUser.apiKeyValidityTime;
    savedApiStatus.value = form.apiInterfaceStatus;
};

const syncCurrentUser = (currentUser: Login.AuthInfo) => {
    form.id = currentUser.id;
    form.name = currentUser.name;
    form.password = '';
    form.oldPassword = '';
    form.mfaStatus = currentUser.mfaStatus || 'Disable';
    form.mfaInterval = currentUser.mfaInterval;
    savedMfaStatus.value = form.mfaStatus;
    mfaDialogOpen.value = false;
    apiDialogOpen.value = false;
    syncApiConfig(currentUser);
};

const ensureApiKey = async () => {
    if (form.apiInterfaceStatus === 'Enable' && !form.apiKey) {
        await generateApiKey().then((res) => {
            form.apiKey = res.data;
        });
    }
};

const resetApiKey = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.apiKeyResetHelper'), i18n.global.t('setting.apiKeyReset'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            loading.value = true;
            await generateApiKey()
                .then((res) => {
                    loading.value = false;
                    form.apiKey = res.data;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            loading.value = false;
        });
};

function checkIPs(rule: any, value: any, callback: any) {
    if (form.ipWhiteList !== '') {
        let addr = form.ipWhiteList.split('\n');
        for (const item of addr) {
            if (item === '') {
                continue;
            }
            if (item.indexOf('/') !== -1) {
                if (item.indexOf(':') !== -1) {
                    if (checkCidrV6(item)) {
                        return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                    }
                } else if (checkCidr(item)) {
                    return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                }
            } else if (checkIpV4V6(item)) {
                return callback(new Error(i18n.global.t('firewall.addressFormatError')));
            }
        }
    }
    callback();
}

const loadComplexitySetting = async () => {
    const res = await getSettingBaseInfo();
    complexityVerification.value = res.data.complexityVerification === 'Enable';
};

const loadMfaCodeBefore = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const validInterval = await formEl.validateField('interval', validateFieldCallback);
    if (!validInterval) {
        return;
    }
    const validTitle = await formEl.validateField('title', validateFieldCallback);
    if (!validTitle) {
        return;
    }
    loadMfaCode();
};

const loadMfaCode = async () => {
    const param = {
        title: mfaForm.title,
        interval: mfaForm.interval,
    };
    const res = await loadMFA(param);
    mfaForm.secret = res.data.secret;
    qrImage.value = res.data.qrImage;
};

function validateFieldCallback(error: any) {
    if (error) {
        return error.message;
    }
    return;
}

const startMfaBinding = async () => {
    mfaForm.interval = form.mfaInterval || 30;
    mfaForm.title = '1Panel-' + form.name;
    mfaForm.code = '';
    await loadMfaCode();
    mfaDialogOpen.value = true;
};

const handleMfaDialogClose = () => {
    mfaDialogOpen.value = false;
    form.mfaStatus = savedMfaStatus.value;
};

const handleApiDialogClose = () => {
    apiDialogOpen.value = false;
    form.apiInterfaceStatus = savedApiStatus.value;
};

const openApiDetail = async () => {
    await ensureApiKey();
    apiDialogOpen.value = true;
};

const openPasskeyDrawer = async () => {
    const settingRes = await loadPasskeySettingInfo();
    hasBindDomain.value = !!settingRes?.data.bindDomain?.trim().length;
    passkeyTrustedProxies.value = settingRes?.data.passkeyTrustedProxies || '';
    passkeyForm.name = '';
    passkeyActiveTab.value = 'keys';
    await checkPasskeyBrowserSupport();

    if (!allPasskeyPrerequisitesMet.value) {
        passkeyPrereqDialogOpen.value = true;
        return;
    }

    passkeyDialogOpen.value = true;
    await loadPasskeys();
};

const handlePasskeyConfigureDomain = () => {
    passkeyPrereqDialogOpen.value = false;
    passkeyDialogOpen.value = false;
    open.value = false;
    router.push({ name: 'Safe' });
};

const handlePasskeyPrereqDialogClose = () => {
    passkeyPrereqDialogOpen.value = false;
};

const handlePasskeyDialogClose = () => {
    passkeyDialogOpen.value = false;
};

const loadPasskeySettingInfo = async () => {
    try {
        return await getSettingInfo();
    } catch (error) {
        return null;
    }
};

const checkPasskeyBrowserSupport = async () => {
    passkeyPrereqBrowser.value = false;
    passkeyPrereqBrowserDetailKey.value = '';
    const credentialApi = window.PublicKeyCredential;
    if (!credentialApi) {
        passkeyPrereqBrowserDetailKey.value = 'setting.passkeyPrereqBrowserDetailWebAuthnUnavailable';
        return;
    }
    if (typeof credentialApi.isUserVerifyingPlatformAuthenticatorAvailable !== 'function') {
        passkeyPrereqBrowserDetailKey.value = 'setting.passkeyPrereqBrowserDetailPlatformCapabilityUnavailable';
        return;
    }
    try {
        passkeyPrereqBrowser.value = await credentialApi.isUserVerifyingPlatformAuthenticatorAvailable();
        if (!passkeyPrereqBrowser.value) {
            passkeyPrereqBrowserDetailKey.value = 'setting.passkeyPrereqBrowserDetailNoPlatformAuthenticator';
        }
    } catch (error) {
        passkeyPrereqBrowserDetailKey.value = 'setting.passkeyPrereqBrowserDetailDetectFailed';
    }
};

const loadPasskeys = async () => {
    passkeyLoading.value = true;
    try {
        const res = await fetchPasskeyList();
        passkeyList.value = res.data || [];
    } catch (error) {
        passkeyList.value = [];
    } finally {
        passkeyLoading.value = false;
    }
};

const registerPasskey = async () => {
    if (!allPasskeyPrerequisitesMet.value) {
        return;
    }
    if (passkeyList.value.length >= passkeyMaxCount) {
        MsgError(i18n.global.t('setting.passkeyLimit'));
        return;
    }
    if (!passkeyForm.name.trim()) {
        MsgError(i18n.global.t('commons.rule.requiredInput'));
        return;
    }
    passkeyLoading.value = true;
    try {
        const res = await passkeyRegisterBegin({ name: passkeyForm.name.trim() });
        const publicKey = normalizePasskeyCreation(res.data.publicKey);
        const credential = (await navigator.credentials.create({ publicKey })) as PublicKeyCredential | null;
        if (!credential) {
            MsgError(i18n.global.t('setting.passkeyFailed'));
            return;
        }
        const payload = buildPasskeyAttestation(credential);
        await passkeyRegisterFinish(payload, res.data.sessionId);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        passkeyForm.name = '';
        await loadPasskeys();
    } finally {
        passkeyLoading.value = false;
    }
};

const removePasskey = async (id: string) => {
    ElMessageBox.confirm(i18n.global.t('setting.passkeyDeleteConfirm'), i18n.global.t('setting.passkey'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            passkeyLoading.value = true;
            await passkeyDelete(id);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            await loadPasskeys();
        })
        .catch(() => {})
        .finally(() => {
            passkeyLoading.value = false;
        });
};

const onSavePasskeyTrustedProxies = async () => {
    savePasskeyProxyLoading.value = true;
    await updateSetting({ key: 'PasskeyTrustedProxies', value: passkeyTrustedProxies.value })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            savePasskeyProxyLoading.value = false;
        });
};

const normalizePasskeyCreation = (publicKey: Record<string, any>): PublicKeyCredentialCreationOptions => {
    const request = { ...publicKey };
    request.challenge = base64UrlToBuffer(request.challenge);
    request.user = { ...request.user, id: base64UrlToBuffer(request.user.id) };
    if (request.excludeCredentials && Array.isArray(request.excludeCredentials)) {
        request.excludeCredentials = request.excludeCredentials.map((item) => {
            return { ...item, id: base64UrlToBuffer(item.id) };
        });
    }
    return request as PublicKeyCredentialCreationOptions;
};

const buildPasskeyAttestation = (credential: PublicKeyCredential) => {
    const response = credential.response as AuthenticatorAttestationResponse;
    return {
        id: credential.id,
        rawId: bufferToBase64Url(credential.rawId),
        type: credential.type,
        response: {
            clientDataJSON: bufferToBase64Url(response.clientDataJSON),
            attestationObject: bufferToBase64Url(response.attestationObject),
        },
        clientExtensionResults: credential.getClientExtensionResults(),
        authenticatorAttachment: credential.authenticatorAttachment,
    };
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const valid = await formEl.validateField(getUserFormFields(), () => {});
    if (!valid) return;
    const needReLogin = form.name !== props.currentUser?.name || !!form.password;
    if (needReLogin) {
        try {
            await ElMessageBox.confirm(i18n.global.t('setting.userChangeHelper'), i18n.global.t('setting.userChange'), {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            });
        } catch {
            return;
        }
    }
    const param: Login.AuthInfoUpdate = {
        id: form.id,
        name: form.name,
        password: form.password,
        oldPassword: form.oldPassword,
    };
    loading.value = true;
    await updateUserInfo(param)
        .then(async () => {
            loading.value = false;
            open.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            if (needReLogin) {
                globalStore.setLogStatus(false);
                globalStore.clearAuthInfo();
                router.push({ name: 'entrance', params: { code: entrance.value } });
                return;
            }
            emit('search');
        })
        .catch(() => {
            loading.value = false;
        });
};

const onBindMFA = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const valid = await formEl.validate(() => {});
    if (!valid) return;
    const param = {
        code: mfaForm.code,
        secret: mfaForm.secret,
        interval: mfaForm.interval,
    };
    loading.value = true;
    await bindMFA(param)
        .then(() => {
            loading.value = false;
            mfaDialogOpen.value = false;
            form.mfaStatus = 'Enable';
            savedMfaStatus.value = 'Enable';
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSaveApi = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const valid = await formEl.validateField(apiFormFields, () => {});
    if (!valid) return;
    const param = {
        apiKey: form.apiKey,
        ipWhiteList: form.ipWhiteList,
        apiInterfaceStatus: form.apiInterfaceStatus,
        apiKeyValidityTime: form.apiKeyValidityTime,
    };
    loading.value = true;
    await updateApiConfig(param)
        .then(() => {
            loading.value = false;
            apiDialogOpen.value = false;
            savedApiStatus.value = form.apiInterfaceStatus;
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleMFA = async () => {
    if (!form.mfaStatus) {
        return;
    }
    if (form.mfaStatus === 'Enable') {
        if (savedMfaStatus.value === 'Enable') {
            return;
        }
        await startMfaBinding();
        return;
    }
    if (savedMfaStatus.value !== 'Enable') {
        mfaDialogOpen.value = false;
        return;
    }
    ElMessageBox.confirm(i18n.global.t('setting.mfaClose'), i18n.global.t('setting.mfa'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            loading.value = true;
            await closeMFA()
                .then(async () => {
                    loading.value = false;
                    mfaDialogOpen.value = false;
                    savedMfaStatus.value = 'Disable';
                    emit('search');
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            form.mfaStatus = 'Enable';
        });
};

const handleApi = async () => {
    if (!form.apiInterfaceStatus) {
        return;
    }
    if (form.apiInterfaceStatus === 'Enable') {
        await ensureApiKey();
        apiDialogOpen.value = true;
        return;
    }
    ElMessageBox.confirm(i18n.global.t('setting.apiInterfaceClose'), i18n.global.t('setting.apiInterface'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            loading.value = true;
            form.apiInterfaceStatus = 'Disable';
            let param = {
                apiKey: form.apiKey,
                ipWhiteList: form.ipWhiteList,
                apiInterfaceStatus: form.apiInterfaceStatus,
                apiKeyValidityTime: form.apiKeyValidityTime,
            };
            await updateApiConfig(param)
                .then(() => {
                    loading.value = false;
                    apiDialogOpen.value = false;
                    savedApiStatus.value = 'Disable';
                    emit('search');
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            form.apiInterfaceStatus = 'Enable';
        });
};

defineExpose({
    openDrawer,
});
</script>

<style scoped>
.hidden-autofill-field {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
    pointer-events: none;
}

.setting-section {
    padding-top: 6px;
    border-top: 1px solid var(--el-border-color-lighter);
}

.setting-action-row {
    display: flex;
    align-items: center;
    gap: 12px;
}

.api-key-input {
    width: calc(100% - 125px);
}

.tooltip-help-list {
    margin: 4px 0 0;
    padding-left: 18px;
    color: inherit;
}

.tooltip-help-block {
    margin-bottom: 6px;
}

.tooltip-help-link {
    padding: 0;
    color: inherit;
    vertical-align: baseline;
}

.tooltip-help-link:hover {
    color: inherit;
}

.label-with-help {
    display: inline-flex;
    align-items: center;
    gap: 4px;
}

.help-icon {
    color: var(--el-text-color-secondary);
    cursor: help;
}

.copy_button {
    border-radius: 0;
    border-left-width: 0;
}

.passkey-prereq-dialog {
    display: flex;
    flex-direction: column;
    gap: 14px;
}

.passkey-prereq-title {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
}

.passkey-prereq-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.passkey-prereq-item {
    display: flex;
    align-items: center;
    gap: 10px;
    line-height: 1.5;
}

.passkey-prereq-item.is-browser {
    align-items: flex-start;
}

.passkey-prereq-status {
    flex-shrink: 0;
    font-size: 16px;
    margin-top: 1px;
}

.passkey-prereq-status.is-ok {
    color: var(--el-color-success);
}

.passkey-prereq-status.is-missing {
    color: var(--el-color-danger);
}

.passkey-prereq-browser-copy {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.passkey-prereq-detail {
    font-size: 12px;
    color: var(--el-text-color-secondary);
}

:deep(.api-key-input .el-input__wrapper) {
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
}
</style>
