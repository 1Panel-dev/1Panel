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
            <div class="flex items-center gap-3 mb-5">
                <el-button :disabled="busy" @click="source = 'local'">
                    {{ $t('database.localUpload') }}
                </el-button>
                <el-button :disabled="busy" @click="onServerSelect">
                    {{ $t('database.hostSelect') }}
                </el-button>
            </div>
            <el-form-item v-if="source === 'server'" :label="$t('container.path')">
                <el-input v-model="form.path" :disabled="busy" @input="handlePathChange" />
            </el-form-item>
            <template v-else>
                <el-upload
                    ref="uploadRef"
                    drag
                    :limit="1"
                    :auto-upload="false"
                    :show-file-list="false"
                    :disabled="busy"
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
                    />
                    <p v-if="uploadedPath" class="input-help break-all">{{ uploadedPath }}</p>
                    <el-button v-if="uploadFailed" :disabled="busy" @click="uploadLocalFile">
                        {{ $t('commons.button.retry') }}
                    </el-button>
                </template>
            </template>
            <el-button v-if="submittedTaskID" link type="primary" @click="openTaskLog">
                {{ $t('commons.button.log') }}
            </el-button>
        </el-form>
        <template #footer>
            <el-button :disabled="busy" @click="loadVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                v-if="source === 'server'"
                :disabled="busy || !!submittedTaskID || !form.paths.length"
                :loading="loading"
                type="primary"
                @click="onSubmit"
            >
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
import { stopChunkUpload } from '@/api/modules/files';
import { loadBaseDir } from '@/api/modules/setting';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgError, MsgSuccess } from '@/utils/message';
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
const loadVisible = ref(false);
const source = ref<'server' | 'local'>('local');
const operateNode = ref('');
const form = reactive({ path: '', paths: [] as string[] });
const localFile = ref<UploadRawFile>();
const uploadPercent = ref(0);
const uploadFailed = ref(false);
const uploadedPath = ref('');
const submittedTaskID = ref('');
let uploadID = '';
let uploadController: AbortController | undefined;
let disposed = false;
const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = () => {
    if (busy.value) return;
    operateNode.value = currentNode.value;
    source.value = 'local';
    form.path = '';
    form.paths = [];
    localFile.value = undefined;
    uploadedPath.value = '';
    uploadPercent.value = 0;
    uploadFailed.value = false;
    submittedTaskID.value = '';
    uploadRef.value?.clearFiles();
    loadVisible.value = true;
};
const beforeClose = (done: () => void) => {
    if (!busy.value) done();
};
const onServerSelect = () => {
    if (busy.value) return;
    source.value = 'server';
    fileRef.value.acceptParams({ dir: false, multiple: true });
};
const onFileChange = (file: UploadFile) => {
    if (!file.raw || busy.value) return;
    if (file.raw.size === 0) {
        MsgError(i18n.global.t('container.imageUploadEmpty'));
        uploadRef.value?.clearFiles();
        return;
    }
    localFile.value = file.raw;
    uploadedPath.value = '';
    submittedTaskID.value = '';
    void uploadLocalFile();
};
const onFileExceed: UploadProps['onExceed'] = (files) => {
    if (busy.value || !files.length) return;
    uploadRef.value?.clearFiles();
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
        controller.signal.throwIfAborted();
        const directory = `${baseDir}/uploads/image/${id}`;
        const count = Math.ceil(file.size / CHUNK_SIZE);
        for (let index = 0; index < count; index++) {
            const data = new FormData();
            data.append('filename', file.name);
            data.append('path', directory);
            data.append('uploadID', id);
            data.append('chunk', file.slice(index * CHUNK_SIZE, (index + 1) * CHUNK_SIZE));
            data.append('chunkIndex', String(index));
            data.append('chunkCount', String(count));
            data.append('offset', String(index * CHUNK_SIZE));
            data.append('fileSize', String(file.size));
            await uploadChunkWithRetry(
                data,
                {
                    headers: { CurrentNode: node },
                    signal: controller.signal,
                    skipErrorMessage: true,
                    timeout: 0,
                    onUploadProgress: (event: { total?: number; loaded: number }) => {
                        const fraction = event.total ? Math.min(1, event.loaded / event.total) : 0;
                        uploadPercent.value = Math.min(99, Math.floor(((index + fraction) * 100) / count));
                    },
                },
                controller.signal,
            );
        }
        controller.signal.throwIfAborted();
        uploadedPath.value = `${directory}/${file.name}`;
        uploadPercent.value = 100;
    } catch {
        if (!controller.signal.aborted) uploadFailed.value = true;
        await stopChunkUpload(id, node).catch(() => {});
        return;
    } finally {
        uploading.value = false;
        uploadController = undefined;
        uploadID = '';
    }
    if (!disposed && currentNode.value === node) await onSubmit();
};

const openTaskLog = () => {
    taskLogRef.value.openWithTaskID(submittedTaskID.value, true, operateNode.value);
};
const onSubmit = async () => {
    const paths = source.value === 'local' ? [uploadedPath.value].filter(Boolean) : [...form.paths];
    if (busy.value || submittedTaskID.value || !paths.length) return;
    loading.value = true;
    const taskID = newUUID();
    const node = operateNode.value;
    submittedTaskID.value = taskID;
    try {
        await imageLoad({ paths, taskID }, node);
        if (disposed || currentNode.value !== node) return;
        loadVisible.value = false;
        openTaskLog();
        MsgSuccess(i18n.global.t('container.imageImportSubmitted'));
    } catch {
        return;
    } finally {
        loading.value = false;
    }
};
const handlePathChange = () => {
    submittedTaskID.value = '';
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
    submittedTaskID.value = '';
    form.paths = [...new Set(Array.isArray(paths) ? paths : [paths])];
    form.path = form.paths.join('; ');
};
watch(currentNode, () => {
    loadVisible.value = false;
    uploadController?.abort();
});
onBeforeUnmount(() => {
    disposed = true;
    uploadController?.abort();
    if (uploadID) void stopChunkUpload(uploadID, operateNode.value).catch(() => {});
});
defineExpose({ acceptParams });
</script>
