<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="30%">
        <template #header>
            <DrawerHeader :header="$t('commons.button.init')" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="initForm" label-position="top" :model="initModel" label-width="100px" :rules="rules">
                    <el-form-item :label="$t('tool.supervisor.primaryConfig')" prop="primaryConfig">
                        <el-input v-model.trim="initModel.primaryConfig"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('tool.supervisor.serviceName')" prop="serviceName">
                        <el-input v-model.trim="initModel.serviceName"></el-input>
                        <span class="input-help">{{ $t('tool.supervisor.serviceNameHelper') }}</span>
                    </el-form-item>
                    <el-alert
                        :title="$t('tool.supervisor.initWarn')"
                        class="common-prompt"
                        :closable="false"
                        type="error"
                    />
                    <el-alert
                        :title="$t('tool.supervisor.restartHelper')"
                        class="common-prompt"
                        :closable="false"
                        type="error"
                    />
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose()" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="openSubmit(initForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
        <ConfirmDialog ref="confirmDialogRef" @confirm="submit(initForm)"></ConfirmDialog>
    </el-drawer>
</template>

<script lang="ts" setup>
import { InitSupervisor } from '@/api/modules/host-tool';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const loading = ref(false);
const initForm = ref<FormInstance>();
const rules = ref({
    primaryConfig: [Rules.requiredInput],
    serviceName: [Rules.requiredInput],
});
const initModel = ref({
    primaryConfig: '',
    serviceName: '',
});
const confirmDialogRef = ref();
const em = defineEmits(['close']);

const acceptParams = (primaryConfig: string, serviceName: string) => {
    initModel.value.primaryConfig = primaryConfig;
    initModel.value.serviceName = serviceName;
    open.value = true;
};

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const openSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('commons.button.init'),
            operationInfo: i18n.global.t('tool.supervisor.restartHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
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
            configPath: initModel.value.primaryConfig,
            serviceName: initModel.value.serviceName,
        })
            .then(() => {
                open.value = false;
                em('close', true);
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
