<template>
    <DrawerPro v-model="open" :header="title" @close="handleClose" size="690">
        <el-form
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
                    <template #prepend><FileList @choose="getPath" :dir="true"></FileList></template>
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
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose(false)" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(fileForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { batchCheckFiles, checkFile, moveFile } from '@/api/modules/files';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance, FormRules } from 'element-plus';
import { ref, reactive, computed, ComputedRef } from 'vue';
import FileList from '@/components/file-list/index.vue';
import { MsgSuccess } from '@/utils/message';
import { getDateStr } from '@/utils/util';
import { File } from '@/api/interface/file';

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

const em = defineEmits(['close']);

const handleClose = (search: boolean) => {
    open.value = false;
    if (fileForm.value) {
        fileForm.value.resetFields();
    }
    em('close', search);
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
    moveFile(addForm)
        .then(() => {
            if (type.value === 'cut') {
                MsgSuccess(i18n.global.t('file.moveSuccess'));
            } else {
                MsgSuccess(i18n.global.t('file.copySuccess'));
            }
            handleClose(true);
        })
        .finally(() => {
            loading.value = false;
        });
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
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
    changeName.value = false;
    addForm.oldPaths = props.oldPaths;
    addForm.type = props.type;
    addForm.newPath = props.path;
    addForm.isDir = props.isDir;
    addForm.name = '';
    addForm.allNames = props.allNames;
    type.value = props.type;
    existFiles.value = [];
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
