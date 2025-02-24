<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            @close="handleClose"
            size="45%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.apiInterface')" :back="handleClose" />
            </template>
            <el-alert class="common-prompt" :closable="false" type="warning">
                <template #default>
                    <ul>
                        <li>
                            <el-text type="danger">{{ $t('setting.apiInterfaceAlert1') }}</el-text>
                        </li>
                        <li>
                            <el-text type="danger">{{ $t('setting.apiInterfaceAlert2') }}</el-text>
                        </li>
                        <li>
                            <el-link :href="apiURL" type="warning" target="_blank" class="mb-0.5 ml-0.5">
                                {{ $t('setting.apiInterfaceAlert3') }}
                            </el-link>
                        </li>
                        <li>
                            <el-link :href="panelURL" type="warning" target="_blank" class="mb-0.5 ml-0.5">
                                {{ $t('setting.apiInterfaceAlert4') }}
                            </el-link>
                        </li>
                    </ul>
                </template>
            </el-alert>
            <el-form
                :model="form"
                ref="formRef"
                @submit.prevent
                v-loading="loading"
                label-position="top"
                :rules="rules"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.apiKey')" prop="apiKey">
                            <el-input v-model="form.apiKey" readonly>
                                <template #suffix>
                                    <CopyButton type="icon" :content="form.apiKey" class="w-30" />
                                </template>
                                <template #append>
                                    <el-button @click="resetApiKey()">
                                        {{ $t('commons.button.reset') }}
                                    </el-button>
                                </template>
                            </el-input>
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
                            <el-input
                                :placeholder="$t('setting.apiKeyValidityTimeEgs')"
                                v-model="form.apiKeyValidityTime"
                            >
                                <template #append>{{ $t('commons.units.minute') }}</template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('setting.apiKeyValidityTimeHelper') }}
                            </span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { generateApiKey, updateApiConfig } from '@/api/modules/setting';
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox, FormInstance } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { checkCidr, checkCidrV6, checkIpV4V6 } from '@/utils/util';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const loading = ref();
const drawerVisible = ref();
const formRef = ref();
const apiURL = `${window.location.protocol}//${window.location.hostname}${
    window.location.port ? `:${window.location.port}` : ''
}/1panel/swagger/index.html`;
const panelURL = `${globalStore.docsUrl}/dev_manual/api_manual/`;

const form = reactive({
    apiKey: '',
    ipWhiteList: '',
    apiInterfaceStatus: '',
    apiKeyValidityTime: 120,
});

const rules = reactive({
    ipWhiteList: [Rules.requiredInput, { validator: checkIPs, trigger: 'blur' }],
    apiKey: [Rules.requiredInput],
    apiKeyValidityTime: [Rules.requiredInput, Rules.integerNumberWith0],
});

interface DialogProps {
    apiInterfaceStatus: string;
    apiKey: string;
    ipWhiteList: string;
    apiKeyValidityTime: number;
}

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
                } else {
                    if (checkCidr(item)) {
                        return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                    }
                }
            } else if (checkIpV4V6(item)) {
                return callback(new Error(i18n.global.t('firewall.addressFormatError')));
            }
        }
    }
    callback();
}

const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = async (params: DialogProps): Promise<void> => {
    form.apiInterfaceStatus = params.apiInterfaceStatus;
    form.apiKey = params.apiKey;
    if (params.apiKey == '') {
        await generateApiKey().then((res) => {
            form.apiKey = res.data;
        });
    }
    form.ipWhiteList = params.ipWhiteList;
    form.apiKeyValidityTime = params.apiKeyValidityTime;
    drawerVisible.value = true;
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

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let param = {
            apiKey: form.apiKey,
            ipWhiteList: form.ipWhiteList,
            apiInterfaceStatus: form.apiInterfaceStatus,
            apiKeyValidityTime: form.apiKeyValidityTime,
        };
        loading.value = true;
        await updateApiConfig(param)
            .then(() => {
                loading.value = false;
                drawerVisible.value = false;
                emit('search');
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
