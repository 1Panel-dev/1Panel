<template>
    <DrawerPro
        v-model="visible"
        :header="title"
        size="large"
        :auto-close="!submitLoading"
        :confirm-before-close="submitLoading"
        @before-close="handleBeforeClose"
        @close="resetForm"
    >
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="form.name" />
            </el-form-item>

            <el-form-item :label="$t('commons.table.type')" prop="type">
                <el-select v-model="form.type" filterable allow-create default-first-option clearable>
                    <el-option
                        v-for="item in isoTypeOptions"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value"
                    />
                </el-select>
            </el-form-item>

            <el-form-item v-if="formMode === 'create'" :label="$t('xpack.vm.isoSource')" prop="sourceMode">
                <el-radio-group v-model="form.sourceMode" @change="changeSourceMode">
                    <el-radio-button label="server">{{ $t('database.hostSelect') }}</el-radio-button>
                    <el-radio-button label="upload">{{ $t('database.localUpload') }}</el-radio-button>
                    <el-radio-button label="download">{{ $t('xpack.vm.downloadFromServer') }}</el-radio-button>
                </el-radio-group>
            </el-form-item>

            <template v-if="formMode === 'edit' || form.sourceMode === 'server'">
                <el-form-item :label="$t('xpack.vm.isoPath')" prop="path">
                    <el-input v-model.trim="form.path">
                        <template #append>
                            <el-button @click="openFileSelector" icon="Folder" />
                        </template>
                    </el-input>
                </el-form-item>
            </template>

            <template v-if="formMode === 'create' && form.sourceMode === 'upload'">
                <el-upload
                    ref="uploadRef"
                    drag
                    :limit="1"
                    :auto-upload="false"
                    :show-file-list="true"
                    :file-list="uploadFiles"
                    accept=".iso"
                    :on-change="handleUploadChange"
                    :on-exceed="handleUploadExceed"
                    :on-remove="handleUploadRemove"
                >
                    <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
                    <div class="el-upload__text">{{ $t('file.dropHelper') }}</div>
                </el-upload>
            </template>

            <template v-if="formMode === 'create' && form.sourceMode === 'download'">
                <el-form-item :label="$t('file.downloadUrl')" prop="downloadUrl">
                    <el-input v-model.trim="form.downloadUrl" @input="changeDownloadUrl" />
                    <span v-if="downloadTargetPath" class="input-help">{{ downloadTargetPath }}</span>
                </el-form-item>
            </template>

            <el-form-item
                v-if="formMode === 'create' && ['upload', 'download'].includes(form.sourceMode)"
                :label="$t('xpack.vm.isoTargetDir')"
                prop="targetDir"
            >
                <el-input v-model.trim="form.targetDir" @blur="changeTargetDir" />
                <span class="input-help">{{ $t('xpack.vm.isoTargetDirHelper') }}</span>
            </el-form-item>

            <el-form-item class="mt-4" :label="$t('commons.table.description')" prop="description">
                <el-input v-model="form.description" type="textarea" :rows="3" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="submitLoading" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                v-permission
                type="primary"
                :disabled="submitLoading || uploading || downloading"
                @click="submit"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <div v-if="visible && submitLoading" class="vm-drawer-loading-mask">
        <div class="vm-submit-progress">
            <el-icon v-if="!showSubmitProgress" class="vm-submit-loading is-loading"><Loading /></el-icon>
            <div class="vm-submit-title">{{ submitStageTitle }}</div>
            <div v-if="submitFileName" class="vm-submit-file">{{ submitFileName }}</div>
            <el-progress
                v-if="showSubmitProgress"
                class="vm-submit-progress-bar"
                text-inside
                :stroke-width="18"
                :indeterminate="submitStage === 'downloading' && downloadPercent <= 0"
                :duration="1"
                :percentage="submitProgress"
            />
            <div v-if="submitStage === 'uploading' && uploadTotalSize" class="vm-submit-size">
                {{ uploadProgressSize }} / {{ computeSize(uploadTotalSize) }}
            </div>
            <div class="vm-submit-helper">{{ $t('xpack.vm.transferInProgress') }}</div>
            <el-button v-if="canCancelTransfer" class="vm-submit-cancel" :loading="cancelling" @click="cancelTransfer">
                {{ $t('commons.button.cancel') }}
            </el-button>
        </div>
    </div>

    <FileList ref="fileRef" @choose="selectFile" />
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref } from 'vue';
import {
    type FormInstance,
    type FormRules,
    type UploadFile,
    type UploadFiles,
    type UploadInstance,
    type UploadProps,
    type UploadRawFile,
} from 'element-plus';
import { Loading, UploadFilled } from '@element-plus/icons-vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess, MsgWarning } from '@/utils/message';
import FileList from '@/components/file-list/index.vue';
import {
    checkFile,
    chunkUploadFileData,
    createFile,
    stopChunkUpload,
    stopWgetFile,
    wgetFile,
} from '@/api/modules/files';
import { TimeoutEnum } from '@/enums/http-enum';
import { VM } from '@/api/interface/vm';
import { isVMResourceUnavailable } from '@/utils/vm';
import { createVMIso, listVMStorages, startVMIsoImport, updateVMIso, updateVMIsoImport } from '@/api/modules/vm';
import { Rules } from '@/global/form-rules';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { checkStreamAuth } from '@/utils/stream-auth';
import { computeSize } from '@/utils/size';

type FormMode = 'create' | 'edit';
type SourceMode = 'server' | 'upload' | 'download';
type SubmitStage = 'idle' | 'checking' | 'uploading' | 'downloading' | 'saving';

const emit = defineEmits(['search']);
const { currentNode: globalCurrentNode } = useGlobalStore();
const operateNode = globalCurrentNode.value || 'local';

const chunkSize = 5 * 1024 * 1024;
const isoTypeOptions = [
    { label: 'Debian', value: 'debian' },
    { label: 'Ubuntu', value: 'ubuntu' },
    { label: 'CentOS', value: 'centos' },
    { label: 'Rocky Linux', value: 'rocky' },
    { label: 'AlmaLinux', value: 'almalinux' },
    { label: 'Fedora', value: 'fedora' },
    { label: 'openSUSE', value: 'opensuse' },
    { label: 'RHEL', value: 'rhel' },
    { label: 'Oracle Linux', value: 'oraclelinux' },
    { label: 'Arch Linux', value: 'archlinux' },
    { label: 'Windows', value: 'windows' },
    { label: 'FreeBSD', value: 'freebsd' },
];

const visible = ref(false);
const submitLoading = ref(false);
const uploading = ref(false);
const uploadPercent = ref(0);
const downloading = ref(false);
const downloadPercent = ref(0);
const submitStage = ref<SubmitStage>('idle');
const cancelling = ref(false);
const cancelRequested = ref(false);
const formMode = ref<FormMode>('create');
const uploadTargetPath = ref('');
const downloadTargetPath = ref('');
const formRef = ref<FormInstance>();
const fileRef = ref();
const uploadRef = ref<UploadInstance>();
const uploadFiles = ref<UploadFiles>([]);
const storageItems = ref<VM.StorageItem[]>([]);
let processSocket: WebSocket | null = null;
let processTimer: ReturnType<typeof setInterval> | null = null;
let processTimeout: ReturnType<typeof setTimeout> | null = null;
let rejectProcess: ((error: Error) => void) | null = null;
let uploadAbortController: AbortController | null = null;
let uploadTaskKey = '';
let downloadTaskKey = '';
let importID = 0;
const title = computed(() =>
    formMode.value === 'create' ? i18n.global.t('commons.button.create') : i18n.global.t('commons.button.edit'),
);
const showSubmitProgress = computed(() => ['uploading', 'downloading'].includes(submitStage.value));
const canCancelTransfer = computed(() => ['uploading', 'downloading'].includes(submitStage.value));
const submitProgress = computed(() => {
    if (submitStage.value === 'uploading') return uploadPercent.value;
    if (submitStage.value === 'downloading') return downloadPercent.value > 0 ? downloadPercent.value : 100;
    return 0;
});
const submitStageTitle = computed(() => {
    switch (submitStage.value) {
        case 'uploading':
            return i18n.global.t('commons.status.uploading');
        case 'downloading':
            return i18n.global.t('commons.status.downloading');
        case 'saving':
            return i18n.global.t('xpack.vm.addingIsoRecord');
        default:
            return i18n.global.t('commons.status.executing');
    }
});
const uploadTotalSize = computed(() => uploadFiles.value[0]?.size || 0);
const uploadProgressSize = computed(() => computeSize((uploadTotalSize.value * uploadPercent.value) / 100));

const form = reactive({
    id: 0,
    type: '',
    name: '',
    path: '',
    targetDir: '',
    sourceMode: 'server' as SourceMode,
    downloadUrl: '',
    downloadName: '',
    description: '',
});
const submitFileName = computed(() => {
    if (submitStage.value === 'uploading') return uploadFiles.value[0]?.name || '';
    if (submitStage.value === 'downloading') return form.downloadName;
    return '';
});

const rules = reactive<FormRules>({
    name: [Rules.requiredInput],
    type: [Rules.requiredSelect],
    path: [Rules.requiredInput],
    downloadUrl: [Rules.requiredInput, { validator: validateDownloadUrl, trigger: 'blur' }],
});

function validateDownloadUrl(_rule: unknown, value: string, callback: (e?: Error) => void) {
    const v = (value || '').trim();
    if (!v) {
        callback();
        return;
    }
    try {
        const u = new URL(v);
        if (u.protocol !== 'http:' && u.protocol !== 'https:') {
            callback(new Error(i18n.global.t('file.wgetUrlInvalid')));
            return;
        }
        callback();
    } catch {
        callback(new Error(i18n.global.t('file.wgetUrlInvalid')));
    }
}

const acceptParams = (row?: VM.IsoItem) => {
    resetForm();
    if (row) {
        formMode.value = 'edit';
        form.id = row.id;
        form.type = row.type;
        form.name = row.name;
        form.path = row.path;
        form.sourceMode = 'server';
        form.description = row.description || '';
        visible.value = true;
        return;
    }
    formMode.value = 'create';
    visible.value = true;
    loadVMStorages();
};

const changeSourceMode = () => {
    form.path = '';
    form.downloadUrl = '';
    form.downloadName = '';
    uploadTargetPath.value = '';
    downloadTargetPath.value = '';
    uploadFiles.value = [];
    uploadRef.value?.clearFiles();
    formRef.value?.clearValidate(['path', 'downloadUrl']);
    if (['upload', 'download'].includes(form.sourceMode) && !form.targetDir) {
        loadVMStorages().then(() => {
            setDefaultTargetDir();
            updateTargetPaths();
        });
    }
    updateTargetPaths();
};

const openFileSelector = () => {
    fileRef.value?.acceptParams({ path: form.path ? dirname(form.path) : '/', dir: false });
};

const selectFile = (path: string) => {
    form.path = path;
    if (!form.name) {
        form.name = filename(path);
    }
};

const handleUploadChange: UploadProps['onChange'] = (file: UploadFile, files: UploadFiles) => {
    uploadFiles.value = files.slice(-1);
    if (!form.name) {
        form.name = file.name;
    }
    updateUploadTargetPath();
};

const handleUploadExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value?.clearFiles();
    const file = files[0] as UploadRawFile;
    uploadRef.value?.handleStart(file);
};

const handleUploadRemove = () => {
    uploadFiles.value = [];
    uploadTargetPath.value = '';
};

const changeDownloadUrl = (url: string) => {
    form.downloadName = getFileNameFromUrl(url);
    if (!form.name) {
        form.name = form.downloadName;
    }
    updateDownloadTargetPath();
};

const changeTargetDir = () => {
    if (!form.targetDir) {
        uploadTargetPath.value = '';
        downloadTargetPath.value = '';
        return;
    }
    form.targetDir = normalizeDir(form.targetDir);
    updateTargetPaths();
};

const updateTargetPaths = () => {
    updateUploadTargetPath();
    updateDownloadTargetPath();
};

const updateUploadTargetPath = () => {
    const file = uploadFiles.value[0];
    uploadTargetPath.value = file?.name && form.targetDir ? joinPath(form.targetDir, file.name) : '';
};

const updateDownloadTargetPath = () => {
    downloadTargetPath.value = form.downloadName && form.targetDir ? joinPath(form.targetDir, form.downloadName) : '';
};

const upload = async () => {
    const file = uploadFiles.value[0];
    if (!file?.raw) {
        return;
    }
    submitStage.value = 'uploading';
    uploading.value = true;
    uploadPercent.value = 0;
    uploadAbortController = new AbortController();
    let completed = false;
    try {
        await uploadLargeFile(file);
        completed = true;
    } finally {
        if (!completed && uploadTaskKey) {
            await stopChunkUpload(uploadTaskKey, operateNode).catch(() => undefined);
        }
        uploading.value = false;
    }
};

const uploadLargeFile = async (file: UploadFile) => {
    const raw = file.raw as UploadRawFile;
    const count = Math.ceil((file.size || 0) / chunkSize);
    for (let index = 0; index < count; index++) {
        const start = index * chunkSize;
        const end = Math.min(start + chunkSize, file.size || 0);
        const formData = new FormData();
        formData.append('filename', file.name);
        formData.append('uploadID', uploadTaskKey);
        formData.append('path', normalizeDir(form.targetDir));
        formData.append('chunk', raw.slice(start, end));
        formData.append('chunkIndex', index.toString());
        formData.append('chunkCount', count.toString());
        formData.append('overwrite', 'True');
        await chunkUploadFileData(formData, {
            headers: { CurrentNode: operateNode },
            onUploadProgress: (progressEvent) => {
                uploadPercent.value = Math.round(
                    ((index + progressEvent.loaded / (progressEvent.total || 1)) * 100) / count,
                );
            },
            timeout: TimeoutEnum.T_60S,
            signal: uploadAbortController?.signal,
        });
    }
    form.path = uploadTargetPath.value;
};

const downloadServerIso = async () => {
    submitStage.value = 'downloading';
    downloading.value = true;
    downloadPercent.value = 0;
    try {
        await ensureTargetDirExists(normalizeDir(form.targetDir));
        const res = await wgetFile(
            {
                url: form.downloadUrl.trim(),
                path: normalizeDir(form.targetDir),
                name: form.downloadName,
                ignoreCertificate: true,
                useProxy: false,
            },
            operateNode,
        );
        downloadTaskKey = res.data.key;
        if (importID) {
            await updateVMIsoImport(
                {
                    id: importID,
                    status: 'downloading',
                    tempFileID: downloadTaskKey,
                    tempFileType: 'download',
                },
                operateNode,
            );
        }
        if (cancelRequested.value) {
            await stopWgetFile(downloadTaskKey, operateNode);
            throw new Error('download cancelled');
        }
        await waitDownloadComplete(res.data.key);
        form.path = downloadTargetPath.value;
    } finally {
        downloading.value = false;
        closeProcessSocket();
    }
};

const cancelTransfer = async () => {
    if (cancelling.value) {
        return;
    }
    cancelling.value = true;
    cancelRequested.value = true;
    try {
        if (submitStage.value === 'uploading') {
            uploadAbortController?.abort();
            return;
        }
        if (submitStage.value === 'downloading' && downloadTaskKey) {
            await stopWgetFile(downloadTaskKey, operateNode);
        }
    } finally {
        if (submitStage.value === 'downloading') {
            closeProcessSocket();
        }
        cancelling.value = false;
    }
};

const waitDownloadComplete = async (key: string) => {
    if (!key) {
        throw new Error(i18n.global.t('commons.msg.operationFailed'));
    }
    const href = window.location.href;
    const protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    const ipLocal = href.split('//')[1].split('/')[0];
    const currentNode = operateNode;
    const url = `${protocol}://${ipLocal}/api/v2/files/wget/process?operateNode=${encodeURIComponent(currentNode)}`;
    const authError = await checkStreamAuth(url, currentNode);
    if (authError) {
        MsgError(authError);
        throw new Error(authError);
    }
    return new Promise<void>((resolve, reject) => {
        let emptyCount = 0;
        let hasProgress = false;
        let settled = false;
        const finish = (error?: Error) => {
            if (settled) return;
            settled = true;
            rejectProcess = null;
            clearProcessTimer();
            if (error) reject(error);
            else resolve();
        };
        const resetTimeout = () => {
            if (processTimeout) clearTimeout(processTimeout);
            processTimeout = setTimeout(() => finish(new Error(i18n.global.t('commons.msg.operationFailed'))), 60_000);
        };
        closeProcessSocket();
        rejectProcess = (error) => finish(error);
        resetTimeout();
        processSocket = new WebSocket(url);
        processSocket.onopen = () => {
            resetTimeout();
            processTimer = setInterval(() => {
                if (processSocket?.readyState === WebSocket.OPEN) {
                    processSocket.send(JSON.stringify({ type: 'wget', keys: [key] }));
                }
            }, 1000);
        };
        processSocket.onmessage = (message) => {
            resetTimeout();
            let list;
            try {
                list = JSON.parse(message.data || '[]');
            } catch {
                finish(new Error(i18n.global.t('commons.msg.operationFailed')));
                return;
            }
            const progress = list?.[0];
            if (progress) {
                if (progress.status === 'Failed' || progress.status === 'Canceled') {
                    const error = progress.error || i18n.global.t('commons.msg.operationFailed');
                    if (!cancelRequested.value) {
                        MsgError(error);
                    }
                    finish(new Error(error));
                    return;
                }
                hasProgress = true;
                emptyCount = 0;
                downloadPercent.value = Math.min(Math.round(Number(progress.percent) || 0), 100);
                if (Number(progress.percent) >= 100) {
                    finish();
                }
                return;
            }
            emptyCount++;
            if (hasProgress && emptyCount >= 3) {
                finish(new Error(i18n.global.t('commons.msg.operationFailed')));
            }
        };
        processSocket.onerror = () => finish(new Error(i18n.global.t('commons.msg.operationFailed')));
        processSocket.onclose = () => {
            finish(new Error(i18n.global.t('commons.msg.operationFailed')));
        };
    });
};

const submit = async () => {
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate();
    submitLoading.value = true;
    submitStage.value = 'checking';
    try {
        if (formMode.value === 'create') {
            if (form.sourceMode === 'upload') {
                if (!form.targetDir) {
                    MsgError(i18n.global.t('commons.rule.requiredInput'));
                    return;
                }
                if (!uploadFiles.value[0]?.raw) {
                    MsgError(i18n.global.t('commons.rule.requiredInput'));
                    return;
                }
                form.targetDir = normalizeDir(form.targetDir);
                updateUploadTargetPath();
                uploadTaskKey = `${Date.now()}-${Math.random().toString(36).slice(2)}`;
                await startImport('uploading', uploadTargetPath.value, uploadTaskKey, 'upload');
                await upload();
            } else if (form.sourceMode === 'download') {
                if (!form.targetDir || !form.downloadUrl || !form.downloadName) {
                    MsgError(i18n.global.t('commons.rule.requiredInput'));
                    return;
                }
                form.targetDir = normalizeDir(form.targetDir);
                updateDownloadTargetPath();
                await startImport('downloading', downloadTargetPath.value);
                await downloadServerIso();
            }
        }

        submitStage.value = 'saving';
        if (formMode.value === 'create') {
            if (importID) {
                await updateVMIsoImport({ id: importID, status: 'ready' }, operateNode);
            } else {
                await createVMIso(
                    {
                        type: form.type,
                        name: form.name,
                        path: form.path,
                        description: form.description,
                    },
                    operateNode,
                );
            }
        } else {
            await updateVMIso(
                {
                    id: form.id,
                    type: form.type,
                    name: form.name,
                    path: form.path,
                    description: form.description,
                },
                operateNode,
            );
        }
        visible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
    } catch (error) {
        if (importID) {
            await updateVMIsoImport(
                {
                    id: importID,
                    status: 'failed',
                    error: error instanceof Error ? error.message : i18n.global.t('commons.msg.operationFailed'),
                },
                operateNode,
            ).catch(() => undefined);
        }
        if (!cancelRequested.value) {
            throw error;
        }
    } finally {
        submitLoading.value = false;
        submitStage.value = 'idle';
        cancelling.value = false;
        cancelRequested.value = false;
        uploadAbortController = null;
        uploadTaskKey = '';
        downloadTaskKey = '';
        importID = 0;
    }
};

const handleBeforeClose = () => {
    MsgWarning(i18n.global.t('xpack.vm.transferInProgress'));
};

const resetForm = () => {
    form.id = 0;
    form.type = '';
    form.name = '';
    form.path = '';
    form.targetDir = '';
    form.sourceMode = 'server';
    form.downloadUrl = '';
    form.downloadName = '';
    form.description = '';
    uploadTargetPath.value = '';
    downloadTargetPath.value = '';
    uploadFiles.value = [];
    uploadPercent.value = 0;
    downloadPercent.value = 0;
    submitStage.value = 'idle';
    cancelling.value = false;
    cancelRequested.value = false;
    uploadAbortController = null;
    uploadTaskKey = '';
    downloadTaskKey = '';
    importID = 0;
    closeProcessSocket();
    uploadRef.value?.clearFiles();
    formRef.value?.clearValidate();
};

const loadVMStorages = async () => {
    if (storageItems.value.length) {
        return;
    }
    const res = await listVMStorages();
    storageItems.value = (res.data || []).filter((item) => item.path && resourceExists(item.status));
};

const resourceExists = (status: string) => !isVMResourceUnavailable(status);

const normalizeDir = (path: string) => {
    const rawPath = (path || '/').replace(/\/+/g, '/').replace(/\/+$/, '') || '/';
    const isAbsolute = rawPath.startsWith('/');
    const parts: string[] = [];
    for (const item of rawPath.split('/')) {
        if (!item || item === '.') {
            continue;
        }
        if (item === '..') {
            if (parts.length) {
                parts.pop();
            } else if (!isAbsolute) {
                parts.push(item);
            }
            continue;
        }
        parts.push(item);
    }
    const normalized = `${isAbsolute ? '/' : ''}${parts.join('/')}`;
    return normalized || (isAbsolute ? '/' : '.');
};
const joinPath = (dir: string, name: string) => `${normalizeDir(dir)}/${name}`.replace(/^\/\/+/, '/');
const filename = (path: string) => path.split('/').filter(Boolean).pop() || path;
const dirname = (path: string) => {
    const parts = path.split('/').filter(Boolean);
    parts.pop();
    return `/${parts.join('/')}`;
};

const getFileNameFromUrl = (url: string) => {
    try {
        const paths = new URL(url).pathname.split('/').filter(Boolean);
        return decodeURIComponent(paths.pop() || '');
    } catch {
        const paths = url.split('?')[0].split('/').filter(Boolean);
        return paths.pop() || '';
    }
};

const setDefaultTargetDir = () => {
    const defaultStorage = storageItems.value.find((item) => item.name === 'default') || storageItems.value[0];
    form.targetDir = defaultStorage?.path ? joinPath(defaultStorage.path, 'vm-iso') : '';
};

const ensureTargetDirExists = async (targetDir: string) => {
    const res = await checkFile(targetDir, false, operateNode);
    if (!res.data) {
        await createFile({ path: targetDir, isDir: true }, operateNode);
    }
};

const startImport = async (
    status: 'uploading' | 'downloading',
    path: string,
    tempFileID = '',
    tempFileType: 'upload' | 'download' | undefined = undefined,
) => {
    const res = await startVMIsoImport(
        {
            type: form.type,
            name: form.name,
            path,
            status,
            tempFileID,
            tempFileType,
            description: form.description,
        },
        operateNode,
    );
    importID = res.data.id;
    if (cancelRequested.value) {
        throw new Error('import cancelled');
    }
};

const clearProcessTimer = () => {
    if (processTimer) {
        clearInterval(processTimer);
        processTimer = null;
    }
    if (processTimeout) {
        clearTimeout(processTimeout);
        processTimeout = null;
    }
};

const closeProcessSocket = () => {
    rejectProcess?.(new Error(i18n.global.t('commons.msg.operationFailed')));
    rejectProcess = null;
    clearProcessTimer();
    const socket = processSocket;
    processSocket = null;
    if (socket) {
        socket.onopen = null;
        socket.onmessage = null;
        socket.onerror = null;
        socket.onclose = null;
        if (socket.readyState < WebSocket.CLOSING) {
            socket.close();
        }
    }
};

onBeforeUnmount(() => {
    if (submitLoading.value) {
        cancelRequested.value = true;
    }
    if (canCancelTransfer.value) {
        void cancelTransfer();
    }
    closeProcessSocket();
});

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.vm-drawer-loading-mask {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    z-index: 3000;
    width: 50%;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: color-mix(in srgb, var(--el-mask-color) 82%, transparent);
}

.vm-submit-progress {
    width: min(420px, 100%);
    padding: 28px 30px;
    text-align: center;
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    border-radius: 10px;
    box-shadow: var(--el-box-shadow-light);
}

.vm-submit-loading {
    margin-bottom: 14px;
    color: var(--el-color-primary);
    font-size: 32px;
}

.vm-submit-title {
    color: var(--el-text-color-primary);
    font-size: 16px;
    font-weight: 500;
}

.vm-submit-file {
    margin-top: 8px;
    overflow: hidden;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.vm-submit-progress-bar {
    margin-top: 20px;
}

.vm-submit-size,
.vm-submit-helper {
    margin-top: 10px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.vm-submit-helper {
    margin-top: 14px;
}

.vm-submit-cancel {
    margin-top: 18px;
}
</style>
