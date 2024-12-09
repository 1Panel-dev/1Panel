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
                                <div class="flex justify-between items-center gap-6">
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
                                    <el-button
                                        v-if="isProductPro"
                                        @click="onChangeThemeColor"
                                        icon="Setting"
                                        class="!h-[34px]"
                                    >
                                        <span>{{ $t('container.custom') }}</span>
                                    </el-button>
                                </div>
                            </el-form-item>

                            <el-form-item :label="$t('setting.menuTabs')" prop="menuTabs">
                                <el-radio-group @change="onSave('MenuTabs', form.menuTabs)" v-model="form.menuTabs">
                                    <el-radio-button value="enable">
                                        <span>{{ $t('commons.button.enable') }}</span>
                                    </el-radio-button>
                                    <el-radio-button value="disable">
                                        <span>{{ $t('commons.button.disable') }}</span>
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
                                    <el-radio value="zh">中文(简体)</el-radio>
                                    <el-radio value="tw">中文(繁體)</el-radio>
                                    <el-radio value="en">English</el-radio>
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

                            <el-form-item :label="$t('setting.defaultNetwork')">
                                <el-input disabled v-model="form.defaultNetworkVal">
                                    <template #append>
                                        <el-button v-show="!show" @click="onChangeNetwork" icon="Setting">
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

                            <el-form-item v-if="!globalStore.isIntl" :label="$t('setting.proxy')" prop="proxyShow">
                                <el-input disabled v-model="form.proxyShow">
                                    <template #append>
                                        <el-button @click="onChangeProxy" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>

                            <el-form-item :label="$t('setting.apiInterface')" prop="apiInterface">
                                <el-switch
                                    @change="onChangeApiInterfaceStatus"
                                    v-model="form.apiInterfaceStatus"
                                    active-value="enable"
                                    inactive-value="disable"
                                />
                                <span class="input-help">{{ $t('setting.apiInterfaceHelper') }}</span>
                                <div v-if="form.apiInterfaceStatus === 'enable'">
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
                                    <el-radio-button value="enable">
                                        <span>{{ $t('commons.button.enable') }}</span>
                                    </el-radio-button>
                                    <el-radio-button value="disable">
                                        <span>{{ $t('commons.button.disable') }}</span>
                                    </el-radio-button>
                                </el-radio-group>
                                <span class="input-help">{{ $t('setting.developerModeHelper') }}</span>
                            </el-form-item>

                            <el-form-item v-if="!globalStore.isIntl" :label="$t('setting.advancedMenuHide')">
                                <el-input disabled v-model="form.proHideMenus">
                                    <template #append>
                                        <el-button v-show="!show" @click="onChangeHideMenus" icon="Setting">
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
        <Proxy ref="proxyRef" @search="search()" />
        <ApiInterface ref="apiInterfaceRef" @search="search()" />
        <Timeout ref="timeoutRef" @search="search()" />
        <Network ref="networkRef" @search="search()" />
        <HideMenu ref="hideMenuRef" @search="search()" />
        <ThemeColor ref="themeColorRef" />
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElForm, ElMessageBox } from 'element-plus';
import { getSettingInfo, updateSetting, getSystemAvailable, updateApiConfig } from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { useI18n } from 'vue-i18n';
import { useTheme } from '@/hooks/use-theme';
import { MsgSuccess } from '@/utils/message';
import Password from '@/views/setting/panel/password/index.vue';
import UserName from '@/views/setting/panel/username/index.vue';
import Timeout from '@/views/setting/panel/timeout/index.vue';
import PanelName from '@/views/setting/panel/name/index.vue';
import SystemIP from '@/views/setting/panel/systemip/index.vue';
import Proxy from '@/views/setting/panel/proxy/index.vue';
import Network from '@/views/setting/panel/default-network/index.vue';
import HideMenu from '@/views/setting/panel/hidemenu/index.vue';
import ThemeColor from '@/views/setting/panel/theme-color/index.vue';
import ApiInterface from '@/views/setting/panel/api-interface/index.vue';
import { storeToRefs } from 'pinia';
import { getXpackSetting, updateXpackSettingByKey } from '@/utils/xpack';
import { setPrimaryColor } from '@/utils/theme';

const loading = ref(false);
const i18n = useI18n();
const globalStore = GlobalStore();

const { isProductPro } = storeToRefs(globalStore);

const { switchTheme } = useTheme();

const mobile = computed(() => {
    return globalStore.isMobile();
});

interface ThemeColor {
    light: string;
    dark: string;
}

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
    themeColor: {} as ThemeColor,
    menuTabs: '',
    language: '',
    complexityVerification: '',
    defaultNetwork: '',
    defaultNetworkVal: '',
    developerMode: '',

    proxyShow: '',
    proxyUrl: '',
    proxyType: '',
    proxyPort: '',
    proxyUser: '',
    proxyPasswd: '',
    proxyPasswdKeep: '',
    proxyDocker: '',

    apiInterfaceStatus: 'disable',
    apiKey: '',
    ipWhiteList: '',

    proHideMenus: ref(i18n.t('setting.unSetting')),
    hideMenuList: '',
});

const show = ref();

const userNameRef = ref();
const passwordRef = ref();
const panelNameRef = ref();
const systemIPRef = ref();
const proxyRef = ref();
const timeoutRef = ref();
const networkRef = ref();
const hideMenuRef = ref();
const themeColorRef = ref();
const apiInterfaceRef = ref();
const unset = ref(i18n.t('setting.unSetting'));

interface Node {
    id: string;
    title: string;
    path?: string;
    label: string;
    isCheck: boolean;
    children?: Node[];
}

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
    form.menuTabs = res.data.menuTabs;
    form.language = res.data.language;
    form.complexityVerification = res.data.complexityVerification;
    form.defaultNetwork = res.data.defaultNetwork;
    form.defaultNetworkVal = res.data.defaultNetwork === 'all' ? i18n.t('commons.table.all') : res.data.defaultNetwork;
    form.proHideMenus = res.data.xpackHideMenu;
    form.hideMenuList = res.data.xpackHideMenu;
    form.developerMode = res.data.developerMode;

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

    const json: Node = JSON.parse(res.data.xpackHideMenu);
    const checkedTitles = getCheckedTitles(json);
    form.proHideMenus = checkedTitles.toString();
    if (isProductPro.value) {
        const xpackRes = await getXpackSetting();
        if (xpackRes) {
            form.theme = xpackRes.data.theme || globalStore.themeConfig.theme || 'light';
            form.themeColor = JSON.parse(xpackRes.data.themeColor);
            globalStore.themeConfig.themeColor = xpackRes.data.themeColor;
            globalStore.themeConfig.theme = form.theme;
            form.proxyDocker = xpackRes.data.proxyDocker;
        }
    } else {
        form.theme = globalStore.themeConfig.theme || res.data.theme || 'light';
    }
};

function extractTitles(node: Node, result: string[]): void {
    if (!node.isCheck && !node.children) {
        result.push(i18n.t(node.title));
    }
    if (node.children) {
        for (const childNode of node.children) {
            extractTitles(childNode, result);
        }
    }
}

function getCheckedTitles(json: Node): string[] {
    let result: string[] = [];
    extractTitles(json, result);
    if (result.length === 0) {
        result.push(i18n.t('setting.unSetting'));
    }
    if (result.length === json.children.length) {
        result = [];
        result.push(i18n.t('setting.hideALL'));
    }
    return result;
}

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

const onChangeApiInterfaceStatus = async () => {
    if (form.apiInterfaceStatus === 'enable') {
        apiInterfaceRef.value.acceptParams({
            apiInterfaceStatus: form.apiInterfaceStatus,
            apiKey: form.apiKey,
            ipWhiteList: form.ipWhiteList,
        });
        return;
    }
    ElMessageBox.confirm(i18n.t('setting.apiInterfaceClose'), i18n.t('setting.apiInterface'), {
        confirmButtonText: i18n.t('commons.button.confirm'),
        cancelButtonText: i18n.t('commons.button.cancel'),
    })
        .then(async () => {
            loading.value = true;
            form.apiInterfaceStatus = 'disable';
            let param = {
                apiKey: form.apiKey,
                ipWhiteList: form.ipWhiteList,
                apiInterfaceStatus: form.apiInterfaceStatus,
            };
            await updateApiConfig(param)
                .then(() => {
                    loading.value = false;
                    search();
                    MsgSuccess(i18n.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            form.apiInterfaceStatus = 'enable';
        });
};
const onChangeNetwork = () => {
    networkRef.value.acceptParams({ defaultNetwork: form.defaultNetwork });
};

const onChangeHideMenus = () => {
    hideMenuRef.value.acceptParams({ menuList: form.hideMenuList });
};

const onChangeThemeColor = () => {
    const themeColor: ThemeColor = JSON.parse(globalStore.themeConfig.themeColor);
    themeColorRef.value.acceptParams({ themeColor: themeColor, theme: globalStore.themeConfig.theme });
};

const onSave = async (key: string, val: any) => {
    loading.value = true;
    if (key === 'Language') {
        i18n.locale.value = val;
        globalStore.updateLanguage(val);
    }
    if (key === 'Theme') {
        globalStore.themeConfig.theme = val;
        switchTheme();
        if (globalStore.isProductPro) {
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
    }
    if (key === 'MenuTabs') {
        globalStore.setOpenMenuTabs(val === 'enable');
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
            await search();
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
