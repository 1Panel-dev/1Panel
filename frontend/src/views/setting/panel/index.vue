<template>
    <div v-loading="loading">
        <LayoutContent :title="$t('setting.panel')" :divider="true">
            <template #main>
                <el-form :model="form" label-position="left" label-width="150px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :xs="24" :sm="20" :md="15" :lg="12" :xl="12">
                            <el-form-item :label="$t('commons.login.username')" prop="userName">
                                <el-input disabled v-model="form.userName">
                                    <template #append>
                                        <el-button @click="onChangeUserName()" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>

                            <el-form-item :label="$t('setting.passwd')" prop="password">
                                <el-input type="password" disabled v-model="form.password">
                                    <template #append>
                                        <el-button icon="Setting" @click="onChangePassword">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>

                            <el-form-item :label="$t('setting.theme')" prop="theme">
                                <el-radio-group @change="onSave('Theme', form.theme)" v-model="form.theme">
                                    <el-radio-button label="light">
                                        <el-icon><Sunny /></el-icon>
                                        {{ $t('setting.light') }}
                                    </el-radio-button>
                                    <el-radio-button label="dark">
                                        <el-icon><Moon /></el-icon>
                                        {{ $t('setting.dark') }}
                                    </el-radio-button>
                                </el-radio-group>
                            </el-form-item>

                            <el-form-item :label="$t('setting.title')" prop="panelName">
                                <el-input disabled v-model="form.panelName">
                                    <template #append>
                                        <el-button icon="Setting" @click="onChangeTitle">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>

                            <el-form-item :label="$t('setting.language')" prop="language">
                                <el-radio-group
                                    style="width: 100%"
                                    @change="onSave('Language', form.language)"
                                    v-model="form.language"
                                >
                                    <el-radio label="zh">中文(简体)</el-radio>
                                    <el-radio label="tw">中文(繁體)</el-radio>
                                    <el-radio label="en">English</el-radio>
                                </el-radio-group>
                            </el-form-item>

                            <el-form-item :label="$t('setting.sessionTimeout')" prop="sessionTimeout">
                                <el-input disabled v-model.number="form.sessionTimeout">
                                    <template #append>
                                        <el-button @click="onChangeTimeout" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <span class="input-help">
                                    {{ $t('setting.sessionTimeoutHelper', [form.sessionTimeout]) }}
                                </span>
                            </el-form-item>

                            <el-form-item :label="$t('setting.timeZone')" prop="timeZone">
                                <el-input disabled v-model.number="form.timeZone">
                                    <template #append>
                                        <el-button @click="onChangeTimeZone" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('setting.systemIP')" prop="systemIP">
                                <el-input disabled v-if="form.systemIP" v-model="form.systemIP">
                                    <template #append>
                                        <el-button @click="onChangeSystemIP" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <el-input disabled v-if="!form.systemIP" v-model="unset">
                                    <template #append>
                                        <el-button @click="onChangeSystemIP" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('setting.syncTime')">
                                <el-input disabled v-model="form.localTime">
                                    <template #append>
                                        <el-button v-show="!show" @click="onChangeNtp" icon="Setting">
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

        <Password ref="passwordRef" />
        <UserName ref="userNameRef" />
        <PanelName ref="panelNameRef" @search="search()" />
        <SystemIP ref="systemIPRef" @search="search()" />
        <Timeout ref="timeoutRef" @search="search()" />
        <TimeZone ref="timezoneRef" @search="search()" />
        <Ntp ref="ntpRef" @search="search()" />
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElForm } from 'element-plus';
import { getSettingInfo, updateSetting, getSystemAvailable } from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { useI18n } from 'vue-i18n';
import { useTheme } from '@/hooks/use-theme';
import { MsgSuccess } from '@/utils/message';
import Password from '@/views/setting/panel/password/index.vue';
import UserName from '@/views/setting/panel/username/index.vue';
import Timeout from '@/views/setting/panel/timeout/index.vue';
import PanelName from '@/views/setting/panel/name/index.vue';
import SystemIP from '@/views/setting/panel/systemip/index.vue';
import TimeZone from '@/views/setting/panel/timezone/index.vue';
import Ntp from '@/views/setting/panel/ntp/index.vue';

const loading = ref(false);
const i18n = useI18n();
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);
const { switchDark } = useTheme();

const form = reactive({
    userName: '',
    password: '',
    email: '',
    sessionTimeout: 0,
    localTime: '',
    timeZone: '',
    ntpSite: '',
    panelName: '',
    systemIP: '',
    theme: '',
    language: '',
    complexityVerification: '',
});

const show = ref();

const userNameRef = ref();
const passwordRef = ref();
const panelNameRef = ref();
const systemIPRef = ref();
const timeoutRef = ref();
const ntpRef = ref();
const timezoneRef = ref();
const unset = ref(i18n.t('setting.unSetting'));

const search = async () => {
    const res = await getSettingInfo();
    form.userName = res.data.userName;
    form.password = '******';
    form.sessionTimeout = Number(res.data.sessionTimeout);
    form.localTime = res.data.localTime;
    form.timeZone = res.data.timeZone;
    form.ntpSite = res.data.ntpSite;
    form.panelName = res.data.panelName;
    form.systemIP = res.data.systemIP;
    form.theme = res.data.theme;
    form.language = res.data.language;
    form.complexityVerification = res.data.complexityVerification;
};

const onChangePassword = () => {
    passwordRef.value.acceptParams({ complexityVerification: form.complexityVerification });
};
const onChangeUserName = () => {
    userNameRef.value.acceptParams({ userName: form.userName });
};
const onChangeTitle = () => {
    panelNameRef.value.acceptParams({ panelName: form.panelName });
};
const onChangeTimeout = () => {
    timeoutRef.value.acceptParams({ sessionTimeout: form.sessionTimeout });
};
const onChangeSystemIP = () => {
    systemIPRef.value.acceptParams({ systemIP: form.systemIP });
};
const onChangeTimeZone = () => {
    timezoneRef.value.acceptParams({ timeZone: form.timeZone });
};
const onChangeNtp = () => {
    ntpRef.value.acceptParams({ localTime: form.localTime, ntpSite: form.ntpSite });
};

const onSave = async (key: string, val: any) => {
    loading.value = true;
    if (key === 'Language') {
        i18n.locale.value = val;
        globalStore.updateLanguage(val);
    }
    if (key === 'Theme') {
        globalStore.setThemeConfig({ ...themeConfig.value, theme: val });
        switchDark();
    }
    let param = {
        key: key,
        value: val + '',
    };
    await updateSetting(param)
        .then(async () => {
            if (param.key === 'Language') {
                location.reload();
            }
            loading.value = false;
            MsgSuccess(i18n.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
    getSystemAvailable();
});
</script>
