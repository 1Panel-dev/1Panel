<template>
    <div>
        <Submenu activeName="about" />
        <el-card style="margin-top: 20px">
            <LayoutContent :header="$t('setting.about')">
                <div style="text-align: center">
                    <div style="justify-self: center">
                        <img style="width: 80px" src="@/assets/images/ko_image.png" />
                    </div>
                    <h3>{{ $t('setting.description') }}</h3>
                    <h3>
                        {{ version }}
                        <el-button type="primary" link @click="onLoadUpgradeInfo">
                            {{ $t('setting.upgradeCheck') }}
                        </el-button>
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
            </LayoutContent>
        </el-card>
        <el-drawer :key="refresh" v-model="drawerShow" size="50%">
            <el-form label-width="120px">
                <el-form-item :label="$t('setting.newVersion')">
                    <el-tag>{{ upgradeInfo.newVersion }}</el-tag>
                </el-form-item>
                <el-form-item :label="$t('setting.tag')">
                    <el-tag>{{ upgradeInfo.tag }}</el-tag>
                </el-form-item>
                <el-form-item :label="$t('setting.upgradeNotes')">
                    <MdEditor style="height: 450px" v-model="upgradeInfo.releaseNote" previewOnly />
                </el-form-item>
                <el-form-item :label="$t('commons.table.createdAt')">
                    <el-tag>{{ upgradeInfo.createdAt }}</el-tag>
                </el-form-item>
                <el-form-item :label="$t('commons.table.publishedAt')">
                    <el-tag>{{ upgradeInfo.publishedAt }}</el-tag>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary">{{ $t('setting.upgradeNow') }}</el-button>
                </el-form-item>
            </el-form>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import { getSettingInfo, loadUpgradeInfo } from '@/api/modules/setting';
import Submenu from '@/views/setting/index.vue';
import { onMounted, ref } from 'vue';
import MdEditor from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';

const version = ref();
const upgradeInfo = ref();
const drawerShow = ref();
const refresh = ref();

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

const onLoadUpgradeInfo = async () => {
    const res = await loadUpgradeInfo();
    upgradeInfo.value = res.data;
    drawerShow.value = true;
};

onMounted(() => {
    search();
});
</script>
