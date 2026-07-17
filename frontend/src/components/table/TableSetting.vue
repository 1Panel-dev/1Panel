<template>
    <div>
        <el-dropdown @command="changeRefresh">
            <el-button class="timer-button">
                {{ refreshRateUnit === 0 ? $t('commons.table.noRefresh') : refreshRateUnit + 's' }}
            </el-button>
            <template #dropdown>
                <el-dropdown-menu>
                    <el-dropdown-item :command="0">{{ $t('commons.table.noRefresh') }}</el-dropdown-item>
                    <el-dropdown-item :command="5">5s</el-dropdown-item>
                    <el-dropdown-item :command="10">10s</el-dropdown-item>
                    <el-dropdown-item :command="30">30s</el-dropdown-item>
                    <el-dropdown-item :command="60">60s</el-dropdown-item>
                    <el-dropdown-item :command="120">120s</el-dropdown-item>
                    <el-dropdown-item :command="300">300s</el-dropdown-item>
                </el-dropdown-menu>
            </template>
        </el-dropdown>
    </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
defineOptions({ name: 'TableSetting' });

const refreshRateUnit = ref<number>(0);
const emit = defineEmits(['search']);
const props = defineProps({
    title: String,
    rate: Number,
});

let timer: NodeJS.Timer | null = null;

const changeRefresh = (command: number) => {
    refreshRateUnit.value = command || 0;
    if (refreshRateUnit.value !== 0) {
        if (timer) {
            clearInterval(Number(timer));
            timer = null;
        }
        timer = setInterval(() => {
            emit('search');
        }, 1000 * refreshRateUnit.value);
    } else {
        if (timer) {
            clearInterval(Number(timer));
            timer = null;
        }
    }
    localStorage.setItem(props.title, refreshRateUnit.value + '');
};

onUnmounted(() => {
    clearInterval(Number(timer));
    timer = null;
    if (props.title) {
        localStorage.setItem(props.title, refreshRateUnit.value + '');
    }
});

onMounted(() => {
    if (props.title && localStorage.getItem(props.title) != null) {
        let rate = Number(localStorage.getItem(props.title));
        refreshRateUnit.value = rate ? Number(rate) : 0;
        changeRefresh(refreshRateUnit.value);
        return;
    }
    if (props.rate) {
        refreshRateUnit.value = props.rate;
        changeRefresh(refreshRateUnit.value);
    }
});
</script>

<style lang="scss" scoped>
.timer-button {
    float: right;
}
</style>
