<template>
    <DrawerPro v-model="drawerVisible" :header="$t('container.createVolume')" @close="handleClose">
        <el-form
            ref="formRef"
            v-loading="loading"
            label-position="top"
            :model="form"
            :rules="rules"
            label-width="80px"
            @submit.prevent
        >
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input clearable v-model.trim="form.name" />
            </el-form-item>
            <el-form-item :label="$t('container.driver')" prop="driver">
                <el-tag type="success">local</el-tag>
            </el-form-item>
            <el-form-item :label="$t('container.nfsEnable')" prop="nfsStatus">
                <el-switch v-model="form.nfsStatus" active-value="enable" inactive-value="disable" />
            </el-form-item>
            <div v-if="form.nfsStatus === 'enable'">
                <el-form-item :label="$t('container.nfsAddress')" prop="nfsAddress">
                    <el-input clearable v-model.trim="form.nfsAddress" :placeholder="$t('commons.rule.hostHelper')" />
                </el-form-item>
                <el-form-item :label="$t('app.version')" prop="nfsVersion">
                    <el-radio-group v-model="form.nfsVersion">
                        <el-radio value="v3">NFS</el-radio>
                        <el-radio value="v4">NFS4</el-radio>
                    </el-radio-group>
                </el-form-item>
                <el-form-item :label="$t('container.mountpoint')" prop="nfsMount">
                    <el-input
                        clearable
                        v-model.trim="form.nfsMount"
                        :placeholder="$t('container.mountpointNFSHelper')"
                    />
                </el-form-item>
                <el-form-item :label="$t('container.options')" prop="nfsOption">
                    <el-input clearable v-model.trim="form.nfsOption" />
                </el-form-item>
            </div>
            <el-form-item :label="$t('container.option')" prop="options">
                <InputTag class="w-full" v-model:tags="form.options" />
            </el-form-item>
            <el-form-item :label="$t('container.tag')" prop="labels">
                <InputTag class="w-full" v-model:tags="form.labels" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { createVolume } from '@/api/modules/container';
import InputTag from '@/components/input-tag/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const drawerVisible = ref(false);
const form = reactive({
    name: '',
    driver: 'local',
    labels: [] as Array<string>,
    nfsStatus: 'disable',
    nfsAddress: '',
    nfsVersion: 'v4',
    nfsMount: '',
    nfsOption: 'rw,noatime,rsize=8192,wsize=8192,tcp,timeo=14',
    options: [] as Array<string>,
});

const acceptParams = (): void => {
    form.name = '';
    form.labels = [];
    form.options = [];
    form.nfsStatus = 'disable';
    form.nfsAddress = '';
    form.nfsVersion = 'v4';
    form.nfsMount = '';
    form.nfsOption = 'rw,noatime,rsize=8192,wsize=8192,tcp,timeo=14';
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const rules = reactive({
    name: [Rules.requiredInput, Rules.volumeName],
    driver: [Rules.requiredSelect],
    nfsAddress: [Rules.host],
    nfsVersion: [Rules.requiredSelect],
    nfsMount: [Rules.requiredInput],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.nfsStatus === 'enable') {
            let typeOption = form.nfsVersion === 'v4' ? 'nfs4' : 'nfs';
            form.options.push('type=' + typeOption);
            form.options.push('o=addr=' + form.nfsAddress + ',' + form.nfsOption);
            let mount = form.nfsMount.startsWith(':') ? form.nfsMount : ':' + form.nfsMount;
            form.options.push('device=' + mount);
        }
        loading.value = true;
        await createVolume(form)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisible.value = false;
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
