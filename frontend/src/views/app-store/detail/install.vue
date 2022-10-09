<template>
    <el-dialog v-model="open" :title="$t('app.install')" width="40%" :before-close="handleClose" @opened="opened">
        <el-form ref="paramForm" label-position="left" :model="form" label-width="150px" :rules="rules">
            <el-form-item :label="$t('app.name')" prop="NAME">
                <el-input v-model="form['NAME']"></el-input>
            </el-form-item>
            <div v-for="(f, index) in installData.params?.formFields" :key="index">
                <el-form-item :label="f.labelZh" :prop="f.envKey">
                    <el-input v-model="form[f.envKey]" v-if="f.type == 'text'" :type="f.type"></el-input>
                    <el-input v-model.number="form[f.envKey]" v-if="f.type == 'number'" :type="f.type"></el-input>
                    <el-input
                        v-model="form[f.envKey]"
                        v-if="f.type == 'password'"
                        :type="f.type"
                        show-password
                    ></el-input>
                    <el-select v-model="form[f.envKey]" v-if="f.type == 'service'">
                        <el-option
                            v-for="service in services"
                            :key="service.label"
                            :value="service.value"
                            :label="service.label"
                        ></el-option>
                    </el-select>
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
import { InstallApp, GetAppService } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import { getRandomStr } from '@/utils/util';
import { FormInstance, FormRules } from 'element-plus';
import { nextTick, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();

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
    NAME: [Rules.requiredInput],
});
let loading = false;
const paramForm = ref<FormInstance>();
const req = reactive({
    appDetailId: 0,
    params: {},
    name: '',
});
let services = ref();

const handleClose = () => {
    open.value = false;
    resetForm();
};

const opened = () => {
    nextTick(() => {
        if (paramForm.value) {
            paramForm.value.clearValidate();
        }
    });
};

const resetForm = () => {
    if (paramForm.value) {
        paramForm.value.resetFields();
    }
};

const acceptParams = (props: InstallRrops): void => {
    installData.value = props;
    const params = installData.value.params;
    if (params?.formFields != undefined) {
        for (const p of params?.formFields) {
            if (p.default == 'random') {
                form[p.envKey] = getRandomStr(6);
            } else {
                form[p.envKey] = p.default;
            }
            if (p.required) {
                rules[p.envKey] = [Rules.requiredInput];
            }
            if (p.key) {
                form[p.envKey] = '';
                getServices(p.envKey, p.key);
            }
        }
    }
    open.value = true;
};

const getServices = async (envKey: string, key: string | undefined) => {
    await GetAppService(key).then((res) => {
        services.value = res.data;
        if (services.value != null) {
            form[envKey] = services.value[0].value;
        }
    });
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        req.appDetailId = installData.value.appDetailId;
        req.params = form;
        req.name = form['NAME'];
        InstallApp(req).then(() => {
            handleClose();
            router.push({ path: '/apps/installed' });
        });
    });
};

defineExpose({
    acceptParams,
});
</script>
