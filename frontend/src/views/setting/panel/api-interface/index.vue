<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            @close="handleClose"
            size="35%"
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
                            {{ $t('setting.apiInterfaceAlert3') }}
                            <el-link :href="apiURL" type="success" target="_blank" class="mb-0.5 ml-0.5">
                                {{ apiURL }}
                            </el-link>
                        </li>
                        <li>
                            {{ $t('setting.apiInterfaceAlert4') }}
                            <el-link :href="panelURL" type="success" target="_blank" class="mb-0.5 ml-0.5">
                                {{ panelURL }}
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
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onBind(formRef)">
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

const loading = ref();
const drawerVisible = ref();
const formRef = ref();
const apiURL = `${window.location.protocol}//${window.location.hostname}${
    window.location.port ? `:${window.location.port}` : ''
}/1panel/swagger/index.html`;
const panelURL = `https://1panel.cn/docs`;

const form = reactive({
    apiKey: '',
    ipWhiteList: '',
    apiInterfaceStatus: '',
});

const rules = reactive({
    ipWhiteList: [Rules.requiredInput],
    apiKey: [Rules.requiredInput],
});

interface DialogProps {
    apiInterfaceStatus: string;
    apiKey: string;
    ipWhiteList: string;
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
    drawerVisible.value = true;
};

const resetApiKey = async () => {
    loading.value = true;
    ElMessageBox.confirm(i18n.global.t('setting.apiKeyResetHelper'), i18n.global.t('setting.apiKeyReset'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
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

const onBind = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let param = {
            apiKey: form.apiKey,
            ipWhiteList: form.ipWhiteList,
            apiInterfaceStatus: form.apiInterfaceStatus,
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
