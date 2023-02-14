<template>
    <el-drawer v-model="open" :before-close="handleClose" :close-on-click-modal="false" width="50%">
        <template #header>
            <DrawerHeader :header="$t('file.setRole')" :back="handleClose" />
        </template>

        <el-row>
            <el-col :span="22" :offset="1">
                <FileRole v-loading="loading" :mode="mode" @get-mode="getMode"></FileRole>
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

let open = ref(false);
let form = ref<File.FileCreate>({ path: '', isDir: false, mode: 0o755 });
let loading = ref<Boolean>(false);
let mode = ref('0755');

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
