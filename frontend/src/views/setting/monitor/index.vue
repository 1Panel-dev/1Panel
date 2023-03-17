<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.monitor')" :divider="true">
            <template #main>
                <el-form :model="form" ref="panelFormRef" label-position="left" label-width="160px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="10">
                            <el-form-item
                                :label="$t('setting.enableMonitor')"
                                :rules="Rules.requiredInput"
                                prop="monitorStatus"
                            >
                                <el-switch
                                    @change="onSave(panelFormRef, 'MonitorStatus', form.monitorStatus)"
                                    v-model="form.monitorStatus"
                                    active-value="enable"
                                    inactive-value="disable"
                                />
                            </el-form-item>
                            <el-form-item
                                :label="$t('setting.storeDays')"
                                :rules="[Rules.integerNumber, checkNumberRange(1, 30)]"
                                prop="monitorStoreDays"
                            >
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
                </el-form>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import LayoutContent from '@/layout/layout-content.vue';
import { cleanMonitors, getSettingInfo, getSystemAvailable, updateSetting } from '@/api/modules/setting';
import { useDeleteData } from '@/hooks/use-delete-data';
import { Rules, checkNumberRange } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const form = reactive({
    monitorStatus: 'disable',
    monitorStoreDays: 30,
});
const panelFormRef = ref<FormInstance>();

const search = async () => {
    const res = await getSettingInfo();
    form.monitorStatus = res.data.monitorStatus;
    form.monitorStoreDays = Number(res.data.monitorStoreDays);
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
            val = val + '';
            break;
    }
    let param = {
        key: key,
        value: val + '',
    };
    loading.value = true;
    await updateSetting(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
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
    getSystemAvailable();
});
</script>
