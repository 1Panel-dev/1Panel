<template>
    <div class="demo-collapse">
        <el-card class="topCard">
            <el-radio-group v-model="activeNames">
                <el-radio-button class="topButton" size="large" label="all">全部</el-radio-button>
                <el-radio-button class="topButton" size="large" label="panel">面板</el-radio-button>
                <el-radio-button class="topButton" size="large" label="safe">安全</el-radio-button>
                <el-radio-button class="topButton" size="large" label="backup">备份</el-radio-button>
                <el-radio-button class="topButton" size="large" label="monitor">监控</el-radio-button>
                <el-radio-button class="topButton" size="large" label="message">通知</el-radio-button>
                <el-radio-button class="topButton" size="large" label="about">关于</el-radio-button>
            </el-radio-group>
        </el-card>
        <Panel v-if="activeNames === 'all' || activeNames === 'panel'" :settingInfo="form" />
        <Safe v-if="activeNames === 'all' || activeNames === 'safe'" :settingInfo="form" />
        <Backup v-if="activeNames === 'all' || activeNames === 'backup'" :settingInfo="form" />
        <Monitor v-if="activeNames === 'all' || activeNames === 'monitor'" :settingInfo="form" />
        <Message v-if="activeNames === 'all' || activeNames === 'message'" :settingInfo="form" />
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { getSettingInfo } from '@/api/modules/setting';
import { Setting } from '@/api/interface/setting';
import Panel from '@/views/setting/tabs/panel.vue';
import Safe from '@/views/setting/tabs/safe.vue';
import Backup from '@/views/setting/tabs/backup.vue';
import Monitor from '@/views/setting/tabs/monitor.vue';
import Message from '@/views/setting/tabs/message.vue';

const activeNames = ref('all');
let form = ref<Setting.SettingInfo>({
    userName: '',
    password: '',
    email: '',
    sessionTimeout: '',
    localTime: '',
    panelName: '',
    theme: '',
    language: '',
    serverPort: '',
    securityEntrance: '',
    passwordTimeOut: '',
    complexityVerification: '',
    mfaStatus: '',
    monitorStatus: '',
    monitorStoreDays: '',
    messageType: '',
    emailVars: '',
    weChatVars: '',
    dingVars: '',
});

const search = async () => {
    const res = await getSettingInfo();
    form.value = res.data;
    form.value.password = '******';
};

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
