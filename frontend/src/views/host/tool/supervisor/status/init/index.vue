<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="30%" :show-close="false">
        <template #header>
            <span>{{ $t('commons.button.init') }}</span>
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="initForm" label-position="top" :model="initModel" label-width="100px" :rules="rules">
                    <el-form-item :label="$t('tool.supervisor.primaryConfig')" prop="primaryConfig">
                        <el-input v-model.trim="initModel.primaryConfig"></el-input>
                    </el-form-item>
                    <el-alert
                        :title="$t('tool.supervisor.initWarn')"
                        class="common-prompt"
                        :closable="false"
                        type="error"
                    />
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button type="primary" @click="submit(initForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
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
});
const initModel = ref({
    primaryConfig: '',
});

const em = defineEmits(['close']);

const acceptParams = (primaryConfig: string) => {
    initModel.value.primaryConfig = primaryConfig;
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        InitSupervisor({ type: 'supervisord', configPath: initModel.value.primaryConfig })
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
