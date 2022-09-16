<template>
    <el-popover placement="right" :width="400" trigger="click" :title="'文件列表'">
        <template #reference>
            <el-button :icon="Folder"></el-button>
        </template>
        <div>
            <BreadCrumbs>
                <BreadCrumbItem @click="jump(-1)" :right="paths.length == 0">root</BreadCrumbItem>
                <BreadCrumbItem
                    v-for="(item, key) in paths"
                    :key="key"
                    @click="jump(key)"
                    :right="key == paths.length - 1"
                >
                    {{ item }}
                </BreadCrumbItem>
            </BreadCrumbs>
        </div>
        <div>
            <el-input :prefix-icon="Search"></el-input>
            <el-table :data="data" highlight-current-row height="40vh">
                <el-table-column width="40" fix>
                    <template #default="{ row }">
                        <el-checkbox v-model="rowName" :true-label="row.name" @click="checkFile(row)" />
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
    </el-popover>
</template>

<script lang="ts" setup>
import { File } from '@/api/interface/file';
import { GetFilesList } from '@/api/modules/files';
import { Folder, Search } from '@element-plus/icons-vue';
import BreadCrumbs from '@/components/bread-crumbs/index.vue';
import BreadCrumbItem from '@/components/bread-crumbs/bread-crumbs-item.vue';
import { onMounted, onUpdated, reactive, ref } from 'vue';

let rowName = ref('');
let data = ref();
let loading = ref(false);
let paths = ref<string[]>([]);
let req = reactive({ path: '/', expand: true, page: 1, pageSize: 20 });

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
    rowName.value = row.name;
    em('choose', row.path);
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
    search(req);
});

onUpdated(() => {
    req.path = props.path;
    search(req);
});
</script>
