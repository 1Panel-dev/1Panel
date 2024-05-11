<template>
    <el-drawer v-model="open" size="40%" :close-on-click-modal="false" :close-on-press-escape="false">
        <template #header>
            <DrawerHeader :header="$t('file.changeOwner')" :resource="name" :back="handleClose" />
        </template>
        <el-row>
            <el-col :span="22" :offset="1">
                <el-alert :title="$t('file.ownerHelper')" type="info" :closable="false" class="common-prompt" />
                <el-form
                    ref="fileForm"
                    label-position="top"
                    :model="addForm"
                    label-width="100px"
                    :rules="rules"
                    v-loading="loading"
                >
                    <el-form-item :label="$t('commons.table.user')" prop="user">
                        <el-input v-model.trim="addForm.user" />
                    </el-form-item>
                    <el-form-item :label="$t('file.group')" prop="group">
                        <el-input v-model.trim="addForm.group" />
                    </el-form-item>
                    <el-form-item v-if="isDir">
                        <el-checkbox v-model="addForm.sub">{{ $t('file.containSub') }}</el-checkbox>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { ChangeOwner } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import { FormInstance, FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';

interface OwnerProps {
    path: string;
    user: string;
    group: string;
    isDir: boolean;
    name: string;
}

const fileForm = ref<FormInstance>();
const loading = ref(false);
const open = ref(false);
const isDir = ref(false);
const name = ref('');

const addForm = reactive({
    path: '',
    user: '',
    group: '',
    sub: false,
});

const rules = reactive<FormRules>({
    user: [Rules.requiredInput],
    group: [Rules.requiredInput],
});

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', false);
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        ChangeOwner(addForm)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const acceptParams = (props: OwnerProps) => {
    addForm.user = props.user;
    addForm.path = props.path;
    addForm.group = props.group;
    isDir.value = props.isDir;
    name.value = props.name;
    open.value = true;
};

defineExpose({ acceptParams });
</script>
