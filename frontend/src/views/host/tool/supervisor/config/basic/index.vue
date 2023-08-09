<template>
    <el-row v-loading="loading">
        <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12" :offset="1">
            <el-form ref="initForm" label-position="top" :model="data" label-width="100px" :rules="rules">
                <el-form-item :label="$t('tool.supervisor.primaryConfig')" prop="configPath">
                    <el-input v-model.trim="data.configPath"></el-input>
                </el-form-item>
                <el-form-item :label="$t('tool.supervisor.serviceName')" prop="serviceName">
                    <el-input v-model.trim="data.serviceName"></el-input>
                    <span class="input-help">{{ $t('tool.supervisor.serviceNameHelper') }}</span>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="submit(initForm)" :disabled="loading">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </el-form-item>
            </el-form>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { HostTool } from '@/api/interface/host-tool';
import { GetSupervisorStatus, InitSupervisor } from '@/api/modules/host-tool';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { onMounted, ref } from 'vue';
const loading = ref(false);
const initForm = ref<FormInstance>();
const rules = ref({
    configPath: [Rules.requiredInput],
    serviceName: [Rules.requiredInput],
});

const data = ref({
    isExist: false,
    version: '',
    status: 'running',
    init: false,
    configPath: '',
    ctlExist: false,
    serviceName: '',
});

const getStatus = async () => {
    try {
        loading.value = true;
        const res = await GetSupervisorStatus();
        data.value = res.data.config as HostTool.Supersivor;
    } catch (error) {}
    loading.value = false;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        InitSupervisor({
            type: 'supervisord',
            configPath: data.value.configPath,
            serviceName: data.value.serviceName,
        })
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

onMounted(() => {
    getStatus();
});
</script>
