<template>
    <el-form :model="form" ref="panelFormRef" label-position="left" label-width="160px">
        <el-card style="margin-top: 10px">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('menu.monitor') }}</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="10">
                    <el-form-item
                        :label="$t('setting.enableMonitor')"
                        :rules="Rules.requiredInput"
                        prop="settingInfo.monitorStatus"
                    >
                        <el-radio-group
                            @change="onSave(panelFormRef, 'MonitorStatus', form.settingInfo.monitorStatus)"
                            v-model="form.settingInfo.monitorStatus"
                        >
                            <el-radio-button label="enable">{{ $t('commons.button.enable') }}</el-radio-button>
                            <el-radio-button label="disable">{{ $t('commons.button.disable') }}</el-radio-button>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item
                        :label="$t('setting.storeDays')"
                        :rules="Rules.number"
                        prop="settingInfo.monitorStoreDays"
                    >
                        <el-input clearable v-model.number="form.settingInfo.monitorStoreDays">
                            <template #append>
                                <el-button
                                    @click="onSave(panelFormRef, 'MonitorStoreDays', form.settingInfo.monitorStoreDays)"
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
import { ref } from 'vue';
import { FormInstance } from 'element-plus';
import { cleanMonitors } from '@/api/modules/setting';
import { useDeleteData } from '@/hooks/use-delete-data';
import { Rules } from '@/global/form-rules';

const emit = defineEmits<{ (e: 'on-save', formEl: FormInstance | undefined, key: string, val: any): void }>();

interface Props {
    settingInfo: any;
}
const form = withDefaults(defineProps<Props>(), {
    settingInfo: {
        monitorStatus: '',
        monitorStoreDays: 30,
    },
});
const panelFormRef = ref<FormInstance>();

function onSave(formEl: FormInstance | undefined, key: string, val: any) {
    emit('on-save', formEl, key, val);
}

const onClean = async () => {
    await useDeleteData(cleanMonitors, {}, 'commons.msg.delete', true);
};
</script>
