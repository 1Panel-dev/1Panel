<template>
    <el-form size="small" :model="form" label-position="left" label-width="120px">
        <el-card style="margin-top: 10px">
            <template #header>
                <div class="card-header">
                    <span>通知</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="8">
                    <el-form-item label="通知方式">
                        <el-radio-group v-model="form.settingInfo.messageType">
                            <el-radio-button label="none">关闭</el-radio-button>
                            <el-radio-button label="email">email</el-radio-button>
                            <el-radio-button label="wechat">企业微信</el-radio-button>
                            <el-radio-button label="dingding">钉钉</el-radio-button>
                        </el-radio-group>
                    </el-form-item>
                    <div v-if="form.settingInfo.messageType === 'none'">
                        <el-form-item>
                            <el-button @click="SaveSetting()">关闭消息通知</el-button>
                        </el-form-item>
                    </div>
                    <div v-if="form.settingInfo.messageType === 'email'">
                        <el-form-item label="邮箱服务名称">
                            <el-input clearable v-model="emailVars.serverName" />
                        </el-form-item>
                        <el-form-item label="邮箱地址">
                            <el-input clearable v-model="emailVars.serverAddr" />
                        </el-form-item>
                        <el-form-item label="邮箱SMTP授权码">
                            <el-input clearable v-model="emailVars.serverSMTP" />
                        </el-form-item>
                        <el-form-item>
                            <el-button @click="SaveSetting()">保存并启用</el-button>
                        </el-form-item>
                    </div>
                    <div v-if="form.settingInfo.messageType === 'wechat'">
                        <el-form-item label="orpid">
                            <el-input clearable v-model="weChatVars.orpid" />
                        </el-form-item>
                        <el-form-item label="corpsecret">
                            <el-input clearable v-model="weChatVars.corpsecret" />
                        </el-form-item>
                        <el-form-item label="touser">
                            <el-input clearable v-model="weChatVars.touser" />
                        </el-form-item>
                        <el-form-item label="agentid">
                            <el-input clearable v-model="weChatVars.agentid" />
                        </el-form-item>
                        <el-form-item>
                            <el-button @click="SaveSetting()">保存并启用</el-button>
                        </el-form-item>
                    </div>
                    <div v-if="form.settingInfo.messageType === 'dingding'">
                        <el-form-item label="webhook token">
                            <el-input clearable v-model="dingVars.webhookToken" />
                        </el-form-item>
                        <el-form-item label="密钥">
                            <el-input clearable v-model="dingVars.secret" />
                        </el-form-item>
                        <el-form-item label="邮箱 SMTP 授权码">
                            <el-input clearable v-model="emailVars.serverSMTP" />
                        </el-form-item>
                        <el-form-item>
                            <el-button @click="SaveSetting()">保存并启用</el-button>
                        </el-form-item>
                    </div>
                </el-col>
            </el-row>
        </el-card>
    </el-form>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { updateSetting } from '@/api/modules/setting';
import i18n from '@/lang';

interface Props {
    settingInfo: any;
}
const form = withDefaults(defineProps<Props>(), {
    settingInfo: {
        messageType: '',
        emailVars: '',
        weChatVars: '',
        dingVars: '',
    },
});

const emailVars = ref({
    serverName: '',
    serverAddr: '',
    serverSMTP: '',
});
const weChatVars = ref({
    orpid: '',
    corpsecret: '',
    touser: '',
    agentid: '',
});
const dingVars = ref({
    webhookToken: '',
    secret: '',
});

const SaveSetting = async () => {
    let settingKey = '';
    let settingVal = '';
    switch (form.settingInfo.messageType) {
        case 'none':
            settingVal = '';
            break;
        case 'email':
            settingVal = JSON.stringify(emailVars.value);
            settingKey = 'EmailVars';
            break;
        case 'wechat':
            settingVal = JSON.stringify(emailVars.value);
            settingKey = 'WeChatVars';
            break;
        case 'dingding':
            settingVal = JSON.stringify(emailVars.value);
            settingKey = 'DingVars';
            break;
    }
    let param = {
        key: settingKey,
        value: settingVal,
    };
    await updateSetting({ key: 'MessageType', value: form.settingInfo.messageType });
    await updateSetting(param);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};
onMounted(() => {
    switch (form.settingInfo.messageType) {
        case 'email':
            emailVars.value = JSON.parse(form.settingInfo.emailVars);
            console.log(emailVars.value);
            break;
        case 'wechat':
            weChatVars.value = JSON.parse(form.settingInfo.weChatVars);
            break;
        case 'dingding':
            dingVars.value = JSON.parse(form.settingInfo.dingVars);
            break;
    }
});
</script>
