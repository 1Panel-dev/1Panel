<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.system')">
            <template #search>
                <LogRouter current="SystemLog" />
            </template>
            <template #leftToolBar>
                <el-select class="p-w-200 mr-2.5" v-model="itemName" @change="search()">
                    <template #prefix>{{ $t('commons.table.date') }}</template>
                    <el-option v-for="(item, index) in fileList" :key="index" :label="item" :value="item" />
                </el-select>
                <el-button>
                    <el-checkbox @change="changeTail" v-model="isWatch">
                        {{ $t('commons.button.watch') }}
                    </el-checkbox>
                </el-button>
                <el-radio-group class="ml-2" @change="search()" v-model="itemType">
                    <el-radio-button :label="$t('logs.agent')" value="Agent" />
                    <el-radio-button :label="$t('logs.core')" value="Core" />
                </el-radio-group>
            </template>
            <template #main>
                <LogFile
                    ref="logRef"
                    :config="logConfig"
                    :default-button="false"
                    v-if="showLog"
                    v-model:loading="loading"
                    v-model:hasContent="hasContent"
                    :height-diff="330"
                />
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import LogFile from '@/components/log/file/index.vue';
import LogRouter from '@/views/log/router/index.vue';
import { nextTick, onMounted, reactive, ref } from 'vue';
import { getSystemFiles } from '@/api/modules/log';

const loading = ref();
const isWatch = ref();
const fileList = ref();
const logRef = ref();

const hasContent = ref(false);
const logConfig = reactive({
    type: 'system',
    name: '',
    colorMode: 'system',
});
const showLog = ref(false);
const itemName = ref();
const itemType = ref('Agent');

const changeTail = () => {
    logRef.value.changeTail(true);
};

const loadFiles = async () => {
    const res = await getSystemFiles();
    fileList.value = res.data || [];
    if (fileList.value) {
        itemName.value = fileList.value[0];
        search();
    }
};

const search = () => {
    logConfig.name = itemType.value === 'Agent' ? itemName.value : 'Core-' + itemName.value;
    showLog.value = false;
    nextTick(() => {
        showLog.value = true;
    });
};

onMounted(() => {
    loadFiles();
});
</script>
