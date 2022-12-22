<template>
    <div>
        <Submenu activeName="monitor" />
        <el-form :model="form" ref="panelFormRef" label-position="left" label-width="160px">
            <el-card style="margin-top: 20px">
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
                            prop="monitorStatus"
                        >
                            <el-radio-group
                                @change="onSave(panelFormRef, 'MonitorStatus', form.monitorStatus)"
                                v-model="form.monitorStatus"
                            >
                                <el-radio-button label="enable">{{ $t('commons.button.enable') }}</el-radio-button>
                                <el-radio-button label="disable">{{ $t('commons.button.disable') }}</el-radio-button>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item :label="$t('setting.storeDays')" :rules="Rules.number" prop="monitorStoreDays">
                            <el-input clearable v-model.number="form.monitorStoreDays">
                                <template #append>
                                    <el-button
                                        @click="onSave(panelFormRef, 'MonitorStoreDays', form.monitorStoreDays)"
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
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, FormInstance } from 'element-plus';
import { cleanMonitors, getSettingInfo, updateSetting } from '@/api/modules/setting';
import { useDeleteData } from '@/hooks/use-delete-data';
import Submenu from '@/views/setting/index.vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';

const form = reactive({
    monitorStatus: '',
    monitorStoreDays: 30,
});
const panelFormRef = ref<FormInstance>();

const search = async () => {
    const res = await getSettingInfo();
    form.monitorStatus = res.data.monitorStatus;
    form.monitorStoreDays = res.data.monitorStoreDays;
};

const onSave = async (formEl: FormInstance | undefined, key: string, val: any) => {
    if (!formEl) return;
    const result = await formEl.validateField(key.replace(key[0], key[0].toLowerCase()), callback);
    if (!result) {
        return;
    }
    if (val === '') {
        return;
    }
    switch (key) {
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
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const onClean = async () => {
    await useDeleteData(cleanMonitors, {}, 'commons.msg.delete');
};

onMounted(() => {
    search();
});
</script>
