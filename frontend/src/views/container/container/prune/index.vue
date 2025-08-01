<template>
    <DialogPro v-model="open" :title="$t('container.containerPrune')" size="small">
        <div>
            <ul class="help-ul">
                <li class="lineClass" style="color: red">{{ $t('container.containerPruneHelper1') }}</li>
                <li class="lineClass">{{ $t('container.containerPruneHelper2') }}</li>
                <li class="lineClass">{{ $t('container.containerPruneHelper3') }}</li>
            </ul>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="open = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onClean()">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
    <TaskLog ref="taskLogRef" width="70%" @close="onSearch" />
</template>

<script lang="ts" setup>
import { containerPrune } from '@/api/modules/container';
import TaskLog from '@/components/log/task/index.vue';
import { ref } from 'vue';
import { newUUID } from '@/utils/util';

const loading = ref(false);
const open = ref<boolean>(false);
const taskLogRef = ref();

const emit = defineEmits<{ (e: 'search'): void }>();

const onClean = async () => {
    loading.value = true;
    let params = {
        taskID: newUUID(),
        pruneType: 'container',
        withTagAll: false,
    };
    await containerPrune(params)
        .then(() => {
            loading.value = false;
            open.value = false;
            openTaskLog(params.taskID);
        })
        .catch(() => {
            loading.value = false;
        });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const onSearch = () => {
    emit('search');
};

const acceptParams = (): void => {
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.lineClass {
    line-height: 30px;
}
</style>
