<template>
    <el-form :model="form" label-position="left" label-width="160px">
        <el-card style="margin-top: 10px">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('menu.monitor') }}</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="10">
                    <el-form-item :label="$t('setting.enableMonitor')">
                        <el-radio-group
                            @change="SaveSetting('MonitorStatus', form.settingInfo.monitorStatus)"
                            v-model="form.settingInfo.monitorStatus"
                        >
                            <el-radio-button label="enable">{{ $t('commons.button.enable') }}</el-radio-button>
                            <el-radio-button label="disable">{{ $t('commons.button.disable') }}</el-radio-button>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item :label="$t('setting.storeDays')">
                        <el-input clearable v-model="form.settingInfo.monitorStoreDays">
                            <template #append>
                                <el-button
                                    @click="SaveSetting('MonitorStoreDays', form.settingInfo.monitorStoreDays)"
                                    icon="Collection"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item>
                        <el-button @click="onClean()" icon="Delete">{{ $t('setting.cleanMonitor') }}</el-button>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-card>
    </el-form>
</template>

<script lang="ts" setup>
import { ElMessage } from 'element-plus';
import { updateSetting, cleanMonitors } from '@/api/modules/setting';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';

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

const onClean = async () => {
    await useDeleteData(cleanMonitors, {}, 'commons.msg.delete', true);
};
</script>
