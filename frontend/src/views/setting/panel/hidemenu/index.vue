<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.menuSetting')" :back="handleClose" size="small">
        <el-alert :closable="false" :title="$t('setting.menuSettingHelper')" type="warning" />
        <ComplexTable :heightDiff="1" :data="treeData.hideMenu" :show-header="false" row-key="id">
            <el-table-column prop="title" :label="$t('setting.menu')">
                <template #default="{ row }">
                    {{ i18n.global.t(row.title) }}
                </template>
            </el-table-column>
            <el-table-column prop="isShow" :label="$t('setting.ifShow')">
                <template #default="{ row }">
                    <el-switch v-if="!row.disabled" v-model="row.isShow" @change="onChangeShow(row)" />
                    <span v-else>-</span>
                </template>
            </el-table-column>
        </ComplexTable>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="saveHideMenus">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { updateMenu } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const drawerVisible = ref();
const loading = ref();
interface DialogProps {
    hideMenu: string;
}
const acceptParams = (params: DialogProps): void => {
    drawerVisible.value = true;
    treeData.hideMenu = JSON.parse(params.hideMenu) || [];
    if (globalStore.isIntl) {
        treeData.hideMenu = removeXAlertDashboard(treeData.hideMenu);
    }
};

const treeData = reactive({
    hideMenu: [],
    checkedData: [],
});

const removeXAlertDashboard = (data: any): any => {
    return data
        .filter((item: { label: string }) => item.label !== 'XAlertDashboard')
        .map((item: { children: any }) => {
            if (Array.isArray(item.children)) {
                item.children = removeXAlertDashboard(item.children);
            }
            return item;
        });
};

const onChangeShow = async (row: any) => {
    if (row.children) {
        for (const item of row.children) {
            item.isShow = row.isShow;
        }
    }
};

const handleClose = () => {
    drawerVisible.value = false;
};

const saveHideMenus = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.confirmMessage'), i18n.global.t('setting.menuSetting'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        const updateJson = JSON.stringify(treeData.hideMenu);
        await updateMenu({ key: 'HideMenu', value: updateJson })
            .then(async () => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loading.value = false;
                drawerVisible.value = false;
                window.location.reload();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
