<template>
    <el-drawer
        v-model="open"
        :before-close="handleClose"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('file.editPermissions')" :resource="name" :back="handleClose" />
        </template>

        <el-row>
            <el-col :span="22" :offset="1">
                <FileRole v-loading="loading" :mode="mode" @get-mode="getMode" :key="open.toString()"></FileRole>
                <el-form-item v-if="form.isDir">
                    <el-checkbox v-model="form.sub">{{ $t('file.containSub') }}</el-checkbox>
                </el-form-item>
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
import { ref } from 'vue';
import { File } from '@/api/interface/file';
import { ChangeFileMode } from '@/api/modules/files';
import i18n from '@/lang';
import FileRole from '@/components/file-role/index.vue';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const form = ref<File.FileCreate>({ path: '', isDir: false, mode: 0o755 });
const loading = ref(false);
const mode = ref('0755');
const name = ref('');

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = (create: File.FileCreate) => {
    open.value = true;
    form.value.isDir = create.isDir;
    form.value.path = create.path;
    form.value.isLink = false;
    form.value.sub = false;
    name.value = create.name;

    mode.value = String(create.mode);
};

const getMode = (val: number) => {
    form.value.mode = val;
};

const submit = async () => {
    loading.value = true;
    ChangeFileMode(form.value)
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
