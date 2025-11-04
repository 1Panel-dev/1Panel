<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            v-model:loading="loading"
            @search="search"
        />

        <div v-if="isExist" :class="{ mask: !isActive }">
            <CardWithHeader :header="$t('menu.container')" class="card-interval">
                <template #body>
                    <span class="count" @click="routerToName('ContainerItem')">{{ countItem.containerCount }}</span>
                </template>
                <template #header-l>
                    <el-tag size="small" class="ml-5" v-if="countItem.containerCount" effect="plain">
                        {{ $t('commons.table.all') }} * {{ countItem.containerCount }}
                    </el-tag>
                    <el-tag size="small" v-if="countItem.running" effect="plain" class="ml-2">
                        {{ $t('commons.status.running') }} * {{ countItem.running }}
                    </el-tag>
                    <el-tag size="small" v-if="countItem.created" effect="plain" class="ml-2">
                        {{ $t('commons.status.created') }} * {{ countItem.created }}
                    </el-tag>
                    <el-tag size="small" v-if="countItem.paused" effect="plain" class="ml-2">
                        {{ $t('commons.status.paused') }} * {{ countItem.paused }}
                    </el-tag>
                    <el-tag size="small" v-if="countItem.restarting" effect="plain" class="ml-2">
                        {{ $t('commons.status.restarting') }} * {{ countItem.restarting }}
                    </el-tag>
                    <el-tag size="small" v-if="countItem.removing" effect="plain" class="ml-2">
                        {{ $t('commons.status.removing') }} * {{ countItem.removing }}
                    </el-tag>
                    <el-tag size="small" v-if="countItem.exited" effect="plain" class="ml-2">
                        {{ $t('commons.status.exited') }} * {{ countItem.exited }}
                    </el-tag>
                    <el-tag size="small" v-if="countItem.dead" effect="plain" class="ml-2">
                        {{ $t('commons.status.dead') }} * {{ countItem.dead }}
                    </el-tag>
                </template>
            </CardWithHeader>
            <el-row :gutter="7" class="card-interval">
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.compose')">
                        <template #body>
                            <span class="count" @click="routerToName('Compose')">{{ countItem.composeCount }}</span>
                        </template>
                    </CardWithHeader>
                </el-col>
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.composeTemplate')">
                        <template #body>
                            <span class="count" @click="routerToName('ComposeTemplate')">
                                {{ countItem.composeTemplateCount }}
                            </span>
                        </template>
                    </CardWithHeader>
                </el-col>
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.image')">
                        <template #body>
                            <span class="count" @click="routerToName('Image')">{{ countItem.imageCount }}</span>
                        </template>
                    </CardWithHeader>
                </el-col>
            </el-row>
            <el-row :gutter="7" class="card-interval">
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.imageRepo')">
                        <template #body>
                            <span class="count" @click="routerToName('Repo')">{{ countItem.repoCount }}</span>
                        </template>
                    </CardWithHeader>
                </el-col>
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.network')">
                        <template #body>
                            <span class="count" @click="routerToName('Network')">{{ countItem.networkCount }}</span>
                        </template>
                    </CardWithHeader>
                </el-col>
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.volume')">
                        <template #body>
                            <span class="count" @click="routerToName('Volume')">{{ countItem.volumeCount }}</span>
                        </template>
                    </CardWithHeader>
                </el-col>
            </el-row>

            <CardWithHeader :header="$t('container.diskUsage')" class="card-interval">
                <template #body>
                    <el-descriptions direction="vertical" align="center" :column="4" class="mt-2">
                        <el-descriptions-item label-width="25%" align="center" :label="$t('container.image')">
                            {{
                                $t('container.usage', [
                                    computeSize2(countItem.imageUsage),
                                    computeSize2(countItem.imageReclaimable),
                                ])
                            }}
                            <el-button
                                v-if="countItem.imageReclaimable"
                                class="-mt-0.5"
                                @click="onClean('image', true)"
                                link
                                type="primary"
                            >
                                {{ $t('container.clean') }}
                            </el-button>
                        </el-descriptions-item>
                        <el-descriptions-item label-width="25%" align="center" :label="$t('menu.container')">
                            {{
                                $t('container.usage', [
                                    computeSize2(countItem.containerUsage),
                                    computeSize2(countItem.containerReclaimable),
                                ])
                            }}
                            <el-button
                                v-if="countItem.containerReclaimable"
                                class="-mt-0.5"
                                @click="onClean('container', false)"
                                link
                                type="primary"
                            >
                                {{ $t('container.clean') }}
                            </el-button>
                        </el-descriptions-item>
                        <el-descriptions-item label-width="25%" align="center" :label="$t('container.localVolume')">
                            {{
                                $t('container.usage', [
                                    computeSize2(countItem.volumeUsage),
                                    computeSize2(countItem.volumeReclaimable),
                                ])
                            }}
                            <el-button
                                v-if="countItem.volumeReclaimable"
                                class="-mt-0.5"
                                @click="onClean('volume', false)"
                                link
                                type="primary"
                            >
                                {{ $t('container.clean') }}
                            </el-button>
                        </el-descriptions-item>
                        <el-descriptions-item label-width="25%" align="center" :label="$t('container.buildCache')">
                            {{
                                $t('container.usage', [
                                    computeSize2(countItem.buildCacheUsage),
                                    computeSize2(countItem.buildCacheUsage),
                                ])
                            }}
                            <el-button
                                v-if="countItem.buildCacheUsage"
                                class="-mt-0.5"
                                @click="onClean('buildcache', false)"
                                link
                                type="primary"
                            >
                                {{ $t('container.clean') }}
                            </el-button>
                        </el-descriptions-item>
                    </el-descriptions>
                </template>
            </CardWithHeader>
            <CardWithHeader :header="$t('container.setting')" class="card-interval">
                <template #body>
                    <el-descriptions :column="1" border>
                        <el-descriptions-item label-width="200" :label="$t('container.sockPath')">
                            {{ countItem.sockPath }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('container.mirrors')">
                            <div v-for="item in countItem.mirrors" :key="item" :value="item" :label="item">
                                <div class="mt-1">
                                    <el-tag>{{ item }}</el-tag>
                                </div>
                            </div>
                        </el-descriptions-item>
                    </el-descriptions>
                    <el-button class="mt-2" type="primary" link @click="routerToName('ContainerSetting')">
                        <el-icon class="mr-1"><Position /></el-icon>
                        {{ $t('container.goSetting') }}
                    </el-button>
                </template>
            </CardWithHeader>
        </div>
        <TaskLog ref="taskLogRef" width="70%" />
    </div>
</template>

<script lang="ts" setup>
import { containerPrune, loadContainerStatus, loadDaemonJson } from '@/api/modules/container';
import DockerStatus from '@/views/container/docker-status/index.vue';
import { getSettingInfo } from '@/api/modules/setting';
import { computeSize2, newUUID } from '@/utils/util';
import TaskLog from '@/components/log/task/index.vue';
import { routerToName } from '@/utils/router';
import { onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';

const taskLogRef = ref();

const loading = ref();
const isActive = ref(false);
const isExist = ref(false);
const countItem = reactive({
    created: 0,
    running: 0,
    paused: 0,
    restarting: 0,
    removing: 0,
    exited: 0,
    dead: 0,

    containerCount: 0,
    composeCount: 0,
    composeTemplateCount: 0,
    imageCount: 0,
    networkCount: 0,
    volumeCount: 0,
    repoCount: 0,

    containerUsage: 0,
    containerReclaimable: 0,
    imageUsage: 0,
    imageReclaimable: 0,
    volumeUsage: 0,
    volumeReclaimable: 0,
    buildCacheUsage: 0,
    buildCacheReclaimable: 0,

    mirrors: [],
    sockPath: '',
});

const search = () => {
    if (!isActive.value || !isExist.value) {
        return;
    }
    loadContainerCount();
    loadContainerSetting();
};

const loadContainerCount = async () => {
    await loadContainerStatus().then((res) => {
        countItem.running = res.data.running;
        countItem.paused = res.data.paused;
        countItem.restarting = res.data.restarting;
        countItem.removing = res.data.removing;
        countItem.created = res.data.created;
        countItem.dead = res.data.dead;
        countItem.exited = res.data.exited;

        countItem.containerCount = res.data.containerCount;
        countItem.composeCount = res.data.composeCount;
        countItem.composeTemplateCount = res.data.composeTemplateCount;
        countItem.imageCount = res.data.imageCount;
        countItem.networkCount = res.data.networkCount;
        countItem.volumeCount = res.data.volumeCount;
        countItem.repoCount = res.data.repoCount;

        countItem.containerUsage = res.data.containerUsage;
        countItem.containerReclaimable = res.data.containerReclaimable;
        countItem.imageUsage = res.data.imageUsage;
        countItem.imageReclaimable = res.data.imageReclaimable;
        countItem.volumeUsage = res.data.volumeUsage;
        countItem.volumeReclaimable = res.data.volumeReclaimable;
        countItem.buildCacheUsage = res.data.buildCacheUsage;
        countItem.buildCacheReclaimable = res.data.buildCacheReclaimable;
    });
};

const loadContainerSetting = async () => {
    const res = await loadDaemonJson();
    countItem.mirrors = res.data.registryMirrors || [];

    const settingRes = await getSettingInfo();
    countItem.sockPath = settingRes.data.dockerSockPath || 'unix:///var/run/docker.sock';
};

const onClean = async (type: string, withAll: boolean) => {
    let msg = '';
    switch (type) {
        case 'image':
            msg = i18n.global.t('container.imageClean');
            break;
        case 'container':
            msg = i18n.global.t('container.containerClean');
            break;
        case 'volume':
            msg = i18n.global.t('container.volumePruneHelper');
            break;
        case 'buildcache':
            msg = i18n.global.t('container.delBuildCacheHelper');
            break;
    }
    ElMessageBox.confirm(msg, i18n.global.t('container.clean'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        loading.value = true;
        let params = {
            taskID: newUUID(),
            pruneType: type,
            withTagAll: withAll,
        };
        await containerPrune(params)
            .then(() => {
                loading.value = false;
                openTaskLog(params.taskID);
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

onMounted(() => {
    search();
});
</script>

<style scoped lang="scss">
.count {
    margin-left: 20px;
    font-size: 25px;
    color: $primary-color;
    font-weight: 500;
    line-height: 32px;
    cursor: pointer;
}
</style>
