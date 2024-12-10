<template>
    <el-drawer
        v-model="open"
        :before-close="handleClose"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('file.editPermissions')" :back="handleClose" />
        </template>

        <el-row>
            <el-col :span="22" :offset="1" v-loading="loading">
                <FileRole :mode="mode" @get-mode="getMode" :key="open.toString()"></FileRole>
                <el-form ref="fileForm" label-position="left" :model="addForm" label-width="100px" :rules="rules">
                    <el-form-item :label="$t('commons.table.user')" prop="user">
                        <el-input v-model.trim="addForm.user" />
                    </el-form-item>
                    <el-form-item :label="$t('file.group')" prop="group">
                        <el-input v-model.trim="addForm.group" />
                    </el-form-item>
                    <el-form-item>
                        <el-checkbox v-model="addForm.sub">{{ $t('file.containSub') }}</el-checkbox>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script setup lang="ts">
import DrawerHeader from '@/components/drawer-header/index.vue';
import { reactive, ref } from 'vue';
import { File } from '@/api/interface/file';
import { BatchChangeRole } from '@/api/modules/files';
import i18n from '@/lang';
import FileRole from '@/components/file-role/index.vue';
import { MsgSuccess } from '@/utils/message';
import { FormRules } from 'element-plus';
import { Rules } from '@/global/form-rules';

interface BatchRoleProps {
    files: File.File[];
}

const open = ref(false);
const loading = ref(false);
const mode = ref('0755');
const files = ref<File.File[]>([]);

const rules = reactive<FormRules>({
    user: [Rules.requiredInput],
    group: [Rules.requiredInput],
});

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    em('close', false);
};

const addForm = reactive({
    paths: [],
    mode: 755,
    user: '',
    group: '',
    sub: false,
});

const acceptParams = (props: BatchRoleProps) => {
    addForm.paths = [];
    files.value = props.files;
    files.value.forEach((file) => {
        addForm.paths.push(file.path);
    });
    addForm.mode = Number.parseInt(String(props.files[0].mode), 8);
    addForm.group = props.files[0].group || props.files[0].gid;
    addForm.user = props.files[0].user || props.files[0].uid;
    addForm.sub = true;

    mode.value = String(props.files[0].mode);
    open.value = true;
};

const getMode = (val: number) => {
    addForm.mode = val;
};

const submit = async () => {
    const regFilePermission = /^[0-7]{3,4}$/;
    if (!regFilePermission.test(addForm.mode.toString(8))) {
        return;
    }
    loading.value = true;

    BatchChangeRole(addForm)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            handleClose();
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({ acceptParams });
</script>
