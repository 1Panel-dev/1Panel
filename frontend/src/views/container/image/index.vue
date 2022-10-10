<template>
    <div v-loading="loading">
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
                <template #toolbar>
                    <el-button type="primary" @click="pullVisiable = true">
                        {{ $t('container.pullFromRepo') }}
                    </el-button>
                    <el-button @click="loadVisiable = true">
                        {{ $t('container.importImage') }}
                    </el-button>
                    <el-button @click="onBatchDelete(null)">
                        {{ $t('container.build') }}
                    </el-button>
                    <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix></el-table-column>
                <el-table-column label="ID" show-overflow-tooltip prop="id" min-width="60" />
                <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="name" min-width="100" />
                <el-table-column :label="$t('container.version')" prop="version" min-width="60" fix />
                <el-table-column :label="$t('container.size')" prop="size" min-width="70" fix />
                <el-table-column :label="$t('commons.table.createdAt')" min-width="80" fix>
                    <template #default="{ row }">
                        {{ dateFromat(0, 0, row.createdAt) }}
                    </template>
                </el-table-column>
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" />
            </ComplexTable>
        </el-card>

        <el-dialog v-model="pullVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('container.imagePull') }}</span>
                </div>
            </template>
            <el-form ref="pullFormRef" :model="pullForm" label-width="80px">
                <el-form-item :label="$t('container.repoName')" :rules="Rules.requiredSelect" prop="repoID">
                    <el-select style="width: 100%" filterable v-model="pullForm.repoID">
                        <el-option
                            v-for="item in repos"
                            :key="item.id"
                            :value="item.id"
                            :label="item.name + ' [ ' + item.downloadUrl + ' ] '"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('container.imageName')" :rules="Rules.requiredInput" prop="imageName">
                    <el-input v-model="pullForm.imageName"></el-input>
                </el-form-item>
                <el-form-item v-if="pullForm.imageName !== ''">
                    <el-tag>docker pull {{ loadDetailInfo(pullForm.repoID) }}/{{ pullForm.imageName }}</el-tag>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="submitPull(pullFormRef)">
                        {{ $t('container.pull') }}
                    </el-button>
                    <el-button @click="pullVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog v-model="pushVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('container.imagePush') }} ({{ pushForm.imageName }})</span>
                </div>
            </template>
            <el-form ref="pushFormRef" :model="pushForm" label-width="80px">
                <el-form-item :label="$t('container.repoName')" :rules="Rules.requiredSelect" prop="repoID">
                    <el-select style="width: 100%" filterable v-model="pushForm.repoID">
                        <el-option
                            v-for="item in repos"
                            :key="item.id"
                            :value="item.id"
                            :label="item.name + ' [ ' + item.downloadUrl + ' ] '"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('container.label')" :rules="Rules.requiredInput" prop="tagName">
                    <el-input v-model="pushForm.tagName"></el-input>
                </el-form-item>
                <el-form-item v-if="pushForm.tagName !== ''">
                    <el-tag>
                        docker tag {{ pushForm.imageName }} {{ loadDetailInfo(pushForm.repoID) }}/{{ pushForm.tagName }}
                    </el-tag>
                </el-form-item>
                <el-form-item v-if="pushForm.tagName !== ''">
                    <el-tag>docker push {{ loadDetailInfo(pushForm.repoID) }}/{{ pushForm.tagName }}</el-tag>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="submitPush(pushFormRef)">
                        {{ $t('container.push') }}
                    </el-button>
                    <el-button @click="pushVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog v-model="saveVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('container.exportImage') }} ({{ saveForm.imageName }})</span>
                </div>
            </template>
            <el-form ref="saveFormRef" :model="saveForm" label-width="80px">
                <el-form-item :label="$t('container.path')" :rules="Rules.requiredSelect" prop="path">
                    <el-input clearable v-model="saveForm.path">
                        <template #append>
                            <FileList @choose="loadSaveDir" :dir="true"></FileList>
                        </template>
                    </el-input>
                </el-form-item>
                <el-form-item :label="$t('container.fileName')" :rules="Rules.requiredInput" prop="name">
                    <el-input v-model="saveForm.name">
                        <template #append>.tar</template>
                    </el-input>
                </el-form-item>
                <el-form-item v-if="saveForm.path !== '' && saveForm.name !== ''">
                    <el-tag>docker save {{ saveForm.imageName }} > {{ saveForm.path }}/{{ saveForm.name }}.tar</el-tag>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="submitSave(saveFormRef)">
                        {{ $t('container.export') }}
                    </el-button>
                    <el-button @click="saveVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog v-model="loadVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('container.importImage') }}</span>
                </div>
            </template>
            <el-form ref="loadFormRef" :model="loadForm" label-width="80px">
                <el-form-item :label="$t('container.path')" :rules="Rules.requiredSelect" prop="path">
                    <el-input clearable v-model="loadForm.path">
                        <template #append>
                            <FileList @choose="loadLoadDir" :dir="false"></FileList>
                        </template>
                    </el-input>
                </el-form-item>
                <el-form-item v-if="loadForm.path !== ''">
                    <el-tag>docker load &lt; {{ loadForm.path }}</el-tag>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="submitLoad(loadFormRef)">{{ $t('container.import') }}</el-button>
                    <el-button @click="loadVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, onMounted, ref } from 'vue';
import FileList from '@/components/file-list/index.vue';
import { dateFromat } from '@/utils/util';
import { Container } from '@/api/interface/container';
import {
    getImagePage,
    getRepoOption,
    imageLoad,
    imagePull,
    imagePush,
    imageRemove,
    imageSave,
} from '@/api/modules/container';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage, ElMessageBox } from 'element-plus';

const loading = ref(false);

const data = ref();
const repos = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    page: 1,
    pageSize: 10,
    total: 0,
});

type FormInstance = InstanceType<typeof ElForm>;
const pullVisiable = ref(false);
const pullFormRef = ref<FormInstance>();
const pullForm = reactive({
    repoID: 1,
    imageName: '',
});

const pushVisiable = ref(false);
const pushFormRef = ref<FormInstance>();
const pushForm = reactive({
    repoID: 1,
    imageName: '',
    tagName: '',
});

const saveVisiable = ref(false);
const saveFormRef = ref<FormInstance>();
const saveForm = reactive({
    imageName: '',
    path: '',
    name: '',
});

const loadVisiable = ref(false);
const loadFormRef = ref<FormInstance>();
const loadForm = reactive({
    path: '',
});

const search = async () => {
    const repoSearch = {
        page: paginationConfig.page,
        pageSize: paginationConfig.pageSize,
    };
    await getImagePage(repoSearch).then((res) => {
        if (res.data) {
            data.value = res.data.items;
        }
        paginationConfig.total = res.data.total;
    });
};
const loadRepos = async () => {
    const res = await getRepoOption();
    repos.value = res.data;
};

const loadSaveDir = async (path: string) => {
    saveForm.path = path;
};
const loadLoadDir = async (path: string) => {
    loadForm.path = path;
};

const submitPull = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            loading.value = true;
            pullVisiable.value = false;
            await imagePull(pullForm);
            loading.value = false;
            search();
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        } catch {
            loading.value = false;
            search();
        }
    });
};

const submitPush = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            loading.value = true;
            pushVisiable.value = false;
            await imagePush(pushForm);
            loading.value = false;
            search();
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        } catch {
            loading.value = false;
            search();
        }
    });
};

const submitLoad = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            loading.value = true;
            loadVisiable.value = false;
            await imageLoad(loadForm);
            loading.value = false;
            search();
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        } catch {
            loading.value = false;
            search();
        }
    });
};

const submitSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        try {
            loading.value = true;
            saveVisiable.value = false;
            await imageSave(saveForm);
            loading.value = false;
            search();
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        } catch {
            loading.value = false;
            search();
        }
    });
};

const onBatchDelete = async (row: Container.ImageInfo | null) => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.delete'), i18n.global.t('commons.msg.deleteTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        if (row) {
            loading.value = true;
            await imageRemove({ imageName: row.name + ':' + row.version });
            loading.value = false;
            search();
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            return;
        }
        let ps = [];
        for (const item of selects.value) {
            ps.push(imageRemove({ imageName: item.name + ':' + item.version }));
        }
        loading.value = true;
        Promise.all(ps)
            .then(() => {
                loading.value = false;
                search();
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
                search();
            });
    });
};

function loadDetailInfo(id: number) {
    for (const item of repos.value) {
        if (item.id === id) {
            return item.downloadUrl;
        }
    }
    return '';
}

const buttons = [
    {
        label: i18n.global.t('container.push'),
        click: (row: Container.ImageInfo) => {
            pushForm.imageName = row.name + ':' + row.version;
            pushVisiable.value = true;
        },
    },
    {
        label: i18n.global.t('container.export'),
        click: (row: Container.ImageInfo) => {
            saveForm.imageName = row.name + ':' + row.version;
            saveVisiable.value = true;
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Container.ImageInfo) => {
            onBatchDelete(row);
        },
    },
];

onMounted(() => {
    search();
    loadRepos();
});
</script>
