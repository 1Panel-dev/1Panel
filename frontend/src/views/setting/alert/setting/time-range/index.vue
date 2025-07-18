<template>
    <DrawerPro v-model="drawerVisible" :header="$t('xpack.alert.sendTimeRange')" @close="handleClose" size="736">
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item
                :label="$t('xpack.alert.dailyAlertNum')"
                :rules="[Rules.integerNumber, checkNumberRange(20, 100)]"
                prop="dailyAlertNum"
            >
                <el-input clearable v-model.number="form.dailyAlertNum" min="20" max="100" />
                <span class="input-help">{{ $t('xpack.alert.dailyAlertNumHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('xpack.alert.sendTimeRange')" prop="sendTimeRange">
                <div class="text-center">
                    <el-transfer
                        v-model="resourceValue"
                        class="text-left inline-block"
                        :titles="[$t('xpack.alert.noticeAlert'), $t('xpack.alert.resourceAlert')]"
                        :format="{
                            noChecked: '${total}',
                            hasChecked: '${checked}/${total}',
                        }"
                        :data="data"
                    >
                        <template #left-footer>
                            <el-form-item prop="noticeTimeRange">
                                <el-time-picker
                                    v-model="noticeTimeRange"
                                    class="!w-[235px] mx-1 mt-1"
                                    is-range
                                    :start-placeholder="$t('xpack.commons.search.timeStart')"
                                    :end-placeholder="$t('xpack.commons.search.结束时间')"
                                />
                                <span class="input-help ml-2">
                                    {{
                                        $t('xpack.alert.sendTimeRangeHelper', [
                                            $t('xpack.alert.noticeAlert').toLowerCase(),
                                        ])
                                    }}
                                </span>
                            </el-form-item>
                        </template>
                        <template #right-footer>
                            <el-form-item prop="resourceTimeRange">
                                <el-time-picker
                                    v-model="resourceTimeRange"
                                    class="!w-[235px] mx-1 mt-1"
                                    is-range
                                    :start-placeholder="$t('xpack.commons.search.timeStart')"
                                    :end-placeholder="$t('xpack.commons.search.结束时间')"
                                />
                                <span class="input-help ml-2">
                                    {{
                                        $t('xpack.alert.sendTimeRangeHelper', [
                                            $t('xpack.alert.resourceAlert').toLowerCase(),
                                        ])
                                    }}
                                </span>
                            </el-form-item>
                        </template>
                    </el-transfer>
                </div>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { computed, ComputedRef, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { UpdateAlertConfig } from '@/api/modules/alert';
import { checkNumberRange, Rules } from '@/global/form-rules';
import { Alert } from '@/api/interface/alert';

const emit = defineEmits<{ (e: 'search'): void }>();

interface Option {
    key: string;
    label: string;
    disabled: boolean;
}

interface NoticeAlert {
    sendTimeRange: string;
    type: string[];
}

interface ResourceAlert {
    sendTimeRange: string;
    type: string[];
}

interface SendTimeRange {
    noticeAlert: NoticeAlert;
    resourceAlert: ResourceAlert;
}

interface DialogProps {
    id: any;
    dailyAlertNum: number;
    sendTimeRange: SendTimeRange;
    isOffline: string;
}

interface ConfigInfo {
    alertDailyNum: number;
    alertSendTimeRange: SendTimeRange;
    isOffline: string;
}

const drawerVisible = ref(false);
const loading = ref(false);
const form = reactive({
    dailyAlertNum: 50,
});
const isOffline = ref();
const id = ref();
const configInfo = ref<ConfigInfo>({
    alertDailyNum: 0,
    isOffline: '',
    alertSendTimeRange: {
        noticeAlert: { sendTimeRange: '', type: [] },
        resourceAlert: { sendTimeRange: '', type: [] },
    },
});
const config = ref<Alert.AlertConfigInfo>({
    id: 0,
    type: '',
    title: '',
    status: '',
    config: '',
});
const resourceValue = ref(['clams', 'cronJob', 'cpu', 'memory', 'load', 'disk']);
const noticeDefaultTime: [Date, Date] = [new Date(0, 0, 1, 8, 0, 0), new Date(0, 0, 1, 23, 59, 59)];
const resourceDefaultTime: [Date, Date] = [new Date(0, 0, 1, 0, 0, 0), new Date(0, 0, 1, 23, 59, 59)];
const noticeTimeRange = ref(noticeDefaultTime);
const resourceTimeRange = ref(resourceDefaultTime);
const generateData = (): Option[] => {
    const data: Option[] = [];
    data.push({ key: 'ssl', label: i18n.global.t('xpack.alert.ssl'), disabled: false });
    data.push({ key: 'siteEndTime', label: i18n.global.t('xpack.alert.siteEndTime'), disabled: false });
    data.push({ key: 'panelPwdEndTime', label: i18n.global.t('xpack.alert.panelPwdEndTime'), disabled: false });
    data.push({ key: 'panelUpdate', label: i18n.global.t('xpack.alert.panelUpdate'), disabled: false });
    data.push({ key: 'clams', label: i18n.global.t('xpack.alert.clams'), disabled: false });
    data.push({ key: 'cronJob', label: i18n.global.t('xpack.alert.cronjob'), disabled: false });
    data.push({ key: 'cpu', label: i18n.global.t('xpack.alert.cpu'), disabled: false });
    data.push({ key: 'memory', label: i18n.global.t('xpack.alert.memory'), disabled: false });
    data.push({ key: 'load', label: i18n.global.t('xpack.alert.load'), disabled: false });
    data.push({ key: 'disk', label: i18n.global.t('xpack.alert.disk'), disabled: false });
    return data;
};

const data = ref(generateData());
const formRef = ref<FormInstance>();
const noticeValue: ComputedRef<string[]> = computed(() => {
    return data.value.filter((item) => !resourceValue.value.includes(item.key)).map((item) => item.key);
});

const acceptParams = (params: DialogProps): void => {
    if (typeof params.sendTimeRange === 'object' && params.sendTimeRange !== null) {
        noticeTimeRange.value = parseTimeRange(params.sendTimeRange.noticeAlert.sendTimeRange);
        resourceTimeRange.value = parseTimeRange(params.sendTimeRange.resourceAlert.sendTimeRange);
        resourceValue.value = params.sendTimeRange.resourceAlert.type;
    }
    form.dailyAlertNum = Number(params.dailyAlertNum);
    isOffline.value = params.isOffline;
    id.value = params.id;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (
            typeof noticeTimeRange.value === 'object' &&
            noticeTimeRange.value !== null &&
            typeof resourceTimeRange.value === 'object' &&
            resourceTimeRange.value !== null
        ) {
            loading.value = true;
            configInfo.value.alertSendTimeRange = {
                noticeAlert: { sendTimeRange: stringifyTimeRange(noticeTimeRange.value), type: noticeValue.value },
                resourceAlert: {
                    sendTimeRange: stringifyTimeRange(resourceTimeRange.value),
                    type: resourceValue.value,
                },
            };
            configInfo.value.isOffline = isOffline.value;
            configInfo.value.alertDailyNum = form.dailyAlertNum;
            try {
                config.value.id = id.value;
                config.value.type = 'common';
                config.value.title = 'xpack.alert.commonConfig';
                config.value.status = 'Enable';
                config.value.config = JSON.stringify(configInfo.value);
                await UpdateAlertConfig(config.value);

                loading.value = false;
                handleClose();
                emit('search');
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            } catch (error) {
                loading.value = false;
            }
        } else {
            MsgError(
                i18n.global.t('commons.msg.confirmNoNull', [i18n.global.t('xpack.alert.timeRange').toLowerCase()]),
            );
        }
    });
};

const parseTimeRange = (timeRangeStr: string): [Date, Date] => {
    return timeRangeStr.split(' - ').map((time) => {
        const [hours, minutes, seconds] = time.split(':').map(Number);
        return new Date(0, 0, 1, hours, minutes, seconds);
    }) as [Date, Date];
};

const stringifyTimeRange = (timeRange: [Date, Date]): string => {
    return timeRange
        .map(
            (date) =>
                `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}:${String(
                    date.getSeconds(),
                ).padStart(2, '0')}`,
        )
        .join(' - ');
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
:deep(.el-transfer) {
    --el-transfer-panel-width: 265px;
}

:deep(.el-transfer__buttons) {
    padding: 10px 30px;
    @media (max-width: 600px) {
        width: 265px;
        text-align: center;
        padding: 10px 0;
        .el-button [class*='el-icon'] svg {
            transform: rotate(90deg);
        }
    }

    @media (min-width: 601px) {
        display: inline-flex;
        flex-direction: column;
        align-items: center;
        gap: 10px;
        width: 40px;
        height: 40px;
        justify-content: center;
        .el-button + .el-button {
            margin-left: 0;
        }
    }
}

:deep(.el-transfer-panel .el-transfer-panel__footer) {
    height: 65px;
}
</style>
