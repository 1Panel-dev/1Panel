<template>
    <el-dialog v-model="open" :title="$t('app.delete')" width="30%" :close-on-click-modal="false">
        <div>
            <el-row>
                <el-col :span="22" :offset="1">
                    <el-alert
                        class="mt-2"
                        :show-icon="true"
                        :type="recycleStatus === 'enable' ? 'warning' : 'error'"
                        :closable="false"
                    >
                        <div class="delete-warn">
                            <span v-if="recycleStatus === 'enable'">{{ $t('file.deleteHelper') }}</span>
                            <span v-else>{{ $t('file.deleteHelper2') }}</span>
                        </div>
                    </el-alert>
                    <div class="mt-4" v-if="recycleStatus === 'enable'">
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
    </el-dialog>
</template>
<script lang="ts" setup>
import i18n from '@/lang';
import { ref } from 'vue';
import { File } from '@/api/interface/file';
import { getIcon } from '@/utils/util';
import { DeleteFile, GetRecycleStatus } from '@/api/modules/files';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { loadBaseDir } from '@/api/modules/setting';

const open = ref(false);
const files = ref();
const loading = ref(false);
const em = defineEmits(['close']);
const forceDelete = ref(false);
const recycleStatus = ref('enable');

const acceptParams = (props: File.File[]) => {
    getStatus();
    files.value = props;
    open.value = true;
    forceDelete.value = false;
};

const getStatus = async () => {
    try {
        const res = await GetRecycleStatus();
        recycleStatus.value = res.data;
        if (recycleStatus.value === 'disable') {
            forceDelete.value = true;
        }
    } catch (error) {}
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

        pros.push(DeleteFile({ path: s['path'], isDir: s['isDir'], forceDelete: forceDelete.value }));
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
