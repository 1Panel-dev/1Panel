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
                                    <li v-if="isProductPro">{{ $t('setting.proxyHelper1') }}</li>
                                    <li v-if="isProductPro">{{ $t('setting.proxyHelper2') }}</li>
                                    <li>{{ $t('setting.proxyHelper3') }}</li>
                                </ul>
                            </template>
                        </el-alert>
                        <el-form-item :label="$t('setting.proxyType')" prop="proxyType">
                            <el-select v-model="form.proxyType" clearable>
                                <el-option value="close" :label="$t('commons.button.close')" />
                                <el-option value="socks5" label="SOCKS5" />
                                <el-option value="http" label="HTTP" />
                                <el-option value="https" label="HTTPS" />
                            </el-select>
                        </el-form-item>
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
                            <el-checkbox v-model="form.proxyPasswdKeepItem" :label="$t('setting.proxyPasswdKeep')" />
                        </el-form-item>
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

const globalStore = GlobalStore();
const emit = defineEmits<{ (e: 'search'): void }>();
const { isProductPro } = storeToRefs(globalStore);

const formRef = ref<FormInstance>();
const rules = reactive({
    proxyType: [Rules.requiredSelect],
    proxyUrl: [Rules.noSpace, Rules.requiredInput],
    proxyPortItem: [Rules.port],
});

const loading = ref(false);
const passwordVisible = ref<boolean>(false);
const form = reactive({
    proxyUrl: '',
    proxyType: '',
    proxyPort: '',
    proxyPortItem: 7890,
    proxyUser: '',
    proxyPasswd: '',
    proxyPasswdKeep: '',
    proxyPasswdKeepItem: false,
});

interface DialogProps {
    url: string;
    type: string;
    port: string;
    user: string;
    passwd: string;
    passwdKeep: string;
}
const acceptParams = (params: DialogProps): void => {
    form.proxyUrl = params.url || '127.0.0.1';
    form.proxyType = params.type || 'close';
    form.proxyPortItem = params.port ? Number(params.port) : 7890;
    form.proxyUser = params.user;
    form.proxyPasswd = params.passwd;
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
        };
        if (!isClose) {
            params.proxyPasswdKeep = form.proxyPasswdKeepItem ? 'Enable' : 'Disable';
        }
        if (form.proxyType === 'http' || form.proxyType === 'https') {
            params.proxyUrl = form.proxyType + '://' + form.proxyUrl;
        }
        loading.value = true;
        await updateProxy(params)
            .then(async () => {
                loading.value = false;
                emit('search');
                passwordVisible.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const handleClose = () => {
    passwordVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
