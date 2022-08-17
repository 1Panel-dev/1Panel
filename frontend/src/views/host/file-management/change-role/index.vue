<template>
    <el-dialog v-model="open" :before-close="handleClose" :title="$t('file.setRole')" width="30%" @open="onOpen">
        <FileRole v-loading="loading" :mode="mode" @get-mode="getMode"></FileRole>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { File } from '@/api/interface/file';
import { ChangeFileMode } from '@/api/modules/files';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import FileRole from '@/components/file-role/index.vue';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    file: {
        type: Object,
        default: function () {
            return {};
        },
    },
});

let form = ref<File.FileCreate>({ path: '', isDir: false, mode: 0o755 });
let loading = ref<Boolean>(false);
let mode = ref('0755');

const em = defineEmits(['close']);
const handleClose = () => {
    em('close', false);
};

const onOpen = () => {
    const f = props.file as File.FileCreate;
    form.value.isDir = f.isDir;
    form.value.path = f.path;
    mode.value = String(f.mode);
};

const getMode = (val: number) => {
    form.value.mode = val;
};

const submit = async () => {
    loading.value = true;
    ChangeFileMode(form.value)
        .then(() => {
            ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
            handleClose();
        })
        .finally(() => {
            loading.value = false;
        });
};
</script>
