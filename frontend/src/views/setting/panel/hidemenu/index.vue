<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.advancedMenuHide')" :back="handleClose" />
            </template>

            <ComplexTable
                :data="treeData.hideMenu"
                :show-header="false"
                style="width: 100%; margin-bottom: 20px"
                row-key="id"
                default-expand-all
            >
                <el-table-column prop="title" :label="$t('setting.menu')">
                    <template #default="{ row }">
                        {{ i18n.global.t(row.title) }}
                    </template>
                </el-table-column>
                <el-table-column prop="isCheck" :label="$t('setting.ifShow')">
                    <template #default="{ row }">
                        <el-switch v-model="row.isCheck" @change="onSaveStatus(row)" />
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
        </el-drawer>
    </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { updateMenu } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';

const drawerVisible = ref();
const loading = ref();
const defaultCheck = ref([]);
const emit = defineEmits<{ (e: 'search'): void }>();
const globalStore = GlobalStore();
interface DialogProps {
    menuList: string;
}
const menuList = ref();

const treeData = reactive({
    hideMenu: [],
    checkedData: [],
});

function loadCheck(data: any, checkList: any) {
    if (data.children === null) {
        if (data.isCheck) {
            checkList.push(data.id);
        }
        return;
    }
    for (const item of data) {
        if (item.isCheck) {
            checkList.push(item.id);
            continue;
        }
        if (item.children) {
            loadCheck(item.children, checkList);
        }
    }
}

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

const onSaveStatus = async (row: any) => {
    if (row.label === '/xpack') {
        if (!row.isCheck) {
            for (const item of treeData.hideMenu[0].children) {
                item.isCheck = false;
            }
        } else {
            let flag = false;
            for (const item of treeData.hideMenu[0].children) {
                if (item.isCheck) {
                    flag = true;
                }
            }
            if (!flag && row.isCheck) {
                for (const item of treeData.hideMenu[0].children) {
                    item.isCheck = true;
                }
            }
        }
    } else {
        let flag = false;
        if (row.isCheck) {
            treeData.hideMenu[0].isCheck = true;
        }
        for (const item of treeData.hideMenu[0].children) {
            if (item.isCheck) {
                flag = true;
            }
        }
        if (!flag) {
            treeData.hideMenu[0].isCheck = false;
        }
    }
};

const acceptParams = (params: DialogProps): void => {
    menuList.value = params.menuList;
    drawerVisible.value = true;
    treeData.hideMenu = [];
    defaultCheck.value = [];
    treeData.hideMenu.push(JSON.parse(menuList.value));
    if (globalStore.isIntl) {
        treeData.hideMenu = removeXAlertDashboard(treeData.hideMenu);
    }
    loadCheck(treeData.hideMenu, defaultCheck.value);
};

const handleClose = () => {
    drawerVisible.value = false;
};

const saveHideMenus = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.confirmMessage'), i18n.global.t('setting.advancedMenuHide'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        const updateJson = JSON.stringify(treeData.hideMenu[0]);
        await updateMenu({ key: 'XpackHideMenu', value: updateJson })
            .then(async () => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loading.value = false;
                drawerVisible.value = false;
                emit('search');
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
