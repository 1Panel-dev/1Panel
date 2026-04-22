<template>
    <DrawerPro
        v-model="visible"
        :header="$t('home.dir')"
        :resource="title"
        size="large"
        :fullScreen="true"
        @close="visible = false"
    >
        <template #content>
            <el-form label-position="top" @submit.prevent>
                <el-form-item>
                    <div v-if="!pathEditing" class="path-breadcrumb" @click="enablePathEditing">
                        <el-breadcrumb separator="/">
                            <el-breadcrumb-item>
                                <el-link type="primary" @click.stop="navigateToPath('/')">
                                    <el-icon><HomeFilled /></el-icon>
                                </el-link>
                            </el-breadcrumb-item>
                            <el-breadcrumb-item v-for="item in pathSegments" :key="item.path">
                                <el-link type="primary" @click.stop="navigateToPath(item.path)">
                                    {{ item.name }}
                                </el-link>
                            </el-breadcrumb-item>
                        </el-breadcrumb>
                    </div>
                    <el-input
                        v-else
                        ref="pathInputRef"
                        v-model="pathInput"
                        class="path-breadcrumb path-input"
                        @blur="cancelPathEditing"
                        @keyup.enter.prevent="submitPathInput"
                    />
                    <el-upload ref="uploadRef" :auto-upload="false" :show-file-list="false" :on-change="onUploadChange">
                        <el-button class="mt-2" :loading="uploading" type="primary" plain>
                            {{ $t('commons.button.upload') }}
                        </el-button>
                    </el-upload>
                    <el-button class="mt-2 ml-3" plain :disabled="selectedRows.length === 0" @click="onBatchDelete">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                    <el-button class="mt-2" plain @click="loadContainerFiles">
                        {{ $t('commons.button.refresh') }}
                    </el-button>
                </el-form-item>
                <ComplexTable
                    :data="containerFiles"
                    size="small"
                    max-height="calc(100vh - 240px)"
                    v-model:selects="selectedRows"
                >
                    <el-table-column type="selection" width="48" />
                    <el-table-column :label="$t('commons.table.name')" min-width="260" show-overflow-tooltip>
                        <template #default="{ row }">
                            <div class="name-cell">
                                <el-icon>
                                    <FolderOpened v-if="row.isDir" />
                                    <Document v-else />
                                </el-icon>
                                <el-link v-if="row.isDir" type="primary" @click="enterDir(row.path)">
                                    {{ row.name }}
                                </el-link>
                                <el-link v-else type="primary" @click="onOpenFile(row.path)">
                                    {{ row.name }}
                                </el-link>
                                <span v-if="row.isLink && row.linkTo" class="symlink-target">-> {{ row.linkTo }}</span>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.size')" width="130">
                        <template #default="{ row }">
                            <span v-if="!row.isDir">{{ formatSize(row.size) }}</span>
                            <el-button
                                v-else-if="!row.sizeLoaded"
                                link
                                type="primary"
                                :loading="row.sizeLoading"
                                @click="onCalcDirSize(row)"
                            >
                                {{ $t('file.calculate') }}
                            </el-button>
                            <span v-else>{{ formatSize(row.size) }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.operate')" width="220">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="onDownloadFile(row)">
                                {{ $t('commons.button.download') }}
                            </el-button>
                            <el-button type="primary" link @click="onDeleteFile(row)">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>
            </el-form>
        </template>
    </DrawerPro>
    <DialogPro v-model="previewVisible" :title="previewTitle" size="w-70" :close-on-click-modal="true">
        <el-alert
            v-if="previewTruncated"
            :title="$t('file.previewTruncated')"
            type="warning"
            :closable="false"
            show-icon
            class="mb-2"
        />
        <pre class="preview-content">{{ previewContent }}</pre>
    </DialogPro>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref } from 'vue';
import {
    deleteContainerFile,
    downloadContainerFile,
    getContainerFileContent,
    getContainerFileSize,
    listContainerFiles,
    uploadContainerFile,
} from '@/api/modules/container';
import { MsgError, MsgSuccess, MsgWarning } from '@/utils/message';
import i18n from '@/lang';
import { ElInput, ElMessageBox, UploadFile, UploadInstance } from 'element-plus';
import { Document, FolderOpened, HomeFilled } from '@element-plus/icons-vue';
import { computeSize2 } from '@/utils/size';
const visible = ref(false);
const title = ref('');
const containerID = ref('');
const filePath = ref('/');
const containerFiles = ref<any[]>([]);
const uploadRef = ref<UploadInstance>();
const pathInputRef = ref<InstanceType<typeof ElInput>>();
const uploading = ref(false);
const previewVisible = ref(false);
const previewTitle = ref('');
const previewContent = ref('');
const previewTruncated = ref(false);
const selectedRows = ref<any[]>([]);
const pathEditing = ref(false);
const pathInput = ref('/');
const pathSegments = computed(() => {
    const parts = filePath.value.split('/').filter((item) => item);
    return parts.map((name, index) => ({
        name,
        path: '/' + parts.slice(0, index + 1).join('/'),
    }));
});

interface DrawerProps {
    containerID: string;
    title: string;
    workingDir?: string;
}

const acceptParams = async (params: DrawerProps): Promise<void> => {
    visible.value = true;
    containerID.value = params.containerID;
    title.value = params.title;
    filePath.value = params.workingDir || '/';
    pathInput.value = filePath.value;
    pathEditing.value = false;
    await loadContainerFiles();
};

const fetchContainerFiles = async (path: string) => {
    if (!containerID.value || !path) {
        return false;
    }
    return await listContainerFiles({
        containerID: containerID.value,
        path,
    })
        .then((res) => {
            containerFiles.value = (res.data || []).map((item) => ({
                ...item,
                sizeLoaded: !item.isDir,
                sizeLoading: false,
            }));
            selectedRows.value = [];
            return true;
        })
        .catch(() => {
            containerFiles.value = [];
            return false;
        });
};

const loadContainerFiles = async () => {
    await fetchContainerFiles(filePath.value);
};

const enterDir = async (path: string) => {
    filePath.value = path;
    await loadContainerFiles();
};

const navigateToPath = async (path: string) => {
    filePath.value = path || '/';
    pathInput.value = filePath.value;
    pathEditing.value = false;
    await loadContainerFiles();
};

const getParentPath = (path: string) => {
    if (!path || path === '/') {
        return '/';
    }
    const segments = path.split('/').filter((item) => item);
    if (segments.length <= 1) {
        return '/';
    }
    return '/' + segments.slice(0, -1).join('/');
};

const enablePathEditing = async () => {
    pathInput.value = filePath.value || '/';
    pathEditing.value = true;
    await nextTick();
    pathInputRef.value?.focus();
};

const cancelPathEditing = () => {
    pathEditing.value = false;
    pathInput.value = filePath.value || '/';
};

const submitPathInput = async () => {
    let targetPath = (pathInput.value || '').trim();
    if (!targetPath) {
        targetPath = '/';
    }
    if (!targetPath.startsWith('/')) {
        targetPath = '/' + targetPath;
    }
    let currentPath = targetPath;
    while (true) {
        const success = await fetchContainerFiles(currentPath);
        if (success) {
            filePath.value = currentPath;
            pathInput.value = currentPath;
            pathEditing.value = false;
            return;
        }
        if (currentPath === '/') {
            filePath.value = '/';
            pathInput.value = '/';
            pathEditing.value = false;
            return;
        }
        currentPath = getParentPath(currentPath);
    }
};

const onUploadChange = async (uploadFile: UploadFile) => {
    if (!uploadFile.raw) {
        return;
    }
    const formData = new FormData();
    formData.append('containerID', containerID.value);
    formData.append('path', filePath.value);
    formData.append('file', uploadFile.raw);
    uploading.value = true;
    await uploadContainerFile(formData)
        .then(async () => {
            MsgSuccess(i18n.global.t('file.uploadSuccess'));
            await loadContainerFiles();
        })
        .catch((err) => {
            MsgError(err?.message || i18n.global.t('commons.msg.operationFailed'));
        })
        .finally(() => {
            uploading.value = false;
            uploadRef.value?.clearFiles();
        });
};

const onOpenFile = async (path: string) => {
    await getContainerFileContent({
        containerID: containerID.value,
        path,
    })
        .then((res) => {
            if (res.data?.isBinary) {
                MsgWarning(i18n.global.t('file.fileCanNotRead'));
                return;
            }
            previewTitle.value = path;
            previewContent.value = res.data?.content || '';
            previewTruncated.value = !!res.data?.truncated;
            previewVisible.value = true;
        })
        .catch((err) => {
            MsgError(err?.message || i18n.global.t('commons.msg.operationFailed'));
        });
};

const onCalcDirSize = async (row: any) => {
    row.sizeLoading = true;
    await getContainerFileSize({
        containerID: containerID.value,
        path: row.path,
    })
        .then((res) => {
            row.size = res.data || 0;
            row.sizeLoaded = true;
        })
        .catch((err) => {
            MsgError(err?.message || i18n.global.t('commons.msg.operationFailed'));
        })
        .finally(() => {
            row.sizeLoading = false;
        });
};

const formatSize = (size: number) => {
    return computeSize2(size || 0);
};

const onDownloadFile = async (row: any) => {
    const blob = await downloadContainerFile({
        containerID: containerID.value,
        path: row.path,
    });
    const downloadUrl = window.URL.createObjectURL(new Blob([blob]));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    const fileName = row.path.split('/').pop() || 'container-file';
    a.download = row.isDir ? `${fileName}.tar.gz` : fileName;
    a.dispatchEvent(new MouseEvent('click'));
};

const onDeleteFile = async (row: any) => {
    try {
        await ElMessageBox.confirm(i18n.global.t('file.deleteHelper2'), i18n.global.t('commons.button.delete'), {
            type: 'warning',
        });
    } catch {
        return;
    }
    await deleteContainerFile({
        containerID: containerID.value,
        paths: [row.path],
    })
        .then(async () => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            await loadContainerFiles();
        })
        .catch((err) => {
            MsgError(err?.message || i18n.global.t('commons.msg.operationFailed'));
        });
};

const onBatchDelete = async () => {
    if (selectedRows.value.length === 0) {
        return;
    }
    try {
        await ElMessageBox.confirm(i18n.global.t('file.deleteHelper2'), i18n.global.t('commons.button.delete'), {
            type: 'warning',
        });
    } catch {
        return;
    }
    await deleteContainerFile({
        containerID: containerID.value,
        paths: selectedRows.value.map((item) => item.path),
    })
        .then(async () => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            selectedRows.value = [];
            await loadContainerFiles();
        })
        .catch((err) => {
            MsgError(err?.message || i18n.global.t('commons.msg.operationFailed'));
        });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.name-cell {
    display: flex;
    align-items: center;
    gap: 6px;
}

.symlink-target {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.path-breadcrumb {
    width: 100%;
    padding: 6px 10px;
    border: 1px solid var(--el-border-color);
    border-radius: 4px;
    overflow-x: auto;
    cursor: text;
}

.path-input {
    overflow: visible;
}

.preview-content {
    margin: 0;
    max-height: 70vh;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-word;
}
</style>
