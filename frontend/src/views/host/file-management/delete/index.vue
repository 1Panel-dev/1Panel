<template>
    <DialogPro
        v-model="open"
        :title="$t('commons.button.delete')"
        :show-close="!loading"
        :close-on-press-escape="!loading"
    >
        <div v-loading="loading">
            <el-row>
                <el-col :span="22" :offset="1">
                    <el-alert
                        class="mt-2"
                        :show-icon="true"
                        :type="recycleStatus === 'Enable' ? 'warning' : 'error'"
                        :closable="false"
                    >
                        <div class="delete-warn">
                            <span v-if="recycleStatus === 'Enable'">{{ $t('file.deleteHelper') }}</span>
                            <span v-else>{{ $t('file.deleteHelper2') }}</span>
                        </div>
                    </el-alert>
                    <div class="mt-4" v-if="recycleStatus === 'Enable'">
                        <el-checkbox v-model="forceDelete" class="force-delete">
                            <span>{{ $t('file.forceDeleteHelper') }}</span>
                        </el-checkbox>
                    </div>

                    <DeleteFileList :files="files" />
                </el-col>
            </el-row>
        </div>

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
import DeleteFileList from '../components/delete-file-list.vue';
import { deleteFilesWithLimit } from '@/utils/file-delete';
import { deleteFile, deleteFileByNode, getRecycleStatus, getRecycleStatusByNode } from '@/api/modules/files';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { loadBaseDir } from '@/api/modules/setting';

const open = ref(false);
const files = ref<File.File[]>([]);
const loading = ref(false);
const em = defineEmits(['close']);
const forceDelete = ref(false);
const recycleStatus = ref('Enable');
const reqNode = ref('');

const acceptParams = (props: File.File[], node = '') => {
    reqNode.value = '';
    if (node != '') {
        reqNode.value = node;
    }
    getStatus();
    files.value = [...props];
    open.value = true;
    forceDelete.value = false;
};

const getStatus = async () => {
    loading.value = true;
    try {
        let res;
        if (reqNode.value != '') {
            res = await getRecycleStatusByNode(reqNode.value);
        } else {
            res = await getRecycleStatus();
        }
        recycleStatus.value = res.data;
        if (recycleStatus.value === 'Disable') {
            forceDelete.value = true;
        }
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const onConfirm = async () => {
    if (loading.value) return;
    loading.value = true;
    const selected = [...files.value];
    const node = reqNode.value;
    const force = forceDelete.value;
    try {
        let baseDir = '';
        if (selected.some((item) => item.isDir)) {
            const pathRes = await loadBaseDir();
            baseDir = pathRes.data;
        }
        // Validate the entire selection before starting any deletion request.
        for (const file of selected) {
            if (!file.isDir) continue;
            if (file.path.includes('.1panel_clash')) {
                MsgWarning(i18n.global.t('file.clashDeleteAlert'));
                return;
            }
            if (file.path === baseDir) {
                MsgWarning(i18n.global.t('file.panelInstallDir'));
                return;
            }
        }
        files.value = await deleteFilesWithLimit(selected, (file) => {
            const params = { path: file.path, isDir: file.isDir, forceDelete: force };
            return node ? deleteFileByNode(params, node) : deleteFile(params);
        });
        if (files.value.length === 0) {
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            open.value = false;
        }
        if (files.value.length < selected.length) em('close');
    } catch {
        // The request interceptor reports errors such as a failed base-directory lookup.
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
.delete-warn {
    line-height: 20px;
    word-wrap: break-word;
}

.force-delete {
    white-space: pre-line;
    word-wrap: break-word;
    line-height: 50px;
}
</style>
