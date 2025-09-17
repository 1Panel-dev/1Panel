<template>
    <div v-loading="loading">
        <LayoutContent :title="$t('setting.panel')" :divider="true">
            <template #main>
                <el-form :model="form" :label-position="mobile ? 'top' : 'left'" label-width="150px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :xs="24" :sm="20" :md="15" :lg="12" :xl="12">
                            <el-form-item :label="$t('setting.user')" prop="userName">
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
                                <div class="flex justify-center items-center sm:gap-6 gap-2">
                                    <div class="sm:contents hidden">
                                        <el-radio-group @change="onSave('Theme', form.theme)" v-model="form.theme">
                                            <el-radio-button value="light">
                                                <span>{{ $t('setting.light') }}</span>
                                            </el-radio-button>
                                            <el-radio-button value="dark">
                                                <span>{{ $t('setting.dark') }}</span>
                                            </el-radio-button>
                                            <el-radio-button value="auto">
                                                <span>{{ $t('setting.auto') }}</span>
                                            </el-radio-button>
                                        </el-radio-group>
                                    </div>
                                    <div class="sm:hidden block w-32 !h-[33.5px]">
                                        <el-select @change="onSave('Theme', form.theme)" v-model="form.theme">
                                            <el-option key="light" value="light" :label="$t('setting.light')">
                                                {{ $t('setting.light') }}
                                            </el-option>
                                            <el-option key="dark" value="dark" :label="$t('setting.dark')">
                                                {{ $t('setting.dark') }}
                                            </el-option>
                                            <el-option key="auto" value="auto" :label="$t('setting.auto')">
                                                {{ $t('setting.auto') }}
                                            </el-option>
                                        </el-select>
                                    </div>
                                    <div>
                                        <el-button
                                            v-if="isMasterProductPro"
                                            @click="onChangeThemeColor"
                                            icon="Setting"
                                            class="!h-[32px] sm:!h-[33.5px]"
                                        >
                                            <span>{{ $t('container.custom') }}</span>
                                        </el-button>
                                    </div>
                                </div>
                            </el-form-item>

                            <el-form-item :label="$t('setting.menuTabs')" prop="menuTabs">
                                <el-radio-group @change="onSave('MenuTabs', form.menuTabs)" v-model="form.menuTabs">
                                    <el-radio-button value="Enable">
                                        <span>{{ $t('commons.button.enable') }}</span>
                                    </el-radio-button>
                                    <el-radio-button value="Disable">
                                        <span>{{ $t('commons.button.disable') }}</span>
                                    </el-radio-button>
                                </el-radio-group>
                            </el-form-item>

                            <el-form-item :label="$t('setting.watermark')" v-if="isMasterProductPro" prop="watermark">
                                <el-radio-group class="w-full" @change="onChangeWatermark" v-model="form.watermarkItem">
                                    <el-radio-button value="Enable">
                                        <span>{{ $t('commons.button.enable') }}</span>
                                    </el-radio-button>
                                    <el-radio-button value="Disable">
                                        <span>{{ $t('commons.button.disable') }}</span>
                                    </el-radio-button>
                                </el-radio-group>
                                <div v-if="form.watermarkItem === 'Enable'">
                                    <div>
                                        <el-button link type="primary" @click="onChangeWatermark">
                                            {{ $t('commons.button.view') }}
                                        </el-button>
                                    </div>
                                </div>
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
                                <el-select
                                    class="sm:!w-1/2 !w-full"
                                    @change="onSave('Language', form.language)"
                                    v-model="form.language"
                                >
                                    <el-option
                                        v-for="option in languageOptions"
                                        :key="option.value"
                                        :value="option.value"
                                        :label="option.label"
                                    >
                                        {{ option.label }}
                                    </el-option>
                                </el-select>
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
                                <span class="input-help">{{ $t('setting.systemIPHelper') }}</span>
                            </el-form-item>

                            <el-form-item :label="$t('setting.proxy')" prop="proxyShow" v-if="isMaster">
                                <el-input disabled v-model="form.proxyShow">
                                    <template #append>
                                        <el-button @click="onChangeProxy" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>

                            <el-form-item :label="$t('setting.apiInterface')" prop="apiInterface" v-if="isMaster">
                                <el-switch
                                    @change="onChangeApiInterfaceStatus"
                                    v-model="form.apiInterfaceStatus"
                                    active-value="Enable"
                                    inactive-value="Disable"
                                />
                                <span class="input-help">{{ $t('setting.apiInterfaceHelper') }}</span>
                                <div v-if="form.apiInterfaceStatus === 'Enable'">
                                    <div>
                                        <el-button link type="primary" @click="onChangeApiInterfaceStatus">
                                            {{ $t('commons.button.view') }}
                                        </el-button>
                                    </div>
                                </div>
                            </el-form-item>

                            <el-form-item :label="$t('setting.developerMode')" prop="developerMode">
                                <el-radio-group
                                    @change="onSave('DeveloperMode', form.developerMode)"
                                    v-model="form.developerMode"
                                >
                                    <el-radio-button value="Enable">
                                        <span>{{ $t('commons.button.enable') }}</span>
                                    </el-radio-button>
                                    <el-radio-button value="Disable">
                                        <span>{{ $t('commons.button.disable') }}</span>
                                    </el-radio-button>
                                </el-radio-group>
                                <span class="input-help">{{ $t('setting.developerModeHelper') }}</span>
                            </el-form-item>

                            <el-form-item :label="$t('setting.menuSetting')">
                                <el-button v-show="!show" @click="onChangeHideMenus" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
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
        <Proxy ref="proxyRef" @search="search()" />
        <ApiInterface ref="apiInterfaceRef" @search="search()" />
        <Timeout ref="timeoutRef" @search="search()" />
        <HideMenu ref="hideMenuRef" @search="search()" />
        <ThemeColor ref="themeColorRef" />
        <Watermark ref="watermarkRef" @search="search()" />
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElForm, ElMessageBox } from 'element-plus';
import {
    getSettingInfo,
    updateSetting,
    getSystemAvailable,
    updateApiConfig,
    getAgentSettingInfo,
} from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { useTheme } from '@/global/use-theme';
import { MsgSuccess } from '@/utils/message';
import ThemeColor from '@/views/setting/panel/theme-color/index.vue';
import ApiInterface from '@/views/setting/panel/api-interface/index.vue';
import Password from '@/views/setting/panel/password/index.vue';
import Watermark from '@/views/setting/panel/watermark/index.vue';
import UserName from '@/views/setting/panel/username/index.vue';
import Timeout from '@/views/setting/panel/timeout/index.vue';
import PanelName from '@/views/setting/panel/name/index.vue';
import SystemIP from '@/views/setting/panel/systemip/index.vue';
import Proxy from '@/views/setting/panel/proxy/index.vue';
import HideMenu from '@/views/setting/panel/hidemenu/index.vue';
import { storeToRefs } from 'pinia';
import { getXpackSetting, updateXpackSettingByKey } from '@/utils/xpack';
import { setPrimaryColor } from '@/utils/theme';
import i18n from '@/lang';

const loading = ref(false);
const globalStore = GlobalStore();

const { isMasterProductPro, isMaster } = storeToRefs(globalStore);

const { switchTheme } = useTheme();
const mobile = computed(() => {
    return globalStore.isMobile();
});

interface ThemeColor {
    light: string;
    dark: string;
    themePredefineColors: {
        light: string[];
        dark: string[];
    };
}

const form = reactive({
    userName: '',
    password: '',
    sessionTimeout: 0,
    panelName: '',
    theme: '',
    watermark: '',
    watermarkItem: '',
    themeColor: {} as ThemeColor,
    menuTabs: '',
    language: '',
    complexityVerification: '',
    developerMode: '',
    systemIP: '',

    proxyShow: '',
    proxyUrl: '',
    proxyType: '',
    proxyPort: '',
    proxyUser: '',
    proxyPasswd: '',
    proxyPasswdKeep: '',
    proxyDocker: '',

    apiInterfaceStatus: 'Disable',
    apiKey: '',
    ipWhiteList: '',
    apiKeyValidityTime: 120,

    hideMenu: '',
});

const show = ref();

const userNameRef = ref();
const passwordRef = ref();
const panelNameRef = ref();
const systemIPRef = ref();
const proxyRef = ref();
const timeoutRef = ref();
const hideMenuRef = ref();
const watermarkRef = ref();
const themeColorRef = ref();
const apiInterfaceRef = ref();
const unset = ref(i18n.global.t('setting.unSetting'));

const languageOptions = ref([
    { value: 'zh', label: '中文(简体)' },
    { value: 'zh-Hant', label: '中文(繁體)' },
    ...(!globalStore.isIntl ? [{ value: 'en', label: 'English' }] : []),
    { value: 'ja', label: '日本語' },
    { value: 'pt-BR', label: 'Português (Brasil)' },
    { value: 'ko', label: '한국어' },
    { value: 'ru', label: 'Русский' },
    { value: 'ms', label: 'Bahasa Melayu' },
    { value: 'tr', label: 'Turkish' },
    { value: 'es-ES', label: 'España - Español' },
]);

if (globalStore.isIntl) {
    languageOptions.value.unshift({ value: 'en', label: 'English' });
}

const search = async () => {
    const agentRes = await getAgentSettingInfo();
    form.systemIP = agentRes.data.systemIP;

    const res = await getSettingInfo();
    form.userName = res.data.userName;
    form.password = '******';
    form.theme = res.data.theme;
    form.menuTabs = res.data.menuTabs;
    form.panelName = res.data.panelName;
    form.language = res.data.language;
    form.sessionTimeout = Number(res.data.sessionTimeout);

    form.proxyUrl = res.data.proxyUrl;
    form.proxyType = res.data.proxyType;
    form.proxyPort = res.data.proxyPort;
    form.proxyShow = form.proxyUrl ? form.proxyUrl + ':' + form.proxyPort : unset.value;
    form.proxyUser = res.data.proxyUser;
    form.proxyPasswd = res.data.proxyPasswd;
    form.proxyPasswdKeep = res.data.proxyPasswdKeep;

    form.apiInterfaceStatus = res.data.apiInterfaceStatus;
    form.apiKey = res.data.apiKey;
    form.ipWhiteList = res.data.ipWhiteList;
    form.apiKeyValidityTime = res.data.apiKeyValidityTime;

    form.developerMode = res.data.developerMode;
    form.hideMenu = res.data.hideMenu;

    form.complexityVerification = res.data.complexityVerification;

    if (isMasterProductPro.value) {
        const xpackRes = await getXpackSetting();
        if (xpackRes) {
            form.theme = xpackRes.data.theme || globalStore.themeConfig.theme || 'light';
            form.themeColor = JSON.parse(xpackRes.data.themeColor || '{"light":"#005eeb","dark":"#F0BE96"}');
            globalStore.themeConfig.themeColor = xpackRes.data.themeColor
                ? xpackRes.data.themeColor
                : '{"light":"#005eeb","dark":"#F0BE96"}';
            globalStore.themeConfig.theme = form.theme;
            form.proxyDocker = xpackRes.data.proxyDocker;
            form.watermark = xpackRes.data.watermark;
            form.watermarkItem = xpackRes.data.watermark ? 'Enable' : 'Disable';
        }
    } else {
        globalStore.themeConfig.theme = form.theme;
    }
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
const onChangeProxy = () => {
    proxyRef.value.acceptParams({
        url: form.proxyUrl,
        type: form.proxyType,
        port: form.proxyPort,
        user: form.proxyUser,
        passwd: form.proxyPasswd,
        passwdKeep: form.proxyPasswdKeep,
        proxyDocker: form.proxyDocker,
    });
};

const onChangeHideMenus = () => {
    hideMenuRef.value.acceptParams({ hideMenu: form.hideMenu });
};

const onChangeThemeColor = () => {
    const themeColor: ThemeColor = JSON.parse(globalStore.themeConfig.themeColor);
    themeColorRef.value.acceptParams({ themeColor: themeColor, theme: globalStore.themeConfig.theme });
};

const onChangeWatermark = async () => {
    if (form.watermarkItem === 'Enable') {
        watermarkRef.value.acceptParams(form.watermark);
        return;
    }
    ElMessageBox.confirm(i18n.global.t('setting.watermarkCloseHelper'), i18n.global.t('setting.watermark'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            loading.value = true;
            await updateXpackSettingByKey('Watermark', '')
                .then(() => {
                    loading.value = false;
                    globalStore.watermark = null;
                    search();
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            form.watermarkItem = 'Enable';
        });
};

const onChangeApiInterfaceStatus = async () => {
    if (form.apiInterfaceStatus === 'Enable') {
        apiInterfaceRef.value.acceptParams({
            apiInterfaceStatus: form.apiInterfaceStatus,
            apiKey: form.apiKey,
            ipWhiteList: form.ipWhiteList,
            apiKeyValidityTime: form.apiKeyValidityTime,
        });
        return;
    }
    ElMessageBox.confirm(i18n.global.t('setting.apiInterfaceClose'), i18n.global.t('setting.apiInterface'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            loading.value = true;
            form.apiInterfaceStatus = 'Disable';
            let param = {
                apiKey: form.apiKey,
                ipWhiteList: form.ipWhiteList,
                apiInterfaceStatus: form.apiInterfaceStatus,
                apiKeyValidityTime: form.apiKeyValidityTime,
            };
            await updateApiConfig(param)
                .then(() => {
                    loading.value = false;
                    search();
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            form.apiInterfaceStatus = 'Enable';
        });
};

const handleThemeChange = async (val: string) => {
    globalStore.themeConfig.theme = val;
    switchTheme();
    if (globalStore.isMasterProductPro) {
        await updateXpackSettingByKey('Theme', val);
        let color: string;
        const themeColor: ThemeColor = JSON.parse(globalStore.themeConfig.themeColor);
        if (val === 'auto') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
            color = prefersDark.matches ? themeColor.dark : themeColor.light;
        } else {
            color = val === 'dark' ? themeColor.dark : themeColor.light;
        }
        globalStore.themeConfig.primary = color;
        setPrimaryColor(color);
    }
};
const onSave = async (key: string, val: any) => {
    loading.value = true;
    let param = {
        key: key,
        value: val + '',
    };
    await updateSetting(param)
        .then(() => {
            if (key === 'Language') {
                i18n.global.locale.value = val;
                globalStore.updateLanguage(val);
                location.reload();
            }
            if (key === 'Theme') {
                handleThemeChange(val);
            }
            if (key === 'MenuTabs') {
                globalStore.setOpenMenuTabs(val === 'Enable');
            }
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
            loading.value = false;
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

<style scoped lang="scss">
:deep(.el-radio-group) {
    min-width: max-content;
}
</style>
