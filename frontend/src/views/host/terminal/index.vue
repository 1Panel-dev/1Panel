<template>
    <div>
        <el-card class="topCard">
            <el-radio-group v-model="active">
                <el-radio-button class="topButton" size="large" @click="routerTo('/hosts/terminal')" label="terminal">
                    {{ $t('menu.terminal') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" @click="routerTo('/hosts/host')" label="host">
                    {{ $t('menu.host') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" @click="routerTo('/hosts/command')" label="command">
                    {{ $t('terminal.quickCommand') }}
                </el-radio-button>
            </el-radio-group>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();
interface MenuProps {
    activeName: string;
}
const props = withDefaults(defineProps<MenuProps>(), {
    activeName: 'terminal',
});

const active = ref('terminal');

onMounted(() => {
    if (props.activeName) {
        active.value = props.activeName;
    }
});

const routerTo = (path: string) => {
    router.push({ path: path });
};
</script>

<style>
.topCard {
    --el-card-border-color: var(--el-border-color-light);
    --el-card-border-radius: 4px;
    --el-card-padding: 0px;
    --el-card-bg-color: var(--el-fill-color-blank);
}
.topButton .el-radio-button__inner {
    display: inline-block;
    line-height: 1;
    white-space: nowrap;
    vertical-align: middle;
    background: var(--el-button-bg-color, var(--el-fill-color-blank));
    border: 0;
    font-weight: 350;
    border-left: 0;
    color: var(--el-button-text-color, var(--el-text-color-regular));
    text-align: center;
    box-sizing: border-box;
    outline: 0;
    margin: 0;
    position: relative;
    cursor: pointer;
    transition: var(--el-transition-all);
    -webkit-user-select: none;
    user-select: none;
    padding: 8px 15px;
    font-size: var(--el-font-size-base);
    border-radius: 0;
}
</style>
