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
                    @submit.prevent
                    ref="alertFormRef"
                    :label-position="mobile ? 'top' : 'left'"
                    label-width="120px"
                >
                    <el-row>
                        <el-col>
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
                <el-alert type="info" :closable="false">
                    <template #title>
                        <div class="flex items-center justify-start">
                            {{ $t('xpack.alert.alertConfigHelper') }}
                            <span v-if="!globalStore.isProductPro">
                                {{ $t('commons.units.semicolon') }}{{ $t('xpack.alert.alertConfigProHelper') }}
                            </span>
                            <el-link
                                class="ml-1 text-xs"
                                type="primary"
                                target="_blank"
                                :href="globalStore.docsUrl + '/user_manual/settings/#3'"
                            >
                                {{ $t('commons.button.helpDoc') }}
                            </el-link>
                        </div>
                    </template>
                </el-alert>
                <div class="grid gap-4 grid-cols-1 md:grid-cols-2 xl:grid-cols-3 mt-3 app">
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
                        <div class="text-sm config-form" v-if="emailConfig.id">
                            <el-form
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
                                <el-form-item :label="$t('commons.login.username')" prop="userName">
                                    {{ emailConfig.config.userName || emailConfig.config.sender }}
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
                        <div v-else class="flex items-center justify-center" style="height: 257px">
                            <el-button size="large" round plain type="primary" @click="onChangeEmail(0)">
                                {{ $t('commons.button.create') }}{{ $t('xpack.alert.emailConfig') }}
                            </el-button>
                        </div>
                    </el-card>
                    <el-card
                        class="rounded-2xl shadow hover:shadow-md transition-all"
                        v-if="globalStore.isProductPro && !globalStore.isIntl"
                    >
                        <div class="flex items-center justify-between mb-2">
                            <div class="text-lg font-semibold">{{ $t('xpack.alert.weCom') }}</div>
                            <div>
                                <el-button
                                    plain
                                    round
                                    size="default"
                                    :disabled="!weComConfig.id"
                                    @click="onChangeWeCom(weComConfig.id)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button
                                    size="default"
                                    plain
                                    round
                                    :disabled="!weComConfig.id"
                                    @click="onDelete(weComConfig.id)"
                                >
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                        <div class="text-sm mb-2">{{ $t('xpack.alert.weComConfigHelper') }}</div>
                        <el-divider class="!mb-2 !mt-3" />
                        <div class="text-sm config-form" v-if="weComConfig.id">
                            <el-form
                                @submit.prevent
                                ref="alertFormRef"
                                :label-position="mobile ? 'top' : 'left'"
                                label-width="110px"
                            >
                                <el-form-item :label="$t('xpack.alert.webhookName')" prop="displayName">
                                    {{ weComConfig.config.displayName }}
                                </el-form-item>
                                <el-form-item :label="$t('xpack.alert.webhookUrl')" prop="url">
                                    <div class="webhook-field">
                                        <template v-if="weComUrlVisible">
                                            <el-tooltip :content="weComConfig.config.url" placement="top" effect="dark">
                                                <span class="webhook-text">
                                                    {{ weComConfig.config.url }}
                                                </span>
                                            </el-tooltip>
                                        </template>
                                        <template v-else>
                                            <span class="webhook-text">****************</span>
                                        </template>
                                        <el-icon class="webhook-icon" @click="weComUrlVisible = !weComUrlVisible">
                                            <Hide v-if="!weComUrlVisible" />
                                            <View v-else />
                                        </el-icon>
                                    </div>
                                </el-form-item>
                            </el-form>
                        </div>
                        <div v-else class="flex items-center justify-center" style="height: 257px">
                            <el-button size="large" round plain type="primary" @click="onChangeWeCom(0)">
                                {{ $t('commons.button.create') }}{{ $t('xpack.alert.weCom') }}
                            </el-button>
                        </div>
                    </el-card>
                    <el-card
                        class="rounded-2xl shadow hover:shadow-md transition-all"
                        v-if="globalStore.isProductPro && !globalStore.isIntl"
                    >
                        <div class="flex items-center justify-between mb-2">
                            <div class="text-lg font-semibold">{{ $t('xpack.alert.dingTalk') }}</div>
                            <div>
                                <el-button
                                    plain
                                    round
                                    size="default"
                                    :disabled="!dingTalkConfig.id"
                                    @click="onChangeDingTalk(dingTalkConfig.id)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button
                                    size="default"
                                    plain
                                    round
                                    :disabled="!dingTalkConfig.id"
                                    @click="onDelete(dingTalkConfig.id)"
                                >
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                        <div class="text-sm mb-2">{{ $t('xpack.alert.dingTalkConfigHelper') }}</div>
                        <el-divider class="!mb-2 !mt-3" />
                        <div class="text-sm config-form" v-if="dingTalkConfig.id">
                            <el-form
                                @submit.prevent
                                ref="alertFormRef"
                                :label-position="mobile ? 'top' : 'left'"
                                label-width="110px"
                            >
                                <el-form-item :label="$t('xpack.alert.webhookName')" prop="displayName">
                                    {{ dingTalkConfig.config.displayName }}
                                </el-form-item>
                                <el-form-item :label="$t('xpack.alert.webhookUrl')" prop="url">
                                    <div class="webhook-field">
                                        <template v-if="dingTalkUrlVisible">
                                            <el-tooltip
                                                :content="dingTalkConfig.config.url"
                                                placement="top"
                                                effect="dark"
                                            >
                                                <span class="webhook-text">
                                                    {{ dingTalkConfig.config.url }}
                                                </span>
                                            </el-tooltip>
                                        </template>
                                        <template v-else>
                                            <span class="webhook-text">****************</span>
                                        </template>
                                        <el-icon class="webhook-icon" @click="dingTalkUrlVisible = !dingTalkUrlVisible">
                                            <Hide v-if="!dingTalkUrlVisible" />
                                            <View v-else />
                                        </el-icon>
                                    </div>
                                </el-form-item>
                            </el-form>
                        </div>
                        <div v-else class="flex items-center justify-center" style="height: 257px">
                            <el-button size="large" round plain type="primary" @click="onChangeDingTalk(0)">
                                {{ $t('commons.button.create') }}{{ $t('xpack.alert.dingTalk') }}
                            </el-button>
                        </div>
                    </el-card>
                    <el-card
                        class="rounded-2xl shadow hover:shadow-md transition-all"
                        v-if="globalStore.isProductPro && !globalStore.isIntl"
                    >
                        <div class="flex items-center justify-between mb-2">
                            <div class="text-lg font-semibold">{{ $t('xpack.alert.feiShu') }}</div>
                            <div>
                                <el-button
                                    plain
                                    round
                                    size="default"
                                    :disabled="!feiShuConfig.id"
                                    @click="onChangeFeiShu(feiShuConfig.id)"
                                >
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                                <el-button
                                    size="default"
                                    plain
                                    round
                                    :disabled="!feiShuConfig.id"
                                    @click="onDelete(feiShuConfig.id)"
                                >
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                        <div class="text-sm mb-2">{{ $t('xpack.alert.feiShuConfigHelper') }}</div>
                        <el-divider class="!mb-2 !mt-3" />
                        <div class="text-sm config-form" v-if="feiShuConfig.id">
                            <el-form
                                @submit.prevent
                                ref="alertFormRef"
                                :label-position="mobile ? 'top' : 'left'"
                                label-width="110px"
                            >
                                <el-form-item :label="$t('xpack.alert.webhookName')" prop="displayName">
                                    {{ feiShuConfig.config.displayName }}
                                </el-form-item>
                                <el-form-item :label="$t('xpack.alert.webhookUrl')" prop="url">
                                    <div class="webhook-field">
                                        <template v-if="feiShuUrlVisible">
                                            <el-tooltip
                                                :content="feiShuConfig.config.url"
                                                placement="top"
                                                effect="dark"
                                            >
                                                <span class="webhook-text">
                                                    {{ feiShuConfig.config.url }}
                                                </span>
                                            </el-tooltip>
                                        </template>
                                        <template v-else>
                                            <span class="webhook-text">****************</span>
                                        </template>
                                        <el-icon class="webhook-icon" @click="feiShuUrlVisible = !feiShuUrlVisible">
                                            <Hide v-if="!feiShuUrlVisible" />
                                            <View v-else />
                                        </el-icon>
                                    </div>
                                </el-form-item>
                            </el-form>
                        </div>
                        <div v-else class="flex items-center justify-center" style="height: 257px">
                            <el-button size="large" round plain type="primary" @click="onChangeFeiShu(0)">
                                {{ $t('commons.button.create') }}{{ $t('xpack.alert.feiShu') }}
                            </el-button>
                        </div>
                    </el-card>
                    <el-card
                        class="rounded-2xl shadow hover:shadow-md transition-all"
                        v-if="globalStore.isProductPro && !globalStore.isIntl"
                    >
                        <div class="flex items-center justify-between mb-2">
                            <div class="text-lg font-semibold">
                                {{ $t('xpack.alert.smsConfig') }}
                            </div>
                            <div>
                                <el-button plain round @click="onChangePhone(smsConfig.id)">
                                    {{ $t('commons.button.edit') }}
                                </el-button>
                            </div>
                        </div>
                        <div class="text-sm mb-2 flex items-center justify-start">
                            {{ $t('xpack.alert.alertSmsHelper', [totalSms, usedSms]) }}
                            <el-link class="ml-1 text-xs" @click="goBuy" type="primary" icon="Position">
                                <span class="ml-0.5">{{ $t('xpack.alert.goBuy') }}</span>
                            </el-link>
                        </div>
                        <el-divider class="!mb-2 !mt-3" />
                        <div class="text-sm config-form">
                            <el-form
                                @submit.prevent
                                ref="alertFormRef"
                                :label-position="mobile ? 'top' : 'left'"
                                label-width="110px"
                            >
                                <el-form-item :label="$t('xpack.alert.phone')">
                                    <span v-if="smsConfig.config.phone">{{ smsConfig.config.phone }}</span>
                                    <span v-else class="label">{{ $t('xpack.alert.defaultPhone') }}</span>
                                </el-form-item>
                                <el-form-item :label="$t('xpack.alert.dailyAlertNum')" prop="dailyAlertNum">
                                    {{ smsConfig.config.alertDailyNum }}
                                </el-form-item>
                            </el-form>
                        </div>
                    </el-card>
                </div>
            </template>
        </LayoutContent>

        <EmailDrawer ref="emailRef" @search="search" />
        <Phone ref="phoneRef" @search="search" />
        <SendTimeRange ref="sendTimeRangeRef" @search="search" />
        <WebhookDrawer ref="webHookRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, Ref } from 'vue';
import { GlobalStore } from '@/store';
import { ListAlertConfigs, DeleteAlertConfig, UpdateAlertConfig } from '@/api/modules/alert';
import { ElMessageBox, FormInstance } from 'element-plus';
import { View, Hide } from '@element-plus/icons-vue';
import Phone from '@/views/setting/alert/setting/phone/index.vue';
import SendTimeRange from '@/views/setting/alert/setting/time-range/index.vue';
import i18n from '@/lang';
import { storeToRefs } from 'pinia';
import { MsgSuccess } from '@/utils/message';
import EmailDrawer from '@/views/setting/alert/setting/email/index.vue';
import WebhookDrawer from '@/views/setting/alert/setting/webhook/index.vue';
import { Alert } from '@/api/interface/alert';
import { getLicenseSmsInfo } from '@/api/modules/setting';

const globalStore = GlobalStore();
const { isMaster } = storeToRefs(globalStore);
const loading = ref(false);

const alertFormRef = ref<FormInstance>();
const phoneRef = ref();
const emailRef = ref();
const webHookRef = ref();
const sendTimeRangeRef = ref();
const sendTimeRangeValue = ref();
const sendTimeRange = ref();

const isInitialized = ref(false);
const defaultEmailConfig: Alert.EmailConfig = {
    id: undefined,
    type: 'email',
    title: 'xpack.alert.emailConfig',
    status: 'Enable',
    config: {
        displayName: '',
        sender: '',
        userName: '',
        password: '',
        host: '',
        port: 25,
        encryption: 'NONE',
        status: '',
        recipient: '',
    },
};
const emailConfig = ref<Alert.EmailConfig>({ ...defaultEmailConfig });

const defaultCommonConfig: Alert.CommonAlertConfig = {
    id: undefined,
    type: 'common',
    title: 'xpack.alert.commonConfig',
    status: 'Enable',
    config: {
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

const commonConfig = ref<Alert.CommonAlertConfig>({ ...defaultCommonConfig });

const defaultSmsConfig: Alert.SmsConfig = {
    id: undefined,
    type: 'sms',
    title: 'xpack.alert.smsConfig',
    status: 'Enable',
    config: {
        phone: '',
        alertDailyNum: 50,
    },
};
const smsConfig = ref<Alert.SmsConfig>({ ...defaultSmsConfig });

const defaultWeComConfig: Alert.WebhookConfig = {
    id: undefined,
    type: 'weCom',
    title: 'xpack.alert.weCom',
    status: 'Enable',
    config: {
        displayName: '',
        url: '',
    },
};
const weComConfig = ref<Alert.WebhookConfig>({ ...defaultWeComConfig });

const defaultDingTalkConfig: Alert.WebhookConfig = {
    id: undefined,
    type: 'dingTalk',
    title: 'xpack.alert.dingTalk',
    status: 'Enable',
    config: {
        displayName: '',
        url: '',
    },
};
const dingTalkConfig = ref<Alert.WebhookConfig>({ ...defaultDingTalkConfig });

const defaultFeiShuConfig: Alert.WebhookConfig = {
    id: undefined,
    type: 'feiShu',
    title: 'xpack.alert.feiShu',
    status: 'Enable',
    config: {
        displayName: '',
        url: '',
    },
};
const feiShuConfig = ref<Alert.WebhookConfig>({ ...defaultFeiShuConfig });

const weComUrlVisible = ref(false);
const dingTalkUrlVisible = ref(false);
const feiShuUrlVisible = ref(false);

const config = ref<Alert.AlertConfigInfo>({
    id: 0,
    type: '',
    title: '',
    status: '',
    config: '',
});
const licenseName = ref('-');
const totalSms = ref(0);
const usedSms = ref(0);
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

        const weComFound = res.data.find((s: any) => s.type === 'weCom');
        assignConfig(weComFound, weComConfig, defaultWeComConfig);

        const dingTalkFound = res.data.find((s: any) => s.type === 'dingTalk');
        assignConfig(dingTalkFound, dingTalkConfig, defaultDingTalkConfig);

        const feiShuFound = res.data.find((s: any) => s.type === 'feiShu');
        assignConfig(feiShuFound, feiShuConfig, defaultFeiShuConfig);
        isInitialized.value = true;
    } finally {
        loading.value = false;
    }
};

const onChangePhone = (id: any) => {
    phoneRef.value.acceptParams({
        id: id,
        phone: smsConfig.value.config.phone,
        dailyAlertNum: smsConfig.value.config.alertDailyNum,
    });
};

const onChangeCommon = (id: any) => {
    sendTimeRangeRef.value.acceptParams({
        id: id,
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

const getSmsInfo = async () => {
    const res = await getLicenseSmsInfo();
    licenseName.value = res.data.licenseName;
    usedSms.value = res.data.smsUsed;
    totalSms.value = res.data.smsTotal;
};

const goBuy = async () => {
    const uri = licenseName.value === '-' ? '' : `${licenseName.value}/buy-sms`;
    window.open('https://www.lxware.cn/uc/cloud/licenses/' + uri, '_blank', 'noopener,noreferrer');
};

const onChangeWeCom = (id: number) => {
    webHookRef.value.acceptParams({
        id: id,
        config: weComConfig.value.config,
        type: 'weCom',
        title: weComConfig.value.title,
    });
};

const onChangeDingTalk = (id: number) => {
    webHookRef.value.acceptParams({
        id: id,
        config: dingTalkConfig.value.config,
        type: 'dingTalk',
        title: dingTalkConfig.value.title,
    });
};

const onChangeFeiShu = (id: number) => {
    webHookRef.value.acceptParams({
        id: id,
        config: feiShuConfig.value.config,
        type: 'feiShu',
        title: feiShuConfig.value.title,
    });
};

onMounted(async () => {
    await search();
    if (globalStore.isProductPro && !globalStore.isIntl) {
        await getSmsInfo();
    }
});
</script>
<style scoped lang="scss">
.app {
    .el-card {
        padding: 0 !important;
        border: var(--panel-border) !important;

        &:hover {
            border: 1px solid var(--el-color-primary) !important;
        }
    }
}
.label {
    color: var(--el-text-color-placeholder);
}
.config-form {
    .el-form-item {
        margin-bottom: 0 !important;
    }
    height: 257px;
}
.webhook-field {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    max-width: 100%;
}
.webhook-text {
    max-width: 100%;
    word-break: break-all;
    white-space: normal;
}
.webhook-icon {
    cursor: pointer;
    color: var(--el-text-color-secondary);
}
.webhook-icon:hover {
    color: var(--el-color-primary);
}
</style>
