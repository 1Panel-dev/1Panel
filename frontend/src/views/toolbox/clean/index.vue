<template>
    <div v-loading="loading">
        <LayoutContent :title="$t('setting.diskClean')" :divider="true">
            <template #leftToolBar>
                <el-button type="primary" @click="scanData">
                    {{ $t('clean.scan') }}
                </el-button>
            </template>
            <template #main>
                <div v-if="scanStatus === 'beforeScan'" class="ml-5">
                    <div v-if="form.lastCleanTime">
                        <el-text class="clean_title">
                            {{ $t('clean.lastCleanTime', [form.lastCleanTime || '-']) }}
                        </el-text>
                        <div class="mt-4">
                            <el-text>
                                {{
                                    $t('clean.lastCleanHelper', [
                                        form.lastCleanData || '-',
                                        form.lastCleanSize ? computeSize(Number(form.lastCleanSize)) : '-',
                                    ])
                                }}
                            </el-text>
                        </div>
                    </div>
                    <div v-else>
                        <el-text class="clean_title">
                            <el-icon>
                                <MagicStick />
                            </el-icon>
                            {{ $t('clean.scanHelper') }}
                        </el-text>
                    </div>
                    <el-row type="flex" justify="center" :gutter="20" class="mb-10">
                        <el-col :xs="24" :sm="24" :md="24" :lg="12" :xl="12" v-for="item in cleanTag" :key="item.title">
                            <el-card class="clean-card">
                                <el-row>
                                    <el-col :span="4">
                                        <el-button
                                            v-if="item.icon.indexOf('p-') === -1"
                                            :icon="item.icon"
                                            link
                                            class="card_icon"
                                        />
                                        <template v-else>
                                            <el-button link class="svg-icon">
                                                <svg-icon :iconName="item.icon"></svg-icon>
                                            </el-button>
                                        </template>
                                    </el-col>
                                    <el-col :span="20">
                                        <div>
                                            <el-text class="mx-1 card_title" type="primary">
                                                {{ $t(item.title) }}
                                            </el-text>
                                        </div>
                                        <span class="input-help">{{ $t(item.description) }}</span>
                                    </el-col>
                                </el-row>
                            </el-card>
                        </el-col>
                    </el-row>
                </div>
                <div v-if="scanStatus === 'afterScan'" class="ml-5">
                    <el-text class="clean_title">{{ $t('clean.cleanSuccessful') }}</el-text>
                    <div class="mt-4">
                        <el-text>
                            {{
                                $t('clean.currentCleanHelper', [
                                    form.lastCleanData,
                                    computeSize(Number(form.lastCleanSize)),
                                ])
                            }}
                        </el-text>
                    </div>
                </div>
                <div v-if="scanStatus === 'scanned'" class="ml-5">
                    <div>
                        <el-text class="clean_title">{{ $t('clean.totalScan') }} {{ computeSize(totalSize) }}</el-text>
                        <el-button
                            type="primary"
                            class="-mt-2 ml-20"
                            :disabled="selectSize <= 0"
                            @click="onSubmitClean"
                        >
                            {{ $t('clean.clean') }}
                        </el-button>
                        <div class="mt-4">
                            <el-text type="info">{{ $t('clean.selectScan') }} {{ computeSize(selectSize) }}</el-text>
                        </div>
                    </div>
                    <el-row type="flex" :gutter="20" class="mb-10">
                        <el-col :xs="24" :sm="24" :md="24" :lg="16" :xl="16">
                            <div class="tree-container">
                                <el-collapse v-model="activeNames" class="tree-collapse">
                                    <el-collapse-item :title="$t('clean.system')" name="system">
                                        <el-tree
                                            ref="systemRef"
                                            :data="cleanData.systemClean"
                                            node-key="id"
                                            :default-checked-keys="systemDefaultCheck"
                                            show-checkbox
                                            :props="defaultProps"
                                            @check="onChange"
                                            class="responsive-tree"
                                            :empty-text="$t('clean.statusEmpty')"
                                        >
                                            <template #default="{ node, data }">
                                                <el-text>
                                                    {{
                                                        load18n(data.label) +
                                                        '     ' +
                                                        computeSize(data.size) +
                                                        '     ' +
                                                        loadTag(node, data)
                                                    }}
                                                </el-text>
                                            </template>
                                        </el-tree>
                                    </el-collapse-item>

                                    <el-collapse-item :title="$t('clean.unknownBackup')" name="backup">
                                        <el-tree
                                            ref="backupRef"
                                            :data="cleanData.backupClean"
                                            node-key="id"
                                            :default-checked-keys="backupDefaultCheck"
                                            show-checkbox
                                            :props="defaultProps"
                                            @check="onChange"
                                            class="responsive-tree"
                                            :empty-text="$t('clean.statusEmpty')"
                                        >
                                            <template #default="{ node, data }">
                                                <el-text>
                                                    {{
                                                        load18n(data.label) +
                                                        '     ' +
                                                        computeSize(data.size) +
                                                        '     ' +
                                                        loadTag(node, data)
                                                    }}
                                                </el-text>
                                            </template>
                                        </el-tree>
                                    </el-collapse-item>

                                    <el-collapse-item :title="$t('clean.docker')" name="docker">
                                        <el-tree
                                            ref="containerRef"
                                            :data="cleanData.containerClean"
                                            node-key="id"
                                            :default-checked-keys="containerDefaultCheck"
                                            show-checkbox
                                            :props="defaultProps"
                                            @check="onChange"
                                            class="responsive-tree"
                                            :empty-text="$t('clean.statusEmpty')"
                                        >
                                            <template #default="{ node, data }">
                                                <el-text>
                                                    {{
                                                        load18n(data.label) +
                                                        '     ' +
                                                        computeSize(data.size) +
                                                        '     ' +
                                                        loadTag(node, data)
                                                    }}
                                                </el-text>
                                            </template>
                                        </el-tree>
                                    </el-collapse-item>
                                    <el-collapse-item :title="$t('clean.upload')" name="upload">
                                        <el-tree
                                            ref="uploadRef"
                                            :data="cleanData.uploadClean"
                                            node-key="id"
                                            :default-checked-keys="uploadDefaultCheck"
                                            show-checkbox
                                            :props="defaultProps"
                                            @check="onChange"
                                            class="responsive-tree"
                                            :empty-text="$t('clean.statusEmpty')"
                                        >
                                            <template #default="{ node, data }">
                                                <el-text>
                                                    {{
                                                        load18n(data.label) +
                                                        '     ' +
                                                        computeSize(data.size) +
                                                        '     ' +
                                                        loadTag(node, data)
                                                    }}
                                                </el-text>
                                            </template>
                                        </el-tree>
                                    </el-collapse-item>
                                    <el-collapse-item :title="$t('clean.download')" name="download">
                                        <el-tree
                                            ref="downloadRef"
                                            :data="cleanData.downloadClean"
                                            node-key="id"
                                            :default-checked-keys="downloadDefaultCheck"
                                            show-checkbox
                                            :props="defaultProps"
                                            @check="onChange"
                                            class="responsive-tree"
                                            :empty-text="$t('clean.statusEmpty')"
                                        >
                                            <template #default="{ node, data }">
                                                <el-text>
                                                    {{
                                                        load18n(data.label) +
                                                        '     ' +
                                                        computeSize(data.size) +
                                                        '     ' +
                                                        loadTag(node, data)
                                                    }}
                                                </el-text>
                                            </template>
                                        </el-tree>
                                    </el-collapse-item>
                                    <el-collapse-item :title="$t('clean.systemLog')" name="system_log">
                                        <el-tree
                                            ref="systemLogRef"
                                            :data="cleanData.systemLogClean"
                                            node-key="id"
                                            :default-checked-keys="systemLogDefaultCheck"
                                            show-checkbox
                                            :props="defaultProps"
                                            @check="onChange"
                                            class="responsive-tree"
                                            :empty-text="$t('clean.statusEmpty')"
                                        >
                                            <template #default="{ node, data }">
                                                <el-text>
                                                    {{
                                                        load18n(data.label) +
                                                        '     ' +
                                                        computeSize(data.size) +
                                                        '     ' +
                                                        loadTag(node, data)
                                                    }}
                                                </el-text>
                                            </template>
                                        </el-tree>
                                    </el-collapse-item>
                                </el-collapse>
                            </div>
                        </el-col>
                    </el-row>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>
<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { computeSize } from '@/utils/util';
import { getSettingInfo } from '@/api/modules/setting';
import { clean, scan } from '@/api/modules/toolbox';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const totalSize = ref<number>(0);
const selectSize = ref<number>(0);

const scanStatus = ref<string>('beforeScan');

const defaultProps = {
    children: 'children',
    label: 'label',
};
const cleanData = reactive({
    systemClean: [],
    backupClean: [],
    uploadClean: [],
    downloadClean: [],
    systemLogClean: [],
    containerClean: [],
});
const systemRef = ref();
const systemDefaultCheck = ref([]);
const backupRef = ref();
const backupDefaultCheck = ref([]);
const uploadRef = ref();
const uploadDefaultCheck = ref([]);
const downloadRef = ref();
const downloadDefaultCheck = ref([]);
const systemLogRef = ref();
const systemLogDefaultCheck = ref([]);
const containerRef = ref();
const containerDefaultCheck = ref([]);
const activeNames = ref(['system', 'backup', 'upload', 'download', 'system_log', 'docker']);

const cleanTag = [
    { icon: 'Setting', title: i18n.global.t('clean.system'), description: i18n.global.t('clean.systemHelper') },
    { icon: 'Upload', title: i18n.global.t('clean.upload'), description: i18n.global.t('clean.uploadHelper') },
    { icon: 'Files', title: i18n.global.t('clean.backup'), description: i18n.global.t('clean.backupHelper') },
    { icon: 'Download', title: i18n.global.t('clean.download'), description: i18n.global.t('clean.downloadHelper') },
    {
        icon: 'p-docker1',
        title: i18n.global.t('clean.docker'),
        description: i18n.global.t('clean.dockerHelper'),
    },
    { icon: 'Document', title: i18n.global.t('clean.systemLog'), description: i18n.global.t('clean.systemLogHelper') },
];

const submitCleans = ref();

const form = reactive({
    lastCleanTime: '',
    lastCleanSize: '',
    lastCleanData: '',
});

const scanData = async () => {
    loading.value = true;
    await scan()
        .then((res) => {
            loading.value = false;
            selectSize.value = 0;
            totalSize.value = 0;
            cleanData.systemClean = res.data.systemClean || [];
            for (const item of cleanData.systemClean) {
                totalSize.value += item.size;
            }
            cleanData.backupClean = res.data.backupClean || [];
            for (const item of cleanData.backupClean) {
                totalSize.value += item.size;
            }
            cleanData.uploadClean = res.data.uploadClean || [];
            for (const item of cleanData.uploadClean) {
                totalSize.value += item.size;
            }
            cleanData.downloadClean = res.data.downloadClean || [];
            for (const item of cleanData.downloadClean) {
                totalSize.value += item.size;
            }
            cleanData.systemLogClean = res.data.systemLogClean || [];
            for (const item of cleanData.systemLogClean) {
                totalSize.value += item.size;
            }
            cleanData.containerClean = res.data.containerClean || [];
            for (const item of cleanData.containerClean) {
                totalSize.value += item.size;
            }
            loadCheck(cleanData.systemClean, systemDefaultCheck.value);
            loadCheck(cleanData.backupClean, backupDefaultCheck.value);
            loadCheck(cleanData.uploadClean, uploadDefaultCheck.value);
            loadCheck(cleanData.downloadClean, downloadDefaultCheck.value);
            loadCheck(cleanData.systemLogClean, systemLogDefaultCheck.value);
            loadCheck(cleanData.containerClean, containerDefaultCheck.value);
            scanStatus.value = 'scanned';
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSubmitClean = async () => {
    ElMessageBox.confirm(i18n.global.t('clean.cleanHelper'), i18n.global.t('clean.clean'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        submitCleans.value = [];
        loadSubmitCheck(cleanData.systemClean);
        loadSubmitCheck(cleanData.backupClean);
        loadSubmitCheck(cleanData.uploadClean);
        loadSubmitCheck(cleanData.downloadClean);
        loadSubmitCheck(cleanData.systemLogClean);
        loadSubmitCheck(cleanData.containerClean);
        await clean(submitCleans.value)
            .then(() => {
                form.lastCleanSize = selectSize.value + '';
                form.lastCleanData = submitCleans.value.length + '';
                scanStatus.value = 'afterScan';
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loading.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const search = async () => {
    const res = await getSettingInfo();
    form.lastCleanTime = res.data.lastCleanTime;
    form.lastCleanSize = res.data.lastCleanSize;
    form.lastCleanData = res.data.lastCleanData;
};

const loadSubmitCheck = (data: any) => {
    if (data.children === null) {
        if (data.isCheck) {
            submitCleans.value.push({ treeType: data.type, name: data.name, size: data.size });
        }
        return;
    }
    for (const item of data) {
        if (item.type === 'unknown_backup' && item.isCheck && item.children) {
            loadSubmitCheck(item.children);
            continue;
        }
        if (item.isCheck && item.type !== 'app_tmp_download') {
            submitCleans.value.push({ treeType: item.type, name: item.name, size: item.size });
            continue;
        }
        if (item.children) {
            loadSubmitCheck(item.children);
        }
    }
};

const changeCheckStatus = (data: any, checked: Array<string>) => {
    data.isCheck = checked.indexOf(data.id) !== -1;
    if (data.children) {
        for (const item of data.children) {
            changeCheckStatus(item, checked);
        }
    }
};

function onChange(data: any, checked: any) {
    let keys = checked.checkedKeys || [];
    changeCheckStatus(data, keys);
    selectSize.value = 0;
    let systemSelects = systemRef.value.getCheckedNodes(false, true);
    for (const item of systemSelects) {
        if (item.children === null) {
            selectSize.value = selectSize.value + Number(item.size);
        }
    }
    let backupSelects = backupRef.value.getCheckedNodes(false, true);
    for (const item of backupSelects) {
        if (item.children === null) {
            selectSize.value = selectSize.value + Number(item.size);
        }
    }
    let uploadSelects = uploadRef.value.getCheckedNodes(false, true);
    for (const item of uploadSelects) {
        if (item.children === null) {
            selectSize.value = selectSize.value + Number(item.size);
        }
    }
    let downloadSelects = downloadRef.value.getCheckedNodes(false, true);
    for (const item of downloadSelects) {
        if (item.children === null) {
            selectSize.value = selectSize.value + Number(item.size);
        }
    }
    let systemLogSelects = systemLogRef.value.getCheckedNodes(false, true);
    for (const item of systemLogSelects) {
        if (item.children === null) {
            selectSize.value = selectSize.value + Number(item.size);
        }
    }
    let containerSelects = containerRef.value.getCheckedNodes(false, true);
    for (const item of containerSelects) {
        if (item.children === null) {
            selectSize.value = selectSize.value + Number(item.size);
        }
    }
}

function loadCheck(data: any, checkList: any) {
    if (data.children === null) {
        if (data.isCheck) {
            checkList.push(data.id);
        }
        return;
    }
    for (const item of data) {
        if (item.isCheck) {
            selectSize.value = selectSize.value + Number(item.size);
            checkList.push(item.id);
            continue;
        }
        if (item.children) {
            loadCheck(item.children, checkList);
        }
    }
}

function loadTag(node: any, data: any) {
    if (node.level !== 1) {
        return '';
    }
    if (data.size === 0) {
        return i18n.global.t('clean.statusClean');
    }
    if (data.label === 'container_images') {
        return i18n.global.t('container.cleanImagesHelper');
    }
    if (data.label === 'container_containers') {
        return i18n.global.t('container.cleanContainersHelper');
    }
    if (data.label === 'container_volumes') {
        return i18n.global.t('container.cleanVolumesHelper');
    }
    if (data.label === 'upgrade') {
        return i18n.global.t('clean.upgradeHelper');
    }
    return data.isRecommend ? i18n.global.t('clean.statusSuggest') : i18n.global.t('clean.statusWarning');
}

function load18n(label: string) {
    switch (label) {
        case '1panel_original':
            return i18n.global.t('clean.panelOriginal');
        case 'upgrade':
            return i18n.global.t('clean.upgrade');
        case 'agent_packages':
            return i18n.global.t('clean.agentPackages');
        case 'snapshot':
            return i18n.global.t('clean.snapshot');
        case 'rollback':
            return i18n.global.t('clean.rollback');
        case 'tmp_backup':
            return i18n.global.t('clean.tmpBackup');
        case 'unknown_app':
            return i18n.global.t('clean.unknownApp');
        case 'unknown_database':
            return i18n.global.t('clean.unknownDatabase');
        case 'unknown_website':
            return i18n.global.t('clean.unknownWebsite');
        case 'unknown_snapshot':
            return i18n.global.t('clean.unknownSnapshot');
        case 'upload':
        case 'upload_tmp':
            return i18n.global.t('clean.upload');
        case 'download':
            return i18n.global.t('clean.download');
        case 'upload_website':
        case 'rollback_website':
        case 'tmp_backup_website':
        case 'download_website':
            return i18n.global.t('menu.website');
        case 'upload_app':
        case 'rollback_app':
        case 'tmp_backup_app':
        case 'download_app':
            return i18n.global.t('app.app');
        case 'upload_database':
        case 'rollback_database':
        case 'tmp_backup_database':
        case 'download_database':
            return i18n.global.t('menu.database');
        case 'upload_directory':
        case 'tmp_backup_directory':
        case 'download_directory':
            return i18n.global.t('clean.directory');
        case 'system_log':
            return i18n.global.t('logs.system');
        case 'tmp_backup_log':
            return i18n.global.t('clean.systemLog');
        case 'script':
            return i18n.global.t('cronjob.library.script');
        case 'docker_log':
            return i18n.global.t('clean.dockerLog');
        case 'task_log':
            return i18n.global.t('logs.task');
        case 'Database':
        case 'Cronjob':
        case 'Container':
        case 'App':
        case 'System':
        case 'Website':
            return i18n.global.t('menu.' + label.toLowerCase());
        case 'Backup':
            return i18n.global.t('logs.detail.backups');
        case 'Tamper':
            return i18n.global.t('xpack.tamper.tamper');
        case 'TaskRollback':
            return i18n.global.t('setting.rollback');
        case 'Script':
            return i18n.global.t('cronjob.library.library');
        case 'ScriptLibrary':
            return i18n.global.t('cronjob.library.remoteLibrary');
        case 'Compose':
            return i18n.global.t('container.compose');
        case 'CustomAppstore':
            return i18n.global.t('xpack.customApp.name');
        case 'NodeFile':
            return i18n.global.t('xpack.exchange.exchange');
        case 'RuntimeExtension':
            return i18n.global.t('website.runtime');
        case 'Image':
            return i18n.global.t('container.image');
        case 'Snapshot':
        case 'tmp_backup_snapshot':
            return i18n.global.t('setting.snapshot');
        case 'AppStore':
            return i18n.global.t('menu.apps');
        case 'AITools':
            return i18n.global.t('menu.aiTools');
        case 'shell':
            return i18n.global.t('clean.shell');
        case 'containerShell':
            return i18n.global.t('clean.containerShell');
        case 'curl':
            return i18n.global.t('clean.curl');
        case 'container_images':
            return i18n.global.t('container.image');
        case 'container_containers':
            return i18n.global.t('menu.container');
        case 'container_volumes':
            return i18n.global.t('clean.volumes');
        case 'build_cache':
            return i18n.global.t('clean.buildCache');
        case 'website_log':
            return i18n.global.t('logs.websiteLog');
        case 'app_tmp_download':
            return i18n.global.t('clean.appTmpDownload');
        case 'unknown_website_log':
            return i18n.global.t('clean.unknownWebsiteLog');
        default:
            return label;
    }
}

onMounted(() => {
    search();
    scanStatus.value = 'beforeScan';
});
</script>

<style lang="scss" scoped>
.clean-card {
    margin-top: 20px;
    cursor: pointer;
    border: var(--panel-border) !important;

    &:hover {
        cursor: pointer;
        border: 1px solid var(--el-color-primary) !important;
    }
}

.card_icon {
    font-size: 36px;
    float: right;
    margin-right: 15px;
}

.card_title {
    font-size: 18px;
}

.clean_title {
    font-size: 22px;
}

.large_button {
    float: right;
    margin-top: -40px;
}

.svg-icon {
    font-size: 14px;
    float: right;
    margin-right: 5px;
}

.tree-container {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: auto;
}
.tree-collapse {
    flex: 1;
    overflow: hidden;
}
.responsive-tree {
    width: 100%;
    max-height: 100%;
    overflow: auto;

    :deep(.el-tree-node .el-tree-node__content) {
        width: fit-content;
        overflow: auto;
        min-width: 300px;
    }
}
</style>
