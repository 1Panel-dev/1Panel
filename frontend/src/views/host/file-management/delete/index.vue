<template>
    <DialogPro v-model="open" :title="$t('commons.button.delete')" size="w-60" class="file-delete-dialog">
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

                    <div class="file-list">
                        <div class="flx-align-center mb-1" v-for="(row, index) in files" :key="index">
                            <div>
                                <svg-icon
                                    v-if="row.isDir"
                                    className="table-icon mr-1 "
                                    iconName="p-file-folder"
                                ></svg-icon>
                                <svg-icon
                                    v-else
                                    className="table-icon mr-1"
                                    :iconName="getIconName(row.extension)"
                                ></svg-icon>
                            </div>
                            <span class="dialog-object-name">{{ row.name }}</span>
                        </div>
                    </div>
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
import { getIcon } from '@/utils/file';
import { deleteFile, deleteFileByNode, getRecycleStatus, getRecycleStatusByNode } from '@/api/modules/files';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { loadBaseDir } from '@/api/modules/setting';

const open = ref(false);
const files = ref<File.File[]>([]);
const loading = ref(false);
const em = defineEmits<{ close: [files: File.File[], node: string] }>();
const forceDelete = ref(false);
const recycleStatus = ref('Enable');
const reqNode = ref('');
let dialogVersion = 0;

const acceptParams = (props: File.File[], node: string) => {
    dialogVersion++;
    reqNode.value = '';
    if (node != '') {
        reqNode.value = node;
    }
    getStatus();
    files.value = props;
    open.value = true;
    forceDelete.value = false;
};

const getStatus = async () => {
    const version = dialogVersion;
    const node = reqNode.value;
    loading.value = true;
    try {
        let res;
        if (node != '') {
            res = await getRecycleStatusByNode(node);
        } else {
            res = await getRecycleStatus();
        }
        if (version !== dialogVersion) {
            return;
        }
        recycleStatus.value = res.data;
        if (recycleStatus.value === 'Disable') {
            forceDelete.value = true;
        }
    } catch (error) {
    } finally {
        if (version === dialogVersion) {
            loading.value = false;
        }
    }
};

const onConfirm = async () => {
    const version = dialogVersion;
    const deletedFiles = files.value.map((file) => ({ ...file }));
    const node = reqNode.value;
    const force = forceDelete.value;
    loading.value = true;
    try {
        const pros = [];
        let baseDir = '';
        if (deletedFiles.some((item) => item.isDir)) {
            const pathRes = await loadBaseDir();
            baseDir = pathRes.data;
        }
        for (const s of deletedFiles) {
            if (s['isDir']) {
                if (s['path'].indexOf('.1panel_clash') > -1) {
                    MsgWarning(i18n.global.t('file.clashDeleteAlert'));
                    return;
                }
                if (s['path'] === baseDir) {
                    MsgWarning(i18n.global.t('file.panelInstallDir'));
                    return;
                }
            }
            if (node != '') {
                pros.push(deleteFileByNode({ path: s['path'], isDir: s['isDir'], forceDelete: force }, node));
            } else {
                pros.push(deleteFile({ path: s['path'], isDir: s['isDir'], forceDelete: force }));
            }
        }
        await Promise.all(pros);
        MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
        if (version === dialogVersion) {
            open.value = false;
        }
        em('close', deletedFiles, node);
    } finally {
        if (version === dialogVersion) {
            loading.value = false;
        }
    }
};

const getIconName = (extension: string) => {
    return getIcon(extension);
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
:global(.el-dialog.file-delete-dialog:not(.is-fullscreen)) {
    width: 60vw !important;
    min-width: 0;
}

.resource {
    margin-top: 10px;
    max-height: 400px;
    overflow: auto;
}

.file-list {
    max-height: 400px;
    overflow-y: auto;
    margin-top: 15px;
}

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
