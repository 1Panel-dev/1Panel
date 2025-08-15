<template>
    <div class="app-warn">
        <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
            <span>{{ $t('app.noAppHelper') }}</span>
            <span class="flex items-center justify-center gap-0.5" @click="openTask">
                <el-icon><Position /></el-icon>
                {{ $t('website.check') }}
            </span>
        </div>
        <div>
            <img src="@/assets/images/no_app.svg" />
        </div>
    </div>
    <TaskList ref="taskListRef" />
</template>

<script lang="ts" setup>
import TaskList from '@/components/task-list/index.vue';
const taskListRef = ref();
const openTask = () => {
    taskListRef.value.acceptParams();
};
</script>
