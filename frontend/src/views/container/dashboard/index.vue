<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            v-model:loading="loading"
            @search="search"
        />

        <div v-if="isExist" :class="{ mask: !isActive }">
            <CardWithHeader :header="$t('menu.container')" class="mt-5">
                <template #body>
                    <span class="count" @click="goRouter('Container')">{{ countItem.containerCount }}</span>
                </template>
                <template #header-l>
                    <el-tag size="small" class="ml-5" v-if="countItem.all" effect="plain">
                        {{ $t('commons.table.all') }} * {{ countItem.all }}
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
            <el-row :gutter="20" class="mt-5">
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.compose')">
                        <template #body>
                            <span class="count" @click="goRouter('Compose')">{{ countItem.composeCount }}</span>
                        </template>
                    </CardWithHeader>
                </el-col>
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.composeTemplate')">
                        <template #body>
                            <span class="count" @click="goRouter('ComposeTemplate')">
                                {{ countItem.composeTemplateCount }}
                            </span>
                        </template>
                    </CardWithHeader>
                </el-col>
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.image')">
                        <template #body>
                            <span class="count" @click="goRouter('Image')">{{ countItem.imageCount }}</span>
                        </template>
                        <template #header-l>
                            <span v-if="countItem.imageSize" effect="plain" class="ml-2 text-xs">
                                {{ $t('commons.status.used') }}: {{ computeSize(countItem.imageSize) }}
                            </span>
                        </template>
                    </CardWithHeader>
                </el-col>
            </el-row>
            <el-row :gutter="20" class="mt-5">
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.imageRepo')">
                        <template #body>
                            <span class="count" @click="goRouter('Repo')">{{ countItem.repoCount }}</span>
                        </template>
                    </CardWithHeader>
                </el-col>
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.network')">
                        <template #body>
                            <span class="count" @click="goRouter('Network')">{{ countItem.networkCount }}</span>
                        </template>
                    </CardWithHeader>
                </el-col>
                <el-col :span="8">
                    <CardWithHeader :header="$t('container.volume')">
                        <template #body>
                            <span class="count" @click="goRouter('Volume')">{{ countItem.volumeCount }}</span>
                        </template>
                    </CardWithHeader>
                </el-col>
            </el-row>
            <CardWithHeader :header="$t('container.setting')" class="mt-5">
                <template #body>
                    <el-descriptions :column="1" border>
                        <el-descriptions-item :label="$t('container.sockPath')">
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
                    <el-button class="mt-2" type="primary" link @click="goRouter('ContainerSetting')">
                        <el-icon class="mr-1"><Position /></el-icon>
                        {{ $t('container.goSetting') }}
                    </el-button>
                </template>
            </CardWithHeader>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { loadContainerStatus, loadDaemonJson } from '@/api/modules/container';
import DockerStatus from '@/views/container/docker-status/index.vue';
import { getSettingInfo } from '@/api/modules/setting';
import router from '@/routers';
import { computeSize } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';

const loading = ref();
const isActive = ref(false);
const isExist = ref(false);
const countItem = reactive({
    all: 0,
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
    imageSize: 0,

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
        countItem.all = res.data.all;
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
        countItem.imageSize = res.data.imageSize;
    });
};

const loadContainerSetting = async () => {
    const res = await loadDaemonJson();
    countItem.mirrors = res.data.registryMirrors || [];

    const settingRes = await getSettingInfo();
    countItem.sockPath = settingRes.data.dockerSockPath || 'unix:///var/run/docker.sock';
};

const goRouter = async (val: string) => {
    router.push({ name: val });
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
