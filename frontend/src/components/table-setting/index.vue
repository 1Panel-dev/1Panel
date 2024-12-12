<template>
    <div>
        <el-popover placement="bottom-start" :width="260" trigger="click">
            <template #reference>
                <el-button round class="timer-button">{{ $t('commons.table.tableSetting') }}</el-button>
            </template>
            <div style="margin-left: 15px">
                <span>{{ $t('commons.table.refreshRate') }}</span>
                <el-select style="margin-left: 5px; width: 120px" v-model="refreshRate" @change="changeRefresh">
                    <el-option :label="$t('commons.table.refreshRateUnit', 0)" :value="0"></el-option>
                    <el-option :label="$t('commons.table.refreshRateUnit', 5)" :value="5"></el-option>
                    <el-option :label="$t('commons.table.refreshRateUnit', 10)" :value="10"></el-option>
                    <el-option :label="$t('commons.table.refreshRateUnit', 30)" :value="30"></el-option>
                    <el-option :label="$t('commons.table.refreshRateUnit', 60)" :value="60"></el-option>
                    <el-option :label="$t('commons.table.refreshRateUnit', 120)" :value="120"></el-option>
                    <el-option :label="$t('commons.table.refreshRateUnit', 300)" :value="300"></el-option>
                </el-select>
            </div>
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
    if (props.title) {
        let rate = Number(localStorage.getItem(props.title));
        refreshRate.value = rate ? Number(rate) : 0;
        changeRefresh();
    }
});
</script>

<style lang="scss" scoped>
.timer-button {
    float: right;
}
</style>
