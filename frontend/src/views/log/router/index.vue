<template>
    <div>
        <el-button
            class="tag-button"
            :class="current != 'OperationLog' ? 'no-active' : ''"
            :type="current === 'OperationLog' ? 'primary' : ''"
            @click="onChangeRoute('OperationLog')"
        >
            {{ $t('logs.operation') }}
        </el-button>
        <el-button
            class="tag-button"
            :class="current != 'LoginLog' ? 'no-active' : ''"
            :type="current === 'LoginLog' ? 'primary' : ''"
            @click="onChangeRoute('LoginLog')"
        >
            {{ $t('logs.login') }}
        </el-button>
        <el-button
            class="tag-button"
            :class="current != 'SystemLog' ? 'no-active' : ''"
            :type="current === 'SystemLog' ? 'primary' : ''"
            @click="onChangeRoute('SystemLog')"
        >
            {{ $t('logs.system') }}
        </el-button>
        <el-button
            class="tag-button"
            :class="current != 'Task' ? 'no-active' : ''"
            :type="current === 'Task' ? 'primary' : ''"
            @click="onChangeRoute('Task')"
        >
            {{ $t('logs.task') }}
        </el-button>
    </div>
</template>
<script setup lang="ts">
import { routerToName } from '@/utils/router';

defineProps({
    current: {
        type: String,
        default: 'LoginLog',
    },
});
const onChangeRoute = async (addr: string) => {
    routerToName(addr);
};
</script>
