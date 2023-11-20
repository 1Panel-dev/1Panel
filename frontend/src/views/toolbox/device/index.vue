<template>
    <div v-loading="loading">
        <LayoutContent :title="$t('toolbox.device.toolbox')" :divider="true">
            <template #main>
                <el-row style="margin-top: 20px">
                    <el-col :span="1"><br /></el-col>
                    <el-col :xs="24" :sm="20" :md="20" :lg="10" :xl="10">
                        <el-form :model="form" label-position="left" ref="formRef" label-width="120px">
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
                                <el-input disabled v-model="form.passwd">
                                    <template #append>
                                        <el-button @click="onChangePasswd" icon="Setting">
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
                                        <el-button @click="onChangeLocalTime" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                        </el-form>
                    </el-col>
                </el-row>
            </template>
        </LayoutContent>

        <Passwd ref="passwdRef" @search="search" />
        <TimeZone ref="timeZoneRef" @search="search" />
        <LocalTime ref="localTimeRef" @search="search" />
        <DNS ref="dnsRef" @search="search" />
        <Hostname ref="hostnameRef" @search="search" />
        <Hosts ref="hostsRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import Passwd from '@/views/toolbox/device/passwd/index.vue';
import TimeZone from '@/views/toolbox/device/time-zone/index.vue';
import LocalTime from '@/views/toolbox/device/local-time/index.vue';
import DNS from '@/views/toolbox/device/dns/index.vue';
import Hostname from '@/views/toolbox/device/hostname/index.vue';
import Hosts from '@/views/toolbox/device/hosts/index.vue';
import { getDeviceBase } from '@/api/modules/toolbox';
import i18n from '@/lang';

const loading = ref(false);

const timeZoneRef = ref();
const localTimeRef = ref();
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
});

const onChangeTimeZone = () => {
    timeZoneRef.value.acceptParams({ timeZone: form.timeZone });
};
const onChangeLocalTime = () => {
    localTimeRef.value.acceptParams({ localTime: form.localTime, ntpSite: form.ntp });
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
};

onMounted(() => {
    search();
});
</script>
