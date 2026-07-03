<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.menuSetting')" @close="handleClose" size="normal">
        <div class="menu-setting-cards">
            <el-card shadow="never" class="menu-setting-card">
                <div class="menu-setting-card__row">
                    <div>
                        <div class="menu-setting-card__label">{{ $t('setting.menuAccordion') }}</div>
                        <span class="input-help">{{ $t('setting.menuAccordionHelper') }}</span>
                    </div>
                    <el-radio-group
                        @change="onSaveSetting('MenuAccordion', form.menuAccordion)"
                        v-model="form.menuAccordion"
                    >
                        <el-radio-button value="Enable">
                            <span>{{ $t('commons.button.enable') }}</span>
                        </el-radio-button>
                        <el-radio-button value="Disable">
                            <span>{{ $t('commons.button.disable') }}</span>
                        </el-radio-button>
                    </el-radio-group>
                </div>
            </el-card>

            <el-card shadow="never" class="menu-setting-card">
                <div class="menu-setting-card__label mb-3">{{ $t('setting.menuHide') }}</div>
                <el-alert :closable="false" :title="$t('setting.menuSettingHelper')" type="warning" />
                <el-tree
                    :data="treeData.hideMenu"
                    :allow-drag="allowDrag"
                    :allow-drop="allowDrop"
                    draggable
                    node-key="id"
                    class="mt-3 menu-hide-tree"
                    :icon="ArrowRight"
                    @node-drop="handleDrop"
                >
                    <template #default="{ node, data }">
                        <div class="grid grid-cols-4 gap-4 items-center w-full py-2 group">
                            <span class="col-span-2" :style="{ paddingLeft: `${(node.level - 1) * 16}px` }">
                                {{ i18n.global.t(data.title) }}
                            </span>
                            <span class="flex justify-center w-[60px]">
                                <el-switch
                                    v-if="!data.disabled"
                                    v-model="data.isShow"
                                    @change="onChangeShow(data)"
                                    @click.stop
                                    @mousedown.stop
                                />
                                <span v-else>-</span>
                            </span>
                            <span
                                class="text-right hidden cursor-move"
                                :class="
                                    data.label == 'Home-Menu' || data.children?.length > 0 ? '' : 'group-hover:block'
                                "
                            >
                                ⋮⋮
                            </span>
                        </div>
                    </template>
                </el-tree>
            </el-card>
        </div>
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
import { defaultMenu, updateMenu, updateSetting } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { ArrowRight } from '@element-plus/icons-vue';
import { sortMenu } from '@/utils/misc';
const { isEE, isIntl, isAdmin, menuAccordion } = useGlobalStore();

const drawerVisible = ref();
const loading = ref();
const em = defineEmits(['search']);
interface DialogProps {
    hideMenu: string;
    menuAccordion: string;
}

const form = reactive({
    menuAccordion: '',
});

const acceptParams = (params: DialogProps): void => {
    drawerVisible.value = true;
    form.menuAccordion = params.menuAccordion || 'Disable';
    let hideMenu = JSON.parse(params.hideMenu);
    sortMenu(hideMenu);
    treeData.hideMenu = hideMenu;
    if (isIntl.value || (isEE.value && !isAdmin.value)) {
        treeData.hideMenu = removeUpage(treeData.hideMenu);
    }
};
type Node = RenderContentContext['node'];

const allowDrag = (draggingNode: Node) => {
    const { label } = draggingNode.data;
    const forbidden = ['Home-Menu'];
    return draggingNode.level < 3 && !forbidden.some((key) => label.includes(key));
};

const allowDrop = (draggingNode: Node, dropNode: Node, type: AllowDropType) => {
    const restricted = ['App-Menu', 'Setting-Menu'];
    const isDraggingFirstLevel = draggingNode.level === 1;
    const isDraggingSecondLevel = draggingNode.level === 2;
    const isDropFirstLevel = dropNode.level === 1;
    const isDropSecondLevel = dropNode.level === 2;
    if (restricted.includes(draggingNode.data.label) && isDropSecondLevel) {
        return false;
    }
    if (dropNode.data.label === 'Home-Menu') {
        return type !== 'prev' && type !== 'inner';
    }

    if (draggingNode.parent && draggingNode.parent.childNodes.length === 1) {
        return false;
    }

    if (
        (isDraggingSecondLevel && isDropFirstLevel) ||
        (isDraggingFirstLevel && isDropSecondLevel && draggingNode.childNodes?.length === 0)
    ) {
        return type === 'prev' || type === 'next';
    }

    if ((isDropFirstLevel && isDraggingFirstLevel) || (isDropSecondLevel && isDraggingSecondLevel)) {
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

const removeUpage = (data: any): any => {
    return data
        .filter((item: { label: string }) => item.label !== 'Upage' && item.label !== 'XApp')
        .map((item: { children: any }) => {
            if (Array.isArray(item.children)) {
                item.children = removeUpage(item.children);
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

const onSaveSetting = async (key: string, val: string) => {
    loading.value = true;
    try {
        await updateSetting({
            key,
            value: val,
        });
        if (key === 'MenuAccordion') {
            menuAccordion.value = val === 'Enable';
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        em('search');
    } finally {
        loading.value = false;
    }
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
    font-size: 14px;
}

:deep(.el-tree-node__content) {
    padding: 8px 8px !important;
    border-bottom: var(--panel-border);
}

.menu-setting-cards {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.menu-setting-card {
    --el-card-padding: 14px 16px;
}

.menu-setting-card__row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.menu-setting-card__label {
    color: var(--el-text-color-primary);
    font-size: var(--el-font-size-base);
}

.menu-setting-card__title {
    display: inline-flex;
    gap: 4px;
    margin-bottom: 12px;
    color: var(--el-text-color-primary);
    font-size: 16px;
    font-weight: 500;
}

@media (max-width: 640px) {
    .menu-setting-card__row {
        align-items: flex-start;
        flex-direction: column;
        gap: 8px;
    }
}
</style>
