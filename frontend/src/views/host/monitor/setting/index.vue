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
                                    active-value="Enable"
                                    inactive-value="Disable"
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
                            <el-form-item :label="$t('monitor.interval')" prop="monitorIntervalItem">
                                <el-input disabled v-model="form.monitorIntervalItem">
                                    <template #append>
                                        <el-button @click="onChangeInterval" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.defaultNetwork')">
                                <el-input disabled v-model="form.defaultNetwork">
                                    <template #append>
                                        <el-button @click="onChangeNetwork" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('monitor.defaultNetworkHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.defaultIO')">
                                <el-input disabled v-model="form.defaultIO">
                                    <template #append>
                                        <el-button @click="onChangeIO" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('monitor.defaultIOHelper') }}</span>
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
        <Network ref="networkRef" @search="search()" />
        <IO ref="ioRef" @search="search()" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { ElMessageBox, FormInstance } from 'element-plus';
import { cleanMonitors, loadMonitorSetting, updateMonitorSetting } from '@/api/modules/host';
import MonitorRouter from '@/views/host/monitor/index.vue';
import Interval from '@/views/host/monitor/setting/interval/index.vue';
import StoreDays from '@/views/host/monitor/setting/days/index.vue';
import Network from '@/views/host/monitor/setting/default-network/index.vue';
import IO from '@/views/host/monitor/setting/default-io/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { splitTimeFromSecond, transTimeUnit } from '@/utils/util';

const loading = ref();
const form = reactive({
    monitorStatus: 'Disable',
    monitorStoreDays: 30,
    monitorInterval: 300,
    timeItem: 5,
    timeUnit: 'm',
    monitorIntervalItem: '',
    defaultNetwork: '',
    defaultIO: '',
});
const panelFormRef = ref<FormInstance>();

const intervalRef = ref();
const daysRef = ref();
const networkRef = ref();
const ioRef = ref();

const search = async () => {
    const res = await loadMonitorSetting();
    form.monitorStatus = res.data.monitorStatus;
    form.monitorInterval = Number(res.data.monitorInterval);
    let item = splitTimeFromSecond(form.monitorInterval);
    form.timeItem = item.timeItem;
    form.timeUnit = item.timeUnit;
    form.monitorIntervalItem = transTimeUnit(form.timeItem + form.timeUnit);

    form.monitorStoreDays = Number(res.data.monitorStoreDays);
    form.defaultNetwork =
        res.data.defaultNetwork === 'all' ? i18n.global.t('commons.table.all') : res.data.defaultNetwork;
    form.defaultIO = res.data.defaultIO === 'all' ? i18n.global.t('commons.table.all') : res.data.defaultIO;
};

const onSaveStatus = async () => {
    loading.value = true;
    await updateMonitorSetting('MonitorStatus', form.monitorStatus)
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
    intervalRef.value.acceptParams({ timeItem: form.timeItem, timeUnit: form.timeUnit });
};
const onChangeNetwork = () => {
    networkRef.value.acceptParams({ defaultNetwork: form.defaultNetwork });
};
const onChangeIO = () => {
    ioRef.value.acceptParams({ defaultIO: form.defaultIO });
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
});
</script>
