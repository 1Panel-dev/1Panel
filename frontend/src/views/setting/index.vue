<template>
    <div class="demo-collapse">
        <el-card class="topCard">
            <el-radio-group v-model="activeNames">
                <el-radio-button class="topButton" size="large" label="all">全部</el-radio-button>
                <el-radio-button class="topButton" size="large" label="user">用户设置</el-radio-button>
                <el-radio-button class="topButton" size="large" label="panel">面板设置</el-radio-button>
                <el-radio-button class="topButton" size="large" label="safe">安全设置</el-radio-button>
                <el-radio-button class="topButton" size="large" label="backup">备份设置</el-radio-button>
                <el-radio-button class="topButton" size="large" label="monitor">监控设置</el-radio-button>
                <el-radio-button class="topButton" size="large" label="message">通知设置</el-radio-button>
                <el-radio-button class="topButton" size="large" label="about">关于</el-radio-button>
            </el-radio-group>
        </el-card>
        <el-form :model="form" label-position="left" label-width="120px">
            <el-card v-if="activeNames === 'all' || activeNames === 'user'" style="margin-top: 20px">
                <template #header>
                    <div class="card-header">
                        <span>用户设置</span>
                    </div>
                </template>
                <el-row>
                    <el-col :span="1"><br /></el-col>
                    <el-col :span="8">
                        <el-form-item label="用户名">
                            <el-input size="small" clearable v-model="form.userName" />
                        </el-form-item>
                        <el-form-item label="密码">
                            <el-input type="password" clearable show-password size="small" v-model="form.password" />
                        </el-form-item>
                        <el-form-item label="主题色">
                            <el-radio-group size="small" v-model="form.theme">
                                <el-radio-button label="black">
                                    <el-icon><Moon /></el-icon>黑金
                                </el-radio-button>
                                <el-tooltip
                                    effect="dark"
                                    placement="top"
                                    content="选择自动设置，将会在晚 6 点到次日早 6 点间自动切换到黑金主题。"
                                >
                                    <el-radio-button label="auto" icon="Sunny">
                                        <el-icon><MagicStick /></el-icon>自动
                                    </el-radio-button>
                                </el-tooltip>
                                <el-radio-button label="write">
                                    <el-icon><Sunny /></el-icon>白金
                                </el-radio-button>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item label="系统语言">
                            <el-radio-group size="small" v-model="form.language">
                                <el-radio-button label="ch">中文 </el-radio-button>
                                <el-radio-button label="en">English </el-radio-button>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item>
                            <el-button size="small" icon="Pointer">更新用户设置</el-button>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-card>
            <el-card v-if="activeNames === 'all' || activeNames === 'panel'" style="margin-top: 10px">
                <template #header>
                    <div class="card-header">
                        <span>面板设置</span>
                    </div>
                </template>
                <el-row>
                    <el-col :span="1"><br /></el-col>
                    <el-col :span="8">
                        <el-form-item label="超时时间">
                            <el-input size="small" v-model="form.sessionTimeout" />
                        </el-form-item>
                        <el-form-item label="同步时间">
                            <el-input size="small" v-model="form.password" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-card>
        </el-form>
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, reactive } from 'vue';
import { getSettingInfo } from '@/api/modules/setting';
import { Setting } from '@/api/interface/setting';

const activeNames = ref('all');
let form = reactive<Setting.SettingInfo>({
    userName: '',
    password: '',
    email: '',
    sessionTimeout: '',
    panelName: '',
    theme: '',
    language: '',
    serverPort: '',
    securityEntrance: '',
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
    console.log(res);
    form = res.data;
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
