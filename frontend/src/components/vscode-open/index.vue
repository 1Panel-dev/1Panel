<template>
    <div>
        <el-dialog
            v-model="dialogVisible"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <el-row>
                <el-col :span="22" :offset="1">
                    <el-alert :closable="false" :title="$t('file.vscodeHelper')" type="info"></el-alert>
                    <el-form
                        ref="vscodeConnectInfoForm"
                        label-position="top"
                        :model="addForm"
                        label-width="100px"
                        class="mt-4"
                    >
                        <el-form-item :label="$t('setting.systemIP')" prop="host">
                            <el-input v-model="addForm.host" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.port')" prop="port">
                            <el-input v-model="addForm.port" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.username')" prop="username">
                            <el-input v-model="addForm.username" />
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="submit(vscodeConnectInfoForm)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';

const dialogVisible = ref();

interface DialogProps {
    path: string;
}
const vscodeConnectInfoForm = ref<FormInstance>();

const addForm = reactive({
    host: '',
    port: 22,
    username: 'root',
    path: '',
});

const em = defineEmits(['close']);
const handleClose = () => {
    dialogVisible.value = false;
    if (vscodeConnectInfoForm.value) {
        vscodeConnectInfoForm.value.resetFields();
    }
    em('close', false);
};
const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        localStorage.setItem('VscodeConnectInfo', JSON.stringify(addForm));
        dialogVisible.value = false;
        const vscodeUrl = `vscode://vscode-remote/ssh-remote+${addForm.username}@${addForm.host}:${addForm.port}${addForm.path}?windowId=_blank`;
        window.open(vscodeUrl);
    });
};
const acceptParams = async (params: DialogProps): Promise<void> => {
    const vscodeConnectInfo = localStorage.getItem('VscodeConnectInfo');

    if (vscodeConnectInfo) {
        try {
            const storedInfo = JSON.parse(vscodeConnectInfo);
            addForm.host = storedInfo.host;
            addForm.port = storedInfo.port;
            addForm.username = storedInfo.username;
        } catch (error) {}
    }

    addForm.path = params.path;
    dialogVisible.value = true;
};

defineExpose({ acceptParams });
</script>
