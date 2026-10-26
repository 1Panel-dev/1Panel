<template>
    <DialogPro
        v-model="open"
        :title="$t('commons.button.delete')"
        :show-close="!loading"
        :close-on-press-escape="!loading"
    >
        <el-row>
            <el-col :span="20" :offset="2">
                <el-alert :title="$t('file.deleteRecycleHelper')" show-icon type="error" :closable="false"></el-alert>
                <DeleteFileList :files="files" />
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onConfirm" :loading="loading" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>
<script lang="ts" setup>
import i18n from '@/lang';
import { ref } from 'vue';
import { File } from '@/api/interface/file';
import DeleteFileList from '../../components/delete-file-list.vue';
import { deleteFilesWithLimit } from '@/utils/file-delete';
import { deleteFile } from '@/api/modules/files';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const files = ref<File.RecycleBin[]>([]);
const loading = ref(false);
const em = defineEmits(['close']);

const acceptParams = (props: File.RecycleBin[]) => {
    files.value = [...props];
    open.value = true;
};

const onConfirm = async () => {
    if (loading.value) return;
    loading.value = true;
    const selected = [...files.value];
    try {
        files.value = await deleteFilesWithLimit(selected, (file) =>
            deleteFile({ path: file.from + '/' + file.rName, isDir: file.isDir, forceDelete: true }),
        );
        if (files.value.length === 0) {
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            open.value = false;
        }
        if (files.value.length < selected.length) em('close');
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
