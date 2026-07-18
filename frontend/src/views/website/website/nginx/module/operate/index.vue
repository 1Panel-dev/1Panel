<template>
    <DrawerPro
        v-model="open"
        :header="$t('nginx.' + mode)"
        size="large"
        :resource="mode === 'update' ? module.name : ''"
        @close="handleClose"
    >
        <el-form ref="moduleForm" label-position="top" :model="module" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="module.name" :disabled="mode === 'update'"></el-input>
            </el-form-item>
            <el-form-item :label="$t('nginx.buildMode')" prop="buildMode">
                <el-radio-group v-model="module.buildMode">
                    <el-radio-button value="auto">{{ $t('nginx.buildModeAuto') }}</el-radio-button>
                    <el-radio-button value="dynamic">{{ $t('nginx.buildModeDynamic') }}</el-radio-button>
                    <el-radio-button value="static">{{ $t('nginx.buildModeStatic') }}</el-radio-button>
                </el-radio-group>
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
            <el-form-item v-if="module.buildMode !== 'static'" :label="$t('nginx.loadOrder')" prop="loadOrder">
                <el-input-number v-model="module.loadOrder" :min="0" :max="9999" />
            </el-form-item>
            <el-alert
                v-if="module.lastError"
                class="!mb-4"
                type="error"
                :title="$t('nginx.buildFailed')"
                :description="module.lastError"
                :closable="false"
                show-icon
            />
        </el-form>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button v-permission type="primary" @click="submit(moduleForm)" :disabled="loading">
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
import { ref } from 'vue';

const moduleForm = ref<FormInstance>();
const open = ref(false);
const em = defineEmits(['close']);
const mode = ref('create');
const loading = ref(false);
type ModuleForm = {
    name: string;
    operate: string;
    script: string;
    enable: boolean;
    params: string;
    packages: string;
    buildMode: Nginx.NginxModule['buildMode'];
    provider: Nginx.NginxModule['provider'];
    dynamicSupport: Nginx.NginxModule['dynamicSupport'];
    loadOrder: number;
    lastError: string;
};
const defaultModule = (): ModuleForm => ({
    name: '',
    operate: 'create',
    script: '',
    enable: true,
    params: '',
    packages: '',
    buildMode: 'auto',
    provider: 'local',
    dynamicSupport: 'unknown',
    loadOrder: 50,
    lastError: '',
});
const module = ref(defaultModule());
const rules = ref({
    name: [Rules.requiredInput, Rules.simpleName],
    params: [Rules.requiredInput],
    buildMode: [Rules.requiredSelect],
});

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = async (operate: string, editModule?: Nginx.NginxModule) => {
    mode.value = operate;
    module.value = defaultModule();
    if (operate === 'update' && editModule) {
        module.value = {
            name: editModule.name,
            script: editModule.script || '',
            enable: editModule.enable,
            params: editModule.params,
            packages: editModule.packages || '',
            buildMode: editModule.buildMode,
            provider: editModule.provider,
            dynamicSupport: editModule.dynamicSupport,
            loadOrder: editModule.loadOrder,
            lastError: editModule.lastError || '',
            operate: 'update',
        };
    }
    open.value = true;
};

const submit = async (form: FormInstance) => {
    await form.validate();
    loading.value = true;
    const data = {
        ...module.value,
        operate: mode.value,
    };
    try {
        await updateNginxModule(data);
        if (mode.value === 'update') {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        } else if (mode.value === 'create') {
            MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
        }
        handleClose();
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
