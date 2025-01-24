<template>
    <DrawerPro
        v-model="open"
        :header="$t('nginx.' + mode)"
        size="large"
        :resource="mode === 'update' ? module.name : ''"
        :back="handleClose"
    >
        <el-form ref="moduleForm" label-position="top" :model="module" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="module.name" :disabled="mode === 'update'"></el-input>
            </el-form-item>
            <el-form-item :label="$t('nginx.params')" prop="params">
                <el-input v-model.trim="module.params" :placeholder="$t('nginx.paramsHelper')"></el-input>
            </el-form-item>
            <el-form-item :label="$t('nginx.packages')" prop="packages">
                <el-input v-model.trim="module.packages" :placeholder="$t('nginx.packagesHelper')"></el-input>
            </el-form-item>
            <el-form-item :label="$t('nginx.script')" prop="script">
                <el-input
                    v-model="module.script"
                    type="textarea"
                    :rows="10"
                    :placeholder="$t('nginx.scriptHelper')"
                ></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit(moduleForm)" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { Nginx } from '@/api/interface/nginx';
import { updateNginxModule } from '@/api/modules/nginx';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';

const moduleForm = ref<FormInstance>();
const open = ref(false);
const em = defineEmits(['close']);
const mode = ref('create');
const loading = ref(false);
const module = ref({
    name: '',
    operate: 'create',
    script: '',
    enable: true,
    params: '',
    packages: '',
});
const rules = ref({
    name: [Rules.requiredInput, Rules.simpleName],
    params: [Rules.requiredInput],
});

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = async (operate: string, editModule: Nginx.NginxModule) => {
    mode.value = operate;
    if (operate === 'update') {
        module.value = {
            name: editModule.name,
            script: editModule.script,
            enable: editModule.enable,
            params: editModule.params,
            packages: editModule.packages,
            operate: 'update',
        };
    }
    open.value = true;
};

const submit = async (form: FormInstance) => {
    await form.validate();
    if (form.validate()) {
        loading.value = true;
        const data = {
            ...module.value,
            operate: mode.value,
        };
        updateNginxModule(data)
            .then(() => {
                if (mode.value === 'update') {
                    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                } else if (mode.value === 'create') {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                }
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    }
};

defineExpose({
    acceptParams,
});
</script>
