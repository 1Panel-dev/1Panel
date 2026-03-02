<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.passkey')" @close="handleClose" size="large">
        <div class="mb-4">
            <el-alert
                v-if="!allPrerequisitesMet"
                :title="$t('setting.passkeyPrereqTitle')"
                type="warning"
                :closable="false"
                class="mb-4"
            >
                <template #default>
                    <div class="flex flex-col gap-2 mt-2">
                        <div class="flex items-center gap-2">
                            <el-icon :color="prereqBindDomain ? '#67c23a' : '#f56c6c'">
                                <Check v-if="prereqBindDomain" />
                                <Close v-else />
                            </el-icon>
                            <span>{{ $t('setting.passkeyPrereqBindDomain') }}</span>
                            <el-button v-if="!prereqBindDomain" link type="primary" @click="goConfigureDomain">
                                {{ $t('setting.passkeyPrereqGoSetup') }}
                            </el-button>
                        </div>
                        <div class="flex items-center gap-2">
                            <el-icon :color="prereqHttps ? '#67c23a' : '#f56c6c'">
                                <Check v-if="prereqHttps" />
                                <Close v-else />
                            </el-icon>
                            <span>{{ $t('setting.passkeyPrereqHttps') }}</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <el-icon :color="prereqBrowser ? '#67c23a' : '#f56c6c'">
                                <Check v-if="prereqBrowser" />
                                <Close v-else />
                            </el-icon>
                            <span>{{ $t('setting.passkeyPrereqBrowser') }}</span>
                        </div>
                    </div>
                </template>
            </el-alert>
        </div>
        <el-tabs v-model="activeTab" type="border-card">
            <el-tab-pane :label="$t('setting.passkeyKeyManagement')" name="keys">
                <el-form label-position="top">
                    <el-form-item :label="$t('setting.passkeyName')">
                        <el-input
                            v-model.trim="passkeyForm.name"
                            :placeholder="$t('setting.passkeyNameHelper')"
                            :disabled="!allPrerequisitesMet"
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
                                :disabled="!allPrerequisitesMet"
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
                        <span class="input-help">
                            {{ $t('setting.passkeyTrustedProxiesHelper') }}
                        </span>
                        <span class="input-help">
                            {{ $t('setting.allowIPEgs') }}
                        </span>
                    </el-form-item>
                </el-form>
            </el-tab-pane>
        </el-tabs>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import { Check, Close } from '@element-plus/icons-vue';
import { ElMessageBox } from 'element-plus';
import {
    getSettingInfo,
    passkeyRegisterBegin,
    passkeyRegisterFinish,
    passkeyList as fetchPasskeyList,
    passkeyDelete,
    updateSetting,
} from '@/api/modules/setting';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Setting } from '@/api/interface/setting';
import { base64UrlToBuffer, bufferToBase64Url } from '@/utils/util';

interface DrawerParams {
    bindDomain: string;
}

const emit = defineEmits(['go-configure-domain']);

const activeTab = ref('keys');
const drawerVisible = ref(false);
const passkeyLoading = ref(false);
const savePasskeyProxyLoading = ref(false);
const passkeyList = ref<Setting.PasskeyInfo[]>([]);
const passkeyForm = reactive({ name: '' });
const passkeyTrustedProxies = ref('');
const hasBindDomain = ref(false);
const passkeyMaxCount = 5;

const prereqBindDomain = computed(() => hasBindDomain.value);
const prereqHttps = computed(() => window.isSecureContext);
const prereqBrowser = computed(() => !!window.PublicKeyCredential);
const allPrerequisitesMet = computed(() => prereqBindDomain.value && prereqHttps.value && prereqBrowser.value);

const passkeyCountText = computed(() => {
    return i18n.global.t('setting.passkeyCount', [passkeyList.value.length, passkeyMaxCount]);
});

const canRegisterPasskey = computed(() => {
    return (
        allPrerequisitesMet.value && passkeyList.value.length < passkeyMaxCount && passkeyForm.name.trim().length > 0
    );
});

const acceptParams = async (params: DrawerParams) => {
    hasBindDomain.value = params.bindDomain.trim().length > 0;
    drawerVisible.value = true;
    await loadPasskeyTrustedProxies();
    await loadPasskeys();
};

const loadPasskeyTrustedProxies = async () => {
    try {
        const res = await getSettingInfo();
        passkeyTrustedProxies.value = res.data.passkeyTrustedProxies || '';
    } catch (error) {
        passkeyTrustedProxies.value = '';
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
    if (!allPrerequisitesMet.value) {
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
    } catch (res: any) {
        if (res?.message) {
            console.log(res.message);
        }
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
            loadPasskeyTrustedProxies();
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

const goConfigureDomain = () => {
    drawerVisible.value = false;
    emit('go-configure-domain');
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
