<template>
    <div>
        <el-popover placement="bottom-start" :width="200" trigger="click">
            <template #reference>
                <el-button round class="timer-button">{{ $t('commons.table.tableSetting') }}</el-button>
            </template>
            <div style="margin-left: 15px">
                <div>
                    <span>{{ $t('commons.table.autoRefresh') }}</span>
                    <el-switch style="margin-left: 5px" v-model="autoRefresh" @change="changeRefresh"></el-switch>
                </div>
                <div>
                    <span>{{ $t('commons.table.refreshRate') }}</span>
                    <el-select style="margin-left: 5px; width: 80px" v-model="refreshRate" @change="changeRefresh">
                        <el-option label="5s" :value="5"></el-option>
                        <el-option label="10s" :value="10"></el-option>
                        <el-option label="30s" :value="30"></el-option>
                        <el-option label="1min" :value="60"></el-option>
                        <el-option label="2min" :value="120"></el-option>
                        <el-option label="5min" :value="300"></el-option>
                    </el-select>
                </div>
            </div>
        </el-popover>
    </div>
</template>

<script setup lang="ts">
import { onUnmounted, ref } from 'vue';
defineOptions({ name: 'TableSetting' });

const autoRefresh = ref<boolean>(false);
const refreshRate = ref<number>(10);
const emit = defineEmits(['search']);

let timer: NodeJS.Timer | null = null;

const changeRefresh = () => {
    if (autoRefresh.value) {
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
};

const startTimer = () => {
    autoRefresh.value = true;
    changeRefresh();
};
const endTimer = () => {
    autoRefresh.value = false;
    changeRefresh();
};

onUnmounted(() => {
    clearInterval(Number(timer));
    timer = null;
});

defineExpose({
    startTimer,
    endTimer,
});
</script>

<style lang="scss" scoped>
.timer-button {
    float: right;
}
</style>
