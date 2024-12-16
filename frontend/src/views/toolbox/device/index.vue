<template>
    <div v-loading="loading">
        <LayoutContent :title="$t('toolbox.device.toolbox')" :divider="true">
            <template #main>
                <el-row style="margin-top: 20px">
                    <el-col :span="1"><br /></el-col>
                    <el-col :xs="24" :sm="20" :md="20" :lg="10" :xl="10">
                        <el-form :model="form" label-position="left" ref="formRef" label-width="130px">
                            <el-form-item label="DNS" prop="dnsItem">
                                <el-input disabled v-model="form.dnsItem">
                                    <template #append>
                                        <el-button @click="onChangeDNS" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item label="Hosts" prop="hosts">
                                <el-input disabled v-model="form.hostItem">
                                    <template #append>
                                        <el-button @click="onChangeHost" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item label="Swap" prop="swap">
                                <el-input disabled v-model="form.swapItem">
                                    <template #append>
                                        <el-button @click="onChangeSwap" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('toolbox.device.hostname')" prop="hostname">
                                <el-input disabled v-model="form.hostname">
                                    <template #append>
                                        <el-button @click="onChangeHostname" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('toolbox.device.passwd')" prop="passwd">
                                <el-input disabled v-model="form.passwd" type="password">
                                    <template #append>
                                        <el-button @click="onChangePasswd" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('toolbox.device.syncSite')" prop="ntp">
                                <el-input disabled v-model="form.ntp">
                                    <template #append>
                                        <el-button @click="onChangeNtp" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('toolbox.device.timeZone')" prop="timeZone">
                                <el-input disabled v-model="form.timeZone">
                                    <template #append>
                                        <el-button @click="onChangeTimeZone" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('toolbox.device.localTime')" prop="localTime">
                                <el-input disabled v-model="form.localTime">
                                    <template #append>
                                        <el-button @click="onChangeLocalTime" icon="Refresh" width="150px">
                                            {{ $t('commons.button.sync') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                        </el-form>
                    </el-col>
                </el-row>
            </template>
        </LayoutContent>

        <Swap ref="swapRef" @search="search" />
        <Passwd ref="passwdRef" @search="search" />
        <TimeZone ref="timeZoneRef" @search="search" />
        <Ntp ref="ntpRef" @search="search" />
        <DNS ref="dnsRef" @search="search" />
        <Hostname ref="hostnameRef" @search="search" />
        <Hosts ref="hostsRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import Swap from '@/views/toolbox/device/swap/index.vue';
import Passwd from '@/views/toolbox/device/passwd/index.vue';
import TimeZone from '@/views/toolbox/device/time-zone/index.vue';
import Ntp from '@/views/toolbox/device/ntp/index.vue';
import DNS from '@/views/toolbox/device/dns/index.vue';
import Hostname from '@/views/toolbox/device/hostname/index.vue';
import Hosts from '@/views/toolbox/device/hosts/index.vue';
import { getDeviceBase, updateDevice } from '@/api/modules/toolbox';
import i18n from '@/lang';
import { computeSize } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const swapRef = ref();
const timeZoneRef = ref();
const ntpRef = ref();
const passwdRef = ref();
const dnsRef = ref();
const hostnameRef = ref();
const hostsRef = ref();

const form = reactive({
    dns: [],
    dnsItem: '',
    hosts: [],
    hostItem: '',
    hostname: '',
    user: '',
    passwd: '******',
    timeZone: '',
    localTime: '',
    ntp: '',

    swapItem: '',
});

const onChangeTimeZone = () => {
    timeZoneRef.value.acceptParams({ timeZone: form.timeZone });
};
const onChangeNtp = () => {
    ntpRef.value.acceptParams({ ntpSite: form.ntp });
};
const onChangeLocalTime = async () => {
    loading.value = true;
    await updateDevice('LocalTime', '')
        .then(() => {
            loading.value = false;
            search();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};
const onChangePasswd = () => {
    passwdRef.value.acceptParams({ user: form.user });
};
const onChangeDNS = () => {
    dnsRef.value.acceptParams({ dns: form.dns });
};
const onChangeHostname = () => {
    hostnameRef.value.acceptParams({ hostname: form.hostname });
};
const onChangeHost = () => {
    hostsRef.value.acceptParams({ hosts: form.hosts });
};
const onChangeSwap = () => {
    swapRef.value.acceptParams();
};

const search = async () => {
    const res = await getDeviceBase();
    form.timeZone = res.data.timeZone;
    form.localTime = res.data.localTime;
    form.hostname = res.data.hostname;
    form.ntp = res.data.ntp;
    form.user = res.data.user;
    form.dns = res.data.dns || [];
    form.dnsItem = form.dns ? i18n.global.t('toolbox.device.dnsHelper') : i18n.global.t('setting.unSetting');
    form.hosts = res.data.hosts || [];
    form.hostItem = form.hosts ? i18n.global.t('toolbox.device.hostsHelper') : i18n.global.t('setting.unSetting');

    form.swapItem = res.data.swapMemoryTotal
        ? computeSize(res.data.swapMemoryTotal)
        : i18n.global.t('setting.unSetting');
};

onMounted(() => {
    search();
});
</script>
