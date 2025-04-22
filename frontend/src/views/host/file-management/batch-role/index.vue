<template>
    <DrawerPro v-model="open" :header="$t('file.setRole')" @close="handleClose" size="large">
        <div v-loading="loading">
            <FileRole :mode="mode" @get-mode="getMode" :key="open.toString()"></FileRole>
            <el-form ref="fileForm" label-position="left" :model="addForm" label-width="100px" :rules="rules">
                <el-form-item :label="$t('commons.table.user')" prop="user">
                    <el-select v-model="addForm.user" @change="handleUserChange" filterable allow-create>
                        <el-option
                            v-for="item in users"
                            :key="item.username"
                            :label="item.username"
                            :value="item.username"
                        />
                    </el-select>
                </el-form-item>

                <el-form-item :label="$t('file.group')" prop="group">
                    <el-select v-model="addForm.group" filterable allow-create>
                        <el-option v-for="group in groups" :key="group" :label="group" :value="group" />
                    </el-select>
                </el-form-item>
                <el-form-item>
                    <el-checkbox v-model="addForm.sub">{{ $t('file.containSub') }}</el-checkbox>
                </el-form-item>
            </el-form>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { File } from '@/api/interface/file';
import { batchChangeRole, searchUserGroup } from '@/api/modules/files';
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
const users = ref<File.UserInfo[]>([]);
const groups = ref<string[]>([]);

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
    addForm.group = props.files[0].group || props.files[0].gid + '';
    addForm.user = props.files[0].user || props.files[0].uid + '';
    addForm.sub = true;

    mode.value = String(props.files[0].mode);
    open.value = true;
};

const getUserAndGroup = async () => {
    try {
        const res = await searchUserGroup();
        users.value = res.data.users;
        groups.value = res.data.groups;
    } catch (error) {
        console.error('Failed to fetch user and group:', error);
    }
};

const handleUserChange = (val: string) => {
    const found = users.value.find((u) => u.username === val);
    if (found) {
        addForm.group = found.group;
    }
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

    batchChangeRole(addForm)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            handleClose();
        })
        .finally(() => {
            loading.value = false;
        });
};
onMounted(() => {
    getUserAndGroup();
});

defineExpose({ acceptParams });
</script>
