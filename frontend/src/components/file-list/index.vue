<template>
    <DialogPro :title="$t('file.list')" size="w-60" v-model="open" @close="handleClose">
        <div>
            <div
                v-show="!searchableStatus"
                @click="searchableStatus = true"
                class="address-bar shadow-md rounded-md px-4 py-2 flex items-center flex-grow"
            >
                <span class="root mr-1">
                    <el-link @click.stop="jump(-1)">
                        <el-icon :size="20"><HomeFilled /></el-icon>
                    </el-link>
                </span>
                <span v-if="paths.length > 0">
                    <span v-for="(_, index) in paths" class="inline-flex items-center" :key="index">
                        <span class="ml-1 mr-1 arrow">></span>
                        <el-tooltip effect="dark" :content="paths[index]" placement="top">
                            <el-link class="path-segment cursor-pointer mr-1 pathname" @click.stop="jump(index)">
                                {{ paths[index].length > 25 ? paths[index].substring(0, 22) + '...' : paths[index] }}
                            </el-link>
                        </el-tooltip>
                    </span>
                </span>
            </div>
            <el-input
                ref="searchableInputRef"
                v-show="searchableStatus"
                v-model="searchablePath"
                @blur="searchableInputBlur"
                class="px-4 py-2 border rounded-md shadow-md"
                @keyup.enter="
                    jumpPath();
                    searchableStatus = false;
                "
            />
        </div>
        <el-button class="mt-4 float-left" link @click="jump(paths.length - 2)" type="primary" size="small">
            {{ $t('file.top') }}
        </el-button>
        <div class="mt-4 float-right">
            <el-button link @click="onAddItem(true)" type="primary" size="small">
                {{ $t('commons.button.createNewFolder') }}
            </el-button>
            <el-button link @click="onAddItem(false)" type="primary" size="small">
                {{ $t('commons.button.createNewFile') }}
            </el-button>
        </div>
        <div>
            <el-table :data="data" highlight-current-row height="40vh" @row-click="openDir" class="cursor-pointer">
                <el-table-column prop="name" show-overflow-tooltip fix>
                    <template #default="{ row }">
                        <svg-icon
                            :class="'table-icon'"
                            :iconName="row.isDir ? 'p-file-folder' : 'p-file-normal'"
                        ></svg-icon>
                        <template v-if="!row.isCreate">
                            {{ row.name }}
                        </template>

                        <template v-else>
                            <el-input
                                ref="rowRefs"
                                v-model="newFolder"
                                class="p-w-200"
                                placeholder="new folder"
                                @input="handleChange(newFolder, row)"
                            ></el-input>
                            <el-button link @click="createFolder(row)" type="primary" size="small" class="ml-2">
                                {{ $t('commons.button.save') }}
                            </el-button>
                            <el-button link @click="cancelFolder(row)" type="primary" size="small" class="!ml-2">
                                {{ $t('commons.button.cancel') }}
                            </el-button>
                        </template>
                    </template>
                </el-table-column>
                <el-table-column prop="size" width="160px" fix>
                    <template #default="{ row }">
                        <el-button
                            type="primary"
                            link
                            small
                            v-if="!row.isCreate"
                            :loading="row.btnLoading"
                            @click="row.isDir ? getDirSize(row.path) : getFileSize(row.path)"
                        >
                            <span v-if="row.isDir">
                                <span v-if="row.dirSize === undefined">
                                    {{ $t('file.calculate') }}
                                </span>
                                <span v-else>{{ computeSize(row.dirSize) }}</span>
                            </span>
                            <span v-else>
                                {{ computeSize(row.size) }}
                            </span>
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
        </div>
        <div class="file-list-bottom">
            <div v-if="selectRow?.path">
                {{ $t('file.currentSelect') }}
                <el-tooltip :content="selectRow.path" placement="top">
                    <el-tag type="success">
                        <div class="path">
                            <span>{{ selectRow.path }}</span>
                        </div>
                    </el-tag>
                </el-tooltip>
            </div>
        </div>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="selectFile" :disabled="disBtn">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { File } from '@/api/interface/file';
import { computeDirSize, createFile, getFileContent, getFilesList } from '@/api/modules/files';
import { onUpdated, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { useSearchableForSelect } from '@/views/host/file-management/hooks/searchable';
import { computeSize } from '@/utils/util';

const data = ref([]);
const loading = ref(false);
const paths = ref<string[]>([]);
const req = reactive({ path: '/', expand: true, page: 1, pageSize: 300, showHidden: true });
const selectRow = ref({ path: '', name: '' });
const rowRefs = ref();
const open = ref(false);
const newFolder = ref();
const disBtn = ref(false);

const { searchableStatus, searchablePath, searchableInputRef, searchableInputBlur } = useSearchableForSelect(paths);
const oldUrl = ref<string>('');

const em = defineEmits(['choose']);

const form = reactive({
    path: '/',
    dir: false,
    isAll: false,
    disabled: false,
});

interface DialogProps {
    path: string;
    dir: boolean;
    isAll: boolean;
    disabled: boolean;
}
const acceptParams = (props: DialogProps): void => {
    form.path = props.path || '/';
    form.dir = props.dir;
    form.isAll = props.isAll;
    form.disabled = props.disabled;
    openPage();
    req.path = form.path;
    oldUrl.value = form.path;
    search(req);
    open.value = true;
};

const selectFile = () => {
    if (selectRow.value) {
        em('choose', selectRow.value.path);
    }
    handleClose();
};

const handleClose = () => {
    open.value = false;
    selectRow.value = { path: '', name: '' };
};

const openPage = () => {
    open.value = true;
    selectRow.value.path = form.dir ? form.path || '/' : '';
};

const openDir = async (row: File.File, column: any, event: any) => {
    if (event?.target?.tagName === 'BUTTON' || event?.target?.tagName === 'SPAN') {
        return;
    }
    if (row.isDir) {
        const name = row.name;
        paths.value.push(name);
        if (req.path === '/') {
            req.path = req.path + name;
        } else {
            req.path = req.path + '/' + name;
        }
        await search(req);
        if (form.isAll || form.dir) {
            selectRow.value.path = req.path;
        } else {
            selectRow.value.path = '';
        }
        return;
    }
    if (!form.isAll && !form.dir) {
        selectRow.value.path = (req.path === '/' ? req.path : req.path + '/') + row.name;
        return;
    }
    selectRow.value.path = '';
};

const jump = async (index: number) => {
    oldUrl.value = req.path;
    let path = '';
    if (index !== -1) {
        const jPaths = paths.value.slice(0, index + 1);
        path = '/' + jPaths.join('/');
    }
    path = path || '/';
    req.path = path;
    selectRow.value.path = form.dir ? req.path : '';
    await search(req);
    open.value = true;
};

const jumpPath = async () => {
    loading.value = true;
    try {
        oldUrl.value = req.path;
        getPaths(searchablePath.value);
        req.path = searchablePath.value || '/';
        search(req);
    } finally {
        loading.value = false;
    }
};

const getFileSize = async (path: string) => {
    let params = {
        path: path,
        expand: true,
        isDetail: true,
        page: 1,
        pageSize: 100,
    };
    updateByPath(path, { btnLoading: true });
    try {
        const res = await getFileContent(params);
        updateByPath(path, { dirSize: res.data.size });
    } finally {
        updateByPath(path, { btnLoading: false });
    }
};

const getDirSize = async (path: string) => {
    const req = {
        path: path,
    };
    updateByPath(path, { btnLoading: true });
    try {
        const res = await computeDirSize(req);
        updateByPath(path, { dirSize: res.data.size });
    } finally {
        updateByPath(path, { btnLoading: false });
    }
};

const updateByPath = (path: string, patch: Partial<(typeof data.value)[0]>) => {
    data.value = data.value.map((item) => (item.path === path ? { ...item, ...patch } : item));
};

const getPaths = (reqPath: string) => {
    const pathArray = reqPath.split('/');
    paths.value = [];
    let base = '/';
    for (const p of pathArray) {
        if (p != '') {
            if (base.endsWith('/')) {
                base = base + p;
            } else {
                base = base + '/' + p;
            }
            paths.value.push(p);
        }
    }
};

const search = async (req: File.ReqFile) => {
    req.dir = form.dir;
    loading.value = true;
    await getFilesList(req)
        .then((res) => {
            if (!res.data.path) {
                req.path = oldUrl.value;
                getPaths(oldUrl.value);
                MsgWarning(i18n.global.t('commons.res.notFound'));
                return;
            }
            data.value = res.data.items || [];
            req.path = res.data.path;
            const pathArray = req.path.split('/');
            paths.value = [];
            for (const p of pathArray) {
                if (p != '') {
                    paths.value.push(p);
                }
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

let addForm = reactive({ path: '', name: '', isDir: true, mode: 0o755, isLink: false, isSymlink: true, linkPath: '' });

const onAddItem = async (isDir: boolean) => {
    const createRow = data.value?.find((row: { isCreate: any }) => row.isCreate);
    if (createRow) {
        MsgWarning(i18n.global.t('commons.msg.creatingInfo'));
        return;
    }
    newFolder.value = isDir ? i18n.global.t('file.noNameFolder') : i18n.global.t('file.noNameFile');
    if (form.dir === isDir) {
        selectRow.value.name = newFolder.value;
        const basePath = req.path === '/' ? req.path : `${req.path}/`;
        selectRow.value.path = `${basePath}${newFolder.value}`;
    }
    data.value?.unshift({
        path: selectRow.value.path,
        isCreate: true,
        isDir: isDir,
        name: newFolder.value,
    });
    disBtn.value = true;
    await nextTick();
    rowRefs.value.focus();
};

const cancelFolder = (row: any) => {
    data.value.shift();
    row.isCreate = false;
    disBtn.value = false;
    selectRow.value.path = form.dir ? req.path : '';
    newFolder.value = '';
};

const handleChange = (value: string, row: any) => {
    row.name = value;
    const basePath = req.path === '/' ? req.path : `${req.path}/`;
    selectRow.value.path = `${basePath}${value}`;
    if (row.isDir) {
        if (form.isAll || form.dir) {
            selectRow.value.path = `${basePath}${value}`;
        } else {
            selectRow.value.path = '';
        }
        return;
    }
    if (form.isAll || !form.dir) {
        selectRow.value.path = `${basePath}${value}`;
        return;
    }
    selectRow.value.path = '';
};

const createFolder = async (row: any) => {
    const basePath = req.path === '/' ? req.path : `${req.path}/`;
    addForm.path = `${basePath}${newFolder.value}`;
    if (addForm.path.indexOf('.1panel_clash') > -1) {
        MsgWarning(i18n.global.t('file.clashDitNotSupport'));
        return;
    }
    addForm.isDir = row.isDir;
    addForm.name = newFolder.value;
    let addItem = {};
    Object.assign(addItem, addForm);
    loading.value = true;
    createFile(addItem as File.FileCreate)
        .then(() => {
            row.isCreate = false;
            disBtn.value = false;
            MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

onUpdated(() => {
    if (form.path != '') {
        req.path = form.path;
    }
    search(req);
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.file-row {
    display: flex;
    align-items: center;
    width: 100%;
}

.address-bar {
    border: var(--el-border);
    .arrow {
        color: #726e6e;
    }
}
.file-list-bottom {
    margin-top: 10px;
    .path {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
}
</style>
