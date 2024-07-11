<template>
    <el-drawer
        v-model="open"
        :destroy-on-close="true"
        :before-close="handleClose"
        size="50%"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
    >
        <template #header>
            <DrawerHeader :header="$t('file.download')" :back="handleClose" />
        </template>
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form
                    ref="fileForm"
                    label-position="top"
                    :model="addForm"
                    label-width="100px"
                    :rules="rules"
                    v-loading="loading"
                >
                    <el-form-item :label="$t('file.downloadUrl')" prop="url">
                        <el-input v-model="addForm.url" @input="getFileName" />
                    </el-form-item>
                    <el-form-item :label="$t('file.path')" prop="path">
                        <el-input v-model="addForm.path">
                            <template #prepend><FileList :path="addForm.path" @choose="getPath"></FileList></template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model="addForm.name"></el-input>
                    </el-form-item>
                    <el-form-item>
                        <el-checkbox v-model="addForm.ignoreCertificate">
                            {{ $t('file.ignoreCertificate') }}
                        </el-checkbox>
                        <span class="input-help">{{ $t('file.ignoreCertificateHelper') }}</span>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose()" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { WgetFile } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance, FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import FileList from '@/components/file-list/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

interface WgetProps {
    path: string;
}

const fileForm = ref<FormInstance>();
const loading = ref(false);
let open = ref(false);
let submitData = ref(false);

const rules = reactive<FormRules>({
    name: [Rules.requiredInput],
    path: [Rules.requiredInput],
    url: [Rules.requiredInput],
});

const addForm = reactive({
    url: '',
    path: '',
    name: '',
    ignoreCertificate: false,
});

const em = defineEmits(['close']);

const handleClose = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    open.value = false;
    em('close', submitData.value);
};

const getPath = (path: string) => {
    addForm.path = path;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        WgetFile(addForm)
            .then(() => {
                MsgSuccess(i18n.global.t('file.downloadStart'));
                submitData.value = true;
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const getFileName = (url: string) => {
    const paths = url.split('/');
    addForm.name = paths[paths.length - 1];
};

const acceptParams = (props: WgetProps) => {
    addForm.path = props.path;
    open.value = true;
    submitData.value = false;
    addForm.ignoreCertificate = false;
};

defineExpose({ acceptParams });
</script>
