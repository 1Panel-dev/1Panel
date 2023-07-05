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
            <el-button :icon="Folder" @click="popoverVisible = true"></el-button>
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
        <div>
            <el-table :data="data" highlight-current-row height="40vh">
                <el-table-column width="40" fix>
                    <template #default="{ row }">
                        <el-checkbox
                            v-model="rowName"
                            :true-label="row.name"
                            :disabled="disabledDir(row)"
                            @change="checkFile(row)"
                        />
                    </template>
                </el-table-column>
                <el-table-column show-overflow-tooltip fix>
                    <template #default="{ row }">
                        <svg-icon v-if="row.isDir" className="table-icon" iconName="p-file-folder"></svg-icon>
                        <svg-icon v-else className="table-icon" iconName="p-file-normal"></svg-icon>
                        <el-link :underline="false" @click="open(row)">{{ row.name }}</el-link>
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
                <el-button type="primary" @click="selectFile">{{ $t('commons.button.confirm') }}</el-button>
            </div>
        </div>
    </el-popover>
</template>

<script lang="ts" setup>
import { File } from '@/api/interface/file';
import { GetFilesList } from '@/api/modules/files';
import { Folder } from '@element-plus/icons-vue';
import BreadCrumbs from '@/components/bread-crumbs/index.vue';
import BreadCrumbItem from '@/components/bread-crumbs/bread-crumbs-item.vue';
import { onMounted, onUpdated, reactive, ref } from 'vue';

const rowName = ref('');
const data = ref();
const loading = ref(false);
const paths = ref<string[]>([]);
const req = reactive({ path: '/', expand: true, page: 1, pageSize: 300 });
const selectRow = ref();
const popoverVisible = ref(false);

const props = defineProps({
    path: {
        type: String,
        default: '/',
    },
    dir: {
        type: Boolean,
        default: false,
    },
});

const em = defineEmits(['choose']);

const checkFile = (row: any) => {
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
    selectRow.value = {};
};

const disabledDir = (row: File.File) => {
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
        search(req);
    }
    rowName.value = '';
};

const jump = async (index: number) => {
    let path = '/';
    if (index != -1) {
        const jPaths = paths.value.slice(0, index + 1);
        for (let i in jPaths) {
            path = path + '/' + jPaths[i];
        }
    }
    req.path = path;
    search(req);
    popoverVisible.value = true;
};

const search = async (req: File.ReqFile) => {
    req.dir = props.dir;
    loading.value = true;
    await GetFilesList(req)
        .then((res) => {
            data.value = res.data.items;
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

onMounted(() => {
    req.path = props.path;
    rowName.value = '';
    search(req);
});

onUpdated(() => {
    req.path = props.path;
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
