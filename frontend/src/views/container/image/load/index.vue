<template>
    <DrawerPro
        v-model="loadVisible"
        :header="$t('container.importImage')"
        :auto-close="!busy"
        :confirm-before-close="true"
        @before-close="beforeClose"
        size="small"
    >
        <el-form @submit.prevent label-position="top">
            <el-alert v-if="taskStatusUnknown" type="warning" :closable="false" class="mb-5">
                <p>{{ $t('container.imageImportUnknown') }}</p>
                <el-button link type="primary" @click="retryTaskStatus">
                    {{ $t('commons.button.retry') }}
                </el-button>
            </el-alert>
            <div class="flex items-center gap-3 mb-5">
                <el-upload
                    ref="uploadRef"
                    :limit="1"
                    :auto-upload="false"
                    :show-file-list="false"
                    :disabled="selectionLocked"
                    :on-change="onFileChange"
                    :on-exceed="onFileExceed"
                >
                    <el-button :disabled="selectionLocked" @click="source = 'local'">
                        {{ $t('database.localUpload') }}
                    </el-button>
                </el-upload>
                <el-button :disabled="busy" @click="onServerSelect">
                    {{ $t('database.hostSelect') }}
                </el-button>
            </div>
            <el-form-item v-if="source === 'server'" :label="$t('container.path')">
                <el-input v-model="form.path" :disabled="busy" @input="handlePathChange" />
            </el-form-item>
            <template v-else>
                <el-upload
                    v-if="!uploadedPath"
                    ref="dragUploadRef"
                    drag
                    :limit="1"
                    :auto-upload="false"
                    :show-file-list="false"
                    :disabled="selectionLocked"
                    :on-change="onFileChange"
                    :on-exceed="onFileExceed"
                >
                    <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
                    <div class="el-upload__text">{{ $t('container.imageUploadDrop') }}</div>
                </el-upload>
                <p class="input-help">{{ $t('container.imageUploadHelper') }}</p>
                <template v-if="localFile">
                    <p class="break-all">{{ localFile.name }} ({{ computeSize(localFile.size) }})</p>
                    <el-progress
                        :percentage="uploadPercent"
                        :status="uploadFailed ? 'exception' : uploadedPath ? 'success' : undefined"
                    >
                        <template v-if="uploadFailed" #default>
                            <el-button
                                link
                                type="danger"
                                icon="CircleClose"
                                :title="$t('commons.button.delete')"
                                :aria-label="$t('commons.button.delete')"
                                :disabled="busy || !!pendingTaskID"
                                @click="removeUpload"
                            />
                        </template>
                    </el-progress>
                    <p v-if="uploadedPath" class="input-help break-all">{{ uploadedPath }}</p>
                    <el-button v-if="uploadFailed" :disabled="busy" @click="uploadLocalFile">
                        {{ $t('commons.button.retry') }}
                    </el-button>
                    <el-button v-if="uploadedPath" :disabled="busy || !!pendingTaskID" @click="removeUpload">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
            </template>
        </el-form>
        <template #footer>
            <el-button :disabled="busy" @click="loadVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="!canImport" :loading="loading" type="primary" @click="onSubmit">
                {{ $t('commons.button.import') }}
            </el-button>
        </template>
    </DrawerPro>
    <FileList v-if="loadVisible" ref="fileRef" @choose="loadLoadDir" />
    <TaskLog ref="taskLogRef" width="70%" @close="emit('search')" />
</template>

<script lang="ts" setup>
import FileList from '@/components/file-list/index.vue';
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { genFileId, type UploadFile, type UploadInstance, type UploadProps, type UploadRawFile } from 'element-plus';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';
import { imageLoad } from '@/api/modules/container';
import { chunkUploadFileData, deleteFileByNode, stopChunkUpload } from '@/api/modules/files';
import { searchTasks } from '@/api/modules/log';
import { loadBaseDir } from '@/api/modules/setting';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgSuccess } from '@/utils/message';
import { computeSize } from '@/utils/size';
import { newUUID } from '@/utils/id';
import { CHUNK_SIZE, uploadChunkWithRetry } from '@/utils/chunk-upload';

const { currentNode } = useGlobalStore();
const loading = ref(false);
const uploading = ref(false);
const busy = computed(() => loading.value || uploading.value);
const fileRef = ref();
const taskLogRef = ref();
const uploadRef = ref<UploadInstance>();
const dragUploadRef = ref<UploadInstance>();
const loadVisible = ref(false);
const source = ref<'server' | 'local'>('local');
const operateNode = ref('');
const form = reactive({ path: '', paths: [] as string[] });
const localFile = ref<UploadRawFile>();
const uploadPercent = ref(0);
const uploadFailed = ref(false);
const uploadedPath = ref('');
const uploadDir = ref('');
const pendingTaskID = ref('');
const taskStatusUnknown = ref(false);
const selectionLocked = computed(() => busy.value || (!!pendingTaskID.value && !taskStatusUnknown.value));
let pendingTaskNode = '';
let taskPollFailures = 0;
let uploadID = '';
let uploadController: AbortController | undefined;
let taskTimer: ReturnType<typeof setTimeout> | undefined;
let disposed = false;
const emit = defineEmits<{ (e: 'search'): void }>();
const canImport = computed(
    () =>
        !busy.value && !pendingTaskID.value && (source.value === 'local' ? !!uploadedPath.value : !!form.paths.length),
);

const resetLocalFile = () => {
    localFile.value = undefined;
    uploadedPath.value = '';
    uploadDir.value = '';
    uploadPercent.value = 0;
    uploadFailed.value = false;
    uploadRef.value?.clearFiles();
    dragUploadRef.value?.clearFiles();
};
const acceptParams = () => {
    if (busy.value) return;
    if (operateNode.value !== currentNode.value) {
        clearPendingTask();
        resetLocalFile();
    }
    operateNode.value = currentNode.value;
    source.value = 'local';
    form.path = '';
    form.paths = [];
    loadVisible.value = true;
};
const beforeClose = (done: () => void) => {
    if (!busy.value) done();
};
const onServerSelect = () => {
    if (busy.value) return;
    if (taskStatusUnknown.value) releaseUnknownTask();
    source.value = 'server';
    fileRef.value.acceptParams({ dir: false, multiple: true });
};
const onFileChange = (file: UploadFile) => {
    if (!file.raw || selectionLocked.value) return;
    if (taskStatusUnknown.value) releaseUnknownTask();
    source.value = 'local';
    uploadedPath.value = '';
    uploadDir.value = '';
    localFile.value = file.raw;
    void uploadLocalFile();
};
const onFileExceed: UploadProps['onExceed'] = (files) => {
    if (selectionLocked.value || !files.length) return;
    uploadRef.value?.clearFiles();
    dragUploadRef.value?.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value?.handleStart(file);
};

const uploadLocalFile = async () => {
    const file = localFile.value;
    if (!file || busy.value) return;
    uploading.value = true;
    uploadFailed.value = false;
    uploadPercent.value = 0;
    const node = operateNode.value;
    const id = newUUID();
    uploadID = id;
    const controller = new AbortController();
    uploadController = controller;
    try {
        const { data: baseDir } = await loadBaseDir(node);
        if (controller.signal.aborted) return;
        const directory = `${baseDir}/uploads/image/${id}`;
        const count = Math.max(1, Math.ceil(file.size / CHUNK_SIZE));
        for (let index = 0; index < count; index++) {
            const data = new FormData();
            data.append('filename', file.name);
            data.append('path', directory);
            data.append('uploadID', id);
            data.append('chunk', file.slice(index * CHUNK_SIZE, (index + 1) * CHUNK_SIZE));
            data.append('chunkIndex', String(index));
            data.append('chunkCount', String(count));
            // The resumable protocol requires a nonempty file. Keep empty files on
            // the legacy single-chunk path so Docker can report the invalid archive.
            if (file.size > 0) {
                data.append('offset', String(index * CHUNK_SIZE));
                data.append('fileSize', String(file.size));
            }
            const config = {
                headers: { CurrentNode: node },
                signal: controller.signal,
                timeout: 0,
                onUploadProgress: (event: { total?: number; loaded: number }) => {
                    const fraction = event.total ? Math.min(1, event.loaded / event.total) : 0;
                    uploadPercent.value = Math.min(99, Math.floor(((index + fraction) * 100) / count));
                },
            };
            if (file.size > 0) {
                await uploadChunkWithRetry(data, config, controller.signal);
            } else {
                await chunkUploadFileData(data, config);
            }
        }
        uploadDir.value = directory;
        uploadedPath.value = `${directory}/${file.name}`;
        uploadPercent.value = 100;
        MsgSuccess(i18n.global.t('file.uploadSuccess'));
    } catch {
        if (!controller.signal.aborted) uploadFailed.value = true;
        await stopChunkUpload(id, node).catch(() => {});
    } finally {
        uploading.value = false;
        uploadController = undefined;
        uploadID = '';
    }
};

const removeUpload = async () => {
    if (busy.value || pendingTaskID.value || !localFile.value) return;
    loading.value = true;
    try {
        if (uploadDir.value) {
            await deleteFileByNode({ path: uploadDir.value, isDir: true, forceDelete: true }, operateNode.value);
        }
        resetLocalFile();
        source.value = 'local';
        form.path = '';
        form.paths = [];
    } finally {
        loading.value = false;
    }
};

const clearPendingTask = () => {
    clearTimeout(taskTimer);
    pendingTaskID.value = '';
    taskStatusUnknown.value = false;
    taskPollFailures = 0;
};
const releaseUnknownTask = () => {
    // Retain the previous upload on the server: its task may still be using it.
    clearPendingTask();
    resetLocalFile();
};
const pollTask = async (taskID: string, node: string) => {
    if (disposed || pendingTaskID.value !== taskID) return;
    let status: string | undefined;
    try {
        const { data } = await searchTasks({ page: 1, pageSize: 1, type: '', status: '', taskID }, node);
        status = data.items?.[0]?.status;
    } catch {
        // Retain the upload while the task status is unknown.
    }
    if (disposed || pendingTaskID.value !== taskID) return;
    if (status === 'Success' || status === 'Failed' || status === 'Canceled') {
        clearPendingTask();
        if (currentNode.value === node) emit('search');
        return;
    }
    taskPollFailures = status === 'Executing' ? 0 : taskPollFailures + 1;
    if (taskPollFailures >= 5) {
        taskStatusUnknown.value = true;
        return;
    }
    const delay = taskPollFailures ? Math.min(30000, 3000 * 2 ** taskPollFailures) : 10000;
    taskTimer = setTimeout(() => void pollTask(taskID, node), delay);
};
const retryTaskStatus = () => {
    if (!pendingTaskID.value || !taskStatusUnknown.value) return;
    taskStatusUnknown.value = false;
    taskPollFailures = 0;
    void pollTask(pendingTaskID.value, pendingTaskNode);
};
const onSubmit = async () => {
    if (!canImport.value) return;
    loading.value = true;
    const taskID = newUUID();
    const node = operateNode.value;
    try {
        await imageLoad({ paths: source.value === 'local' ? [uploadedPath.value] : form.paths, taskID }, node);
        if (disposed) return;
        clearPendingTask();
        pendingTaskID.value = taskID;
        pendingTaskNode = node;
        loadVisible.value = false;
        taskLogRef.value.openWithTaskID(taskID, true, node);
        MsgSuccess(i18n.global.t('container.imageImportSubmitted'));
        void pollTask(taskID, node);
    } finally {
        loading.value = false;
    }
};
const handlePathChange = () => {
    form.paths = [
        ...new Set(
            form.path
                .split(';')
                .map((path) => path.trim())
                .filter(Boolean),
        ),
    ];
};
const loadLoadDir = (paths: string | string[]) => {
    if (currentNode.value !== operateNode.value) return;
    source.value = 'server';
    form.paths = [...new Set(Array.isArray(paths) ? paths : [paths])];
    form.path = form.paths.join('; ');
};
watch(currentNode, () => {
    loadVisible.value = false;
});
onBeforeUnmount(() => {
    disposed = true;
    clearTimeout(taskTimer);
    uploadController?.abort();
    if (uploadID) void stopChunkUpload(uploadID, operateNode.value).catch(() => {});
});
defineExpose({ acceptParams });
</script>
