<template>
    <DrawerPro v-model="dialogVisible" :header="$t('container.imagePrune')" @close="handleClose" size="small">
        <el-form ref="deleteForm" v-loading="loading">
            <el-form-item>
                <el-radio-group class="w-full" v-model="scope" @change="changeScope">
                    <el-radio value="untag">{{ $t('container.imagePruneSome') }}</el-radio>
                    <el-radio value="unused">{{ $t('container.imagePruneAll') }}</el-radio>
                </el-radio-group>
                <span class="input-help">{{ showMsg }}</span>
                <el-checkbox
                    class="w-full"
                    v-if="data.length !== 0"
                    v-model="checkAll"
                    :indeterminate="isIndeterminate"
                    @change="handleCheckAllChange"
                >
                    {{ $t('commons.table.all') }}
                </el-checkbox>
                <el-checkbox-group v-model="checkedLists" @change="handleCheckedChange">
                    <el-checkbox class="w-full" v-for="(item, index) in data" :key="index" :value="item.id">
                        {{
                            item.tags && item.tags[0]
                                ? item.tags[0]
                                : item.id.replaceAll('sha256:', '').substring(0, 12)
                        }}
                    </el-checkbox>
                </el-checkbox-group>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" :disabled="data.length === 0 || loading" @click="onClean">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" width="70%" @close="onSearch" />
</template>

<script lang="ts" setup>
import { containerPrune, imageRemove, listAllImage } from '@/api/modules/container';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';
import { newUUID } from '@/utils/util';
import { ref } from 'vue';

const dialogVisible = ref(false);
const scope = ref('untag');
const showMsg = ref();
const loading = ref();
const unTagList = ref([]);
const unUsedList = ref([]);
const data = ref([]);
const taskLogRef = ref();

const checkAll = ref(false);
const isIndeterminate = ref(false);
const checkedLists = ref([]);

const acceptParams = async (): Promise<void> => {
    const res = await listAllImage();
    let list = res.data || [];
    unTagList.value = [];
    unUsedList.value = [];
    for (const item of list) {
        if (
            !item.tags ||
            item.tags.length === 0 ||
            (item.tags.length === 1 && item.tags[0].indexOf('<none>') !== -1 && !item.isUsed)
        ) {
            unTagList.value.push(item);
        }
        if (!item.isUsed) {
            unUsedList.value.push(item);
        }
    }
    dialogVisible.value = true;
    scope.value = 'untag';
    changeScope();
};

const emit = defineEmits<{ (e: 'search'): void }>();

const changeScope = () => {
    if (scope.value === 'untag') {
        data.value = unTagList.value || [];
        showMsg.value =
            data.value.length === 0
                ? i18n.global.t('container.imagePruneSomeHelper')
                : i18n.global.t('container.imagePruneSomeEmpty');
        return;
    }
    data.value = unUsedList.value || [];
    showMsg.value =
        data.value.length === 0
            ? i18n.global.t('container.imagePruneAllHelper')
            : i18n.global.t('container.imagePruneAllEmpty');
    return;
};

const handleCheckAllChange = (val: boolean) => {
    checkedLists.value = [];
    if (!val) {
        isIndeterminate.value = false;
        return;
    }
    for (const item of data.value) {
        checkedLists.value.push(item.id);
    }
};
const handleCheckedChange = (value: string[]) => {
    const checkedCount = value.length;
    checkAll.value = checkedCount === unUsedList.value.length;
    isIndeterminate.value = checkedCount > 0 && checkedCount < unUsedList.value.length;
};

const handleClose = () => {
    dialogVisible.value = false;
};

const onSearch = () => {
    emit('search');
};

const onClean = async () => {
    loading.value = true;
    if (checkAll.value) {
        prune();
        return;
    }
    removeImage();
};

const prune = async () => {
    let params = {
        taskID: newUUID(),
        pruneType: 'image',
        withTagAll: scope.value === 'unused',
    };
    await containerPrune(params)
        .then(() => {
            loading.value = false;
            dialogVisible.value = false;
            openTaskLog(params.taskID);
        })
        .catch(() => {
            loading.value = false;
        });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const removeImage = async () => {
    let params = {
        taskID: newUUID(),
        names: checkedLists.value,
    };
    await imageRemove(params)
        .then(() => {
            loading.value = false;
            dialogVisible.value = false;
            openTaskLog(params.taskID);
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
