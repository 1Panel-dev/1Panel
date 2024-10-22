<template>
    <el-popover
        placement="right"
        :width="400"
        trigger="click"
        :title="$t('file.list')"
        :visible="popoverVisible"
        popper-class="file-list"
    >
        <template #reference>
            <el-button :icon="Folder" :disabled="disabled" @click="openPage()"></el-button>
        </template>
        <div>
            <el-button class="close" link @click="closePage">
                <el-icon><Close /></el-icon>
            </el-button>
            <BreadCrumbs>
                <BreadCrumbItem @click="jump(-1)" :right="paths.length == 0">
                    <el-icon><HomeFilled /></el-icon>
                </BreadCrumbItem>
                <template v-if="paths.length > 2">
                    <BreadCrumbItem>
                        <el-dropdown ref="dropdown1" trigger="click" @command="jump($event)">
                            <span class="el-dropdown-link">...</span>
                            <template #dropdown>
                                <el-dropdown-menu>
                                    <el-dropdown-item
                                        v-for="(item, key) in paths.slice(0, -1)"
                                        :key="key"
                                        :command="key"
                                    >
                                        {{ item }}
                                    </el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                    </BreadCrumbItem>
                    <BreadCrumbItem @click="jump(paths.length - 1)" :right="true">
                        <span class="sle" style="max-width: 200px">{{ paths[paths.length - 1] }}</span>
                    </BreadCrumbItem>
                </template>
                <template v-else>
                    <BreadCrumbItem
                        v-for="(item, key) in paths"
                        :key="key"
                        @click="jump(key)"
                        :right="key == paths.length - 1"
                    >
                        <span class="sle" style="max-width: 200px">{{ item }}</span>
                    </BreadCrumbItem>
                </template>
            </BreadCrumbs>
        </div>
        <div class="mt-4">
            <el-button link @click="onAddItem(true)" type="primary" size="small">
                {{ $t('commons.button.createNewFolder') }}
            </el-button>
            <el-button link @click="onAddItem(false)" type="primary" size="small">
                {{ $t('commons.button.createNewFile') }}
            </el-button>
        </div>
        <div>
            <el-table :data="data" highlight-current-row height="40vh">
                <el-table-column width="40" fix>
                    <template #default="{ row }">
                        <el-checkbox
                            v-model="rowName"
                            :true-value="row.name"
                            :disabled="disabledDir(row)"
                            @change="checkFile(row)"
                        />
                    </template>
                </el-table-column>
                <el-table-column show-overflow-tooltip fix>
                    <template #default="{ row }">
                        <div>
                            <svg-icon
                                :class="'table-icon'"
                                :iconName="row.isDir ? 'p-file-folder' : 'p-file-normal'"
                            ></svg-icon>

                            <template v-if="!row.isCreate">
                                <el-link :underline="false" @click="open(row)">
                                    {{ row.name }}
                                </el-link>
                            </template>

                            <template v-else>
                                <el-input
                                    ref="rowRefs"
                                    v-model="newFolder"
                                    style="width: 200px"
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
                        </div>
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
            <div class="button">
                <el-button @click="closePage">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="selectFile" :disabled="disBtn">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </div>
        </div>
    </el-popover>
</template>

<script lang="ts" setup>
import { File } from '@/api/interface/file';
import { CreateFile, GetFilesList } from '@/api/modules/files';
import { Folder, HomeFilled, Close } from '@element-plus/icons-vue';
import BreadCrumbs from '@/components/bread-crumbs/index.vue';
import BreadCrumbItem from '@/components/bread-crumbs/bread-crumbs-item.vue';
import { onMounted, onUpdated, reactive, ref, nextTick } from 'vue';
import i18n from '@/lang';
import { MsgSuccess, MsgWarning } from '@/utils/message';

const rowName = ref('');
const data = ref([]);
const loading = ref(false);
const paths = ref<string[]>([]);
const req = reactive({ path: '/', expand: true, page: 1, pageSize: 300, showHidden: true });
const selectRow = ref({ path: '', name: '' });
const rowRefs = ref();
const popoverVisible = ref(false);
const newFolder = ref();
const disBtn = ref(false);

const props = defineProps({
    path: {
        type: String,
        default: '/',
    },
    dir: {
        type: Boolean,
        default: false,
    },
    isAll: {
        type: Boolean,
        default: false,
    },
    disabled: {
        type: Boolean,
        default: false,
    },
});

const em = defineEmits(['choose']);

const checkFile = (row: any) => {
    disBtn.value = row.isCreate;
    selectRow.value = row;
    rowName.value = selectRow.value.name;
};

const selectFile = () => {
    if (selectRow.value) {
        em('choose', selectRow.value.path);
    }
    closePage();
};

const closePage = () => {
    popoverVisible.value = false;
    selectRow.value = { path: '', name: '' };
};

const openPage = () => {
    popoverVisible.value = true;
    selectRow.value.path = props.dir ? props.path || '/' : '';
    rowName.value = '';
};

const disabledDir = (row: File.File) => {
    if (props.isAll) {
        return false;
    }
    if (props.dir !== row.isDir) {
        return true;
    }
    if (!props.dir) {
        return row.isDir;
    }
    return false;
};

const open = async (row: File.File) => {
    if (row.isDir) {
        const name = row.name;
        paths.value.push(name);
        if (req.path === '/') {
            req.path = req.path + name;
        } else {
            req.path = req.path + '/' + name;
        }
        await search(req);
    }
    selectRow.value.path = props.dir ? req.path : '';
    rowName.value = '';
};

const jump = async (index: number) => {
    let path = '';
    if (index != -1) {
        if (index !== -1) {
            const jPaths = paths.value.slice(0, index + 1);
            path = '/' + jPaths.join('/');
        }
    }
    path = path || '/';
    req.path = path;
    selectRow.value.path = props.dir ? req.path : '';
    rowName.value = '';
    await search(req);
    popoverVisible.value = true;
};

const search = async (req: File.ReqFile) => {
    req.dir = props.dir;
    loading.value = true;
    await GetFilesList(req)
        .then((res) => {
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
    if (props.dir === isDir) {
        rowName.value = newFolder.value;
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
    selectRow.value.path = props.dir ? req.path : '';
    rowName.value = '';
    newFolder.value = '';
};

const handleChange = (value: string, row: any) => {
    if (rowName.value === row.name) {
        selectRow.value.name = value;
        rowName.value = value;
        row.name = value;
        const basePath = req.path === '/' ? req.path : `${req.path}/`;
        selectRow.value.path = `${basePath}${value}`;
    }
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
    CreateFile(addItem as File.FileCreate)
        .then(() => {
            row.isCreate = false;
            disBtn.value = false;
            MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    if (props.path != '') {
        req.path = props.path;
    }
    rowName.value = '';
    search(req);
});

onUpdated(() => {
    if (props.path != '') {
        req.path = props.path;
    }
    search(req);
});
</script>

<style lang="scss">
.file-list {
    position: relative;
    .close {
        position: absolute;
        right: 10px;
        top: 10px;
    }
}
.file-list-bottom {
    margin-top: 10px;
    .path {
        width: 250px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    .button {
        margin-top: 10px;
        float: right;
    }
}
</style>
