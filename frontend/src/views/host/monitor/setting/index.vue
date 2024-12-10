<template>
    <div>
        <MonitorRouter />

        <LayoutContent v-loading="loading" :title="$t('menu.settings', 2)" :divider="true">
            <template #main>
                <el-form :model="form" @submit.prevent ref="panelFormRef" label-position="left" label-width="160px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="12">
                            <el-form-item :label="$t('monitor.enableMonitor')" prop="monitorStatus">
                                <el-switch
                                    @change="onSaveStatus"
                                    v-model="form.monitorStatus"
                                    active-value="enable"
                                    inactive-value="disable"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('monitor.storeDays')" prop="monitorStoreDays">
                                <el-input disabled v-model="form.monitorStoreDays">
                                    <template #append>
                                        <el-button @click="onChangeStoreDays" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.interval')" prop="monitorInterval">
                                <el-input disabled v-model="form.monitorInterval">
                                    <template #append>
                                        <el-button @click="onChangeInterval" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item>
                                <el-button @click="onClean()" icon="Delete">{{ $t('monitor.cleanMonitor') }}</el-button>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>

        <Interval ref="intervalRef" @search="search" />
        <StoreDays ref="daysRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { ElMessageBox, FormInstance } from 'element-plus';
import { cleanMonitors } from '@/api/modules/host';
import { getSettingInfo, getSystemAvailable, updateSetting } from '@/api/modules/setting';
import MonitorRouter from '@/views/host/monitor/index.vue';
import Interval from '@/views/host/monitor/setting/interval/index.vue';
import StoreDays from '@/views/host/monitor/setting/days/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const form = reactive({
    monitorStatus: 'disable',
    monitorStoreDays: 30,
    monitorInterval: 1,
});
const panelFormRef = ref<FormInstance>();

const intervalRef = ref();
const daysRef = ref();

const search = async () => {
    const res = await getSettingInfo();
    form.monitorStatus = res.data.monitorStatus;
    form.monitorInterval = Number(res.data.monitorInterval);
    form.monitorStoreDays = Number(res.data.monitorStoreDays);
};

const onSaveStatus = async () => {
    let param = {
        key: 'MonitorStatus',
        value: form.monitorStatus,
    };
    loading.value = true;
    await updateSetting(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const onChangeStoreDays = () => {
    daysRef.value.acceptParams({ monitorStoreDays: form.monitorStoreDays });
};
const onChangeInterval = () => {
    intervalRef.value.acceptParams({ monitorInterval: form.monitorInterval });
};

const onClean = async () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.clean'), i18n.global.t('monitor.cleanMonitor'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await cleanMonitors()
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

onMounted(() => {
    search();
    getSystemAvailable();
});
</script>
