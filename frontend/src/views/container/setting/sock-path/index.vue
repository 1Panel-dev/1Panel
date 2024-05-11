<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('container.sockPath')" :back="handleClose" />
            </template>
            <el-form
                ref="formRef"
                label-position="top"
                :model="form"
                :rules="rules"
                @submit.prevent
                v-loading="loading"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('container.sockPath')" prop="dockerSockPath">
                            <el-input v-model="form.dockerSockPath">
                                <template #prepend>unix://</template>
                                <template #append>
                                    <FileList @choose="loadBuildDir"></FileList>
                                </template>
                            </el-input>
                            <span class="input-help">{{ $t('container.sockPathHelper1') }}</span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { updateSetting } from '@/api/modules/setting';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { ElMessageBox, FormInstance } from 'element-plus';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    dockerSockPath: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    dockerSockPath: '',
    currentPath: '',
});
const formRef = ref<FormInstance>();
const rules = reactive({
    dockerSockPath: [{ required: true, validator: checkSockPath, trigger: 'blur' }],
});

function checkSockPath(rule: any, value: any, callback: any) {
    if (!value.endsWith('.sock')) {
        return callback(new Error(i18n.global.t('container.sockPathErr')));
    }
    callback();
}

const acceptParams = (params: DialogProps): void => {
    form.dockerSockPath = params.dockerSockPath.replaceAll('unix://', '');
    form.currentPath = params.dockerSockPath.replaceAll('unix://', '');
    drawerVisible.value = true;
};

const loadBuildDir = async (path: string) => {
    form.dockerSockPath = path;
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('container.sockPathMsg'), i18n.global.t('container.sockPath'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            loading.value = true;
            let params = {
                key: 'DockerSockPath',
                value: form.dockerSockPath.startsWith('unix://')
                    ? form.dockerSockPath
                    : 'unix://' + form.dockerSockPath,
            };
            await updateSetting(params)
                .then(() => {
                    loading.value = false;
                    handleClose();
                    emit('search');
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
