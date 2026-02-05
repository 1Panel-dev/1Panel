<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.passkey')" @close="handleClose" size="large">
        <div>
            <el-form label-position="top">
                <el-form-item :label="$t('setting.passkeyName')">
                    <el-input v-model.trim="passkeyForm.name" :placeholder="$t('setting.passkeyNameHelper')" />
                </el-form-item>
                <el-button type="primary" @click="registerPasskey" :disabled="!canRegisterPasskey">
                    {{ $t('setting.passkeyAdd') }}
                </el-button>
                <span class="text-xs text-gray-500 ml-3">{{ passkeyCountText }}</span>
            </el-form>
        </div>
        <el-divider />
        <el-table :data="passkeyList" v-loading="passkeyLoading">
            <el-table-column prop="name" :label="$t('setting.passkeyName')" min-width="120" />
            <el-table-column prop="createdAt" :label="$t('setting.passkeyCreatedAt')" min-width="160" />
            <el-table-column :label="$t('setting.passkeyLastUsedAt')" min-width="160">
                <template #default="scope">
                    <span>{{ scope.row.lastUsedAt || '-' }}</span>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.operate')" width="120">
                <template #default="scope">
                    <el-button link type="danger" @click="removePasskey(scope.row.id)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
            </el-table-column>
        </el-table>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import {
    passkeyRegisterBegin,
    passkeyRegisterFinish,
    passkeyList as fetchPasskeyList,
    passkeyDelete,
} from '@/api/modules/setting';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Setting } from '@/api/interface/setting';
import { base64UrlToBuffer, bufferToBase64Url } from '@/utils/util';

interface DrawerParams {
    sslStatus: string;
    supported: boolean;
}

const drawerVisible = ref(false);
const passkeyLoading = ref(false);
const passkeyList = ref<Setting.PasskeyInfo[]>([]);
const passkeyForm = reactive({ name: '' });
const passkeySupported = ref(false);
const sslStatus = ref('Disable');
const passkeyMaxCount = 5;

const passkeyCountText = computed(() => {
    return i18n.global.t('setting.passkeyCount', [passkeyList.value.length, passkeyMaxCount]);
});

const canRegisterPasskey = computed(() => {
    return (
        sslStatus.value !== 'Disable' &&
        passkeySupported.value &&
        passkeyList.value.length < passkeyMaxCount &&
        passkeyForm.name.trim().length > 0
    );
});

const acceptParams = async (params: DrawerParams) => {
    sslStatus.value = params.sslStatus;
    passkeySupported.value = params.supported;
    drawerVisible.value = true;
    await loadPasskeys();
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
    if (sslStatus.value === 'Disable') {
        MsgError(i18n.global.t('setting.passkeyRequireSSL'));
        return;
    }
    if (!passkeySupported.value) {
        MsgError(i18n.global.t('setting.passkeyNotSupported'));
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

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
