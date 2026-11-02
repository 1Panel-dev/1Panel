<template>
    <DialogPro v-model="open" :title="$t('commons.button.delete')" size="w-60" class="recycle-file-dialog">
        <el-row>
            <el-col :span="20" :offset="2">
                <el-alert :title="$t('file.deleteRecycleHelper')" show-icon type="error" :closable="false"></el-alert>
                <div class="flx-align-center mb-1 mt-1" v-for="(row, index) in files" :key="index">
                    <div>
                        <svg-icon v-if="row.isDir" className="table-icon mr-1 " iconName="p-file-folder"></svg-icon>
                        <svg-icon v-else className="table-icon mr-1" :iconName="getIconName(row.extension)"></svg-icon>
                    </div>
                    <span class="dialog-object-name">{{ row.name }}</span>
                </div>
            </el-col>
        </el-row>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onConfirm" v-loading="loading">
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
import { getIcon } from '@/utils/file';
import { deleteFile } from '@/api/modules/files';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const files = ref();
const loading = ref(false);
const em = defineEmits(['close']);
const forceDelete = ref(false);

const acceptParams = (props: File.RecycleBin[]) => {
    files.value = props;
    open.value = true;
    forceDelete.value = false;
};

const onConfirm = () => {
    const pros = [];
    for (const s of files.value) {
        pros.push(deleteFile({ path: s.from + '/' + s.rName, isDir: s.isDir, forceDelete: true }));
    }
    loading.value = true;
    Promise.all(pros)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            open.value = false;
            em('close');
        })
        .finally(() => {
            loading.value = false;
        });
};

const getIconName = (extension: string) => {
    return getIcon(extension);
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
:global(.el-dialog.recycle-file-dialog:not(.is-fullscreen)) {
    width: 60vw !important;
    min-width: 0;
}

.resource {
    margin-top: 10px;
    max-height: 400px;
    overflow: auto;
}
</style>
