<template>
    <div>
        <el-drawer v-model="drawerVisible" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.advancedMenuShow')" :back="handleClose" />
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
import { DialogProps } from 'element-plus';
import i18n from '@/lang';
import { updateSetting } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

const drawerVisible = ref();
const loading = ref();
const defaultCheck = ref([]);
const emit = defineEmits<{ (e: 'search'): void }>();
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

const onSaveStatus = async (row: any) => {
    if (row.label === '/xpack') {
        if (!row.isCheck) {
            // 如果父节点的开关关闭，子节点全部关闭
            for (const item of treeData.hideMenu[0].children) {
                item.isCheck = false;
            }
        } else {
            // 如果全部子节点关闭，父节点打开，则所有子节点自动打开
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
        // 如果有一个子节点打开，则父节点也打开
        if (row.isCheck) {
            treeData.hideMenu[0].isCheck = true;
        }
        // 如果全部子开关关闭，则父节点也关闭
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
    loadCheck(treeData.hideMenu, defaultCheck.value);
};

const handleClose = () => {
    drawerVisible.value = false;
};

const saveHideMenus = async () => {
    const updateJson = JSON.stringify(treeData.hideMenu[0]);
    await updateSetting({ key: 'XpackHideMenu', value: updateJson })
        .then(async () => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            loading.value = false;
            drawerVisible.value = false;
            emit('search');
            window.location.reload();
            return;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss"></style>
