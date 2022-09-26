<template>
    <el-dialog v-model="open" :title="$t('app.install')" width="30%">
        <el-form ref="paramForm" label-position="left" :model="form" label-width="150px" :rules="rules">
            <el-form-item :label="$t('app.name')" prop="name">
                <el-input v-model="req.name"></el-input>
            </el-form-item>
            <div v-for="(f, index) in installData.params?.formFields" :key="index">
                <el-form-item :label="f.labelZh" :prop="f.envKey">
                    <el-input
                        v-model="form[f.envKey]"
                        v-if="f.type == 'text' || f.type == 'number'"
                        :type="f.type"
                    ></el-input>
                    <el-input
                        v-model="form[f.envKey]"
                        v-if="f.type == 'password'"
                        :type="f.type"
                        show-password
                    ></el-input>
                </el-form-item>
            </div>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(paramForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup name="appInstall">
import { App } from '@/api/interface/app';
import { InstallApp } from '@/api/modules/app';
import { Rules } from '@/global/form-rues';
import { FormInstance, FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
interface InstallRrops {
    appDetailId: number;
    params?: App.AppParams;
}

const installData = ref<InstallRrops>({
    appDetailId: 0,
});
let open = ref(false);
let form = reactive<{ [key: string]: any }>({});
let rules = reactive<FormRules>({
    name: [Rules.requiredInput],
});
let loading = false;
const paramForm = ref<FormInstance>();
const req = reactive({
    appDetailId: 0,
    params: {},
    name: '',
});
const em = defineEmits(['close']);
const handleClose = () => {
    if (paramForm.value) {
        paramForm.value.resetFields();
    }
    open.value = false;
    em('close', open);
};

const acceptParams = (props: InstallRrops): void => {
    installData.value = props;
    const params = installData.value.params;
    if (params?.formFields != undefined) {
        for (const p of params?.formFields) {
            form[p.envKey] = p.default;
            if (p.required) {
                rules[p.envKey] = [Rules.requiredInput];
            }
        }
    }
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        req.appDetailId = installData.value.appDetailId;
        req.params = form;
        InstallApp(req).then((res) => {
            console.log(res);
        });
    });
};

defineExpose({
    acceptParams,
});
</script>
