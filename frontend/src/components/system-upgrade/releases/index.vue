<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.release')" @close="handleClose" size="large">
        <div class="note">
            <el-collapse
                v-if="notes && notes.length !== 0"
                v-model="currentVersion"
                :accordion="true"
                v-loading="loading"
            >
                <div v-for="(item, index) in notes" :key="index">
                    <el-collapse-item :name="index">
                        <template #title>
                            <div>
                                <span class="version">{{ item.version }}</span>
                                <span class="date">{{ item.createdAt }}</span>
                            </div>
                            <svg-icon class="icon" iconName="p-featureshitu"></svg-icon>
                            <span class="icon-span">{{ item.newCount }}</span>
                            <svg-icon class="icon" iconName="p-youhuawendang"></svg-icon>
                            <span class="icon-span">{{ item.optimizationCount }}</span>
                            <svg-icon class="icon" iconName="p-bug"></svg-icon>
                            <span class="icon-span">{{ item.fixCount }}</span>
                        </template>
                        <div class="panel-MdEditor">
                            <MdEditor v-model="item.content" previewOnly :theme="isDarkTheme ? 'dark' : 'light'" />
                        </div>
                    </el-collapse-item>
                </div>
            </el-collapse>
            <el-empty v-else>
                <template #description>
                    <span class="input-help">
                        {{ $t('setting.releaseHelper') }}
                        <el-link class="pageRoute" icon="Position" type="primary">
                            {{ $t('firewall.quickJump') }}
                        </el-link>
                    </span>
                </template>
            </el-empty>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { listReleases } from '@/api/modules/setting';
import MdEditor from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import { ref } from 'vue';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';

const globalStore = GlobalStore();
const { isDarkTheme } = storeToRefs(globalStore);

const drawerVisible = ref(false);
const currentVersion = ref(0);
const notes = ref([]);
const loading = ref();

const acceptParams = (): void => {
    search();
    drawerVisible.value = true;
};

const handleClose = () => {
    drawerVisible.value = false;
};

const search = async () => {
    loading.value = true;
    await listReleases()
        .then((res) => {
            notes.value = res.data || [];
            loading.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.version {
    margin-left: 10px;
    display: inline-block;
    width: 50px;
}
.date {
    margin-left: 20px;
    margin-right: 40px;
    display: inline-block;
    width: 100px;
}
.icon-span {
    display: inline-block;
    width: 10px;
}
.panel-MdEditor {
    :deep(.md-editor-preview) {
        font-size: 12px;
    }
}
:deep(.md-editor-dark) {
    background-color: var(--panel-main-bg-color-9);
}
:deep(.el-collapse-item__content) {
    padding: 0px;
}
.icon {
    font-size: 7px;
    margin-left: 50px;
}
.pageRoute {
    font-size: 12px;
    margin-left: 5px;
    margin-top: -4px;
}
</style>
