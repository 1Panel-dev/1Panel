<template>
    <el-form :model="mesForm" label-position="left" label-width="160px">
        <el-card style="margin-top: 10px">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('setting.message') }}</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="10">
                    <el-form-item :label="$t('setting.messageType')">
                        <el-radio-group v-model="mesForm.messageType">
                            <el-radio-button label="none">{{ $t('commons.button.close') }}</el-radio-button>
                            <el-radio-button label="email">{{ $t('setting.email') }}</el-radio-button>
                            <el-radio-button label="wechat">{{ $t('setting.wechat') }}</el-radio-button>
                            <el-radio-button label="dingding">{{ $t('setting.dingding') }}</el-radio-button>
                        </el-radio-group>
                    </el-form-item>
                    <div v-if="mesForm.messageType === 'none'">
                        <el-form-item>
                            <el-button @click="SaveSetting()">{{ $t('setting.closeMessage') }}</el-button>
                        </el-form-item>
                    </div>
                    <div v-if="mesForm.messageType === 'email'">
                        <el-form-item :label="$t('setting.emailServer')">
                            <el-input clearable v-model="mesForm.emailVars.serverName" />
                        </el-form-item>
                        <el-form-item :label="$t('setting.emailAddr')">
                            <el-input clearable v-model="mesForm.emailVars.serverAddr" />
                        </el-form-item>
                        <el-form-item :label="$t('setting.emailSMTP')">
                            <el-input clearable v-model="mesForm.emailVars.serverSMTP" />
                        </el-form-item>
                        <el-form-item>
                            <el-button @click="SaveSetting()">{{ $t('commons.button.saveAndEnable') }}</el-button>
                        </el-form-item>
                    </div>
                    <div v-if="mesForm.messageType === 'wechat'">
                        <el-form-item label="orpid">
                            <el-input clearable v-model="mesForm.weChatVars.orpid" />
                        </el-form-item>
                        <el-form-item label="corpsecret">
                            <el-input clearable v-model="mesForm.weChatVars.corpsecret" />
                        </el-form-item>
                        <el-form-item label="touser">
                            <el-input clearable v-model="mesForm.weChatVars.touser" />
                        </el-form-item>
                        <el-form-item label="agentid">
                            <el-input clearable v-model="mesForm.weChatVars.agentid" />
                        </el-form-item>
                        <el-form-item>
                            <el-button @click="SaveSetting()">{{ $t('commons.button.saveAndEnable') }}</el-button>
                        </el-form-item>
                    </div>
                    <div v-if="mesForm.messageType === 'dingding'">
                        <el-form-item label="webhook token">
                            <el-input clearable v-model="mesForm.dingVars.webhookToken" />
                        </el-form-item>
                        <el-form-item :label="$t('setting.secret')">
                            <el-input clearable v-model="mesForm.dingVars.secret" />
                        </el-form-item>
                        <el-form-item>
                            <el-button @click="SaveSetting()">{{ $t('commons.button.saveAndEnable') }}</el-button>
                        </el-form-item>
                    </div>
                </el-col>
            </el-row>
        </el-card>
    </el-form>
</template>

<script lang="ts" setup>
import { reactive, watch } from 'vue';
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

const mesForm = reactive({
    messageType: '',
    emailVars: {
        serverName: '',
        serverAddr: '',
        serverSMTP: '',
    },
    weChatVars: {
        orpid: '',
        corpsecret: '',
        touser: '',
        agentid: '',
    },
    dingVars: {
        webhookToken: '',
        secret: '',
    },
});

watch(form, (val: any) => {
    if (val.settingInfo.messageType) {
        mesForm.messageType = form.settingInfo.messageType;
        mesForm.emailVars = val.settingInfo.emailVars
            ? JSON.parse(val.settingInfo.emailVars)
            : { serverName: '', serverAddr: '', serverSMTP: '' };
        mesForm.weChatVars = val.settingInfo.weChatVars
            ? JSON.parse(val.settingInfo.weChatVars)
            : { orpid: '', corpsecret: '', touser: '', agentid: '' };
        mesForm.dingVars = val.settingInfo.dingVars
            ? JSON.parse(val.settingInfo.dingVars)
            : { webhookToken: '', secret: '' };
    }
});

const SaveSetting = async () => {
    let settingKey = '';
    let settingVal = '';
    switch (mesForm.messageType) {
        case 'none':
            settingVal = '';
            break;
        case 'email':
            settingVal = JSON.stringify(mesForm.emailVars);
            settingKey = 'EmailVars';
            break;
        case 'wechat':
            settingVal = JSON.stringify(mesForm.weChatVars);
            settingKey = 'WeChatVars';
            break;
        case 'dingding':
            settingVal = JSON.stringify(mesForm.dingVars);
            settingKey = 'DingVars';
            break;
    }
    let param = {
        key: settingKey,
        value: settingVal,
    };
    await updateSetting({ key: 'MessageType', value: mesForm.messageType });
    await updateSetting(param);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};
</script>
