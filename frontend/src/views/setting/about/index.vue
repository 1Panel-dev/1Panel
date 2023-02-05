<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.about')" :divider="true">
            <template #main>
                <div style="text-align: center; margin-top: 20px">
                    <div style="justify-self: center">
                        <img style="width: 80px" src="@/assets/images/1panel-small.png" />
                    </div>
                    <h3>{{ $t('setting.description') }}</h3>
                    <h3>
                        {{ version }}
                        <el-button v-if="version !== 'Waiting'" type="primary" link @click="onLoadUpgradeInfo">
                            {{ $t('setting.upgradeCheck') }}
                        </el-button>
                        <el-tag v-else round style="margin-left: 10px">{{ $t('setting.upgrading') }}</el-tag>
                    </h3>
                    <div style="margin-top: 10px">
                        <el-link @click="toGithub">
                            <svg-icon style="font-size: 7px; margin-bottom: 3px" iconName="p-huaban88"></svg-icon>
                            <span style="line-height: 20px">{{ $t('setting.project') }}</span>
                        </el-link>
                        <el-link @click="toIssue" style="margin-left: 15px">
                            <svg-icon style="font-size: 7px; margin-bottom: 3px" iconName="p-bug"></svg-icon>
                            <span>{{ $t('setting.issue') }}</span>
                        </el-link>
                        <el-link @click="toTalk" style="margin-left: 15px">
                            <svg-icon style="font-size: 7px; margin-bottom: 3px" iconName="p-taolun"></svg-icon>
                            <span>{{ $t('setting.chat') }}</span>
                        </el-link>
                        <el-link @click="toGithubStar" style="margin-left: 15px">
                            <svg-icon style="font-size: 7px; margin-bottom: 3px" iconName="p-star"></svg-icon>
                            <span>{{ $t('setting.star') }}</span>
                        </el-link>
                    </div>
                </div>
            </template>
        </LayoutContent>
        <el-drawer :key="refresh" v-model="drawerVisiable" size="50%">
            <template #header>
                <DrawerHeader :header="$t('setting.upgrade')" :back="handleClose" />
            </template>
            <el-form label-width="120px">
                <el-form-item :label="$t('setting.newVersion')">
                    <el-tag>{{ upgradeInfo.newVersion }}</el-tag>
                </el-form-item>
                <el-form-item :label="$t('commons.table.createdAt')">
                    <el-tag>{{ upgradeInfo.createdAt }}</el-tag>
                </el-form-item>
                <el-form-item :label="$t('setting.upgradeNotes')">
                    <MdEditor style="height: calc(100vh - 330px)" v-model="upgradeInfo.releaseNote" previewOnly />
                </el-form-item>
                <el-form-item :label="$t('setting.source')">
                    <el-radio-group v-model="Source">
                        <el-radio label="gitee">Gitee</el-radio>
                        <el-radio label="github">GitHub</el-radio>
                    </el-radio-group>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="onUpgrade">{{ $t('setting.upgradeNow') }}</el-button>
                </el-form-item>
            </el-form>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import { getSettingInfo, loadUpgradeInfo, upgrade } from '@/api/modules/setting';
import { onMounted, ref } from 'vue';
import MdEditor from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import { ElMessage, ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';

const version = ref();
const upgradeInfo = ref();
const Source = ref('gitee');
const drawerVisiable = ref();
const refresh = ref();

const loading = ref();
const search = async () => {
    const res = await getSettingInfo();
    version.value = res.data.systemVersion;
};

const toGithub = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank');
};
const toIssue = () => {
    window.open('https://github.com/1Panel-dev/1Panel/issues', '_blank');
};
const toTalk = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank');
};
const toGithubStar = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank');
};

const handleClose = () => {
    drawerVisiable.value = false;
};

const onLoadUpgradeInfo = async () => {
    const res = await loadUpgradeInfo();
    if (!res.data) {
        ElMessage.info(i18n.global.t('setting.noUpgrade'));
        return;
    }
    upgradeInfo.value = res.data;
    drawerVisiable.value = true;
};
const onUpgrade = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.upgradeHelper', i18n.global.t('setting.upgrade')), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(() => {
        loading.value = true;
        let param = {
            version: upgradeInfo.value.newVersion,
            source: Source.value,
        };
        upgrade(param)
            .then(() => {
                loading.value = false;
                drawerVisiable.value = false;
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

onMounted(() => {
    search();
});
</script>
