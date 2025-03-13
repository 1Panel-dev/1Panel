<template>
    <div>
        <el-popover placement="bottom-start" :width="200" trigger="click">
            <template #reference>
                <el-button class="timer-button" icon="Clock"></el-button>
            </template>
            <el-select v-model="refreshRate" @change="changeRefresh">
                <template #prefix>{{ $t('commons.table.refreshRate') }}</template>
                <el-option :label="$t('commons.table.noRefresh')" :value="0"></el-option>
                <el-option :label="$t('commons.table.refreshRateUnit', [5])" :value="5"></el-option>
                <el-option :label="$t('commons.table.refreshRateUnit', [10])" :value="10"></el-option>
                <el-option :label="$t('commons.table.refreshRateUnit', [30])" :value="30"></el-option>
                <el-option :label="$t('commons.table.refreshRateUnit', [60])" :value="60"></el-option>
                <el-option :label="$t('commons.table.refreshRateUnit', [120])" :value="120"></el-option>
                <el-option :label="$t('commons.table.refreshRateUnit', [300])" :value="300"></el-option>
            </el-select>
        </el-popover>
    </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
defineOptions({ name: 'TableSetting' });

const refreshRate = ref<number>(0);
const emit = defineEmits(['search']);
const props = defineProps({
    title: String,
    rate: Number,
});

let timer: NodeJS.Timer | null = null;

const changeRefresh = () => {
    if (refreshRate.value !== 0) {
        if (timer) {
            clearInterval(Number(timer));
            timer = null;
        }
        timer = setInterval(() => {
            emit('search');
        }, 1000 * refreshRate.value);
    } else {
        if (timer) {
            clearInterval(Number(timer));
            timer = null;
        }
    }
    localStorage.setItem(props.title, refreshRate.value + '');
};

onUnmounted(() => {
    clearInterval(Number(timer));
    timer = null;
    if (props.title) {
        localStorage.setItem(props.title, refreshRate.value + '');
    }
});

onMounted(() => {
    if (props.title && localStorage.getItem(props.title) != null) {
        let rate = Number(localStorage.getItem(props.title));
        refreshRate.value = rate ? Number(rate) : 0;
        changeRefresh();
        return;
    }
    if (props.rate) {
        refreshRate.value = props.rate;
        changeRefresh();
    }
});
</script>

<style lang="scss" scoped>
.timer-button {
    float: right;
}
</style>
