<template>
    <div>
        <el-card class="topCard">
            <el-radio-group v-model="activeNames">
                <el-radio-button class="topButton" size="large" label="all">{{ $t('setting.all') }}</el-radio-button>
                <el-radio-button class="topButton" size="large" label="panel">
                    {{ $t('setting.panel') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="safe">{{ $t('setting.safe') }}</el-radio-button>
                <el-radio-button class="topButton" size="large" label="backup">
                    {{ $t('setting.backup') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="monitor">
                    {{ $t('menu.monitor') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="about">
                    {{ $t('setting.about') }}
                </el-radio-button>
            </el-radio-group>
        </el-card>
        <Panel
            v-if="activeNames === 'all' || activeNames === 'panel'"
            :settingInfo="form"
            @on-save="SaveSetting"
            @search="search"
        />
        <Safe
            v-if="activeNames === 'all' || activeNames === 'safe'"
            :settingInfo="form"
            @on-save="SaveSetting"
            @search="search"
        />
        <Backup v-if="activeNames === 'all' || activeNames === 'backup'" :settingInfo="form" @on-save="SaveSetting" />
        <Monitor v-if="activeNames === 'all' || activeNames === 'monitor'" :settingInfo="form" @on-save="SaveSetting" />
        <About v-if="activeNames === 'all' || activeNames === 'about'" />
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue';
import { getSettingInfo, updateSetting } from '@/api/modules/setting';
import { Setting } from '@/api/interface/setting';
import Panel from '@/views/setting/tabs/panel.vue';
import Safe from '@/views/setting/tabs/safe.vue';
import Backup from '@/views/setting/tabs/backup.vue';
import Monitor from '@/views/setting/tabs/monitor.vue';
import About from '@/views/setting/tabs/about.vue';
import { GlobalStore } from '@/store';
import { useTheme } from '@/hooks/use-theme';
import { useI18n } from 'vue-i18n';
import { ElMessage, FormInstance } from 'element-plus';

const i18n = useI18n();
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);

const activeNames = ref('all');
let form = ref<Setting.SettingInfo>({
    userName: '',
    password: '',
    email: '',
    sessionTimeout: 86400,
    localTime: '',
    panelName: '',
    theme: '',
    language: '',
    serverPort: 8888,
    securityEntrance: '',
    expirationDays: 0,
    expirationTime: '',
    complexityVerification: '',
    mfaStatus: '',
    mfaSecret: '',
    monitorStatus: '',
    monitorStoreDays: 30,
    messageType: '',
    emailVars: '',
    weChatVars: '',
    dingVars: '',
});

const search = async () => {
    const res = await getSettingInfo();
    form.value = res.data;
    form.value.password = '******';
    form.value.expirationTime = form.value.expirationDays === 0 ? '-' : form.value.expirationTime;
};

const { switchDark } = useTheme();

const SaveSetting = async (formEl: FormInstance | undefined, key: string, val: any) => {
    if (!formEl) return;
    const result = await formEl.validateField('settingInfo.' + key.replace(key[0], key[0].toLowerCase()), callback);
    if (!result) {
        return;
    }
    if (val === '') {
        return;
    }
    switch (key) {
        case 'Language':
            i18n.locale.value = val;
            globalStore.updateLanguage(val);
            break;
        case 'Theme':
            globalStore.setThemeConfig({ ...themeConfig.value, theme: val });
            switchDark();
            break;
        case 'PanelName':
            globalStore.setThemeConfig({ ...themeConfig.value, panelName: val });
            break;
        case 'MonitorStoreDays':
        case 'ServerPort':
            val = val + '';
            break;
    }
    let param = {
        key: key,
        value: val + '',
    };
    await updateSetting(param);
    ElMessage.success(i18n.t('commons.msg.operationSuccess'));
    search();
};

function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

onMounted(() => {
    search();
});
</script>

<style>
.topCard {
    --el-card-border-color: var(--el-border-color-light);
    --el-card-border-radius: 4px;
    --el-card-padding: 0px;
    --el-card-bg-color: var(--el-fill-color-blank);
}
.topButton .el-radio-button__inner {
    display: inline-block;
    line-height: 1;
    white-space: nowrap;
    vertical-align: middle;
    background: var(--el-button-bg-color, var(--el-fill-color-blank));
    border: 0;
    font-weight: 350;
    border-left: 0;
    color: var(--el-button-text-color, var(--el-text-color-regular));
    text-align: center;
    box-sizing: border-box;
    outline: 0;
    margin: 0;
    position: relative;
    cursor: pointer;
    transition: var(--el-transition-all);
    -webkit-user-select: none;
    user-select: none;
    padding: 8px 15px;
    font-size: var(--el-font-size-base);
    border-radius: 0;
}
</style>
