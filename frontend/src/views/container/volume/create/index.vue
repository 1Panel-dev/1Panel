<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="30%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.createVolume')" :back="handleClose" />
        </template>
        <el-row type="flex" justify="center">
            <el-col :span="22">
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
                            <el-input
                                clearable
                                v-model.trim="form.nfsAddress"
                                :placeholder="$t('commons.rule.hostHelper')"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('container.version')" prop="nfsVersion">
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
                    <el-form-item :label="$t('container.option')" prop="optionStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :rows="3"
                            v-model="form.optionStr"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('container.tag')" prop="labelStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :rows="3"
                            v-model="form.labelStr"
                        />
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
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
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { createVolume } from '@/api/modules/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const drawerVisible = ref(false);
const form = reactive({
    name: '',
    driver: 'local',
    labelStr: '',
    labels: [] as Array<string>,
    nfsStatus: 'disable',
    nfsAddress: '',
    nfsVersion: 'v4',
    nfsMount: '',
    nfsOption: 'rw,noatime,rsize=8192,wsize=8192,tcp,timeo=14',
    optionStr: '',
    options: [] as Array<string>,
});

const acceptParams = (): void => {
    form.name = '';
    form.labels = [];
    form.labelStr = '';
    form.options = [];
    form.optionStr = '';
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
        if (form.labelStr !== '') {
            form.labels = form.labelStr.split('\n');
        }
        if (form.optionStr !== '') {
            form.options = form.optionStr.split('\n');
        }
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
