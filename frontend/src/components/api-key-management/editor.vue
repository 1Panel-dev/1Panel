<template>
    <DrawerPro
        v-model="visible"
        :header="item ? $t('commons.button.edit') : $t('apiKeyManagement.create')"
        size="min(640px, 100vw)"
        :auto-close="false"
        :confirm-before-close="true"
        @before-close="beforeClose"
    >
        <template v-if="created">
            <el-alert
                :title="$t(secret ? 'apiKeyManagement.saveSecret' : 'apiKeyManagement.alreadyCreated')"
                :type="secret ? 'warning' : 'info'"
                :closable="false"
                show-icon
                class="mb-4"
            />
            <el-form label-position="top">
                <el-form-item :label="$t('commons.table.name')">{{ created.name }}</el-form-item>
                <el-form-item v-if="secret" :label="$t('setting.apiKey')">
                    <el-input :model-value="secret" readonly autocomplete="off">
                        <template #append>
                            <CopyButton :content="secret" :is-icon="false" />
                        </template>
                    </el-input>
                </el-form-item>
            </el-form>
            <p v-if="created.allowAppBinding" class="input-help mb-4 leading-5">
                {{ $t('apiKeyManagement.bindingException') }}
            </p>
            <el-checkbox v-if="secret" v-model="saved">{{ $t('apiKeyManagement.saved') }}</el-checkbox>
        </template>
        <el-form v-else ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="submit">
            <el-alert v-if="!item" type="warning" :closable="false" class="!mb-4">
                <ul class="m-0 list-disc space-y-1 pl-4 break-words leading-6">
                    <li>
                        <span class="text-[var(--panel-alert-error-text-color,var(--el-color-danger))]">
                            {{ $t('setting.apiInterfaceAlert1') }}
                        </span>
                    </li>
                    <li>
                        <span class="text-[var(--panel-alert-error-text-color,var(--el-color-danger))]">
                            {{ $t('setting.apiInterfaceAlert2') }}
                        </span>
                    </li>
                    <li>
                        <el-link
                            href="/1panel/swagger/index.html"
                            target="_blank"
                            rel="noopener noreferrer"
                            type="warning"
                        >
                            {{ $t('setting.apiInterfaceAlert3') }}
                        </el-link>
                    </li>
                    <li v-if="!isFxplay">
                        <el-link
                            :href="`${docsUrl}/dev_manual/api_manual/`"
                            target="_blank"
                            rel="noopener noreferrer"
                            type="warning"
                        >
                            {{ $t('setting.apiInterfaceAlert4') }}
                        </el-link>
                    </li>
                </ul>
            </el-alert>
            <el-alert
                v-if="uncertain"
                :title="$t('apiKeyManagement.uncertain')"
                type="warning"
                :closable="false"
                class="mb-4"
            />
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="form.name" :maxlength="64" :disabled="busy || uncertain || isLegacy" />
            </el-form-item>
            <el-form-item v-if="!isLegacy" :label="$t('commons.table.description')" prop="description">
                <el-input
                    v-model="form.description"
                    type="textarea"
                    :maxlength="256"
                    :rows="2"
                    :disabled="busy || uncertain"
                />
            </el-form-item>
            <el-form-item :label="$t('setting.ipWhiteList')" prop="ipWhiteList">
                <el-radio-group v-model="ipMode" :disabled="busy || uncertain" class="mb-2">
                    <el-radio value="restricted">{{ $t('apiKeyManagement.specifiedIPs') }}</el-radio>
                    <el-radio value="any">{{ $t('apiKeyManagement.anyIP') }}</el-radio>
                </el-radio-group>
                <el-input
                    v-if="ipMode === 'restricted'"
                    v-model="form.ipWhiteList"
                    type="textarea"
                    :rows="4"
                    :maxlength="4096"
                    :disabled="busy || uncertain"
                    :placeholder="$t('setting.ipWhiteListEgs')"
                />
                <span class="input-help mt-2 leading-5">{{ $t('apiKeyManagement.ipHelp') }}</span>
            </el-form-item>
            <el-form-item :label="$t('setting.apiTrustedProxies')" prop="apiTrustedProxies">
                <el-input
                    v-model="form.apiTrustedProxies"
                    type="textarea"
                    :rows="2"
                    :maxlength="4096"
                    :disabled="busy || uncertain"
                />
                <span class="input-help mt-2 leading-5">{{ $t('setting.apiTrustedProxiesHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('apiKeyManagement.signatureWindow')" prop="apiKeyValidityTime">
                <el-input-number
                    v-model="form.apiKeyValidityTime"
                    :min="0"
                    :max="isLegacy ? undefined : 1440"
                    :precision="0"
                    :disabled="busy || uncertain"
                />
                <span class="ml-2">{{ $t('commons.units.minute') }}</span>
                <span class="input-help mt-2 leading-5">{{ $t('setting.apiKeyValidityTimeHelper') }}</span>
            </el-form-item>
            <el-form-item v-if="!isLegacy" :label="$t('apiKeyManagement.expiresAt')" prop="expiresAt">
                <el-checkbox v-model="neverExpires" :disabled="busy || uncertain" class="w-full mb-2">
                    {{ $t('apiKeyManagement.never') }}
                </el-checkbox>
                <el-date-picker
                    v-if="!neverExpires"
                    v-model="form.expiresAt"
                    type="datetime"
                    :disabled="busy || uncertain"
                    :clearable="false"
                    class="!w-full"
                />
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="form.allowAppBinding" :disabled="busy || uncertain">
                    {{ $t('apiKeyManagement.allowAppBinding') }}
                </el-checkbox>
                <span class="input-help mt-2 leading-5">{{ $t('apiKeyManagement.bindingException') }}</span>
                <span class="input-help mt-2 leading-5">{{ $t('apiKeyManagement.bindingDisableHelp') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button v-if="!created" :disabled="busy" @click="beforeClose(() => (visible = false))">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button v-if="created" type="primary" :disabled="!!secret && !saved" @click="finish">
                {{ $t('commons.button.confirm') }}
            </el-button>
            <el-button v-else type="primary" :loading="busy" @click="submit">
                {{ $t(item ? 'commons.button.save' : 'apiKeyManagement.create') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { ElMessageBox, type FormInstance } from 'element-plus';
import { isAxiosError } from 'axios';
import { v4 as uuidv4 } from 'uuid';
import { onBeforeRouteLeave } from 'vue-router';
import DrawerPro from '@/components/drawer-pro/index.vue';
import type { APIKey } from '@/api/interface/api-key';
import { createAPIKey, updateAPIKey } from '@/api/modules/api-key';
import { ANY_API_KEY_IP, defaultAPIKeyExpiry, isAnyAPIKeyIP, normalizeAPIKeyIPs } from '@/utils/api-key';
import { checkCidr, checkCidrV6, checkIpV4V6 } from '@/utils/validate';
import { Rules } from '@/global/form-rules';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import i18n from '@/lang';

const emit = defineEmits<{ changed: []; created: [item: APIKey.Item] }>();
const { globalStore, docsUrl, isFxplay } = useGlobalStore();
const visible = ref(false);
const busy = ref(false);
const uncertain = ref(false);
const item = ref<APIKey.Item>();
const created = ref<APIKey.Item>();
const secret = ref('');
const saved = ref(false);
const formRef = ref<FormInstance>();
const ipMode = ref('restricted');
const neverExpires = ref(false);
const isLegacy = computed(() => item.value?.kind === 'legacy');
const form = reactive<APIKey.Editable>({
    name: '',
    description: '',
    ipWhiteList: '',
    apiTrustedProxies: '',
    apiKeyValidityTime: 120,
    expiresAt: null,
    allowAppBinding: false,
});
let requestID = '';
let pendingCreate: (APIKey.Editable & { requestID: string }) | undefined;
let editorVersion = 0;

const discardEditor = () => {
    editorVersion++;
    secret.value = '';
    created.value = undefined;
    visible.value = false;
    busy.value = false;
};

const validateIPs = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
    const entries = normalizeAPIKeyIPs(value).split('\n').filter(Boolean);
    const invalid = entries.some((entry) =>
        entry.includes('/') ? (entry.includes(':') ? checkCidrV6(entry) : checkCidr(entry)) : checkIpV4V6(entry),
    );
    callback(invalid ? new Error(i18n.global.t('firewall.addressFormatError')) : undefined);
};
const rules = {
    name: [Rules.requiredInput],
    ipWhiteList: [
        {
            validator: (rule: unknown, value: string, callback: (error?: Error) => void) => {
                if (ipMode.value === 'any') return callback();
                if (!normalizeAPIKeyIPs(value)) return callback(new Error(i18n.global.t('commons.rule.requiredInput')));
                validateIPs(rule, value, callback);
            },
            trigger: 'blur',
        },
    ],
    apiTrustedProxies: [{ validator: validateIPs, trigger: 'blur' }],
    apiKeyValidityTime: [Rules.requiredInput, Rules.integerNumberWith0],
    expiresAt: [
        {
            validator: (_rule: unknown, value: string | null, callback: (error?: Error) => void) => {
                callback(
                    isLegacy.value || neverExpires.value || (value && new Date(value).getTime() > Date.now())
                        ? undefined
                        : new Error(i18n.global.t('apiKeyManagement.futureExpiry')),
                );
            },
            trigger: 'change',
        },
    ],
};

const open = (existing?: APIKey.Item, fromApp = false) => {
    editorVersion++;
    item.value = existing;
    created.value = undefined;
    secret.value = '';
    saved.value = false;
    uncertain.value = false;
    pendingCreate = undefined;
    requestID = uuidv4();
    Object.assign(
        form,
        existing
            ? { ...existing }
            : {
                  name: '',
                  description: '',
                  ipWhiteList: '',
                  apiTrustedProxies: '',
                  apiKeyValidityTime: 120,
                  expiresAt: defaultAPIKeyExpiry(),
                  allowAppBinding: fromApp,
              },
    );
    if (isLegacy.value) form.name = i18n.global.t('apiKeyManagement.legacy');
    form.ipWhiteList = form.ipWhiteList.replace(/,/g, '\n');
    form.apiTrustedProxies = form.apiTrustedProxies.replace(/,/g, '\n');
    neverExpires.value = !form.expiresAt;
    ipMode.value = isAnyAPIKeyIP(form.ipWhiteList) ? 'any' : 'restricted';
    visible.value = true;
    nextTick(() => formRef.value?.clearValidate());
};

const finish = () => {
    secret.value = '';
    visible.value = false;
    if (created.value) emit('created', created.value);
};
const beforeClose = async (done: () => void) => {
    if (busy.value) return;
    if (secret.value && !saved.value) {
        await ElMessageBox.alert(i18n.global.t('apiKeyManagement.saveSecret'), i18n.global.t('apiKeyManagement.title'));
        return;
    }
    if (created.value) finish();
    else if (uncertain.value) emit('changed');
    done();
};

const submit = async () => {
    if (busy.value) return;
    busy.value = true;
    const ticket = editorVersion;
    try {
        if (!(await formRef.value?.validate().catch(() => false)) || ticket !== editorVersion) return;
        const payload: APIKey.Editable = {
            name: form.name.trim(),
            description: form.description,
            ipWhiteList: ipMode.value === 'any' ? ANY_API_KEY_IP : normalizeAPIKeyIPs(form.ipWhiteList),
            apiTrustedProxies: normalizeAPIKeyIPs(form.apiTrustedProxies),
            apiKeyValidityTime: form.apiKeyValidityTime,
            expiresAt: isLegacy.value || neverExpires.value ? null : new Date(form.expiresAt!).toISOString(),
            allowAppBinding: form.allowAppBinding,
        };
        if (item.value) {
            const response = await updateAPIKey({ ...payload, id: item.value.id, revision: item.value.revision });
            if (ticket !== editorVersion) return;
            if (response.data?.terminalClosePending) MsgWarning(i18n.global.t('apiKeyManagement.closePending'));
            else MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            visible.value = false;
        } else {
            pendingCreate ??= { ...payload, requestID };
            const response = await createAPIKey(pendingCreate);
            if (ticket !== editorVersion) return;
            created.value = response.data.item;
            secret.value = response.data.apiKey;
            uncertain.value = false;
        }
        emit('changed');
    } catch (error) {
        if (ticket === editorVersion && !item.value) {
            uncertain.value = isAxiosError(error);
            if (!uncertain.value) pendingCreate = undefined;
        }
    } finally {
        if (ticket === editorVersion) busy.value = false;
    }
};
watch(
    () => globalStore.isLogin,
    (loggedIn) => {
        if (!loggedIn) discardEditor();
    },
    { flush: 'sync' },
);
onBeforeUnmount(discardEditor);
onBeforeRouteLeave(async (to) => {
    if (!globalStore.isLogin || ['entrance', 'Expired', 'EnterpriseLicenseRequired'].includes(String(to.name))) {
        discardEditor();
        return true;
    }
    if (busy.value) return false;
    if (secret.value && !saved.value) {
        await ElMessageBox.alert(i18n.global.t('apiKeyManagement.saveSecret'), i18n.global.t('apiKeyManagement.title'));
        return false;
    }
    discardEditor();
    return true;
});
defineExpose({ open });
</script>
