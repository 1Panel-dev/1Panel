<template>
    <div>
        <el-dropdown @command="changeRefresh">
            <el-badge
                badge-style="background-color: transparent; font-size: 12px; border: none; color: black"
                :offset="[-12, 7]"
                :value="refreshRateUnit === 0 ? '' : refreshRateUnit + 's'"
                class="item"
            >
                <el-button class="timer-button" icon="Clock"></el-button>
            </el-badge>
            <template #dropdown>
                <el-dropdown-menu>
                    <el-dropdown-item :command="0">{{ $t('commons.table.noRefresh') }}</el-dropdown-item>
                    <el-dropdown-item :command="5">{{ $t('commons.table.refreshRateUnit', [5]) }}</el-dropdown-item>
                    <el-dropdown-item :command="10">{{ $t('commons.table.refreshRateUnit', [10]) }}</el-dropdown-item>
                    <el-dropdown-item :command="30">{{ $t('commons.table.refreshRateUnit', [30]) }}</el-dropdown-item>
                    <el-dropdown-item :command="60">{{ $t('commons.table.refreshRateUnit', [60]) }}</el-dropdown-item>
                    <el-dropdown-item :command="120">{{ $t('commons.table.refreshRateUnit', [120]) }}</el-dropdown-item>
                    <el-dropdown-item :command="300">{{ $t('commons.table.refreshRateUnit', [300]) }}</el-dropdown-item>
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
