<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.menuSetting')" @close="handleClose" size="normal">
        <el-alert :closable="false" :title="$t('setting.menuSettingHelper')" type="warning" />
        <el-tree
            :data="treeData.hideMenu"
            :allow-drag="allowDrag"
            :allow-drop="allowDrop"
            draggable
            node-key="id"
            class="mt-3"
            :icon="ArrowRight"
            @node-drop="handleDrop"
        >
            <template #default="{ node, data }">
                <div class="grid grid-cols-4 gap-4 items-center w-full py-2 group">
                    <span class="col-span-2" :style="{ paddingLeft: `${(node.level - 1) * 16}px` }">
                        {{ i18n.global.t(data.title) }}
                    </span>
                    <span class="flex justify-center w-[60px]">
                        <el-switch v-if="!data.disabled" v-model="data.isShow" @change="onChangeShow(data)" />
                        <span v-else>-</span>
                    </span>
                    <span
                        class="text-right hidden cursor-move"
                        :class="data.label == 'Home-Menu' || data.children?.length > 0 ? '' : 'group-hover:block'"
                    >
                        ⋮⋮
                    </span>
                </div>
            </template>
        </el-tree>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="defaultHideMenus">{{ $t('commons.button.setDefault') }}</el-button>
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
import { AllowDropType, ElMessageBox, RenderContentContext } from 'element-plus';
import i18n from '@/lang';
import { defaultMenu, updateMenu } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
import { ArrowRight } from '@element-plus/icons-vue';
import { sortMenu } from '@/utils/util';
const globalStore = GlobalStore();

const drawerVisible = ref();
const loading = ref();
interface DialogProps {
    hideMenu: string;
}

const acceptParams = (params: DialogProps): void => {
    drawerVisible.value = true;
    let hideMenu = JSON.parse(params.hideMenu);
    sortMenu(hideMenu);
    treeData.hideMenu = hideMenu;
    if (globalStore.isIntl) {
        treeData.hideMenu = removeXAlertDashboard(treeData.hideMenu);
    }
};
type Node = RenderContentContext['node'];

const allowDrag = (draggingNode: Node) => {
    return !draggingNode.data.label.includes('Home-Menu') || draggingNode.level < 3;
};

const allowDrop = (draggingNode: Node, dropNode: Node, type: AllowDropType) => {
    if (dropNode.data.label === 'Home-Menu') {
        return type !== 'prev' && type !== 'inner';
    }
    const draggingHasChildren = draggingNode.childNodes?.length > 0;
    const isDraggingTooDeep = draggingNode.level > 2;
    const isDraggingFirstLevel = draggingNode.level === 1;

    const isDropFirstLevel = dropNode.level === 1;
    const isDropSecondLevel = dropNode.level === 2;
    const dropHasChildren = dropNode.childNodes?.length > 0;

    if (draggingNode.parent && draggingNode.parent.childNodes.length === 1) {
        return false;
    }

    if (isDropFirstLevel && dropHasChildren) {
        return type === 'inner' || type === 'prev' || type === 'next';
    }

    if (isDraggingFirstLevel && draggingNode.childNodes?.length === 0 && dropNode.level !== 2) {
        return type === 'prev' || type === 'next';
    }

    if (isDraggingFirstLevel && draggingNode.childNodes?.length > 0 && dropNode.level !== 2) {
        return type === 'inner' || type === 'prev' || type === 'next';
    }

    if (draggingHasChildren || isDraggingTooDeep) {
        return false;
    }

    if (isDropSecondLevel && draggingNode.level === 2) {
        return type === 'prev' || type === 'next';
    }

    return false;
};

const handleDrop = (draggingNode: Node, dropNode: Node) => {
    const siblingNodes = dropNode.level == 2 ? dropNode.parent.parent.data : dropNode.parent.data;
    siblingNodes.forEach((node, index) => {
        node.sort = (index + 1) * 100;
    });

    const updateChildSort = (nodes) => {
        nodes.forEach((node, index) => {
            node.sort = (index + 1) * 100;
            if (node.children && node.children.length) {
                updateChildSort(node.children);
            }
        });
    };

    if (siblingNodes.length) {
        siblingNodes.forEach((node) => {
            if (node.children && node.children.length) {
                updateChildSort(node.children);
            }
        });
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
        return;
    }
    for (const item of treeData.hideMenu) {
        if (!item.children) {
            continue;
        }
        let allHide = true;
        for (const item2 of item.children) {
            if (item2.isShow) {
                allHide = false;
            }
            if (item2.id === row.id && item2.isShow) {
                item.isShow = true;
                return;
            }
        }
        if (allHide) {
            item.isShow = false;
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
const defaultHideMenus = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.recoverMessage'), i18n.global.t('setting.menuSetting'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        await defaultMenu()
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
<style scoped lang="scss">
:deep(.el-tree) {
    --el-tree-node-content-height: 26px;
    font-size: 16px;
}
:deep(.el-tree-node__content) {
    padding: 8px 8px !important;
    border-bottom: var(--panel-border);
}
</style>
