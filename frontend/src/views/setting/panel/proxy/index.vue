<template>
    <div>
        <el-drawer
            v-model="passwordVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.proxy')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" :rules="rules" v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-alert class="common-prompt" :closable="false" type="warning">
                            <template #default>
                                {{ $t('setting.proxyHelper') }}
                                <ul style="margin-left: -20px">
                                    <li>{{ $t('setting.proxyHelper1') }}</li>
                                    <li>{{ $t('setting.proxyHelper2') }}</li>
                                    <li>{{ $t('setting.proxyHelper4') }}</li>
                                    <li>{{ $t('setting.proxyHelper3') }}</li>
                                </ul>
                            </template>
                        </el-alert>
                        <el-form-item :label="$t('setting.proxyType')" prop="proxyType">
                            <el-select v-model="form.proxyType" clearable>
                                <el-option value="close" :label="$t('commons.button.off')" />
                                <el-option value="socks5" label="SOCKS5" />
                                <el-option value="http" label="HTTP" />
                                <el-option value="https" label="HTTPS" />
                            </el-select>
                        </el-form-item>
                        <div v-if="form.proxyType !== 'close'">
                            <el-form-item :label="$t('setting.proxyUrl')" prop="proxyUrl">
                                <el-input
                                    clearable
                                    v-model.trim="form.proxyUrl"
                                    v-if="form.proxyType == 'http' || form.proxyType === 'https'"
                                >
                                    <template #prepend>
                                        <span>{{ form.proxyType }}</span>
                                    </template>
                                </el-input>
                                <el-input clearable v-model.trim="form.proxyUrl" v-else />
                            </el-form-item>
                            <el-form-item :label="$t('setting.proxyPort')" prop="proxyPortItem">
                                <el-input clearable type="number" v-model.number="form.proxyPortItem" />
                            </el-form-item>
                            <el-form-item :label="$t('commons.login.username')" prop="proxyUser">
                                <el-input clearable v-model.trim="form.proxyUser" />
                            </el-form-item>
                            <el-form-item :label="$t('commons.login.password')" prop="proxyPasswd">
                                <el-input type="password" show-password clearable v-model.trim="form.proxyPasswd" />
                            </el-form-item>
                            <el-form-item>
                                <el-checkbox
                                    v-model="form.proxyPasswdKeepItem"
                                    :label="$t('setting.proxyPasswdKeep')"
                                />
                            </el-form-item>
                            <el-form-item v-if="isProductPro">
                                <el-checkbox v-model="form.proxyDocker" :label="$t('setting.proxyDocker')" />
                                <span class="input-help">{{ $t('setting.proxyDockerHelper') }}</span>
                            </el-form-item>
                        </div>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="passwordVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="submitChangePassword(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit" />
    </div>
</template>

<script lang="ts" setup>
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { updateProxy } from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
import { updateXpackSettingByKey } from '@/utils/xpack';
import { updateDaemonJson } from '@/api/modules/container';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { escapeProxyURL } from '@/utils/util';

const globalStore = GlobalStore();
const emit = defineEmits<{ (e: 'search'): void }>();
const { isProductPro } = storeToRefs(globalStore);

const confirmDialogRef = ref();
const formRef = ref<FormInstance>();
const rules = reactive({
    proxyType: [Rules.requiredSelect],
    proxyUrl: [Rules.noSpace, Rules.requiredInput],
    proxyPortItem: [Rules.port],
});

const loading = ref(false);
const passwordVisible = ref<boolean>(false);
const proxyDockerVisible = ref<boolean>(false);
const form = reactive({
    proxyUrl: '',
    proxyType: '',
    proxyPort: '',
    proxyPortItem: 7890,
    proxyUser: '',
    proxyPasswd: '',
    proxyPasswdKeep: '',
    proxyPasswdKeepItem: false,
    proxyDocker: false,
});

interface DialogProps {
    url: string;
    type: string;
    port: string;
    user: string;
    passwd: string;
    passwdKeep: string;
    proxyDocker: string;
}
const acceptParams = (params: DialogProps): void => {
    if (params.url) {
        if (params.type === 'http' || params.type === 'https') {
            form.proxyUrl = params.url.replaceAll(params.type + '://', '');
        } else {
            form.proxyUrl = params.url;
        }
    } else {
        form.proxyUrl = '127.0.0.1';
    }
    form.proxyType = params.type || 'close';
    form.proxyPortItem = params.port ? Number(params.port) : 7890;
    form.proxyUser = params.user;
    form.proxyPasswd = params.passwd;
    form.proxyDocker = params.proxyDocker !== '';
    proxyDockerVisible.value = params.proxyDocker !== '';
    passwordVisible.value = true;
    form.proxyPasswdKeepItem = params.passwdKeep === 'Enable';
};

const submitChangePassword = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let isClose = form.proxyType === '' || form.proxyType === 'close';
        let params = {
            proxyType: isClose ? '' : form.proxyType,
            proxyUrl: isClose ? '' : form.proxyUrl,
            proxyPort: isClose ? '' : form.proxyPortItem + '',
            proxyUser: isClose ? '' : form.proxyUser,
            proxyPasswd: isClose ? '' : form.proxyPasswd,
            proxyPasswdKeep: '',
            proxyDocker: isClose ? false : form.proxyDocker,
        };
        if (!isClose) {
            params.proxyPasswdKeep = form.proxyPasswdKeepItem ? 'Enable' : 'Disable';
        }
        if (form.proxyType === 'http' || form.proxyType === 'https') {
            params.proxyUrl = form.proxyType + '://' + form.proxyUrl;
        }
        if (
            isProductPro.value &&
            (params.proxyDocker ||
                (proxyDockerVisible.value && isClose) ||
                (proxyDockerVisible.value && !isClose) ||
                (proxyDockerVisible.value && !params.proxyDocker))
        ) {
            let confirmParams = {
                header: i18n.global.t('setting.confDockerProxy'),
                operationInfo: i18n.global.t('setting.restartNowHelper'),
                submitInputInfo: i18n.global.t('setting.restartNow'),
            };
            confirmDialogRef.value!.acceptParams(confirmParams);
        } else {
            loading.value = true;
            await updateProxy(params)
                .then(async () => {
                    loading.value = false;
                    emit('search');
                    passwordVisible.value = false;
                    if (isClose) {
                        await updateDaemonJson(`${form.proxyType}-proxy`, '');
                    }
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        }
    });
};

const onSubmit = async () => {
    try {
        loading.value = true;
        let isClose = form.proxyType === '' || form.proxyType === 'close';
        let params = {
            proxyType: isClose ? '' : form.proxyType,
            proxyUrl: isClose ? '' : form.proxyUrl,
            proxyPort: isClose ? '' : form.proxyPortItem + '',
            proxyUser: isClose ? '' : form.proxyUser,
            proxyPasswd: isClose ? '' : form.proxyPasswd,
            proxyPasswdKeep: '',
            proxyDocker: isClose ? false : form.proxyDocker,
        };
        if (!isClose) {
            params.proxyPasswdKeep = form.proxyPasswdKeepItem ? 'Enable' : 'Disable';
        }
        let proxyPort = params.proxyPort ? `:${params.proxyPort}` : '';
        let proxyUser = params.proxyUser ? `${escapeProxyURL(params.proxyUser)}` : '';
        let proxyPasswd = '';
        if (params.proxyUser) {
            proxyPasswd = params.proxyPasswd ? `:${escapeProxyURL(params.proxyPasswd)}@` : '@';
        }

        let proxyUrl = form.proxyType + '://' + proxyUser + proxyPasswd + form.proxyUrl + proxyPort;
        if (form.proxyType === 'http' || form.proxyType === 'https') {
            params.proxyUrl = form.proxyType + '://' + form.proxyUrl;
        }
        await updateProxy(params);
        if (isClose || params.proxyDocker === false) {
            proxyUrl = '';
        }
        await updateXpackSettingByKey('ProxyDocker', proxyUrl);
        await updateDaemonJson(`${form.proxyType}-proxy`, proxyUrl);
        emit('search');
        handleClose();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch (error) {
        console.error(error);
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    passwordVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
