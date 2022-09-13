<template>
    <el-form size="small" :model="form" label-position="left" label-width="120px">
        <el-card style="margin-top: 10px">
            <template #header>
                <div class="card-header">
                    <span>安全</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="8">
                    <el-form-item label="面板端口">
                        <el-input clearable v-model="form.settingInfo.serverPort">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('ServerPort', form.settingInfo.serverPort)"
                                    icon="Collection"
                                >
                                    保存
                                </el-button>
                            </template>
                            <el-tooltip
                                class="box-item"
                                effect="dark"
                                content="Top Left prompts info"
                                placement="top-start"
                            >
                                <el-icon style="font-size: 14px; margin-top: 8px"><WarningFilled /></el-icon>
                            </el-tooltip>
                        </el-input>
                        <div>
                            <span class="input-help">
                                建议端口范围8888 - 65535，注意：有安全组的服务器请提前在安全组放行新端口
                            </span>
                        </div>
                    </el-form-item>
                    <el-form-item label="安全入口">
                        <el-input clearable v-model="form.settingInfo.securityEntrance">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('SecurityEntrance', form.settingInfo.securityEntrance)"
                                    icon="Collection"
                                >
                                    保存
                                </el-button>
                            </template>
                        </el-input>
                        <div>
                            <span class="input-help">
                                面板管理入口，设置后只能通过指定安全入口登录面板,如: /89dc6ae8
                            </span>
                        </div>
                    </el-form-item>
                    <el-form-item label="密码过期时间">
                        <el-input clearable v-model="form.settingInfo.passwordTimeOut">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('Password', form.settingInfo.passwordTimeOut)"
                                    icon="Collection"
                                >
                                    保存
                                </el-button>
                            </template>
                        </el-input>
                        <div>
                            <span class="input-help">为面板密码设置过期时间，过期后需要重新设置密码</span>
                        </div>
                    </el-form-item>
                    <el-form-item label="密码复杂度验证">
                        <el-switch
                            v-model="form.settingInfo.complexityVerification"
                            active-value="enable"
                            inactive-value="disable"
                            @change="SaveSetting('ComplexityVerification', form.settingInfo.complexityVerification)"
                        />
                        <div>
                            <span class="input-help">
                                密码必须满足密码长度大于8位且大写字母、小写字母、数字、特殊字符至少3项组合
                            </span>
                        </div>
                    </el-form-item>
                    <el-form-item label="两步验证">
                        <el-switch
                            v-model="form.settingInfo.mfaStatus"
                            active-value="enable"
                            inactive-value="disable"
                            @change="SaveSetting('MFAStatus', form.settingInfo.mfaStatus)"
                        />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-card>
    </el-form>
</template>

<script lang="ts" setup>
import { ElMessage, ElForm } from 'element-plus';
import { updateSetting } from '@/api/modules/setting';
import i18n from '@/lang';

interface Props {
    settingInfo: any;
}
const form = withDefaults(defineProps<Props>(), {
    settingInfo: {
        serverPort: '',
        securityEntrance: '',
        passwordTimeOut: '',
        complexityVerification: '',
        mfaStatus: '',
    },
});

const SaveSetting = async (key: string, val: string) => {
    let param = {
        key: key,
        value: val,
    };
    await updateSetting(param);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};
</script>
