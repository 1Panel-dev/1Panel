<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.about')" :divider="true">
            <template #main>
                <div style="text-align: center; margin-top: 20px">
                    <div style="justify-self: center">
                        <img style="width: 80px" src="@/assets/images/1panel-logo-light.png" />
                    </div>
                    <h3>{{ $t('setting.description') }}</h3>

                    <el-upload
                        action="#"
                        ref="uploadRef"
                        class="upload-demo"
                        :limit="1"
                        :on-exceed="handleExceed"
                        :auto-upload="false"
                        v-model:file-list="uploaderFiles"
                    >
                        <template #trigger>
                            <el-button type="primary" v-if="license.name === ''">上传 License 文件</el-button>
                            <el-button type="primary" v-else>更新 License 文件</el-button>
                        </template>
                        <el-button
                            style="margin-left: 0.75rem"
                            type="primary"
                            @click="submit()"
                            :disabled="loading || uploaderFiles.length == 0"
                        >
                            {{ $t('commons.button.confirm') }}
                        </el-button>
                    </el-upload>

                    <div class="flx-center" v-if="license.name !== ''">
                        <div class="version">{{ '版本类型：' + license.name }}</div>
                        <div class="version" style="margin-left: 20px">
                            {{ '有效期至：' + dateFormat(0, 0, license.expiresAt) }}
                        </div>
                    </div>

                    <h3>
                        <SystemUpgrade />
                    </h3>

                    <div style="margin-top: 10px">
                        <el-link @click="toDoc">
                            <el-icon><Document /></el-icon>
                            <span>{{ $t('setting.doc') }}</span>
                        </el-link>
                        <el-link @click="toGithub" class="system-link">
                            <svg-icon iconName="p-huaban88"></svg-icon>
                            <span>{{ $t('setting.project') }}</span>
                        </el-link>
                        <el-link @click="toIssue" class="system-link">
                            <svg-icon iconName="p-bug"></svg-icon>
                            <span>{{ $t('setting.issue') }}</span>
                        </el-link>
                        <el-link @click="toGithubStar" class="system-link">
                            <svg-icon iconName="p-star"></svg-icon>
                            <span>{{ $t('setting.star') }}</span>
                        </el-link>
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { getLicense, getSettingInfo, getSystemAvailable, UploadFileData } from '@/api/modules/setting';
import { onMounted, ref, reactive } from 'vue';
import { genFileId } from 'element-plus';
import SystemUpgrade from '@/components/system-upgrade/index.vue';
import { UploadInstance, UploadProps, UploadRawFile, UploadFiles } from 'element-plus';
import { MsgError } from '@/utils/message';
import { dateFormat } from '@/utils/util';

const version = ref();
const loading = ref();
const uploadRef = ref();

const search = async () => {
    const res = await getSettingInfo();
    version.value = res.data.systemVersion;
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value!.handleStart(file);
};

const get = async () => {
    const res = await getLicense(1);
    loading.value = false;
    if (res.data !== undefined) {
        license.name = res.data.name;
        license.expiresAt = res.data.expiresAt;
    }
};

const submit = async () => {
    const files = uploaderFiles.value.slice();
    if (files.length !== 1) {
        MsgError('只能上传一个 License 文件');
    }
    const file = files[0];
    loading.value = true;
    const formData = new FormData();
    formData.append('file', file.raw);
    const res = await UploadFileData(formData);
    loading.value = false;
    license.appName = res.data.appName;
    await get();
};

const toDoc = () => {
    window.open('https://1panel.cn/docs/', '_blank', 'noopener,noreferrer');
};
const toGithub = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank', 'noopener,noreferrer');
};
const toIssue = () => {
    window.open('https://github.com/1Panel-dev/1Panel/issues', '_blank', 'noopener,noreferrer');
};
const toGithubStar = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank', 'noopener,noreferrer');
};

onMounted(() => {
    search();
    getSystemAvailable();
    get();
});
</script>

<style lang="scss" scoped>
.system-link {
    margin-left: 15px;

    .svg-icon {
        font-size: 7px;
    }
    span {
        line-height: 20px;
    }
}

.version {
    font-size: 14px;
    color: #858585;
    text-decoration: none;
    letter-spacing: 0.5px;
}
</style>
