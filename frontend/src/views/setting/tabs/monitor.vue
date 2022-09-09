<template>
    <el-form size="small" :model="form" label-position="left" label-width="120px">
        <el-card style="margin-top: 10px">
            <template #header>
                <div class="card-header">
                    <span>监控</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="8">
                    <el-form-item label="开启监控">
                        <el-switch
                            v-model="form.settingInfo.monitorStatus"
                            active-value="enable"
                            inactive-value="disable"
                            @change="SaveSetting('MonitorStatus', form.settingInfo.monitorStatus)"
                        />
                    </el-form-item>
                    <el-form-item label="过期时间">
                        <el-input clearable v-model="form.settingInfo.monitorStoreDays">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('MonitorStoreDays', form.settingInfo.monitorStoreDays)"
                                    icon="Collection"
                                >
                                    保存
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item>
                        <el-button icon="Delete"> 清空监控记录 </el-button>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-card>
    </el-form>
</template>

<script lang="ts" setup>
import { ElMessage } from 'element-plus';
import { updateSetting } from '@/api/modules/setting';
import i18n from '@/lang';

interface Props {
    settingInfo: any;
}
const form = withDefaults(defineProps<Props>(), {
    settingInfo: {
        monitorStatus: '',
        monitorStoreDays: '',
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
