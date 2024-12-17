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
                                <el-radio-group @change="onSave('Theme', form.theme)" v-model="form.theme">
                                    <el-radio-button v-if="isMasterProductPro" value="dark-gold">
                                        <span>{{ $t('setting.darkGold') }}</span>
                                    </el-radio-button>
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

                            <el-form-item :label="$t('setting.proxy')" prop="proxyShow">
                                <el-input disabled v-model="form.proxyShow">
                                    <template #append>
                                        <el-button @click="onChangeProxy" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
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

                            <el-form-item :label="$t('setting.advancedMenuHide')">
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
        <Proxy ref="proxyRef" @search="search()" />
        <Timeout ref="timeoutRef" @search="search()" />
        <HideMenu ref="hideMenuRef" @search="search()" />
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElForm } from 'element-plus';
import { getSettingInfo, updateSetting, getSystemAvailable } from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { useI18n } from 'vue-i18n';
import { useTheme } from '@/global/use-theme';
import { MsgSuccess } from '@/utils/message';
import Password from '@/views/setting/panel/password/index.vue';
import UserName from '@/views/setting/panel/username/index.vue';
import Timeout from '@/views/setting/panel/timeout/index.vue';
import PanelName from '@/views/setting/panel/name/index.vue';
import Proxy from '@/views/setting/panel/proxy/index.vue';
import HideMenu from '@/views/setting/panel/hidemenu/index.vue';
import { storeToRefs } from 'pinia';
import { getXpackSetting, updateXpackSettingByKey } from '@/utils/xpack';

const loading = ref(false);
const i18n = useI18n();
const globalStore = GlobalStore();

const { isMasterProductPro } = storeToRefs(globalStore);

const { switchTheme } = useTheme();
const mobile = computed(() => {
    return globalStore.isMobile();
});

const form = reactive({
    userName: '',
    password: '',
    sessionTimeout: 0,
    panelName: '',
    theme: '',
    menuTabs: '',
    language: '',
    complexityVerification: '',
    developerMode: '',

    proxyShow: '',
    proxyUrl: '',
    proxyType: '',
    proxyPort: '',
    proxyUser: '',
    proxyPasswd: '',
    proxyPasswdKeep: '',

    proHideMenus: ref(i18n.t('setting.unSetting')),
    hideMenuList: '',
});

const show = ref();

const userNameRef = ref();
const passwordRef = ref();
const panelNameRef = ref();
const proxyRef = ref();
const timeoutRef = ref();
const hideMenuRef = ref();
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
    form.panelName = res.data.panelName;
    form.menuTabs = res.data.menuTabs;
    form.language = res.data.language;
    form.complexityVerification = res.data.complexityVerification;
    form.developerMode = res.data.developerMode;

    form.proxyUrl = res.data.proxyUrl;
    form.proxyType = res.data.proxyType;
    form.proxyPort = res.data.proxyPort;
    form.proxyShow = form.proxyUrl ? form.proxyUrl + ':' + form.proxyPort : unset.value;
    form.proxyUser = res.data.proxyUser;
    form.proxyPasswd = res.data.proxyPasswd;
    form.proxyPasswdKeep = res.data.proxyPasswdKeep;

    const json: Node = JSON.parse(res.data.xpackHideMenu);
    if (json.isCheck === false) {
        json.children.forEach((child: any) => {
            if (child.isCheck === true) {
                child.isCheck = false;
            }
        });
    }
    form.proHideMenus = JSON.stringify(json);
    form.hideMenuList = JSON.stringify(json);
    const checkedTitles = getCheckedTitles(json);
    form.proHideMenus = checkedTitles.toString();

    if (isMasterProductPro.value) {
        const xpackRes = await getXpackSetting();
        if (xpackRes) {
            form.theme = xpackRes.data.theme === 'dark-gold' ? 'dark-gold' : res.data.theme;
            return;
        }
    }
    form.theme = res.data.theme;
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
const onChangeProxy = () => {
    proxyRef.value.acceptParams({
        url: form.proxyUrl,
        type: form.proxyType,
        port: form.proxyPort,
        user: form.proxyUser,
        passwd: form.proxyPasswd,
        passwdKeep: form.proxyPasswdKeep,
    });
};

const onChangeHideMenus = () => {
    hideMenuRef.value.acceptParams({ menuList: form.hideMenuList });
};

const onSave = async (key: string, val: any) => {
    loading.value = true;
    if (key === 'Language') {
        i18n.locale.value = val;
        globalStore.updateLanguage(val);
    }
    if (key === 'Theme') {
        if (val === 'dark-gold') {
            globalStore.themeConfig.isGold = true;
        } else {
            globalStore.themeConfig.isGold = false;
            globalStore.themeConfig.theme = val;
        }
        switchTheme();
        if (globalStore.isMasterProductPro) {
            updateXpackSettingByKey('Theme', val === 'dark-gold' ? 'dark-gold' : '');
            if (val === 'dark-gold') {
                MsgSuccess(i18n.t('commons.msg.operationSuccess'));
                loading.value = false;
                search();
                return;
            }
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
