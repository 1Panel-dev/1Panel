<template>
    <DialogPro v-model="open" :title="$t('commons.button.delete')" size="small">
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
                            <span class="sle">{{ row.name }}</span>
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
                <el-button type="primary" @click="onConfirm" :disabled="loading">
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
import { getIcon } from '@/utils/util';
import { deleteFile, deleteFileByNode, getRecycleStatus, getRecycleStatusByNode } from '@/api/modules/files';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { loadBaseDir } from '@/api/modules/setting';

const open = ref(false);
const files = ref();
const loading = ref(false);
const em = defineEmits(['close']);
const forceDelete = ref(false);
const recycleStatus = ref('Enable');
const reqNode = ref('');

const acceptParams = (props: File.File[], node: string) => {
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
    const pros = [];
    for (const s of files.value) {
        if (s['isDir']) {
            if (s['path'].indexOf('.1panel_clash') > -1) {
                MsgWarning(i18n.global.t('file.clashDeleteAlert'));
                return;
            }
            const pathRes = await loadBaseDir();
            if (s['path'] === pathRes.data) {
                MsgWarning(i18n.global.t('file.panelInstallDir'));
                return;
            }
        }
        if (reqNode.value != '') {
            pros.push(
                deleteFileByNode({ path: s['path'], isDir: s['isDir'], forceDelete: forceDelete.value }, reqNode.value),
            );
        } else {
            pros.push(deleteFile({ path: s['path'], isDir: s['isDir'], forceDelete: forceDelete.value }));
        }
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
