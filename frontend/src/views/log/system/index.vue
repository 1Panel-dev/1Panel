<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.system')">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button class="tag-button no-active" @click="onChangeRoute('OperationLog')">
                            {{ $t('logs.operation') }}
                        </el-button>
                        <el-button class="tag-button no-active" @click="onChangeRoute('LoginLog')">
                            {{ $t('logs.login') }}
                        </el-button>
                        <el-button class="tag-button" type="primary" @click="onChangeRoute('SystemLog')">
                            {{ $t('logs.system') }}
                        </el-button>
                    </el-col>
                </el-row>
            </template>
            <template #search>
                <div class="flex flex-wrap gap-2 sm:gap-4">
                    <el-select class="p-w-200" v-model="logConfig.name" @change="search()">
                        <template #prefix>{{ $t('commons.button.log') }}</template>
                        <el-option v-for="(item, index) in fileList" :key="index" :label="item" :value="item" />
                    </el-select>
                    <el-checkbox border @change="changeTail" v-model="isWatch">
                        {{ $t('commons.button.watch') }}
                    </el-checkbox>
                </div>
            </template>
            <template #main>
                <LogFile
                    ref="logRef"
                    :config="logConfig"
                    v-if="showLog"
                    :default-button="false"
                    v-model:loading="loading"
                    v-model:hasContent="hasContent"
                    :style="'height: calc(100vh - 370px);min-height: 200px'"
                />
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { getSystemFiles } from '@/api/modules/log';
import LogFile from '@/components/log-file/index.vue';

const router = useRouter();
const loading = ref();
const isWatch = ref();
const fileList = ref();
const logRef = ref();

const hasContent = ref(false);
const logConfig = reactive({
    type: 'system',
    name: '',
});
const showLog = ref(false);

const changeTail = () => {
    logRef.value.changeTail(true);
};

const loadFiles = async () => {
    const res = await getSystemFiles();
    fileList.value = res.data || [];
    if (fileList.value) {
        logConfig.name = fileList.value[0];
        search();
    }
};

const search = () => {
    showLog.value = false;
    nextTick(() => {
        showLog.value = true;
    });
};

const onChangeRoute = async (addr: string) => {
    router.push({ name: addr });
};

onMounted(() => {
    loadFiles();
});
</script>

<style scoped lang="scss">
.watchCheckbox {
    margin-top: 2px;
    margin-bottom: 10px;
    float: left;
    margin-left: 20px;
}
</style>
