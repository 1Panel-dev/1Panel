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
import { onActivated, onDeactivated, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { panelCacheKey, readPanelCache, writePanelCache } from '@/utils/panel-cache';

defineOptions({ name: 'TableSetting' });
const refreshRateUnit = ref(0);
const emit = defineEmits(['search']);
const props = defineProps({ title: String, rate: Number });
const route = useRoute();
// Capture the owning page; useRoute continues to change while a page is deactivated.
const pagePath = route.path;
const cacheKey = () => panelCacheKey(`refresh:${pagePath}:${props.title || ''}`);
let timer: ReturnType<typeof setInterval> | undefined;
let active = false;
const stopTimer = () => {
    clearInterval(timer);
    timer = undefined;
};
const startTimer = () => {
    stopTimer();
    if (active && refreshRateUnit.value > 0) {
        timer = setInterval(() => emit('search'), 1000 * refreshRateUnit.value);
    }
};
const restore = () => {
    const rate = readPanelCache<number>('refresh', cacheKey());
    refreshRateUnit.value = Number.isFinite(rate) && rate >= 0 ? rate : props.rate || 0;
    startTimer();
};
const changeRefresh = (rate: number) => {
    refreshRateUnit.value = rate;
    writePanelCache('refresh', cacheKey(), rate);
    startTimer();
};
const activate = () => {
    if (active) return;
    active = true;
    restore();
};
watch(cacheKey, () => {
    if (active) restore();
});
onMounted(activate);
onActivated(activate);
onDeactivated(() => {
    active = false;
    stopTimer();
});
onUnmounted(stopTimer);
</script>

<style lang="scss" scoped>
.timer-button {
    float: right;
}
</style>
