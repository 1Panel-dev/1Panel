<template>
    <DrawerPro v-model="open" :header="title" @close="handleClose" size="large">
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
                            {{ taskInfo?.name || $t('file.compress') }}
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
                ref="fileForm"
                label-position="top"
                :model="form"
                label-width="100px"
                :rules="rules"
                v-loading="loading"
            >
                <el-form-item :label="$t('file.compressType')" prop="type">
                    <el-select v-model="form.type">
                        <el-option v-for="item in options" :key="item" :label="item" :value="item" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input v-model="form.name">
                        <template #append>{{ extension }}</template>
                    </el-input>
                </el-form-item>
                <el-form-item :label="$t('file.compressDst')" prop="dst">
                    <el-input v-model="form.dst">
                        <template #prepend>
                            <el-button icon="Folder" @click="fileRef.acceptParams({ dir: true, path: form.dst })" />
                        </template>
                    </el-input>
                </el-form-item>
                <el-form-item :label="$t('setting.compressPassword')" prop="secret" v-if="form.type === 'tar.gz'">
                    <el-input v-model="form.secret"></el-input>
                </el-form-item>
                <el-form-item>
                    <el-checkbox v-model="form.replace" :label="$t('file.replace')"></el-checkbox>
                </el-form-item>
            </el-form>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <template v-if="!currentTaskID">
                    <el-button :type="loading ? 'danger' : 'default'" :loading="canceling" @click="handleClose">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button v-if="!loading" type="primary" @click="submit(fileForm)">
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
                    <el-button v-if="currentTaskID" @click="openTaskLog">
                        {{ $t('commons.button.log') }}
                    </el-button>
                    <el-button type="default" @click="closeDrawer">
                        {{ $t('commons.button.close') }}
                    </el-button>
                </template>
            </span>
        </template>
    </DrawerPro>
    <FileList ref="fileRef" @choose="getLinkPath" />
    <TaskLog ref="taskLogRef" />
</template>

<script setup lang="ts">
import i18n from '@/lang';
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { File } from '@/api/interface/file';
import { Log } from '@/api/interface/log';
import { FormInstance, FormRules } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { CompressExtension, CompressType } from '@/enums/files';
import { compressFile, stopCompressFile } from '@/api/modules/files';
import { searchTasks } from '@/api/modules/log';
import FileList from '@/components/file-list/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getErrorMessage } from '@/utils/misc';
import { newUUID } from '@/utils/id';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { currentNode } = useGlobalStore();

interface CompressProps {
    files: Array<any>;
    dst: string;
    name: string;
    operate: string;
}

const rules = reactive<FormRules>({
    type: [Rules.requiredSelect],
    dst: [Rules.requiredInput],
    name: [Rules.requiredInput],
});

const fileForm = ref<FormInstance>();
const loading = ref(false);
const canceling = ref(false);
const stopping = ref(false);
const abortController = ref<AbortController | null>(null);
const form = ref<File.FileCompress>({ files: [], type: 'zip', dst: '', name: '', replace: false, secret: '' });
const options = ref<string[]>([]);
const open = ref(false);
const title = ref('');
const operate = ref('compress');
const fileRef = ref();
const taskLogRef = ref<InstanceType<typeof TaskLog> | null>(null);
const currentTaskID = ref('');
const taskInfo = ref<Log.Task | null>(null);
let taskTimer: ReturnType<typeof setInterval> | null = null;
const compressTaskKey = 'file-management-compress-task';

const em = defineEmits<{
    (e: 'close', value: boolean): void;
    (
        e: 'task-change',
        value: {
            taskID: string;
            status: string;
        },
    ): void;
}>();

const extension = computed(() => {
    return CompressExtension[form.value.type];
});

const isTaskExecuting = computed(() => taskInfo.value?.status === 'Executing');
const showTaskStatus = computed(() =>
    Boolean(currentTaskID.value || loading.value || canceling.value || stopping.value),
);

const emitTaskChange = () => {
    const payload = {
        taskID: currentTaskID.value,
        status: taskInfo.value?.status || (currentTaskID.value || loading.value ? 'Executing' : ''),
    };
    if (payload.taskID && payload.status) {
        localStorage.setItem(compressTaskKey, JSON.stringify(payload));
    } else {
        localStorage.removeItem(compressTaskKey);
    }
    em('task-change', payload);
};

const stopTaskPolling = () => {
    if (taskTimer) {
        clearInterval(taskTimer);
        taskTimer = null;
    }
};

const loadTaskInfo = async () => {
    if (!currentTaskID.value) {
        taskInfo.value = null;
        emitTaskChange();
        return;
    }
    try {
        const res = await searchTasks(
            {
                taskID: currentTaskID.value,
                type: '',
                status: '',
                page: 1,
                pageSize: 1,
            },
            currentNode.value,
        );
        taskInfo.value = res.data.items?.[0] || null;
        emitTaskChange();
        if (!taskInfo.value || taskInfo.value.status !== 'Executing') {
            stopTaskPolling();
            resetDrawerState();
            em('close', false);
        }
    } catch (error) {
        stopTaskPolling();
        resetDrawerState();
        em('close', false);
    }
};

const startTaskPolling = () => {
    stopTaskPolling();
    void loadTaskInfo();
    taskTimer = setInterval(() => {
        void loadTaskInfo();
    }, 1500);
};

const resetDrawerState = () => {
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    abortController.value = null;
    canceling.value = false;
    loading.value = false;
    taskInfo.value = null;
    currentTaskID.value = '';
    stopTaskPolling();
    emitTaskChange();
    open.value = false;
};

const closeDrawer = () => {
    if (currentTaskID.value && isTaskExecuting.value) {
        stopTaskPolling();
        open.value = false;
        em('close', false);
        return;
    }
    resetDrawerState();
    em('close', false);
};

const resumeTask = () => {
    if (!currentTaskID.value) {
        return;
    }
    open.value = true;
};

const stopCurrentTask = async () => {
    if (stopping.value || (!loading.value && !currentTaskID.value)) {
        return;
    }
    stopping.value = true;
    try {
        if (loading.value) {
            if (abortController.value) {
                canceling.value = true;
                abortController.value.abort();
            }
        } else if (currentTaskID.value) {
            await stopCompressFile(currentTaskID.value);
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        open.value = false;
        em('close', false);
    } catch (err) {
        MsgError(getErrorMessage(err));
    } finally {
        stopping.value = false;
    }
};

const handleClose = () => {
    if (loading.value) {
        stopCurrentTask();
        return;
    }
    closeDrawer();
};

const openTaskLog = () => {
    if (!currentTaskID.value) {
        return;
    }
    taskLogRef.value?.openWithTaskID(currentTaskID.value, true, currentNode.value);
};

const getLinkPath = (path: string) => {
    form.value.dst = path;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        let addItem = {};
        Object.assign(addItem, form.value);
        addItem['name'] = form.value.name + extension.value;
        const taskID = newUUID();
        addItem['taskID'] = taskID;
        canceling.value = false;
        abortController.value = new AbortController();
        loading.value = true;
        compressFile(addItem as File.FileCompress, {
            signal: abortController.value?.signal,
        })
            .then(() => {
                currentTaskID.value = taskID;
                loading.value = false;
                taskInfo.value = null;
                emitTaskChange();
                startTaskPolling();
            })
            .catch((error) => {
                const message = getErrorMessage(error).toLowerCase();
                if (canceling.value || message.includes('canceled') || message.includes('errshutdown')) {
                    currentTaskID.value = '';
                    emitTaskChange();
                    return;
                }
                currentTaskID.value = '';
                emitTaskChange();
                MsgError(getErrorMessage(error));
            })
            .finally(() => {
                loading.value = false;
                const wasCanceling = canceling.value;
                abortController.value = null;
                canceling.value = false;
                if (wasCanceling) {
                    closeDrawer();
                }
            });
    });
};

const acceptParams = (props: CompressProps) => {
    if (currentTaskID.value) {
        open.value = true;
        if (!taskTimer) {
            startTaskPolling();
        }
        emitTaskChange();
        return;
    }

    form.value.files = props.files;
    form.value.dst = props.dst;
    form.value.name = props.name;
    form.value.secret = '';
    form.value.replace = false;
    abortController.value = null;
    canceling.value = false;
    stopping.value = false;

    operate.value = props.operate;
    options.value = [];
    for (const t in CompressType) {
        options.value.push(CompressType[t]);
    }
    taskInfo.value = null;
    currentTaskID.value = '';
    emitTaskChange();
    stopTaskPolling();
    open.value = true;

    title.value = i18n.global.t('file.' + props.operate);
};

const restoreRunningTask = () => {
    const taskText = localStorage.getItem(compressTaskKey);
    if (!taskText) {
        return;
    }
    try {
        const task = JSON.parse(taskText) as {
            taskID?: string;
            status?: string;
        };
        if (!task.taskID) {
            localStorage.removeItem(compressTaskKey);
            return;
        }
        currentTaskID.value = task.taskID;
        if (task.status === 'Executing') {
            emitTaskChange();
        } else {
            localStorage.removeItem(compressTaskKey);
        }
    } catch {
        localStorage.removeItem(compressTaskKey);
    }
};

onMounted(() => {
    restoreRunningTask();
});

watch(open, (val) => {
    if (val && currentTaskID.value) {
        startTaskPolling();
    } else {
        stopTaskPolling();
    }
});

onUnmounted(() => {
    stopTaskPolling();
});

defineExpose({ acceptParams, resumeTask });
</script>
