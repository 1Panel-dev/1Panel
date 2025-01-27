<template>
    <DrawerPro v-model="open" :header="$t('runtime.extension')" size="large" :back="handleClose">
        <el-descriptions title="" border>
            <el-descriptions-item :label="$t('runtime.loadedExtension')" width="100px">
                <el-tag v-for="(ext, index) in extensions" :key="index" type="info" class="mr-1 mt-1">{{ ext }}</el-tag>
            </el-descriptions-item>
        </el-descriptions>
        <div class="mt-5">
            <el-text>{{ $t('runtime.popularExtension') }}</el-text>
        </div>
        <ComplexTable :data="supportExtensions" @search="search()" :heightDiff="350" v-loading="loading">
            <el-table-column prop="name" :label="$t('commons.table.name')" />
            <el-table-column prop="installed" :label="$t('commons.table.status')">
                <template #default="{ row }">
                    <el-icon v-if="row.installed" color="green"><Select /></el-icon>
                    <el-icon v-else><CloseBold /></el-icon>
                </template>
            </el-table-column>
            <fu-table-operations
                :ellipsis="2"
                width="100px"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
    </DrawerPro>
    <TaskLog ref="taskLogRef" @close="search()" />
</template>

<script setup lang="ts">
import { Runtime } from '@/api/interface/runtime';
import { GetPHPExtensions, InstallPHPExtension, UnInstallPHPExtension } from '@/api/modules/runtime';
import i18n from '@/lang';
import { ref } from 'vue';
import { newUUID } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const runtime = ref();
const extensions = ref([]);
const supportExtensions = ref([]);
const loading = ref(false);
const taskLogRef = ref();

const handleClose = () => {
    open.value = false;
};

const buttons = [
    {
        label: i18n.global.t('commons.button.install'),
        click: function (row: Runtime.SupportExtension) {
            installExtension(row);
        },
        show: function (row: Runtime.SupportExtension) {
            return !row.installed;
        },
    },
    {
        label: i18n.global.t('commons.commons.button.uninstall'),
        click: function (row: Runtime.SupportExtension) {
            unInstallPHPExtension(row);
        },
        show: function (row: Runtime.SupportExtension) {
            return row.installed;
        },
    },
];

const installExtension = async (row: Runtime.SupportExtension) => {
    ElMessageBox.confirm(i18n.global.t('runtime.installExtension', [row.name]), i18n.global.t('runtime.extension'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        const req = {
            id: runtime.value.id,
            name: row.name,
            taskID: newUUID(),
        };
        loading.value = true;
        try {
            await InstallPHPExtension(req);
            taskLogRef.value.openWithTaskID(req.taskID);

            loading.value = false;
        } catch (error) {}
    });
};

const unInstallPHPExtension = async (row: Runtime.SupportExtension) => {
    ElMessageBox.confirm(i18n.global.t('runtime.uninstallExtension', [row.name]), i18n.global.t('runtime.extension'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        const req = {
            id: runtime.value.id,
            name: row.name,
        };
        loading.value = true;
        try {
            await UnInstallPHPExtension(req);
            MsgSuccess(i18n.global.t('commons.msg.uninstallSuccess'));
            loading.value = false;
            search();
        } catch (error) {}
    });
};

const search = async () => {
    try {
        const res = await GetPHPExtensions(runtime.value.id);
        extensions.value = res.data.extensions;
        supportExtensions.value = res.data.supportExtensions;
    } catch (error) {}
};

const acceptParams = (req: Runtime.Runtime): void => {
    open.value = true;
    runtime.value = req;
    search();
};

defineExpose({ acceptParams });
</script>
