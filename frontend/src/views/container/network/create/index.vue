<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <DrawerHeader :header="$t('container.createNetwork')" :back="handleClose" />
        </template>
        <el-form ref="formRef" label-position="top" v-loading="loading" :model="form" :rules="rules" label-width="80px">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('container.networkName')" prop="name">
                        <el-input clearable v-model.trim="form.name" />
                    </el-form-item>
                    <el-form-item :label="$t('container.driver')" prop="driver">
                        <el-select v-model="form.driver">
                            <el-option label="bridge" value="bridge" />
                            <el-option label="ipvlan" value="ipvlan" />
                            <el-option label="macvlan" value="macvlan" />
                            <el-option label="overlay" value="overlay" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('container.option')" prop="optionStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :autosize="{ minRows: 2, maxRows: 4 }"
                            v-model="form.optionStr"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('container.subnet')" prop="subnet">
                        <el-input clearable v-model.trim="form.subnet" />
                    </el-form-item>
                    <el-form-item :label="$t('container.gateway')" prop="gateway">
                        <el-input clearable v-model.trim="form.gateway" />
                    </el-form-item>
                    <el-form-item :label="$t('container.scope')" prop="scope">
                        <el-input clearable v-model.trim="form.scope" />
                    </el-form-item>
                    <el-form-item :label="$t('container.tag')" prop="labelStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :autosize="{ minRows: 2, maxRows: 4 }"
                            v-model="form.labelStr"
                        />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { createNetwork } from '@/api/modules/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const drawerVisiable = ref(false);
const form = reactive({
    name: '',
    labelStr: '',
    labels: [] as Array<string>,
    optionStr: '',
    options: [] as Array<string>,
    driver: '',
    subnet: '',
    gateway: '',
    scope: '',
});

const acceptParams = (): void => {
    form.name = '';
    form.labelStr = '';
    form.labels = [];
    form.optionStr = '';
    form.options = [];
    form.driver = '';
    form.subnet = '';
    form.gateway = '';
    form.scope = '';
    drawerVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisiable.value = false;
};

const rules = reactive({
    name: [Rules.requiredInput],
    driver: [Rules.requiredSelect],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.labelStr !== '') {
            form.labels = form.labelStr.split('\n');
        }
        if (form.optionStr !== '') {
            form.options = form.optionStr.split('\n');
        }
        loading.value = true;
        await createNetwork(form)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisiable.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
