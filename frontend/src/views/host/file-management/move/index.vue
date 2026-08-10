<template>
    <DrawerPro v-model="open" :header="title" @close="handleClose" size="690">
        <div class="space-y-4">
            <div
                v-if="showTaskStatus"
                class="rounded-lg border border-[var(--el-border-color-light)] bg-[var(--el-fill-color-light)] px-4 py-3"
            >
                <div class="flex items-center justify-between gap-3">
                    <div class="space-y-1">
                        <div class="text-sm font-medium text-[var(--el-text-color-secondary)]">
                            {{ $t('commons.status.executing') }}
                        </div>
                        <div class="text-xs text-[var(--el-text-color-secondary)]">
                            {{ taskInfo?.name || title }}
                        </div>
                    </div>
                    <el-button link type="primary" @click="openTaskLog" :disabled="!currentTaskID">
                        {{ $t('commons.button.log') }}
                    </el-button>
                </div>
                <el-progress
                    class="mt-3"
                    :percentage="100"
                    :indeterminate="true"
                    :duration="1"
                    :stroke-width="8"
                    :show-text="false"
                />
            </div>

            <el-form
                v-if="!currentTaskID"
                @submit.prevent
                ref="fileForm"
                label-position="top"
                :model="addForm"
                :rules="rules"
                v-loading="loading"
            >
                <el-alert
                    v-if="type == 'cut' && existFiles?.length == 0 && addForm.cover && changeName"
                    show-icon
                    type="warning"
                    :closable="false"
                >
                    <template #title>
                        <span class="whitespace-break-spaces">{{ $t('file.coverDirHelper') }}</span>
                    </template>
                </el-alert>
                <el-form-item :label="$t('file.path')" prop="newPath">
                    <el-input v-model="addForm.newPath">
                        <template #prepend>
                            <el-button icon="Folder" @click="fileRef.acceptParams({ dir: true })" />
                        </template>
                    </el-input>
                </el-form-item>
                <div v-if="changeName">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model="addForm.name" :disabled="addForm.cover"></el-input>
                    </el-form-item>
                    <el-radio-group v-model="addForm.cover" @change="changeType">
                        <el-radio :value="true" size="large">{{ $t('file.replace') }}</el-radio>
                        <el-radio :value="false" size="large">{{ $t('file.rename') }}</el-radio>
                    </el-radio-group>
                </div>
                <div v-if="existFiles.length > 0 && !changeName" class="text-center">
                    <el-alert show-icon type="warning" :closable="false">
                        <template #title>
                            <span class="whitespace-break-spaces">
                                {{ $t('file.existFileDirHelper') + $t('file.coverDirHelper') }}
                            </span>
                        </template>
                    </el-alert>
                    <el-transfer
                        v-model="skipFiles"
                        class="text-left inline-block mt-4"
                        :titles="[$t('commons.button.cover'), $t('commons.button.skip')]"
                        :format="{
                            noChecked: '${total}',
                            hasChecked: '${checked}/${total}',
                        }"
                        :data="transferData"
                    />
                </div>
            </el-form>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <template v-if="!currentTaskID">
                    <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="submit(fileForm)" :disabled="loading">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </template>
                <template v-else>
                    <el-button
                        v-if="isTaskExecuting || loading"
                        type="danger"
                        :loading="stopping"
                        @click="stopCurrentTask"
                    >
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button @click="openTaskLog">{{ $t('commons.button.log') }}</el-button>
                    <el-button type="default" @click="closeDrawer">{{ $t('commons.button.close') }}</el-button>
                </template>
            </span>
        </template>
    </DrawerPro>
    <FileList ref="fileRef" @choose="getPath" />
    <TaskLog ref="taskLogRef" />
</template>

<script lang="ts" setup>
import { batchCheckFiles, checkFile, moveFile, stopMoveFile } from '@/api/modules/files';
import { File } from '@/api/interface/file';
import { Log } from '@/api/interface/log';
import { searchTasks } from '@/api/modules/log';
import TaskLog from '@/components/log/task/index.vue';
import FileList from '@/components/file-list/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance, FormRules } from 'element-plus';
import { ref, reactive, computed, ComputedRef, onMounted, onUnmounted } from 'vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getDateStr } from '@/utils/date';
import { getErrorMessage } from '@/utils/misc';
import { newUUID } from '@/utils/id';

interface MoveProps {
    oldPaths: Array<string>;
    allNames: Array<string>;
    type: string;
    path: string;
    name: string;
    isDir: boolean;
}

const fileForm = ref<FormInstance>();
const loading = ref(false);
const open = ref(false);
const type = ref('cut');
const changeName = ref(false);
const oldName = ref('');
const existFiles = ref<File.ExistFileInfo[]>([]);
const skipFiles = ref([]);
const transferData = ref([]);
const fileRef = ref();
const taskLogRef = ref<InstanceType<typeof TaskLog> | null>(null);
const currentTaskID = ref('');
const currentTaskNode = ref('');
const taskInfo = ref<Log.Task | null>(null);
const stopping = ref(false);
const canceling = ref(false);
let taskTimer: ReturnType<typeof setInterval> | null = null;
const moveTaskKey = 'file-management-move-task';
const { currentNode } = useGlobalStore();

const getMoveTaskKey = (node: string) => `${moveTaskKey}:${node}`;

const title = computed(() => {
    if (type.value === 'cut') {
        return i18n.global.t('file.move');
    } else {
        return i18n.global.t('commons.button.copy');
    }
});

const addForm = reactive({
    oldPaths: [] as string[],
    newPath: '',
    type: '',
    name: '',
    allNames: [] as string[],
    isDir: false,
    cover: false,
    coverPaths: [] as string[],
});

const rules = reactive<FormRules>({
    newPath: [Rules.requiredInput],
    name: [Rules.requiredInput],
});

const em = defineEmits(['close', 'loading']);

const isTaskExecuting = computed(() => {
    return !!currentTaskID.value && (!taskInfo.value || taskInfo.value.status === 'Executing');
});

const showTaskStatus = computed(() => {
    return !!currentTaskID.value && (loading.value || isTaskExecuting.value);
});

const syncTaskStorage = () => {
    const status = taskInfo.value?.status || (currentTaskID.value || loading.value ? 'Executing' : '');
    const node = currentTaskNode.value || currentNode.value;
    const storageKey = getMoveTaskKey(node);
    if (currentTaskID.value && status === 'Executing') {
        localStorage.setItem(
            storageKey,
            JSON.stringify({ taskID: currentTaskID.value, status, type: type.value, node }),
        );
        return;
    }
    localStorage.removeItem(storageKey);
};

const stopTaskPolling = () => {
    if (taskTimer) {
        clearInterval(taskTimer);
        taskTimer = null;
    }
};

const clearCurrentTask = () => {
    stopTaskPolling();
    localStorage.removeItem(getMoveTaskKey(currentTaskNode.value || currentNode.value));
    currentTaskID.value = '';
    currentTaskNode.value = '';
    taskInfo.value = null;
    loading.value = false;
    stopping.value = false;
    canceling.value = false;
};

const resetDrawerState = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    changeName.value = false;
    oldName.value = '';
    existFiles.value = [];
    skipFiles.value = [];
    transferData.value = [];
    addForm.oldPaths = [];
    addForm.newPath = '';
    addForm.type = '';
    addForm.name = '';
    addForm.allNames = [];
    addForm.isDir = false;
    addForm.cover = false;
    addForm.coverPaths = [];
    clearCurrentTask();
    open.value = false;
};

const closeDrawer = () => {
    if (isTaskExecuting.value) {
        open.value = false;
        return;
    }
    resetDrawerState();
    em('close', false);
};

const handleClose = () => {
    if (loading.value) {
        void stopCurrentTask();
        return;
    }
    closeDrawer();
};

const loadTaskInfo = async () => {
    if (!currentTaskID.value) {
        return;
    }
    const taskID = currentTaskID.value;
    const taskNode = currentTaskNode.value || currentNode.value;
    try {
        const res = await searchTasks(
            {
                taskID,
                type: '',
                status: '',
                page: 1,
                pageSize: 1,
            },
            taskNode,
        );
        if (currentTaskID.value !== taskID || currentTaskNode.value !== taskNode) {
            return;
        }
        const item = res.data.items?.[0];
        if (!item) {
            return;
        }
        taskInfo.value = item;
        syncTaskStorage();
        if (item.status === 'Executing') {
            return;
        }
        stopTaskPolling();
        if (item.status === 'Success') {
            if (type.value === 'cut') {
                MsgSuccess(i18n.global.t('file.moveSuccess'));
            } else {
                MsgSuccess(i18n.global.t('file.copySuccess'));
            }
            resetDrawerState();
            em('close', true);
            return;
        }
        if (item.status === 'Canceled') {
            resetDrawerState();
            em('close', true);
            return;
        }
        const errorMessage = item.errorMsg || i18n.global.t('commons.msg.operationFailed');
        clearCurrentTask();
        MsgError(errorMessage);
    } catch (error) {
        console.error(error);
    }
};

const startTaskPolling = () => {
    stopTaskPolling();
    void loadTaskInfo();
    taskTimer = setInterval(() => {
        void loadTaskInfo();
    }, 1500);
};

const openTaskLog = () => {
    if (!currentTaskID.value) {
        return;
    }
    taskLogRef.value?.openWithTaskID(currentTaskID.value, true, currentTaskNode.value || currentNode.value);
};

const stopCurrentTask = async () => {
    if (!currentTaskID.value || stopping.value) {
        return;
    }
    if (loading.value) {
        canceling.value = true;
        return;
    }
    stopping.value = true;
    try {
        await stopMoveFile(currentTaskID.value, currentTaskNode.value || currentNode.value);
    } catch (error) {
        MsgError(getErrorMessage(error));
    } finally {
        stopping.value = false;
    }
};

const getFileName = (filePath: string) => {
    if (filePath.endsWith('/')) {
        filePath = filePath.slice(0, -1);
    }

    return filePath.split('/').pop();
};

const coverFiles: ComputedRef<string[]> = computed(() => {
    const existingNames = new Set(
        existFiles.value.filter((item) => !skipFiles.value.includes(item.name) && item.isDir).map((item) => item.name),
    );
    return addForm.oldPaths.filter((path) => existingNames.has(getFileName(path)));
});

const mvFiles: ComputedRef<string[]> = computed(() => {
    const skipSet = new Set(skipFiles.value);
    const coverSet = new Set(coverFiles.value.map(getFileName));

    return addForm.oldPaths.filter((path) => {
        const name = getFileName(path);
        return !skipSet.has(name) && !coverSet.has(name);
    });
});

const getPath = (path: string) => {
    addForm.newPath = path;
};

const changeType = () => {
    if (addForm.cover) {
        addForm.name = oldName.value;
    } else {
        addForm.name = renameFileWithSuffix(oldName.value, addForm.isDir);
    }
};

const mvFile = () => {
    const taskID = newUUID();
    currentTaskID.value = taskID;
    currentTaskNode.value = currentNode.value;
    taskInfo.value = null;
    open.value = true;
    loading.value = true;
    em('loading', true);
    syncTaskStorage();
    moveFile({ ...addForm, taskID })
        .then(async () => {
            loading.value = false;
            if (canceling.value) {
                stopping.value = true;
                try {
                    await stopMoveFile(taskID, currentTaskNode.value || currentNode.value);
                } catch (error) {
                    MsgError(getErrorMessage(error));
                } finally {
                    stopping.value = false;
                }
            }
            startTaskPolling();
        })
        .catch((error) => {
            clearCurrentTask();
            MsgError(getErrorMessage(error));
        })
        .finally(() => {
            loading.value = false;
            em('loading', false);
        });
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        addForm.coverPaths = coverFiles.value;
        addForm.oldPaths = mvFiles.value;
        mvFile();
    });
};

const getCompleteExtension = (filename: string): string => {
    const compoundExtensions = [
        '.tar.gz',
        '.tar.bz2',
        '.tar.xz',
        '.tar.lzma',
        '.tar.Z',
        '.tar.zst',
        '.tar.lzo',
        '.tar.sz',
        '.tgz',
        '.tbz2',
        '.txz',
        '.tzst',
    ];
    const foundExtension = compoundExtensions.find((ext) => filename.endsWith(ext));
    if (foundExtension) {
        return foundExtension;
    }
    const match = filename.match(/\.[a-zA-Z0-9]+$/);
    return match ? match[0] : '';
};

const renameFileWithSuffix = (fileName: string, isDir: boolean): string => {
    const insertStr = '-' + getDateStr();
    const completeExt = isDir ? '' : getCompleteExtension(fileName);
    if (!completeExt) {
        return `${fileName}${insertStr}`;
    } else {
        const baseName = fileName.slice(0, fileName.length - completeExt.length);
        return `${baseName}${insertStr}${completeExt}`;
    }
};

const handleFilePaths = async (fileNames: string[], newPath: string) => {
    const uniqueFiles = [...new Set(fileNames)];
    const fileNamesWithPath = uniqueFiles.map((file) => newPath + '/' + file);
    const existData = await batchCheckFiles(fileNamesWithPath);
    existFiles.value = existData.data;
    transferData.value = existData.data.map((file) => ({
        key: file.name,
        label: file.name,
    }));
};

const acceptParams = async (props: MoveProps) => {
    if (currentTaskID.value) {
        if (isTaskExecuting.value) {
            open.value = true;
            if (!taskTimer) {
                startTaskPolling();
            }
            return;
        }
        resetDrawerState();
    }
    changeName.value = false;
    addForm.oldPaths = props.oldPaths;
    addForm.type = props.type;
    addForm.newPath = props.path;
    addForm.isDir = props.isDir;
    addForm.name = '';
    addForm.allNames = props.allNames;
    type.value = props.type;
    existFiles.value = [];
    addForm.coverPaths = [];
    if (props.name && props.name != '') {
        oldName.value = props.name;
        const res = await checkFile(props.path + '/' + props.name, false);
        if (res.data) {
            changeName.value = true;
            addForm.cover = false;
            addForm.name = renameFileWithSuffix(props.name, addForm.isDir);
            open.value = true;
        } else {
            mvFile();
        }
    } else if (props.allNames && props.allNames.length > 0) {
        await handleFilePaths(addForm.allNames, addForm.newPath);
        if (existFiles.value.length > 0) {
            changeName.value = false;
            open.value = true;
        } else {
            mvFile();
        }
    } else {
        mvFile();
    }
};

const restoreTask = () => {
    const node = currentNode.value;
    const storageKey = getMoveTaskKey(node);
    const taskText = localStorage.getItem(storageKey);
    if (!taskText) {
        return;
    }
    try {
        const task = JSON.parse(taskText) as { taskID?: string; status?: string; type?: string; node?: string };
        if (!task.taskID || task.status !== 'Executing' || task.node !== node) {
            localStorage.removeItem(storageKey);
            return;
        }
        currentTaskID.value = task.taskID;
        currentTaskNode.value = node;
        if (task.type === 'copy' || task.type === 'cut') {
            type.value = task.type;
        }
        taskInfo.value = null;
        startTaskPolling();
    } catch {
        localStorage.removeItem(storageKey);
    }
};

onMounted(() => {
    restoreTask();
});

onUnmounted(() => {
    stopTaskPolling();
});

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
:deep(.el-transfer) {
    --el-transfer-panel-width: 260px;
    .el-button {
        padding: 4px 7px;
    }
}

:deep(.el-transfer__buttons) {
    padding: 5px 15px;
    @media (max-width: 600px) {
        width: 250px;
        text-align: center;
        padding: 10px 0;
        .el-button [class*='el-icon'] svg {
            transform: rotate(90deg);
        }
    }

    @media (min-width: 601px) {
        display: inline-flex;
        flex-direction: column;
        align-items: center;
        gap: 10px;
        width: 40px;
        height: 40px;
        justify-content: center;
        .el-button + .el-button {
            margin-left: 0;
        }
    }
}

:deep(.el-transfer-panel .el-transfer-panel__footer) {
    height: 65px;
}
</style>
