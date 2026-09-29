<template>
    <div>
        <MonitorRouter />

        <LayoutContent v-loading="loading" :title="$t('menu.settings', 2)">
            <template #main>
                <el-form :model="form" @submit.prevent label-position="left" label-width="160px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="12">
                            <h3 class="mb-4">{{ $t('monitor.hostMonitor') }}</h3>
                            <el-form-item :label="$t('monitor.enableMonitor')" prop="monitorStatus">
                                <el-switch
                                    v-permission
                                    v-node-admin
                                    @change="onSaveStatus('MonitorStatus', form.monitorStatus)"
                                    v-model="form.monitorStatus"
                                    active-value="Enable"
                                    inactive-value="Disable"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('monitor.storeDays')" prop="monitorStoreDays">
                                <el-input disabled v-model="form.monitorStoreDays">
                                    <template #append>
                                        <el-button
                                            v-permission
                                            v-node-admin
                                            @click="onChangeStoreDays('MonitorStoreDays', form.monitorStoreDays)"
                                            icon="Setting"
                                        >
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.interval')" prop="monitorIntervalItem">
                                <el-input disabled v-model="form.monitorIntervalItem">
                                    <template #append>
                                        <el-button
                                            v-permission
                                            v-node-admin
                                            @click="onChangeInterval('MonitorInterval', form.monitorInterval)"
                                            icon="Setting"
                                        >
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.defaultNetwork')">
                                <el-input disabled v-model="form.defaultNetwork">
                                    <template #append>
                                        <el-button v-permission v-node-admin @click="onChangeNetwork" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('monitor.defaultNetworkHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.defaultIO')">
                                <el-input disabled v-model="form.defaultIO">
                                    <template #append>
                                        <el-button v-permission v-node-admin @click="onChangeIO" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">{{ $t('monitor.defaultIOHelper') }}</span>
                            </el-form-item>
                            <el-form-item>
                                <el-button v-permission v-node-admin @click="onClean('host')" icon="Delete">
                                    {{ $t('monitor.cleanMonitor') }}
                                </el-button>
                            </el-form-item>
                            <el-divider />
                            <h3 class="mb-4">{{ $t('monitor.gpuMonitor') }}</h3>
                            <el-form-item :label="$t('monitor.enableMonitor')" prop="gpuMonitorStatus">
                                <el-switch
                                    v-permission
                                    v-node-admin
                                    @change="onSaveStatus('GPUMonitorStatus', form.gpuMonitorStatus)"
                                    v-model="form.gpuMonitorStatus"
                                    active-value="Enable"
                                    inactive-value="Disable"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('monitor.storeDays')" prop="gpuMonitorStoreDays">
                                <el-input disabled v-model="form.gpuMonitorStoreDays">
                                    <template #append>
                                        <el-button
                                            v-permission
                                            v-node-admin
                                            @click="onChangeStoreDays('GPUMonitorStoreDays', form.gpuMonitorStoreDays)"
                                            icon="Setting"
                                        >
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.interval')" prop="gpuMonitorIntervalItem">
                                <el-input disabled v-model="form.gpuMonitorIntervalItem">
                                    <template #append>
                                        <el-button
                                            v-permission
                                            v-node-admin
                                            @click="onChangeInterval('GPUMonitorInterval', form.gpuMonitorInterval)"
                                            icon="Setting"
                                        >
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item>
                                <el-button v-permission v-node-admin @click="onClean('gpu')" icon="Delete">
                                    {{ $t('monitor.cleanMonitor') }}
                                </el-button>
                            </el-form-item>
                            <el-divider />
                            <h3 class="mb-4">{{ $t('monitor.vllmMonitor') }}</h3>
                            <el-form-item :label="$t('monitor.enableMonitor')" prop="vllmMonitorStatus">
                                <el-switch
                                    v-permission
                                    v-node-admin
                                    @change="onSaveStatus('VLLMMonitorStatus', form.vllmMonitorStatus)"
                                    v-model="form.vllmMonitorStatus"
                                    active-value="Enable"
                                    inactive-value="Disable"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('monitor.storeDays')" prop="vllmMonitorStoreDays">
                                <el-input disabled v-model="form.vllmMonitorStoreDays">
                                    <template #append>
                                        <el-button
                                            v-permission
                                            v-node-admin
                                            @click="
                                                onChangeStoreDays('VLLMMonitorStoreDays', form.vllmMonitorStoreDays)
                                            "
                                            icon="Setting"
                                        >
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.interval')" prop="vllmMonitorIntervalItem">
                                <el-input disabled v-model="form.vllmMonitorIntervalItem">
                                    <template #append>
                                        <el-button
                                            v-permission
                                            v-node-admin
                                            @click="onChangeInterval('VLLMMonitorInterval', form.vllmMonitorInterval)"
                                            icon="Setting"
                                        >
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
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
import { ElMessageBox } from 'element-plus';
import { cleanMonitors, loadMonitorSetting, updateMonitorSetting } from '@/api/modules/host';
import MonitorRouter from '@/views/host/monitor/index.vue';
import Interval from '@/views/host/monitor/setting/interval/index.vue';
import StoreDays from '@/views/host/monitor/setting/days/index.vue';
import Network from '@/views/host/monitor/setting/default-network/index.vue';
import IO from '@/views/host/monitor/setting/default-io/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { splitTimeFromSecond, transTimeUnit } from '@/utils/validate';
const loading = ref();
const form = reactive({
    gpuMonitorStatus: 'Disable',
    gpuMonitorStoreDays: 30,
    gpuMonitorInterval: 300,
    gpuMonitorIntervalItem: '',
    vllmMonitorStatus: 'Disable',
    vllmMonitorStoreDays: 30,
    vllmMonitorInterval: 300,
    vllmMonitorIntervalItem: '',
    monitorStatus: 'Disable',
    monitorStoreDays: 30,
    monitorInterval: 300,
    monitorIntervalItem: '',
    defaultNetwork: '',
    defaultIO: '',
});

const intervalRef = ref();
const daysRef = ref();
const networkRef = ref();
const ioRef = ref();

const search = async () => {
    const res = await loadMonitorSetting();
    form.gpuMonitorStatus = res.data.gpuMonitorStatus;
    form.gpuMonitorStoreDays = Number(res.data.gpuMonitorStoreDays);
    form.gpuMonitorInterval = Number(res.data.gpuMonitorInterval);
    const gpuInterval = splitTimeFromSecond(form.gpuMonitorInterval);
    form.gpuMonitorIntervalItem = transTimeUnit(gpuInterval.timeItem + gpuInterval.timeUnit);
    form.vllmMonitorStatus = res.data.vllmMonitorStatus;
    form.vllmMonitorStoreDays = Number(res.data.vllmMonitorStoreDays);
    form.vllmMonitorInterval = Number(res.data.vllmMonitorInterval);
    const vllmInterval = splitTimeFromSecond(form.vllmMonitorInterval);
    form.vllmMonitorIntervalItem = transTimeUnit(vllmInterval.timeItem + vllmInterval.timeUnit);
    form.monitorStatus = res.data.monitorStatus;
    form.monitorInterval = Number(res.data.monitorInterval);
    const item = splitTimeFromSecond(form.monitorInterval);
    form.monitorIntervalItem = transTimeUnit(item.timeItem + item.timeUnit);

    form.monitorStoreDays = Number(res.data.monitorStoreDays);
    form.defaultNetwork =
        res.data.defaultNetwork === 'all' ? i18n.global.t('commons.table.all') : res.data.defaultNetwork;
    form.defaultIO = res.data.defaultIO === 'all' ? i18n.global.t('commons.table.all') : res.data.defaultIO;
};

const onSaveStatus = async (key: 'MonitorStatus' | 'GPUMonitorStatus' | 'VLLMMonitorStatus', value: string) => {
    loading.value = true;
    await updateMonitorSetting(key, value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(async () => {
            loading.value = false;
            await search();
        });
};

const onChangeStoreDays = (key: 'MonitorStoreDays' | 'GPUMonitorStoreDays' | 'VLLMMonitorStoreDays', value: number) => {
    daysRef.value.acceptParams({ key, monitorStoreDays: value });
};
const onChangeInterval = (key: 'MonitorInterval' | 'GPUMonitorInterval' | 'VLLMMonitorInterval', value: number) => {
    const interval = splitTimeFromSecond(value);
    intervalRef.value.acceptParams({ key, timeItem: interval.timeItem, timeUnit: interval.timeUnit });
};
const onChangeNetwork = () => {
    networkRef.value.acceptParams({ defaultNetwork: form.defaultNetwork });
};
const onChangeIO = () => {
    ioRef.value.acceptParams({ defaultIO: form.defaultIO });
};

const onClean = async (type: 'host' | 'gpu') => {
    const name = i18n.global.t(type === 'host' ? 'monitor.hostMonitor' : 'monitor.gpuMonitor');
    ElMessageBox.confirm(i18n.global.t('monitor.cleanHelper', [name]), i18n.global.t('monitor.cleanMonitor'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await cleanMonitors(type)
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
