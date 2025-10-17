<template>
    <div>
        <DrawerPro v-model="open" :header="$t('file.formatConvert')" @close="handleClose" size="60%">
            <template #content>
                <el-tabs v-model="activeName" type="card" @tab-click="handleClick">
                    <el-tab-pane :label="$t('file.formatConvert')" name="convert">
                        <el-form v-loading="loading">
                            <div class="flex-row justify-start items-center">
                                <el-button type="primary" plain :icon="Plus" @click="addFile" />
                            </div>
                            <ComplexTable v-model:selects="selects" :data="data">
                                <el-table-column fixed type="selection" width="30" />
                                <el-table-column
                                    :label="$t('file.sourceFile')"
                                    show-overflow-tooltip
                                    prop="inputFile"
                                    min-width="150"
                                ></el-table-column>
                                <el-table-column
                                    :label="$t('file.sourceFormat')"
                                    show-overflow-tooltip
                                    prop="extension"
                                    width="90"
                                ></el-table-column>
                                <el-table-column prop="outputFile" width="160">
                                    <template #header>
                                        <div class="flex items-center justify-start gap-x-1">
                                            <span>{{ $t('file.converting') }}</span>
                                            <el-select
                                                v-model="handleOutputFile"
                                                placeholder="Select"
                                                size="small"
                                                style="width: 90px"
                                                class="font-normal"
                                                @change="handleHeaderSelectChange"
                                            >
                                                <el-option
                                                    v-for="item in options"
                                                    :key="item.value"
                                                    :label="item.label"
                                                    :value="item.value"
                                                ></el-option>
                                            </el-select>
                                        </div>
                                    </template>

                                    <template #default="{ row }">
                                        <el-select
                                            v-model="row.outputFormat"
                                            style="width: 135px"
                                            :value="row.outputFormat || getOptionsByType(row.type)[0]?.value"
                                        >
                                            <el-option
                                                v-for="item in getOptionsByType(row.type)"
                                                :key="item.value"
                                                :label="item.label"
                                                :value="item.value"
                                            ></el-option>
                                        </el-select>
                                    </template>
                                </el-table-column>
                                <fu-table-operations
                                    :buttons="buttons"
                                    :label="$t('commons.table.operate')"
                                    fix
                                    width="100"
                                />
                            </ComplexTable>
                            <div class="flex-row justify-start items-center pt-6">
                                <div class="flex justify-start items-center">
                                    <el-breadcrumb-item>{{ $t('file.saveDir') + ':' }}</el-breadcrumb-item>
                                    <el-breadcrumb separator="/">
                                        <el-breadcrumb-item>{{ outputPath }}</el-breadcrumb-item>
                                    </el-breadcrumb>
                                </div>
                                <el-checkbox v-model="deleteSource" :label="$t('file.deleteSourceFile')" size="large" />
                            </div>
                        </el-form>
                    </el-tab-pane>
                    <el-tab-pane :label="$t('file.convertLogs')" name="log">
                        <ComplexTable :pagination-config="paginationConfig" @search="search()" :data="logs">
                            <el-table-column
                                :label="$t('commons.table.date')"
                                show-overflow-tooltip
                                prop="date"
                                width="160"
                            ></el-table-column>
                            <el-table-column
                                :label="$t('commons.table.type')"
                                show-overflow-tooltip
                                prop="type"
                                width="70"
                            ></el-table-column>
                            <el-table-column :label="$t('commons.button.log')" prop="log" min-width="160">
                                <template #default="{ row }">
                                    <el-tooltip :content="row.log" placement="top">
                                        {{ row.log }}
                                    </el-tooltip>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('commons.table.message')" prop="status" width="100">
                                <template #default="{ row }">
                                    <el-tooltip v-if="row.status === 'FAILED'" :content="row.message" placement="top">
                                        <el-button type="danger" link>
                                            {{ $t('commons.status.' + row.status.toLowerCase()) }}
                                        </el-button>
                                    </el-tooltip>

                                    <el-button v-else :type="row.status === 'SUCCESS' ? 'success' : 'default'" link>
                                        {{ $t('commons.status.' + row.status.toLowerCase()) }}
                                    </el-button>
                                </template>
                            </el-table-column>
                        </ComplexTable>
                    </el-tab-pane>
                </el-tabs>
            </template>
            <template #footer v-if="activeName === 'convert'">
                <span class="dialog-footer">
                    <el-button @click="removeSelected" :disabled="selects.length === 0">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                    <el-button type="primary" @click="convert" :disabled="selects.length === 0">
                        {{ $t('commons.button.handle') }}
                    </el-button>
                </span>
            </template>
        </DrawerPro>
        <FileList ref="chooseFileRef" @choose="chooseFile" />
        <TaskLog ref="taskLogRef" />
    </div>
</template>

<script setup lang="ts">
import i18n from '@/lang';
import { reactive, ref } from 'vue';
import { File } from '@/api/interface/file';
import FileList from '@/components/file-list/index.vue';
import { Plus } from '@element-plus/icons-vue';
import { MsgWarning, MsgSuccess } from '@/utils/message';
import { convertFiles, convertLogs, getFileContent } from '@/api/modules/files';
import { getFileType, isConvertible } from '@/utils/util';
import { v4 as uuidv4 } from 'uuid';
import TaskLog from '@/components/log/task/index.vue';
import { TabsPaneContext } from 'element-plus';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const activeName = ref('convert');
const open = ref(false);
const loading = ref(false);
const deleteSource = ref(false);
const handleOutputFile = ref('png');
const chooseFileRef = ref();
const outputPath = ref('');
const data = ref<File.ConvertFile[]>([]);
const em = defineEmits(['close']);
let selects = ref<any>([]);
const taskLogRef = ref();
const logs = ref([]);

const CONVERT_KEY = 'convert-files:' + globalStore.currentNode;
const handleClose = () => {
    open.value = false;
    handleOutputFile.value = 'png';
    activeName.value = 'convert';
    em('close', false);
};
interface DialogProps {
    files: File.ConvertFile[];
    outputPath: string;
}

const options = [
    { value: 'png', label: 'png', type: 'image' },
    { value: 'jpg', label: 'jpg', type: 'image' },
    { value: 'webp', label: 'webp', type: 'image' },
    { value: 'gif', label: 'gif', type: 'image' },
    { value: 'jpeg', label: 'jpeg', type: 'image' },
    { value: 'bmp', label: 'bmp', type: 'image' },
    { value: 'tiff', label: 'tiff', type: 'image' },

    { value: 'mp4', label: 'mp4', type: 'video' },
    { value: 'avi', label: 'avi', type: 'video' },
    { value: 'mov', label: 'mov', type: 'video' },
    { value: 'mkv', label: 'mkv', type: 'video' },

    { value: 'mp3', label: 'mp3', type: 'audio' },
    { value: 'wav', label: 'wav', type: 'audio' },
    { value: 'flac', label: 'flac', type: 'audio' },
    { value: 'aac', label: 'aac', type: 'audio' },
];

let req = reactive({
    name: '',
    page: 1,
    pageSize: 10,
    orderBy: 'favorite',
    order: 'descending',
});

const getOptionsByType = (type: string) => {
    return options.filter((opt) => opt.type === type);
};

const initTableData = () => {
    data.value.map((row) => {
        const matchedOption = getOptionsByType(row.type).find((item) => row.extension.concat(item.value));
        const defaultOption = getOptionsByType(row.type)[0]?.value;
        row.outputFormat = matchedOption?.value || defaultOption;
    });
};

const handleHeaderSelectChange = (value: string) => {
    if (value && data.value.length) {
        data.value.forEach((row) => {
            const isValid = getOptionsByType(row.type).some((item) => item.value === value);
            row.outputFormat = isValid ? value : row.outputFormat;
        });
    }
};
const addFile = () => {
    chooseFileRef.value?.acceptParams({ dir: false });
};

const paginationConfig = reactive({
    cacheSizeKey: 'file-convert-log-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('file-convert-log-page-size')) || 20,
    total: 0,
});

const search = async () => {
    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;
    loading.value = true;
    data.value = [];
    await convertLogs(req)
        .then((res) => {
            logs.value = res.data.items;
            paginationConfig.total = res.data.total;
        })
        .finally(() => {
            loading.value = false;
        });
};

const chooseFile = async (path: string) => {
    await getFileContent({ path: path, expand: false, page: 1, pageSize: 1, isDetail: true }).then((res) => {
        if (!isConvertible(res.data.extension, res.data.mimeType)) {
            MsgWarning(i18n.global.t('file.fileCanNotConvert'));
            return;
        }
        updateFiles([
            {
                type: getFileType(res.data.extension),
                inputFile: res.data.name,
                extension: res.data.extension,
                path: getDirPath(res.data.path, res.data.name),
                outputFormat: res.data.extension.slice(1),
            },
        ]);
    });
};

const getDirPath = (fullPath: string, fileName: string): string => {
    if (!fullPath || !fileName) return fullPath;
    if (fullPath.endsWith(fileName)) {
        const dir = fullPath.slice(0, fullPath.length - fileName.length);
        return dir.replace(/\/$/, '');
    }
    return fullPath;
};

const delFile = async (path: string, name: string) => {
    ElMessageBox.confirm(i18n.global.t('file.deleteHelper2'), i18n.global.t('commons.msg.remove'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        try {
            loading.value = true;
            removeByPathAndName(path, name);
        } catch (error) {
        } finally {
            loading.value = false;
        }
    });
};

const removeByPathAndName = (path: string, name: string) => {
    const fullPath = `${path}/${name}`;
    data.value = data.value.filter((item) => `${item.path}/${item.inputFile}` !== fullPath);
    saveFiles();
};

const convert = async () => {
    if (selects.value.length > 0) {
        ElMessageBox.confirm(i18n.global.t('file.convertHelper'), i18n.global.t('file.convert'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            loading.value = true;
            const reqFiles = selects.value;
            const output = outputPath.value;
            const deleteSrc = deleteSource.value;
            const taskID = uuidv4();
            const singleReq = {
                files: reqFiles,
                outputPath: output,
                deleteSource: deleteSrc,
                taskID: taskID,
            };
            await convertFiles(singleReq)
                .then(() => {
                    removeSelected();
                    deleteSource.value = false;
                    MsgSuccess(i18n.global.t('file.execConvert'));
                    openTaskLog(taskID);
                })
                .finally(() => {
                    loading.value = false;
                });
        });
    } else {
        MsgWarning(i18n.global.t('file.convertHelper1'));
    }
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: any) => {
            delFile(row.path, row.inputFile);
        },
    },
];

const loadFiles = () => {
    const stored = localStorage.getItem(CONVERT_KEY);
    if (stored) {
        try {
            data.value = JSON.parse(stored);
        } catch {
            data.value = [];
        }
    } else {
        data.value = [];
    }
};

const saveFiles = () => {
    localStorage.setItem(CONVERT_KEY, JSON.stringify(data.value));
};

const updateFiles = (newFiles: File.ConvertFile[]) => {
    const stored = localStorage.getItem(CONVERT_KEY);
    if (stored) {
        try {
            const oldFiles: File.ConvertFile[] = JSON.parse(stored);
            const merged = [...oldFiles];
            for (const file of newFiles) {
                if (!merged.some((item) => item.path === file.path && item.inputFile === file.inputFile)) {
                    merged.push(file);
                }
            }
            data.value = merged;
        } catch {
            data.value = [...newFiles];
        }
    } else {
        data.value = [...newFiles];
    }
    saveFiles();
};

const handleClick = (tab: TabsPaneContext) => {
    if (tab.paneName == 'log') {
        search();
    } else {
        loadFiles();
    }
};

const removeSelected = () => {
    const selectedKeys = new Set(selects.value.map((s) => `${s.path}/${s.inputFile}`));
    data.value = data.value.filter((d) => !selectedKeys.has(`${d.path}/${d.inputFile}`));
    saveFiles();
};

const acceptParams = (param: DialogProps): void => {
    open.value = true;
    outputPath.value = param.outputPath;
    loadFiles();
    updateFiles(param.files);
    initTableData();
};

defineExpose({ acceptParams });
</script>
