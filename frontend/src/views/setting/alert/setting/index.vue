<template>
    <div>
        <LayoutContent :title="$t('commons.button.set')" v-loading="loading" :divider="true">
            <template #title>
                <div class="flex items-center justify-between">
                    <span>{{ $t('xpack.alert.commonConfig') }}</span>
                    <el-button plain round size="default" @click="onChangeCommon(commonConfig.id)">
                        {{ $t('commons.button.edit') }}
                    </el-button>
                </div>
            </template>
            <template #main>
                <el-form
                    :model="form"
                    @submit.prevent
                    ref="alertFormRef"
                    :label-position="mobile ? 'top' : 'left'"
                    label-width="120px"
                >
                    <el-row>
                        <el-col>
                            <el-form-item :label="$t('xpack.alert.dailyAlertNum')" prop="dailyAlertNum">
                                {{ commonConfig.config.alertDailyNum }}
                            </el-form-item>

                            <el-form-item :label="$t('xpack.alert.sendTimeRange')" prop="sendTimeRange">
                                {{ sendTimeRange }}
                            </el-form-item>
                            <div v-if="!isMaster">
                                <el-form-item :label="$t('xpack.alert.offline')" prop="isOffline">
                                    <el-switch
                                        @change="onChangeOffline"
                                        v-model="commonConfig.config.isOffline"
                                        active-value="Enable"
                                        inactive-value="Disable"
                                    ></el-switch>
                                    <span class="input-help">{{ $t('xpack.alert.offlineHelper') }}</span>
                                </el-form-item>
                            </div>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>
        <LayoutContent :title="$t('commons.button.set')" v-loading="loading" :divider="true">
            <template #title>{{ $t('xpack.alert.methodConfig') }}</template>
            <template #main>
                <div class="grid gap-4 grid-cols-1 md:grid-cols-2 xl:grid-cols-3">
                    <el-card class="rounded-2xl shadow hover:shadow-md transition-all">
                        <div class="flex items-center justify-between mb-2">
                            <div class="text-lg font-semibold">{{ $t('xpack.alert.emailConfig') }}</div>
                            <div>
                                <el-button
                                    plain
                                    round
                                    size="default"
                                    :disabled="!emailConfig.id"
                                    @click="onChangeEmail(emailConfig.id)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button
                                    size="default"
                                    plain
                                    round
                                    :disabled="!emailConfig.id"
                                    @click="onDelete(emailConfig.id)"
                                >
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                        <div class="text-sm mb-2">{{ $t('xpack.alert.emailConfigHelper') }}</div>
                        <el-divider class="!mb-2 !mt-3" />
                        <div class="text-sm email-form" v-if="emailConfig.id">
                            <el-form
                                :model="form"
                                @submit.prevent
                                ref="alertFormRef"
                                :label-position="mobile ? 'top' : 'left'"
                                label-width="110px"
                            >
                                <el-form-item :label="$t('xpack.alert.displayName')" prop="displayName">
                                    {{ emailConfig.config.displayName }}
                                </el-form-item>
                                <el-form-item :label="$t('xpack.alert.sender')" prop="sender">
                                    {{ emailConfig.config.sender }}
                                </el-form-item>
                                <el-form-item :label="$t('xpack.alert.host')" prop="host">
                                    {{ emailConfig.config.host }}
                                </el-form-item>
                                <el-form-item :label="$t('xpack.alert.port')" prop="port">
                                    {{ emailConfig.config.port }}
                                </el-form-item>
                                <el-form-item :label="$t('xpack.alert.encryption')" prop="encryption">
                                    {{ emailConfig.config.encryption }}
                                </el-form-item>
                                <el-form-item :label="$t('xpack.alert.recipient')" prop="recipient">
                                    {{ emailConfig.config.recipient }}
                                </el-form-item>
                            </el-form>
                        </div>
                        <el-alert v-else center class="alert" style="height: 257px" :closable="false">
                            <el-button size="large" round plain type="primary" @click="onChangeEmail(0)">
                                {{ $t('commons.button.create') }}{{ $t('xpack.alert.emailConfig') }}
                            </el-button>
                        </el-alert>
                    </el-card>
                    <el-card
                        class="rounded-2xl shadow hover:shadow-md transition-all"
                        v-if="globalStore.isProductPro && !globalStore.isIntl"
                    >
                        <div class="flex items-center justify-between mb-2">
                            <div class="text-lg font-semibold">{{ $t('xpack.alert.smsConfig') }}</div>
                            <div>
                                <el-button plain round @click="onChangePhone(smsConfig.id)">
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                            </div>
                        </div>
                        <div class="text-sm mb-2">{{ $t('xpack.alert.smsConfigHelper') }}</div>
                        <el-divider class="!mb-2 !mt-3" />
                        <el-form-item :label="$t('xpack.alert.phone')">
                            <span v-if="smsConfig.config.phone">{{ smsConfig.config.phone }}</span>
                            <span v-else class="label">{{ $t('xpack.alert.defaultPhone') }}</span>
                        </el-form-item>
                    </el-card>
                </div>
            </template>
        </LayoutContent>

        <EmailDrawer ref="emailRef" @search="search" />
        <Phone ref="phoneRef" @search="search" />
        <SendTimeRange ref="sendTimeRangeRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { GlobalStore } from '@/store';
import { ListAlertConfigs, DeleteAlertConfig, UpdateAlertConfig } from '@/api/modules/alert';
import { ElMessageBox, FormInstance } from 'element-plus';
import Phone from '@/views/setting/alert/setting/phone/index.vue';
import SendTimeRange from '@/views/setting/alert/setting/time-range/index.vue';
import i18n from '@/lang';
import { storeToRefs } from 'pinia';
import { MsgSuccess } from '@/utils/message';
import EmailDrawer from '@/views/setting/alert/setting/email/index.vue';
import { Alert } from '@/api/interface/alert';

const globalStore = GlobalStore();
const { isMaster } = storeToRefs(globalStore);
const loading = ref(false);

const alertFormRef = ref<FormInstance>();
const phoneRef = ref();
const emailRef = ref();
const sendTimeRangeRef = ref();
const sendTimeRangeValue = ref();
const sendTimeRange = ref();

const isInitialized = ref(false);
export interface EmailConfig {
    id?: number;
    type: string;
    title: string;
    status: string;
    config: {
        status?: string;
        sender?: string;
        password?: string;
        displayName?: string;
        host?: string;
        port?: number;
        encryption?: string;
        recipient?: string;
    };
}
const defaultEmailConfig: EmailConfig = {
    id: undefined,
    type: 'email',
    title: 'xpack.alert.emailConfig',
    status: 'Enable',
    config: {
        displayName: '',
        sender: '',
        password: '',
        host: '',
        port: 25,
        encryption: 'NONE',
        status: '',
        recipient: '',
    },
};
const emailConfig = ref<EmailConfig>({ ...defaultEmailConfig });

export interface CommonConfig {
    id?: number;
    type: string;
    title: string;
    status: string;
    config: {
        isOffline?: string;
        alertDailyNum?: number;
        alertSendTimeRange?: string;
    };
}
const defaultCommonConfig: CommonConfig = {
    id: undefined,
    type: 'common',
    title: 'xpack.alert.commonConfig',
    status: 'Ena      ble',
    config: {
        alertDailyNum: 50,
        alertSendTimeRange:
            i18n.global.t('xpack.alert.noticeAlert') +
            ': ' +
            '08:00:00 - 23:59:59' +
            ' | ' +
            i18n.global.t('xpack.alert.resourceAlert') +
            ': ' +
            '00:00:00 - 23:59:59',
        isOffline: 'Disable',
    },
};

const commonConfig = ref<CommonConfig>({ ...defaultCommonConfig });

export interface SmsConfig {
    id?: number;
    type: string;
    title: string;
    status: string;
    config: {
        phone?: string;
    };
}
const defaultSmsConfig: SmsConfig = {
    id: undefined,
    type: 'sms',
    title: 'xpack.alert.smsConfig',
    status: 'Enable',
    config: {
        phone: '',
    },
};
const smsConfig = ref<SmsConfig>({ ...defaultSmsConfig });

const config = ref<Alert.AlertConfigInfo>({
    id: 0,
    type: '',
    title: '',
    status: '',
    config: '',
});
const mobile = computed(() => {
    return globalStore.isMobile();
});

function parseConfig<T extends object>(raw: any, fallback: T): T {
    try {
        const parsed = JSON.parse(raw.config || '{}');
        return {
            ...fallback,
            ...parsed,
        };
    } catch (err) {
        return { ...fallback };
    }
}

function assignConfig<T extends { config: any }>(raw: any, target: Ref<T>, fallback: T) {
    if (raw) {
        target.value = {
            ...(fallback as any),
            id: raw.id,
            type: raw.type,
            title: raw.title,
            status: raw.status,
            config: parseConfig(raw, fallback.config),
        };
    } else {
        target.value = { ...fallback };
    }
}

const search = async () => {
    loading.value = true;
    try {
        const res = await ListAlertConfigs();
        const emailFound = res.data.find((s: any) => s.type === 'email');
        assignConfig(emailFound, emailConfig, defaultEmailConfig);

        const commonFound = res.data.find((s: any) => s.type === 'common');
        assignConfig(commonFound, commonConfig, defaultCommonConfig);

        const smsFound = res.data.find((s: any) => s.type === 'sms');
        assignConfig(smsFound, smsConfig, defaultSmsConfig);
        sendTimeRangeValue.value = commonConfig.value.config.alertSendTimeRange;
        const noticeTimeRange = sendTimeRangeValue.value.noticeAlert.sendTimeRange || '08:00:00 - 23:59:59';
        const resourceTimeRange = sendTimeRangeValue.value.resourceAlert.sendTimeRange || '00:00:00 - 23:59:59';
        sendTimeRange.value =
            i18n.global.t('xpack.alert.noticeAlert') +
            ': ' +
            noticeTimeRange +
            ' | ' +
            i18n.global.t('xpack.alert.resourceAlert') +
            ': ' +
            resourceTimeRange;
        isInitialized.value = true;
    } finally {
        loading.value = false;
    }
};

const onChangePhone = (id: any) => {
    phoneRef.value.acceptParams({ id: id, phone: smsConfig.value.config.phone });
};

const onChangeCommon = (id: any) => {
    sendTimeRangeRef.value.acceptParams({
        id: id,
        dailyAlertNum: commonConfig.value.config.alertDailyNum,
        sendTimeRange: sendTimeRangeValue.value,
        isOffline: commonConfig.value.config.isOffline,
    });
};

const onChangeEmail = (id: number) => {
    emailRef.value.acceptParams({ id: id, config: emailConfig.value.config });
};

const onChangeOffline = async () => {
    if (!isInitialized.value) return;
    if (!isMaster.value && commonConfig.value.config.isOffline != '') {
        let title =
            commonConfig.value.config.isOffline == 'Enable'
                ? i18n.global.t('xpack.alert.offlineOff')
                : i18n.global.t('xpack.alert.offlineClose');
        let content =
            commonConfig.value.config.isOffline == 'Enable'
                ? i18n.global.t('xpack.alert.offlineOffHelper')
                : i18n.global.t('xpack.alert.offlineCloseHelper');
        ElMessageBox.confirm(content, title, {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        })
            .then(async () => {
                loading.value = true;
                try {
                    config.value.id = commonConfig.value.id;
                    config.value.type = 'common';
                    config.value.title = 'xpack.alert.commonConfig';
                    config.value.status = 'Enable';
                    config.value.config = JSON.stringify(commonConfig.value.config);
                    await UpdateAlertConfig(config.value);
                    loading.value = false;
                    await search();
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                } catch (error) {
                    loading.value = false;
                }
            })
            .catch(() => {
                commonConfig.value.config.isOffline =
                    commonConfig.value.config.isOffline == 'Enable' ? 'Disable' : 'Enable';
            });
    }
};

const onDelete = (id: number) => {
    ElMessageBox.confirm(i18n.global.t('xpack.alert.deleteConfigMsg'), i18n.global.t('xpack.alert.deleteConfigTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        await DeleteAlertConfig({ id: id });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    });
};
onMounted(async () => {
    await search();
});
</script>
<style scoped lang="scss">
.label {
    color: var(--el-text-color-placeholder);
}
.email-form {
    .el-form-item {
        margin-bottom: 0 !important;
    }
}
</style>
